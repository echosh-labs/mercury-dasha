package youtube

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtubeanalytics/v2"
)

// FinancialDailyRow captures daily revenue, ad performance, and CPM metrics.
type FinancialDailyRow struct {
	Day                        string  `json:"day"`
	EstimatedRevenue           float64 `json:"estimated_revenue"`
	EstimatedAdRevenue         float64 `json:"estimated_ad_revenue"`
	EstimatedRedPartnerRevenue float64 `json:"estimated_red_partner_revenue"`
	GrossRevenue               float64 `json:"gross_revenue"`
	CPM                        float64 `json:"cpm"`
	PlaybackBasedCPM           float64 `json:"playback_based_cpm"`
	MonetizedPlaybacks         int64   `json:"monetized_playbacks"`
}

// FinancialReport aggregates YouTube monetization metrics for AMRA correlation.
type FinancialReport struct {
	StartDate                  string              `json:"start_date"`
	EndDate                    string              `json:"end_date"`
	Monetized                  bool                `json:"monetized"`
	Currency                   string              `json:"currency"`
	TotalEstimatedRevenue      float64             `json:"total_estimated_revenue"`
	TotalEstimatedAdRevenue    float64             `json:"total_estimated_ad_revenue"`
	TotalEstimatedRedRevenue   float64             `json:"total_estimated_red_revenue"`
	TotalGrossRevenue          float64             `json:"total_gross_revenue"`
	AverageCPM                 float64             `json:"average_cpm"`
	AveragePlaybackBasedCPM    float64             `json:"average_playback_based_cpm"`
	TotalMonetizedPlaybacks    int64               `json:"total_monetized_playbacks"`
	StatusMessage              string              `json:"status_message,omitempty"`
	DailyRows                  []FinancialDailyRow `json:"daily_rows"`
}

// GetFinancialReport queries YouTube Analytics for monetization, ad revenue, and CPM metrics.
func (c *Client) GetFinancialReport(ctx context.Context, startDate, endDate string) (*FinancialReport, error) {
	if startDate == "" {
		startDate = time.Now().UTC().AddDate(0, 0, -28).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")
	}

	tokenSource, err := c.auth.TokenSource(ctx)
	if err != nil {
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	service, err := youtubeanalytics.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("failed to init youtube analytics service: %w", err)
	}

	call := service.Reports.Query().
		Ids("channel==MINE").
		StartDate(startDate).
		EndDate(endDate).
		Metrics("estimatedRevenue,estimatedAdRevenue,estimatedRedPartnerRevenue,grossRevenue,cpm,playbackBasedCpm,monetizedPlaybacks").
		Dimensions("day").
		Sort("day")

	call.Context(ctx)
	resp, err := call.Do()
	if err != nil {
		// Non-monetized channels or permissions missing: return zeroed report gracefully
		errMsg := err.Error()
		if strings.Contains(errMsg, "403") || strings.Contains(errMsg, "400") || strings.Contains(errMsg, "forbidden") {
			return &FinancialReport{
				StartDate:     startDate,
				EndDate:       endDate,
				Monetized:     false,
				Currency:      "USD",
				StatusMessage: "Channel is not yet enrolled in the YouTube Partner Program (YPP) or monetization reporting is pending.",
				DailyRows:     make([]FinancialDailyRow, 0),
			}, nil
		}
		return nil, fmt.Errorf("failed to query financial report: %w", err)
	}

	c.uploader.trackQuota(1)

	report := &FinancialReport{
		StartDate: startDate,
		EndDate:   endDate,
		Monetized: true,
		Currency:  "USD",
		DailyRows: make([]FinancialDailyRow, 0),
	}

	var sumCPM float64
	var sumPlaybackCPM float64
	cpmCount := 0

	// Column order:
	// day (0), estimatedRevenue (1), estimatedAdRevenue (2), estimatedRedPartnerRevenue (3),
	// grossRevenue (4), cpm (5), playbackBasedCpm (6), monetizedPlaybacks (7)
	for _, row := range resp.Rows {
		if len(row) < 8 {
			continue
		}

		day := fmt.Sprintf("%v", row[0])
		estRev := toFloat64(row[1])
		adRev := toFloat64(row[2])
		redRev := toFloat64(row[3])
		grossRev := toFloat64(row[4])
		cpm := toFloat64(row[5])
		pbCPM := toFloat64(row[6])
		monPlays := toInt64(row[7])

		report.TotalEstimatedRevenue += estRev
		report.TotalEstimatedAdRevenue += adRev
		report.TotalEstimatedRedRevenue += redRev
		report.TotalGrossRevenue += grossRev
		report.TotalMonetizedPlaybacks += monPlays

		if cpm > 0 {
			sumCPM += cpm
			cpmCount++
		}
		if pbCPM > 0 {
			sumPlaybackCPM += pbCPM
		}

		report.DailyRows = append(report.DailyRows, FinancialDailyRow{
			Day:                        day,
			EstimatedRevenue:           estRev,
			EstimatedAdRevenue:         adRev,
			EstimatedRedPartnerRevenue: redRev,
			GrossRevenue:               grossRev,
			CPM:                        cpm,
			PlaybackBasedCPM:           pbCPM,
			MonetizedPlaybacks:         monPlays,
		})
	}

	if cpmCount > 0 {
		report.AverageCPM = sumCPM / float64(cpmCount)
		report.AveragePlaybackBasedCPM = sumPlaybackCPM / float64(cpmCount)
	}

	return report, nil
}
