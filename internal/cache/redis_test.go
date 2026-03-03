package cache

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

type RedisCacheTestSuite struct {
	suite.Suite
	container *redis.RedisContainer
	cache     *RedisCache
	ctx       context.Context
}

func (s *RedisCacheTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := redis.Run(s.ctx,
		"redis:8.6.1-alpine",
	)
	require.NoError(s.T(), err)
	s.container = container

	connStr, err := container.ConnectionString(s.ctx)
	require.NoError(s.T(), err)

	addr := strings.TrimPrefix(connStr, "redis://")

	cache, err := NewRedisCache(addr, 1*time.Minute)
	require.NoError(s.T(), err)
	s.cache = cache
}

func (s *RedisCacheTestSuite) SetupTest() {
	err := s.cache.client.FlushDB(s.ctx).Err()
	require.NoError(s.T(), err)
}

func (s *RedisCacheTestSuite) TearDownSuite() {
	if err := s.cache.Close(); err != nil {
		s.T().Logf("failed to close cache: %v", err)
	}
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Logf("failed to terminate container: %v", err)
	}
}

func (s *RedisCacheTestSuite) TestSet_Get_ReturnsLink() {
	link, err := domain.NewLink("https://google.com", "abc123", nil, nil)
	require.NoError(s.T(), err)

	err = s.cache.Set(s.ctx, "abc123", link)
	require.NoError(s.T(), err)

	result, err := s.cache.Get(s.ctx, "abc123")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), link, result)
}

func (s *RedisCacheTestSuite) TestGet_MissingKey_ReturnsCacheMiss() {
	_, err := s.cache.Get(s.ctx, "nonexistent")
	assert.ErrorIs(s.T(), err, ErrCacheMiss)
}

func (s *RedisCacheTestSuite) TestGet_ExpiredKey_ReturnsCacheMiss() {
	connStr, err := s.container.ConnectionString(s.ctx)
	require.NoError(s.T(), err)

	cache, err := NewRedisCache(strings.TrimPrefix(connStr, "redis://"), 1*time.Millisecond)
	require.NoError(s.T(), err)
	s.T().Cleanup(func() {
		assert.NoError(s.T(), cache.Close())
	})

	link, err := domain.NewLink("https://google.com", "abc123", nil, nil)
	require.NoError(s.T(), err)

	err = cache.Set(s.ctx, "abc123", link)
	require.NoError(s.T(), err)

	time.Sleep(10 * time.Millisecond)

	_, err = cache.Get(s.ctx, "abc123")
	assert.ErrorIs(s.T(), err, ErrCacheMiss)
}

func (s *RedisCacheTestSuite) TestClose_ReturnsNoError() {
	connStr, err := s.container.ConnectionString(s.ctx)
	require.NoError(s.T(), err)

	cache, err := NewRedisCache(strings.TrimPrefix(connStr, "redis://"), 1*time.Minute)
	require.NoError(s.T(), err)

	assert.NoError(s.T(), cache.Close())
}

func TestRedisCacheTestSuite(t *testing.T) {
	suite.Run(t, new(RedisCacheTestSuite))
}
