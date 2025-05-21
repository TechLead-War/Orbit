package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type LeetCodeService struct {
	client    *http.Client
	rateLimit *time.Ticker
}

type LeetCodeUserProfile struct {
	Data struct {
		MatchedUser struct {
			Profile struct {
				RealName string `json:"realName"`
				Ranking  int    `json:"ranking"`
			} `json:"profile"`
			SubmitStats struct {
				AcSubmissionNum []struct {
					Difficulty string `json:"difficulty"`
					Count      int    `json:"count"`
				} `json:"acSubmissionNum"`
			} `json:"submitStats"`
		} `json:"matchedUser"`
	} `json:"data"`
}

type ContestRankingInfo struct {
	Data struct {
		UserContestRanking struct {
			AttendedContestsCount int     `json:"attendedContestsCount"`
			Rating                float64 `json:"rating"`
			GlobalRanking         int     `json:"globalRanking"`
			TopPercentage         float64 `json:"topPercentage"`
		} `json:"userContestRanking"`
		UserContestRankingHistory []struct {
			Contest struct {
				Title string `json:"title"`
			} `json:"contest"`
			Rating              float64 `json:"rating"`
			Ranking             int     `json:"ranking"`
			Attended            bool    `json:"attended"`
			TrendDirection      string  `json:"trendDirection"`
			ProblemsSolved      int     `json:"problemsSolved"`
			FinishTimeInSeconds int     `json:"finishTimeInSeconds"`
		} `json:"userContestRankingHistory"`
	} `json:"data"`
}

func NewLeetCodeService() *LeetCodeService {
	return &LeetCodeService{
		client:    &http.Client{Timeout: 10 * time.Second},
		rateLimit: time.NewTicker(500 * time.Millisecond),
	}
}

func (s *LeetCodeService) GetUserProfile(username string) (*LeetCodeUserProfile, error) {
	<-s.rateLimit.C

	query := `{
		"query": "query getUserProfile($username: String!) { matchedUser(username: $username) { profile { realName ranking } submitStats { acSubmissionNum { difficulty count } } } }",
		"variables": { "username": "%s" }
	}`

	jsonStr := []byte(fmt.Sprintf(query, username))
	req, err := http.NewRequest("POST", "https://leetcode.com/graphql", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API returned non-200 status code: %d", resp.StatusCode)
	}

	var profile LeetCodeUserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &profile, nil
}

func (s *LeetCodeService) GetUserStats(username string) (map[string]int, error) {
	profile, err := s.GetUserProfile(username)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	stats["ranking"] = profile.Data.MatchedUser.Profile.Ranking

	for _, submission := range profile.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		stats[submission.Difficulty] = submission.Count
	}

	return stats, nil
}

func (s *LeetCodeService) GetContestRanking(username string) (*ContestRankingInfo, error) {
	<-s.rateLimit.C

	query := `{
		"query": "query userContestRankingInfo($username: String!) { userContestRanking(username: $username) { attendedContestsCount rating globalRanking topPercentage } userContestRankingHistory(username: $username) { contest { title } rating ranking attended trendDirection problemsSolved finishTimeInSeconds } }",
		"variables": { "username": "%s" }
	}`

	jsonStr := []byte(fmt.Sprintf(query, username))
	req, err := http.NewRequest("POST", "https://leetcode.com/graphql", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API returned non-200 status code: %d", resp.StatusCode)
	}

	var contestInfo ContestRankingInfo
	if err := json.NewDecoder(resp.Body).Decode(&contestInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &contestInfo, nil
}
