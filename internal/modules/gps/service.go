package gps

import (
	"context"
	"time"
)

type IService interface {
	Record(ctx context.Context, req CreateGPSRequest) error
	GetHistory(ctx context.Context, robotID uint, from, to time.Time) ([]GPSResponse, error)
}

type service struct {
	repo IRepository
}

func NewService(repo IRepository) IService {
	return &service{repo: repo}
}

func (s *service) Record(ctx context.Context, req CreateGPSRequest) error {
	return s.repo.Save(ctx, &GPSReading{
		Time:      req.Time,
		RobotID:   req.RobotID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
}

func (s *service) GetHistory(ctx context.Context, robotID uint, from, to time.Time) ([]GPSResponse, error) {
	readings, err := s.repo.FindByRobotAndTimeRange(ctx, robotID, from, to)
	if err != nil {
		return nil, err
	}

	resp := make([]GPSResponse, len(readings))
	for i := range readings {
		resp[i] = toGPSResponse(&readings[i])
	}
	return resp, nil
}
