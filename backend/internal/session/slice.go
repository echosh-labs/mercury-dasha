package session

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// WavInfo captures parsed header details from a PCM WAV file.
type WavInfo struct {
	Channels      int
	SampleRate    int
	ByteRate      int
	BlockAlign    int
	BitsPerSample int
	DataOffset    int64
	DataSize      int64
	DurationSec   float64
}

// GenerateWavHeader builds a canonical 44-byte RIFF/WAVE header for linear PCM audio.
func GenerateWavHeader(numChannels, sampleRate, bitsPerSample int, dataBytes int64) []byte {
	blockAlign := numChannels * (bitsPerSample / 8)
	byteRate := sampleRate * blockAlign
	chunkSize := uint32(36 + dataBytes)

	buf := make([]byte, 44)

	// "RIFF" chunk descriptor
	copy(buf[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(buf[4:8], chunkSize)
	copy(buf[8:12], []byte("WAVE"))

	// "fmt " subchunk
	copy(buf[12:16], []byte("fmt "))
	binary.LittleEndian.PutUint32(buf[16:20], 16) // Subchunk1Size = 16 for PCM
	binary.LittleEndian.PutUint16(buf[20:22], 1)  // AudioFormat = 1 (PCM)
	binary.LittleEndian.PutUint16(buf[22:24], uint16(numChannels))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(buf[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(buf[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(buf[34:36], uint16(bitsPerSample))

	// "data" subchunk
	copy(buf[36:40], []byte("data"))
	binary.LittleEndian.PutUint32(buf[40:44], uint32(dataBytes))

	return buf
}

// ParseWavHeader parses a RIFF/WAVE stream to locate format specifications and the exact data chunk offset.
func ParseWavHeader(r io.ReaderAt, fileSize int64) (*WavInfo, error) {
	if fileSize < 44 {
		return nil, fmt.Errorf("file size %d too small for a valid WAV header", fileSize)
	}

	headerBuf := make([]byte, 12)
	if _, err := r.ReadAt(headerBuf, 0); err != nil {
		return nil, fmt.Errorf("failed to read RIFF header: %w", err)
	}

	if string(headerBuf[0:4]) != "RIFF" || string(headerBuf[8:12]) != "WAVE" {
		return nil, fmt.Errorf("not a valid RIFF/WAVE file")
	}

	info := &WavInfo{
		Channels:      2,
		SampleRate:    48000,
		BitsPerSample: 16,
	}

	offset := int64(12)
	chunkHeader := make([]byte, 8)
	foundFmt := false
	foundData := false

	for offset+8 <= fileSize {
		if _, err := r.ReadAt(chunkHeader, offset); err != nil {
			break
		}
		chunkID := string(chunkHeader[0:4])
		chunkSize := int64(binary.LittleEndian.Uint32(chunkHeader[4:8]))
		offset += 8

		switch chunkID {
		case "fmt ":
			fmtData := make([]byte, chunkSize)
			if _, err := r.ReadAt(fmtData, offset); err != nil {
				return nil, fmt.Errorf("failed to read fmt chunk: %w", err)
			}
			if len(fmtData) >= 14 {
				audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
				if audioFormat != 1 && audioFormat != 3 { // 1=PCM, 3=IEEE Float
					return nil, fmt.Errorf("unsupported WAV audio format %d (only PCM/Float supported)", audioFormat)
				}
				info.Channels = int(binary.LittleEndian.Uint16(fmtData[2:4]))
				info.SampleRate = int(binary.LittleEndian.Uint32(fmtData[4:8]))
				info.ByteRate = int(binary.LittleEndian.Uint32(fmtData[8:12]))
				info.BlockAlign = int(binary.LittleEndian.Uint16(fmtData[12:14]))
			}
			if len(fmtData) >= 16 {
				info.BitsPerSample = int(binary.LittleEndian.Uint16(fmtData[14:16]))
			}
			foundFmt = true

		case "data":
			info.DataOffset = offset
			info.DataSize = chunkSize
			if info.DataOffset+info.DataSize > fileSize {
				info.DataSize = fileSize - info.DataOffset
			}
			foundData = true
			break
		}

		if foundData {
			break
		}
		// Chunks in RIFF are word-aligned (2 bytes)
		if chunkSize%2 != 0 {
			chunkSize++
		}
		offset += chunkSize
	}

	if !foundFmt || !foundData {
		return nil, fmt.Errorf("invalid WAV: missing fmt or data chunk")
	}

	if info.BlockAlign == 0 && info.Channels > 0 && info.BitsPerSample > 0 {
		info.BlockAlign = info.Channels * (info.BitsPerSample / 8)
	}

	if info.ByteRate == 0 && info.SampleRate > 0 && info.BlockAlign > 0 {
		info.ByteRate = info.SampleRate * info.BlockAlign
	}

	if info.ByteRate > 0 {
		info.DurationSec = float64(info.DataSize) / float64(info.ByteRate)
	}

	return info, nil
}

// ExtractPrecisionSlice streams a sample-accurate slice from an open WAV file directly to a writer.
// Zero transcoding loss: raw PCM frames are extracted with sample-accurate microsecond alignment.
func ExtractPrecisionSlice(r io.ReaderAt, fileSize int64, startMs, endMs int64, w io.Writer) (*AudioSlice, error) {
	if startMs < 0 || endMs <= startMs {
		return nil, ErrInvalidSliceRange
	}

	info, err := ParseWavHeader(r, fileSize)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source WAV: %w", err)
	}

	bytesPerSample := info.BitsPerSample / 8
	if bytesPerSample <= 0 {
		bytesPerSample = 2
	}
	frameSize := int64(info.Channels * bytesPerSample)

	totalFrames := info.DataSize / frameSize
	sessionDurationMs := int64(info.DurationSec * 1000)

	if startMs >= sessionDurationMs && sessionDurationMs > 0 {
		return nil, fmt.Errorf("start time %d ms exceeds session duration %d ms", startMs, sessionDurationMs)
	}

	startFrame := (startMs * int64(info.SampleRate)) / 1000
	if startFrame > totalFrames {
		startFrame = totalFrames
	}

	endFrame := (endMs * int64(info.SampleRate)) / 1000
	if endFrame > totalFrames {
		endFrame = totalFrames
	}

	if endFrame <= startFrame {
		return nil, fmt.Errorf("requested slice contains 0 audio frames")
	}

	sliceFrames := endFrame - startFrame
	sliceDataBytes := sliceFrames * frameSize
	startByteOffset := info.DataOffset + (startFrame * frameSize)

	// Synthesize valid 44-byte RIFF header for this exact slice
	header := GenerateWavHeader(info.Channels, info.SampleRate, info.BitsPerSample, sliceDataBytes)
	if _, err := w.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write slice WAV header: %w", err)
	}

	// Stream slice audio samples in 64KB buffers
	buf := make([]byte, 64*1024)
	bytesRemaining := sliceDataBytes
	currentOffset := startByteOffset

	for bytesRemaining > 0 {
		readSize := int64(len(buf))
		if readSize > bytesRemaining {
			readSize = bytesRemaining
		}

		n, err := r.ReadAt(buf[:readSize], currentOffset)
		if n > 0 {
			if _, writeErr := w.Write(buf[:n]); writeErr != nil {
				return nil, fmt.Errorf("failed writing slice data: %w", writeErr)
			}
			bytesRemaining -= int64(n)
			currentOffset += int64(n)
		}
		if err != nil {
			if err == io.EOF && bytesRemaining > 0 {
				break
			}
			if err != io.EOF {
				return nil, fmt.Errorf("error reading source PCM data: %w", err)
			}
		}
	}

	slice := &AudioSlice{
		ID:          fmt.Sprintf("slice-%d-%d", startMs, endMs),
		StartMs:     startMs,
		EndMs:       endMs,
		DurationSec: float64(sliceFrames) / float64(info.SampleRate),
		SampleRate:  info.SampleRate,
		Channels:    info.Channels,
		BitDepth:    info.BitsPerSample,
		Format:      "wav",
		SizeBytes:   44 + sliceDataBytes,
		CreatedAt:   time.Now(),
	}

	return slice, nil
}

// SavePrecisionSlice extracts and saves a standalone slice to a file on disk.
func SavePrecisionSlice(sourcePath, destPath string, startMs, endMs int64, label, category string) (*AudioSlice, error) {
	file, err := os.Open(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open source audio file: %w", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to stat source audio file: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create destination directory: %w", err)
	}

	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create slice file: %w", err)
	}
	defer destFile.Close()

	slice, err := ExtractPrecisionSlice(file, fi.Size(), startMs, endMs, destFile)
	if err != nil {
		os.Remove(destPath)
		return nil, err
	}

	slice.Label = label
	slice.Category = category
	slice.FilePath = destPath

	return slice, nil
}

// AppendChunkToFile writes an incoming audio chunk directly to disk.
// If the file does not exist, it creates it.
func AppendChunkToFile(filePath string, chunk []byte) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return 0, fmt.Errorf("failed to create directory %s: %w", filepath.Dir(filePath), err)
	}

	f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return 0, fmt.Errorf("failed to open chunk file: %w", err)
	}
	defer f.Close()

	if _, err := f.Write(chunk); err != nil {
		return 0, fmt.Errorf("failed to write chunk: %w", err)
	}

	fi, err := f.Stat()
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}

