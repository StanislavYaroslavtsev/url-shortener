package grpc

import (
	"context"

	"github.com/StanislavYaroslavtsev/url-shortener/services/stats/internal/repository"
	pb "github.com/StanislavYaroslavtsev/url-shortener/shared/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type StatsServer struct {
	pb.UnimplementedStatsServiceServer
	repo repository.EventRepository
}

func NewStatsServer(repo repository.EventRepository) *StatsServer {
	return &StatsServer{repo: repo}
}

func (s *StatsServer) GetStats(ctx context.Context, req *pb.GetStatsRequest) (*pb.GetStatsResponse, error) {
	stats, err := s.repo.GetStats(ctx, req.Code)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
	}

	topCountries := make([]*pb.CountryStats, len(stats.TopCountries))
	for i, c := range stats.TopCountries {
		topCountries[i] = &pb.CountryStats{
			Country: c.Country,
			Clicks:  c.Clicks,
		}
	}

	return &pb.GetStatsResponse{
		Code:          req.Code,
		TotalClicks:   stats.TotalClicks,
		LastClickedAt: stats.LastClickedAt.String(),
		TopCountries:  topCountries,
	}, nil
}
