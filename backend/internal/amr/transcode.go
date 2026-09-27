package amr

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/echosh-labs/mercury-dasha/internal/amr/amrnb"
)

// Magic header for AMR storage format (single channel narrowband).
var amrHeader = []byte("#!AMR\n")

// Frame sizes in bytes for FT 0..15 in AMR-NB.
var amrFrameSizes = [16]int{
	12, // Mode 0: 4.75 kbps
	13, // Mode 1: 5.15 kbps
	15, // Mode 2: 5.90 kbps
	17, // Mode 3: 6.70 kbps
	19, // Mode 4: 7.40 kbps
	20, // Mode 5: 7.95 kbps
	26, // Mode 6: 10.2 kbps
	31, // Mode 7: 12.2 kbps
	5,  // Mode 8: SID (Comfort Noise)
	0, 0, 0, 0, 0, 0, // Modes 9..14: Reserved
	0, // Mode 15: No Data
}

// IsAMR checks whether the provided byte slice begins with the standard AMR magic header.
func IsAMR(data []byte) bool {
	return bytes.HasPrefix(data, amrHeader)
}

// DecodeAMR reads an AMR-NB stream from r and decodes it into 16-bit 8000Hz mono PCM samples.
func DecodeAMR(r io.Reader) ([]int16, error) {
	header := make([]byte, len(amrHeader))
	if _, err := io.ReadFull(r, header); err != nil {
		return nil, fmt.Errorf("amr: failed to read header: %w", err)
	}
	if !bytes.Equal(header, amrHeader) {
		return nil, errors.New("amr: invalid magic header (expected #!AMR\\n)")
	}

	dec := amrnb.NewDecoder(amrnb.DecoderConfig{OctetAligned: true})
	var pcmSamples []int16

	tocBuf := make([]byte, 1)
	for {
		_, err := io.ReadFull(r, tocBuf)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return nil, fmt.Errorf("amr: error reading toc byte: %w", err)
		}

		toc := tocBuf[0]
		ft := (int(toc) >> 3) & 0x0F
		if ft >= len(amrFrameSizes) {
			break
		}

		frameLen := amrFrameSizes[ft]
		frameData := make([]byte, frameLen)
		if frameLen > 0 {
			if _, err := io.ReadFull(r, frameData); err != nil {
				// EOF in middle of frame, end gracefully
				break
			}
		}

		// Build RFC 4867 single-frame octet-aligned packet: CMR (0) + ToC (F=0) + frameData
		packet := make([]byte, 2+frameLen)
		packet[0] = 0x00        // CMR
		packet[1] = toc & 0x7F // F=0
		if frameLen > 0 {
			copy(packet[2:], frameData)
		}

		samples, derr := dec.Decode(packet)
		if derr != nil {
			// Skip or stop on unrecoverable frame error
			break
		}
		pcmSamples = append(pcmSamples, samples...)
	}

	return pcmSamples, nil
}

// EncodeWAV serializes 16-bit PCM samples into a standard RIFF WAVE byte buffer.
func EncodeWAV(samples []int16, sampleRate int, numChannels int) []byte {
	dataSize := uint32(len(samples) * 2)
	fileSize := 36 + dataSize

	buf := new(bytes.Buffer)
	buf.Grow(int(44 + dataSize))

	// RIFF chunk descriptor
	buf.WriteString("RIFF")
	_ = binary.Write(buf, binary.LittleEndian, fileSize)
	buf.WriteString("WAVE")

	// "fmt " sub-chunk
	buf.WriteString("fmt ")
	_ = binary.Write(buf, binary.LittleEndian, uint32(16))               // Subchunk1Size (16 for PCM)
	_ = binary.Write(buf, binary.LittleEndian, uint16(1))                // AudioFormat (1 = PCM)
	_ = binary.Write(buf, binary.LittleEndian, uint16(numChannels))     // NumChannels
	_ = binary.Write(buf, binary.LittleEndian, uint32(sampleRate))      // SampleRate
	byteRate := uint32(sampleRate * numChannels * 2)
	_ = binary.Write(buf, binary.LittleEndian, byteRate)                // ByteRate
	blockAlign := uint16(numChannels * 2)
	_ = binary.Write(buf, binary.LittleEndian, blockAlign)              // BlockAlign
	_ = binary.Write(buf, binary.LittleEndian, uint16(16))              // BitsPerSample (16-bit)

	// "data" sub-chunk
	buf.WriteString("data")
	_ = binary.Write(buf, binary.LittleEndian, dataSize)
	_ = binary.Write(buf, binary.LittleEndian, samples)

	return buf.Bytes()
}

// TranscodeAMRToWAV decodes an AMR-NB stream and returns standard 16-bit 8000Hz mono WAV bytes.
func TranscodeAMRToWAV(r io.Reader) ([]byte, error) {
	samples, err := DecodeAMR(r)
	if err != nil {
		return nil, err
	}
	return EncodeWAV(samples, amrnb.SampleRate, 1), nil
}

// TranscodeAMRFileToWAVFile reads an AMR file from srcPath and writes decoded WAV to dstPath.
func TranscodeAMRFileToWAVFile(srcPath, dstPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	wavBytes, err := TranscodeAMRToWAV(f)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dstPath), 0755); err != nil {
		return err
	}

	tmpDst := dstPath + ".tmp"
	if err := os.WriteFile(tmpDst, wavBytes, 0644); err != nil {
		return err
	}
	return os.Rename(tmpDst, dstPath)
}

// GetOrCreateCachedWAV ensures a cached WAV exists for the specified source AMR file.
// If valid cached WAV exists and is newer than srcFilePath, it returns immediately.
func GetOrCreateCachedWAV(cacheDir, relPath, srcFilePath string) (string, error) {
	if cacheDir == "" {
		cacheDir = filepath.Join(".data", "cache", "audio")
	}

	// Deterministic cached filename with sha256 prefix to prevent collision
	hash := sha256.Sum256([]byte(relPath))
	baseName := stringsTrimExt(filepath.Base(relPath))
	cachedFileName := fmt.Sprintf("%s_%x.wav", baseName, hash[:6])
	cachedPath := filepath.Join(cacheDir, cachedFileName)

	srcStat, err := os.Stat(srcFilePath)
	if err != nil {
		return "", fmt.Errorf("amr: source file stat failed: %w", err)
	}

	if cacheStat, err := os.Stat(cachedPath); err == nil {
		if cacheStat.ModTime().After(srcStat.ModTime()) && cacheStat.Size() > 44 {
			return cachedPath, nil
		}
	}

	if err := TranscodeAMRFileToWAVFile(srcFilePath, cachedPath); err != nil {
		return "", fmt.Errorf("amr: transcoding failed: %w", err)
	}

	return cachedPath, nil
}

func stringsTrimExt(fileName string) string {
	ext := filepath.Ext(fileName)
	if ext == "" {
		return fileName
	}
	return fileName[:len(fileName)-len(ext)]
}
