package playback

import (
	"context"
	"time"
)

type IService interface {
	Record(ctx context.Context, robotID uint, t time.Time, data []byte) error
	GetHistory(ctx context.Context, robotID uint, from, to time.Time) ([]PlaybackFrameResponse, error)
}

type service struct {
	repo    IRepository
	storage Storage
}

func NewService(repo IRepository, storage Storage) IService {
	return &service{repo: repo, storage: storage}
}

// Record writes the raw frame to storage, then saves a metadata row pointing at it.
// Time comes from the source payload, same as the other modules.
func (s *service) Record(ctx context.Context, robotID uint, t time.Time, data []byte) error {
	path, err := s.storage.Save(robotID, t, data)
	if err != nil {
		return err
	}

	return s.repo.Save(ctx, &PlaybackFrame{
		Time:        t,
		RobotID:     robotID,
		StoragePath: path,
		SizeBytes:   int64(len(data)),
	})
}

func (s *service) GetHistory(ctx context.Context, robotID uint, from, to time.Time) ([]PlaybackFrameResponse, error) {
	frames, err := s.repo.FindByRobotAndTimeRange(ctx, robotID, from, to)
	if err != nil {
		return nil, err
	}

	resp := make([]PlaybackFrameResponse, len(frames))
	for i := range frames {
		resp[i] = toPlaybackFrameResponse(&frames[i])
	}
	return resp, nil
}
