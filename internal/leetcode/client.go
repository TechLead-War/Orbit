package leetcode

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
}

type UserStats struct {
	EasyCount   int `json:"easySolved"`
	MediumCount int `json:"mediumSolved"`
	HardCount   int `json:"hardSolved"`
	TotalSolved int `json:"totalSolved"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) GetUserStats(username string) (*UserStats, error) {
	url := fmt.Sprintf("https://leetcode.com/api/users/%s", username)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API returned status %d", resp.StatusCode)
	}

	var result struct {
		SubmitStats struct {
			AcSubmissionNum []struct {
				Count      int    `json:"count"`
				Difficulty string `json:"difficulty"`
			} `json:"acSubmissionNum"`
		} `json:"submitStats"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	stats := &UserStats{}
	for _, submission := range result.SubmitStats.AcSubmissionNum {
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

	return stats, nil
}
