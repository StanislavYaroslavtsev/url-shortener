package repository

import (
	"context"
	"testing"
	"time"

	"github.com/StanislavYaroslavtsev/url-shortener/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresRepoTestSuite struct {
	suite.Suite
	container *postgres.PostgresContainer
	repo      *PostgresRepository
	ctx       context.Context
}

func (s *PostgresRepoTestSuite) SetupSuite() {
	s.ctx = context.Background()

	container, err := postgres.Run(s.ctx,
		"postgres:18.2-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)

	require.NoError(s.T(), err)
	s.container = container

	connStr, err := container.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	repo, err := NewPostgresRepositoryFromDSN(connStr)
	require.NoError(s.T(), err)
	s.repo = repo

	err = Migrate(s.repo, "../../migrations")
	require.NoError(s.T(), err)
}

func (s *PostgresRepoTestSuite) SetupTest() {
	_, err := s.repo.pool.Exec(s.ctx, "TRUNCATE TABLE links")
	require.NoError(s.T(), err)
}

func (s *PostgresRepoTestSuite) TestSave_SavesLink() {
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(s.T(), err)

	err = s.repo.Save(s.ctx, link)
	require.NoError(s.T(), err)

	saved, err := s.repo.Get(s.ctx, "abc123")
	require.NoError(s.T(), err)

	assert.Equal(s.T(), link, saved)
}

func (s *PostgresRepoTestSuite) TestSave_CodeExists_ReturnsError() {
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(s.T(), err)

	err = s.repo.Save(s.ctx, link)
	require.NoError(s.T(), err)

	err = s.repo.Save(s.ctx, link)
	assert.ErrorIs(s.T(), err, ErrCodeExists)
}

func (s *PostgresRepoTestSuite) TestGet_ReturnsLink() {
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(s.T(), err)

	err = s.repo.Save(s.ctx, link)
	require.NoError(s.T(), err)

	saved, err := s.repo.Get(s.ctx, "abc123")
	require.NoError(s.T(), err)
	assert.Equal(s.T(), link, saved)
}

func (s *PostgresRepoTestSuite) TestGet_NoSuchCode_ReturnsError() {
	_, err := s.repo.Get(s.ctx, "abc123")
	assert.ErrorIs(s.T(), err, ErrLinkNotFound)
}

func (s *PostgresRepoTestSuite) TestDelete_RemovesLink() {
	link, err := domain.NewLink("https://google.com", "abc123", nil)
	require.NoError(s.T(), err)

	err = s.repo.Save(s.ctx, link)
	require.NoError(s.T(), err)

	err = s.repo.Delete(s.ctx, "abc123")
	require.NoError(s.T(), err)

	_, err = s.repo.Get(s.ctx, "abc123")
	assert.ErrorIs(s.T(), err, ErrLinkNotFound)
}

func (s *PostgresRepoTestSuite) TestDelete_NoSuchCode_ReturnsError() {
	err := s.repo.Delete(s.ctx, "abc123")
	assert.ErrorIs(s.T(), err, ErrLinkNotFound)
}

func (s *PostgresRepoTestSuite) TearDownSuite() {
	if err := s.repo.Close(); err != nil {
		s.T().Fatalf("failed to close repo: %v", err)
	}
	if err := s.container.Terminate(s.ctx); err != nil {
		s.T().Fatalf("failed to terminate container: %v", err)
	}
}

func TestPostgresRepoTestSuite(t *testing.T) {
	suite.Run(t, new(PostgresRepoTestSuite))
}
