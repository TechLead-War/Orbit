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

// UserStats represents a user's LeetCode statistics
type UserStats struct {
	TotalSolved    int    `json:"total_solved"`
	EasyCount      int    `json:"easy_count"`
	MediumCount    int    `json:"medium_count"`
	HardCount      int    `json:"hard_count"`
	ContestRating  int    `json:"contest_rating"`
	ContestRanking int    `json:"contest_ranking"`
	Error          string `json:"error,omitempty"`
}

// UserProfile represents a user's LeetCode profile
type UserProfile struct {
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

// ContestInfo represents a user's LeetCode contest information
type ContestInfo struct {
	Data struct {
		UserContestRanking struct {
			Rating        int `json:"rating"`
			GlobalRanking int `json:"globalRanking"`
		} `json:"userContestRanking"`
	} `json:"data"`
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

// GetUserStats retrieves a user's LeetCode statistics
func (c *Client) GetUserStats(username string) (*UserStats, error) {
	// Get user profile for problem counts
	profile, err := c.GetUserProfile(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get contest ranking for contest stats
	contestInfo, err := c.GetContestRanking(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest ranking: %w", err)
	}

	stats := &UserStats{
		TotalSolved:    0,
		EasyCount:      0,
		MediumCount:    0,
		HardCount:      0,
		ContestRating:  contestInfo.Data.UserContestRanking.Rating,
		ContestRanking: contestInfo.Data.UserContestRanking.GlobalRanking,
	}

	// Map submission stats
	for _, submission := range profile.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		count := submission.Count
		switch submission.Difficulty {
		case "Easy":
			stats.EasyCount = count
		case "Medium":
			stats.MediumCount = count
		case "Hard":
			stats.HardCount = count
		}
	}
	stats.TotalSolved = stats.EasyCount + stats.MediumCount + stats.HardCount

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

// GetUserProfile retrieves a user's LeetCode profile
func (c *Client) GetUserProfile(username string) (*UserProfile, error) {
	// Wait for rate limiter
	err := c.rateLimiter.Wait(context.Background())
	if err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	url := "https://leetcode.com/graphql"
	query := `{
		"query": "query getUserProfile($username: String!) { matchedUser(username: $username) { submitStats { acSubmissionNum { difficulty count } } } }",
		"variables": {"username": "` + username + `"}
	}`

	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to look more like a browser request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("Referer", "https://leetcode.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("leetcode user not found: %s", username)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("leetcode API rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API returned status %d", resp.StatusCode)
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Verify we got valid data
	if len(profile.Data.MatchedUser.SubmitStats.AcSubmissionNum) == 0 {
		return nil, fmt.Errorf("no submission data found for user: %s", username)
	}

	return &profile, nil
}

// GetContestRanking retrieves a user's LeetCode contest ranking
func (c *Client) GetContestRanking(username string) (*ContestInfo, error) {
	// Wait for rate limiter
	err := c.rateLimiter.Wait(context.Background())
	if err != nil {
		return nil, fmt.Errorf("rate limit exceeded: %w", err)
	}

	url := "https://leetcode.com/graphql"
	query := `{
		"query": "query getUserContestInfo($username: String!) { userContestRanking(username: $username) { rating globalRanking } }",
		"variables": {"username": "` + username + `"}
	}`

	req, err := http.NewRequest("POST", url, strings.NewReader(query))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add headers to look more like a browser request
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.114 Safari/537.36")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Origin", "https://leetcode.com")
	req.Header.Set("Referer", "https://leetcode.com/")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get contest ranking: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("leetcode user not found: %s", username)
	}

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("leetcode API rate limit exceeded")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API returned status %d", resp.StatusCode)
	}

	var contestInfo ContestInfo
	if err := json.NewDecoder(resp.Body).Decode(&contestInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &contestInfo, nil
}
