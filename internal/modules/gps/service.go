package gps

import (
	"context"
	"time"
)

type IService interface {
	Record(ctx context.Context, req CreateGPSRequest) error
}

type service struct {
	repo IRepository
}

func NewService(repo IRepository) IService {
	return &service{repo: repo}
}

func (s *service) Record(ctx context.Context, req CreateGPSRequest) error {
	return s.repo.Save(ctx, &GPSReading{
		Time:      time.Now().UTC(),
		RobotID:   req.RobotID,
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
	})
}
