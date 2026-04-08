package runtime

import (
	"context"

	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/repository"
	"github.com/nexctl/nexctl/server/pkg/errcode"
)

// Service handles node runtime state updates.
type Service struct {
	runtime repository.RuntimeStateRepository
	nodes   repository.NodeRepository
}

// NewService creates a runtime service.
func NewService(runtime repository.RuntimeStateRepository, nodes repository.NodeRepository) *Service {
	return &Service{runtime: runtime, nodes: nodes}
}

// Update stores the current runtime state for the node and refreshes heartbeat.
func (s *Service) Update(ctx context.Context, nodeID int64, req UpdateStateRequest) *errcode.AppError {
	now := req.ReportedAt()
	state := &model.NodeRuntimeState{
		NodeID:        nodeID,
		CPUPercent:    req.CPUPercent,
		MemoryPercent: req.MemoryPercent,
		DiskPercent:   req.DiskPercent,
		NetworkRxBps:  req.NetworkRxBps,
		NetworkTxBps:  req.NetworkTxBps,
		Load1:         req.Load1,
		Load5:         req.Load5,
		Load15:        req.Load15,
		UptimeSeconds: req.UptimeSeconds,
		ProcessCount:  req.ProcessCount,
		ReportedAt:    now,
	}

	if err := s.runtime.Upsert(ctx, state); err != nil {
		return errcode.Wrap(errcode.Internal, "update runtime state failed", err)
	}
	if err := s.nodes.UpdateHeartbeat(ctx, nodeID, now, model.NodeStatusOnline); err != nil {
		return errcode.Wrap(errcode.Internal, "update node heartbeat failed", err)
	}
	return nil
}

// Get returns the latest runtime state by node ID.
func (s *Service) Get(ctx context.Context, nodeID int64) (*model.NodeRuntimeState, *errcode.AppError) {
	state, err := s.runtime.GetByNodeID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "get runtime state failed", err)
	}
	return state, nil
}
