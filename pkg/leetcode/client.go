package leetcode

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

// Client represents a LeetCode API client
type Client struct {
	httpClient  *http.Client
	rateLimiter *rate.Limiter
}

// UserStats represents LeetCode user statistics
type UserStats struct {
	EasyCount   int    `json:"easySolved"`
	MediumCount int    `json:"mediumSolved"`
	HardCount   int    `json:"hardSolved"`
	TotalSolved int    `json:"totalSolved"`
	Error       string `json:"error,omitempty"`
}

// NewClient creates a new LeetCode client with rate limiting
// Rate limit: 2 requests per second with burst of 5
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		rateLimiter: rate.NewLimiter(rate.Every(500*time.Millisecond), 5), // 2 requests per second, burst of 5
	}
}

// GetUserStats fetches a user's LeetCode statistics with rate limiting and error handling
func (c *Client) GetUserStats(username string) (*UserStats, error) {
	// Wait for rate limiter
	err := c.rateLimiter.Wait(context.Background())
	if err != nil {
		return &UserStats{Error: "rate limit exceeded"}, fmt.Errorf("rate limit exceeded: %v", err)
	}

	url := fmt.Sprintf("https://leetcode.com/graphql")
	query := `{
		"query": "query getUserProfile($username: String!) { matchedUser(username: $username) { submitStats { acSubmissionNum { difficulty count } } } }",
		"variables": {"username": "` + username + `"}
	}`

	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	if err != nil {
		return &UserStats{Error: "failed to create request"}, fmt.Errorf("failed to create request: %v", err)
	}

	// Add headers to look more like a browser request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &UserStats{Error: "failed to get user stats"}, fmt.Errorf("failed to get user stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return &UserStats{Error: "user not found"}, fmt.Errorf("leetcode user not found: %s", username)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return &UserStats{Error: "rate limit exceeded"}, fmt.Errorf("leetcode API rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return &UserStats{Error: fmt.Sprintf("API returned status %d", resp.StatusCode)},
			fmt.Errorf("leetcode API returned status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			MatchedUser struct {
				SubmitStats struct {
					AcSubmissionNum []struct {
						Count      int    `json:"count"`
						Difficulty string `json:"difficulty"`
					} `json:"acSubmissionNum"`
				} `json:"submitStats"`
			} `json:"matchedUser"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return &UserStats{Error: "failed to decode response"}, fmt.Errorf("failed to decode response: %v", err)
	}

	stats := &UserStats{}
	for _, submission := range result.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		switch submission.Difficulty {
		case "Easy":
			stats.EasyCount = submission.Count
		case "Medium":
			stats.MediumCount = submission.Count
		case "Hard":
			stats.HardCount = submission.Count
		case "All":
			stats.TotalSolved = submission.Count
		}
	}

	// If total solved is not set from "All" category, calculate it
	if stats.TotalSolved == 0 {
		stats.TotalSolved = stats.EasyCount + stats.MediumCount + stats.HardCount
	}

	return stats, nil
}

// IsValidUsername checks if a LeetCode username exists
func (c *Client) IsValidUsername(username string) bool {
	stats, _ := c.GetUserStats(username)
	return stats.Error == ""
}

// GetWeeklyProgress calculates the progress made in the current week
func (c *Client) GetWeeklyProgress(username string, startDate, endDate time.Time) (*UserStats, error) {
	// TODO: Implement weekly progress calculation by comparing with previous stats
	// This will require storing historical data and comparing with current stats
	return c.GetUserStats(username)
}
