package service

import (
	"github.com/ayush/ORBIT/internal/leetcode"
)

type LeetCodeService struct {
	service *leetcode.Service
}

func NewLeetCodeService() *LeetCodeService {
	return &LeetCodeService{
		service: leetcode.NewService(),
	}
}

func (s *LeetCodeService) GetUserStats(username string) (map[string]int, error) {
	return s.service.GetUserStats(username)
}

func (s *LeetCodeService) GetContestRanking(username string) (*leetcode.ContestRankingInfo, error) {
	return s.service.GetContestRanking(username)
}

func (s *LeetCodeService) GetUserProfile(username string) (*leetcode.UserProfile, error) {
	return s.service.GetUserProfile(username)
}
