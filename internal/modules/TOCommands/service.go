package TOCommands

import (
	"context"
	"time"
)

type IService interface {
	Record(ctx context.Context, req CreateCommandRequest) error
	GetHistory(ctx context.Context, robotID uint, from, to time.Time) ([]CommandResponse, error)
}

type service struct {
	repo IRepository
}

func NewService(repo IRepository) IService {
	return &service{repo: repo}
}

func (s *service) Record(ctx context.Context, req CreateCommandRequest) error {
	cmd := &TeleOpCommand{
		Time:    req.Time,
		RobotID: req.RobotID,
		UserID:  req.UserID,
		Command: req.Command,
		MsgID:   req.MsgID,
		T1:      req.T1,
		Lx:      req.Lx,
		Ly:      req.Ly,
		Az:      req.Az,
	}
	return s.repo.Save(ctx, cmd)
}

func (s *service) GetHistory(ctx context.Context, robotID uint, from, to time.Time) ([]CommandResponse, error) {
	cmds, err := s.repo.FindByRobotAndTimeRange(ctx, robotID, from, to)
	if err != nil {
		return nil, err
	}

	resp := make([]CommandResponse, len(cmds))
	for i := range cmds {
		resp[i] = toCommandResponse(&cmds[i])
	}
	return resp, nil
}
