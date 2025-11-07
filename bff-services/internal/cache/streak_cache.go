package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// StreakCacheService handles Redis caching for streak and activity data
type StreakCacheService struct {
	redisClient *redis.Client
	streakTTL   time.Duration
	activityTTL time.Duration
}

// StreakData represents cached streak information
type StreakData struct {
	CurrentLen int     `json:"current_len"`
	LongestLen int     `json:"longest_len"`
	LastDay    *string `json:"last_day"`
}

// ActivityData represents cached activity information
type ActivityData struct {
	ActivityDate     string `json:"activity_dt"`
	LessonsCompleted int    `json:"lessons_completed"`
	QuizzesCompleted int    `json:"quizzes_completed"`
	Minutes          int    `json:"minutes"`
}

// WeekActivityData represents the cached weekly activity response
type WeekActivityData struct {
	Data []ActivityData `json:"data"`
}

// NewStreakCacheService creates a new streak cache service
func NewStreakCacheService(redisClient *redis.Client) *StreakCacheService {
	return &StreakCacheService{
		redisClient: redisClient,
		streakTTL:   5 * time.Minute,  // Cache streak for 5 minutes
		activityTTL: 10 * time.Minute, // Cache weekly activity for 10 minutes
	}
}

// GetStreakCacheKey returns the cache key for a user's streak
func (s *StreakCacheService) GetStreakCacheKey(userID string) string {
	return fmt.Sprintf("streak:%s", userID)
}

// GetWeekActivityCacheKey returns the cache key for a user's weekly activity
func (s *StreakCacheService) GetWeekActivityCacheKey(userID string) string {
	return fmt.Sprintf("week_activity:%s", userID)
}

// GetCachedStreak retrieves cached streak data for a user
func (s *StreakCacheService) GetCachedStreak(ctx context.Context, userID string) (*StreakData, error) {
	if s.redisClient == nil {
		return nil, nil // Return nil without error if Redis is not available
	}

	key := s.GetStreakCacheKey(userID)
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss, not an error
		}
		return nil, err // Redis error
	}

	var streakData StreakData
	if err := json.Unmarshal([]byte(val), &streakData); err != nil {
		return nil, err
	}

	return &streakData, nil
}

// CacheStreak stores streak data for a user
func (s *StreakCacheService) CacheStreak(ctx context.Context, userID string, streak *StreakData) error {
	if s.redisClient == nil {
		return nil // Skip if Redis is not available
	}

	key := s.GetStreakCacheKey(userID)
	data, err := json.Marshal(streak)
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, string(data), s.streakTTL).Err()
}

// GetCachedWeekActivity retrieves cached weekly activity data for a user
func (s *StreakCacheService) GetCachedWeekActivity(ctx context.Context, userID string) ([]ActivityData, error) {
	if s.redisClient == nil {
		return nil, nil // Return nil without error if Redis is not available
	}

	key := s.GetWeekActivityCacheKey(userID)
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss, not an error
		}
		return nil, err // Redis error
	}

	var activityData WeekActivityData
	if err := json.Unmarshal([]byte(val), &activityData); err != nil {
		return nil, err
	}

	return activityData.Data, nil
}

// CacheWeekActivity stores weekly activity data for a user
func (s *StreakCacheService) CacheWeekActivity(ctx context.Context, userID string, activities []ActivityData) error {
	if s.redisClient == nil {
		return nil // Skip if Redis is not available
	}

	key := s.GetWeekActivityCacheKey(userID)
	data, err := json.Marshal(WeekActivityData{Data: activities})
	if err != nil {
		return err
	}

	return s.redisClient.Set(ctx, key, string(data), s.activityTTL).Err()
}

// InvalidateStreakCache removes streak cache for a user
func (s *StreakCacheService) InvalidateStreakCache(ctx context.Context, userID string) error {
	if s.redisClient == nil {
		return nil
	}

	key := s.GetStreakCacheKey(userID)
	return s.redisClient.Del(ctx, key).Err()
}

// InvalidateWeekActivityCache removes weekly activity cache for a user
func (s *StreakCacheService) InvalidateWeekActivityCache(ctx context.Context, userID string) error {
	if s.redisClient == nil {
		return nil
	}

	key := s.GetWeekActivityCacheKey(userID)
	return s.redisClient.Del(ctx, key).Err()
}

// InvalidateAllUserCache invalidates all cache for a user
func (s *StreakCacheService) InvalidateAllUserCache(ctx context.Context, userID string) error {
	if s.redisClient == nil {
		return nil
	}

	streakErr := s.InvalidateStreakCache(ctx, userID)
	weekErr := s.InvalidateWeekActivityCache(ctx, userID)

	if streakErr != nil {
		return streakErr
	}
	return weekErr
}