// UpdateWavHeaderSizes updates the 44-byte RIFF header of an existing WAV file on disk with accurate final sizes.
func UpdateWavHeaderSizes(filePath string) error {
	f, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return err
	}
	fileSize := fi.Size()
	if fileSize < 44 {
		return nil
	}

	dataSize := uint32(fileSize - 44)
	riffChunkSize := uint32(fileSize - 8)

	// Verify it has RIFF header
	magic := make([]byte, 4)
	if _, err := f.ReadAt(magic, 0); err != nil || string(magic) != "RIFF" {
		return nil // Not a standard WAV file, leave alone
	}

	riffBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(riffBuf, riffChunkSize)
	if _, err := f.WriteAt(riffBuf, 4); err != nil {
		return err
	}

	dataBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataBuf, dataSize)
	if _, err := f.WriteAt(dataBuf, 40); err != nil {
		return err
	}

	return nil
}

// SynthesizeSilentWav creates a minimal valid PCM WAV file buffer for tests or placeholders.
func SynthesizeSilentWav(channels, sampleRate, bitsPerSample int, durationSec float64) []byte {
	blockAlign := channels * (bitsPerSample / 8)
	totalSamples := int64(float64(sampleRate) * durationSec)
	dataBytes := totalSamples * int64(blockAlign)

	header := GenerateWavHeader(channels, sampleRate, bitsPerSample, dataBytes)
	pcm := make([]byte, dataBytes)

	buf := bytes.NewBuffer(header)
	buf.Write(pcm)
	return buf.Bytes()
}
