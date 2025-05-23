package leetcode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Service struct {
	client    *http.Client
	rateLimit *time.Ticker
}

type UserProfile struct {
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

func NewService() *Service {
	return &Service{
		client:    &http.Client{Timeout: 10 * time.Second},
		rateLimit: time.NewTicker(500 * time.Millisecond),
	}
}

func (s *Service) GetUserProfile(username string) (*UserProfile, error) {
	<-s.rateLimit.C

	query := map[string]interface{}{
		"query": `
			query getUserProfile($username: String!) {
				matchedUser(username: $username) {
					profile {
						realName
						ranking
					}
					submitStats {
						acSubmissionNum {
							difficulty
							count
						}
					}
				}
			}
		`,
		"variables": map[string]interface{}{
			"username": username,
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequest("POST", "https://leetcode.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("leetcode API returned non-200 status code: %d", resp.StatusCode)
	}

	var profile UserProfile
	if err := json.NewDecoder(resp.Body).Decode(&profile); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if profile.Data.MatchedUser.Profile.Ranking == 0 {
		return nil, fmt.Errorf("user not found or has no submissions")
	}

	return &profile, nil
}

func (s *Service) GetUserStats(username string) (map[string]int, error) {
	profile, err := s.GetUserProfile(username)
	if err != nil {
		return nil, err
	}

	stats := make(map[string]int)
	stats["ranking"] = profile.Data.MatchedUser.Profile.Ranking

	total := 0
	for _, submission := range profile.Data.MatchedUser.SubmitStats.AcSubmissionNum {
		stats[submission.Difficulty] = submission.Count
		total += submission.Count
	}
	stats["All"] = total

	return stats, nil
}

func (s *Service) GetContestRanking(username string) (*ContestRankingInfo, error) {
	<-s.rateLimit.C

	query := map[string]interface{}{
		"query": `
			query userContestRankingInfo($username: String!) {
				userContestRanking(username: $username) {
					attendedContestsCount
					rating
					globalRanking
					topPercentage
				}
				userContestRankingHistory(username: $username) {
					contest {
						title
					}
					rating
					ranking
					attended
					trendDirection
					problemsSolved
					finishTimeInSeconds
				}
			}
		`,
		"variables": map[string]interface{}{
			"username": username,
		},
	}

	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequest("POST", "https://leetcode.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0")

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
