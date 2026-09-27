package youtube

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/youtubeanalytics/v2"
)

// AnalyticsDailyRow represents single day metrics from YouTube Studio.
type AnalyticsDailyRow struct {
	Day                     string  `json:"day"`
	Views                   int64   `json:"views"`
	MinutesWatched          float64 `json:"minutes_watched"`
	AverageViewDurationSecs float64 `json:"average_view_duration_secs"`
	SubscribersGained       int64   `json:"subscribers_gained"`
	SubscribersLost         int64   `json:"subscribers_lost"`
	Likes                   int64   `json:"likes"`
}

// AnalyticsReport synthesizes viewer retention and performance feedback loops.
type AnalyticsReport struct {
	StartDate               string              `json:"start_date"`
	EndDate                 string              `json:"end_date"`
	TotalViews              int64               `json:"total_views"`
	TotalMinutesWatched     float64             `json:"total_minutes_watched"`
	AverageViewDurationSecs float64             `json:"average_view_duration_secs"`
	SubscribersGained       int64               `json:"subscribers_gained"`
	SubscribersLost         int64               `json:"subscribers_lost"`
	NetSubscribers          int64               `json:"net_subscribers"`
	TotalComments           int64               `json:"total_comments"`
	TotalLikes              int64               `json:"total_likes"`
	DailyRows               []AnalyticsDailyRow `json:"daily_rows"`
}

// GetAnalyticsReport queries YouTube Analytics API v2 for viewer performance metrics.
func (c *Client) GetAnalyticsReport(ctx context.Context, startDate, endDate string) (*AnalyticsReport, error) {
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
		Metrics("views,estimatedMinutesWatched,averageViewDuration,subscribersGained,subscribersLost,comments,likes").
		Dimensions("day").
		Sort("day")

	call.Context(ctx)
	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to query youtube analytics: %w", err)
	}

	c.uploader.trackQuota(1)

	report := &AnalyticsReport{
		StartDate: startDate,
		EndDate:   endDate,
		DailyRows: make([]AnalyticsDailyRow, 0),
	}

	// Column indices from headers:
	// day (0), views (1), estimatedMinutesWatched (2), averageViewDuration (3),
	// subscribersGained (4), subscribersLost (5), comments (6), likes (7)
	for _, row := range resp.Rows {
		if len(row) < 8 {
			continue
		}

		day := fmt.Sprintf("%v", row[0])
		views := toInt64(row[1])
		mins := toFloat64(row[2])
		avd := toFloat64(row[3])
		subG := toInt64(row[4])
		subL := toInt64(row[5])
		comments := toInt64(row[6])
		likes := toInt64(row[7])

		report.TotalViews += views
		report.TotalMinutesWatched += mins
		report.SubscribersGained += subG
		report.SubscribersLost += subL
		report.TotalComments += comments
		report.TotalLikes += likes

		report.DailyRows = append(report.DailyRows, AnalyticsDailyRow{
			Day:                     day,
			Views:                   views,
			MinutesWatched:          mins,
			AverageViewDurationSecs: avd,
			SubscribersGained:       subG,
			SubscribersLost:         subL,
			Likes:                   likes,
		})
	}

	report.NetSubscribers = report.SubscribersGained - report.SubscribersLost
	if report.TotalViews > 0 {
		report.AverageViewDurationSecs = (report.TotalMinutesWatched * 60) / float64(report.TotalViews)
	}

	return report, nil
}

func toInt64(val any) int64 {
	switch v := val.(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func toFloat64(val any) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0.0
	}
}
