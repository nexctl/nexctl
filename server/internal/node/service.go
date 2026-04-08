package node

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/internal/model"
	"github.com/nexctl/nexctl/server/internal/repository"
	"github.com/nexctl/nexctl/server/pkg/errcode"
)

// Service implements node registration and query business logic.
type Service struct {
	cfg           config.NodeConfig
	installTokens repository.InstallTokenRepository
	nodes         repository.NodeRepository
	runtime       repository.RuntimeStateRepository
	audit         *audit.Service
	externalURL   string
}

// NewService creates a node service.
func NewService(cfg config.NodeConfig, installTokens repository.InstallTokenRepository, nodes repository.NodeRepository, runtime repository.RuntimeStateRepository, audit *audit.Service, externalURL string) *Service {
	return &Service{
		cfg:           cfg,
		installTokens: installTokens,
		nodes:         nodes,
		runtime:       runtime,
		audit:         audit,
		externalURL:   strings.TrimRight(externalURL, "/"),
	}
}

// Register creates or completes node registration: enrollment_token (console pre-created node) or install_token (legacy).
func (s *Service) Register(ctx context.Context, req RegisterRequest) (*RegisterResponse, *errcode.AppError) {
	if strings.TrimSpace(req.EnrollmentToken) != "" {
		return s.registerWithEnrollment(ctx, req)
	}
	return s.registerWithInstallToken(ctx, req)
}

func (s *Service) registerWithInstallToken(ctx context.Context, req RegisterRequest) (*RegisterResponse, *errcode.AppError) {
	if strings.TrimSpace(req.InstallToken) == "" || strings.TrimSpace(req.NodeKey) == "" {
		return nil, errcode.New(errcode.InvalidArgument, "install_token and node_key are required")
	}

	token, err := s.installTokens.FindByToken(ctx, req.InstallToken)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "query install token failed", err)
	}
	if !repository.IsUsable(token, time.Now().UTC()) {
		return nil, errcode.New(errcode.InstallTokenInvalid, "install token is invalid")
	}

	agentID, err := randomHex(12)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate agent_id failed", err)
	}
	agentSecret, err := randomHex(24)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate agent_secret failed", err)
	}

	now := time.Now().UTC()
	record := &model.Node{
		AgentID:         agentID,
		AgentSecret:     agentSecret,
		NodeKey:         req.NodeKey,
		Name:            req.Name,
		Hostname:        req.Hostname,
		Platform:        req.Platform,
		PlatformVersion: req.PlatformVersion,
		Arch:            req.Arch,
		AgentVersion:    req.AgentVersion,
		Status:          model.NodeStatusOnline,
		LastHeartbeatAt: now,
		LastOnlineAt:    now,
	}

	if err := s.nodes.Create(ctx, record); err != nil {
		return nil, errcode.Wrap(errcode.Internal, "create node failed", err)
	}
	if err := s.installTokens.IncrementUsedCount(ctx, token.ID); err != nil {
		return nil, errcode.Wrap(errcode.Internal, "update install token failed", err)
	}

	detailJSON := "{}"
	if b, err := json.Marshal(map[string]string{"node_key": record.NodeKey}); err == nil {
		detailJSON = string(b)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "agent",
		ActorID:      record.AgentID,
		ActorName:    record.Name,
		Action:       "node.register",
		ResourceType: "node",
		ResourceID:   strconv.FormatInt(record.ID, 10),
		Detail:       detailJSON,
	})

	return &RegisterResponse{
		NodeID:      record.ID,
		AgentID:     record.AgentID,
		AgentSecret: record.AgentSecret,
		WSURL:       fmt.Sprintf("%s/api/v1/agents/ws", s.externalURL),
	}, nil
}

func (s *Service) registerWithEnrollment(ctx context.Context, req RegisterRequest) (*RegisterResponse, *errcode.AppError) {
	if strings.TrimSpace(req.NodeKey) == "" {
		return nil, errcode.New(errcode.InvalidArgument, "node_key is required")
	}

	hash := hashEnrollmentToken(req.EnrollmentToken)
	row, err := s.nodes.GetByEnrollmentTokenHash(ctx, hash)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "query enrollment node failed", err)
	}
	if row == nil || row.Status != model.NodeStatusPending || strings.TrimSpace(row.EnrollmentTokenHash) == "" {
		return nil, errcode.New(errcode.EnrollmentTokenInvalid, "enrollment token is invalid or expired")
	}

	agentID, err := randomHex(12)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate agent_id failed", err)
	}
	agentSecret, err := randomHex(24)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate agent_secret failed", err)
	}
	now := time.Now().UTC()

	updated := *row
	updated.AgentID = agentID
	updated.AgentSecret = agentSecret
	updated.NodeKey = req.NodeKey
	updated.Hostname = req.Hostname
	updated.Platform = req.Platform
	updated.PlatformVersion = req.PlatformVersion
	updated.Arch = req.Arch
	updated.AgentVersion = req.AgentVersion
	updated.Status = model.NodeStatusOnline
	updated.LastHeartbeatAt = now
	updated.LastOnlineAt = now

	if err := s.nodes.CompleteEnrollment(ctx, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.New(errcode.EnrollmentTokenInvalid, "enrollment token is invalid or already used")
		}
		return nil, errcode.Wrap(errcode.Internal, "complete enrollment failed", err)
	}

	detailJSON := "{}"
	if b, err := json.Marshal(map[string]string{"node_key": updated.NodeKey, "mode": "enrollment"}); err == nil {
		detailJSON = string(b)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "agent",
		ActorID:      updated.AgentID,
		ActorName:    updated.Name,
		Action:       "node.register",
		ResourceType: "node",
		ResourceID:   strconv.FormatInt(updated.ID, 10),
		Detail:       detailJSON,
	})

	return &RegisterResponse{
		NodeID:      updated.ID,
		AgentID:     updated.AgentID,
		AgentSecret: updated.AgentSecret,
		WSURL:       fmt.Sprintf("%s/api/v1/agents/ws", s.externalURL),
	}, nil
}

// CreatePendingNode pre-creates a node in the console and returns a one-time enrollment token for the agent.
func (s *Service) CreatePendingNode(ctx context.Context, req CreatePendingNodeRequest, actorUserID, actorUsername string) (*CreatePendingNodeResponse, *errcode.AppError) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.New(errcode.InvalidArgument, "name is required")
	}

	expiresIn := 7 * 24 * time.Hour
	if req.ExpiresInHours > 0 {
		expiresIn = time.Duration(req.ExpiresInHours) * time.Hour
	}
	expiresAt := time.Now().UTC().Add(expiresIn)

	plainToken, err := randomHex(32)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate enrollment token failed", err)
	}
	tokenHash := hashEnrollmentToken(plainToken)

	pendID, err := randomHex(8)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate pending id failed", err)
	}
	keyRand, err := randomHex(8)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate pending node_key failed", err)
	}
	secretPlaceholder, err := randomHex(24)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate placeholder secret failed", err)
	}

	now := time.Now().UTC()
	record := &model.Node{
		AgentID:         "pend-" + pendID,
		AgentSecret:     secretPlaceholder,
		NodeKey:         "pend-key-" + keyRand,
		Name:            name,
		Hostname:        "",
		Platform:        "",
		PlatformVersion: "",
		Arch:            "",
		AgentVersion:    "0.0.0",
		Status:          model.NodeStatusPending,
		LastHeartbeatAt: now,
		LastOnlineAt:    now,
	}

	if err := s.nodes.CreatePendingEnrollment(ctx, record, tokenHash, &expiresAt); err != nil {
		return nil, errcode.Wrap(errcode.Internal, "create pending node failed", err)
	}

	detailJSON := "{}"
	if b, err := json.Marshal(map[string]string{"name": name}); err == nil {
		detailJSON = string(b)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "user",
		ActorID:      actorUserID,
		ActorName:    actorUsername,
		Action:       "node.create_pending",
		ResourceType: "node",
		ResourceID:   strconv.FormatInt(record.ID, 10),
		Detail:       detailJSON,
	})

	return &CreatePendingNodeResponse{
		ID:                  record.ID,
		Name:                name,
		Status:              model.NodeStatusPending,
		EnrollmentToken:     plainToken,
		EnrollmentExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

// RegenerateEnrollmentToken issues a new enrollment token for a pending node (replaces the stored hash).
func (s *Service) RegenerateEnrollmentToken(ctx context.Context, nodeID int64, actorUserID, actorUsername string) (*CreatePendingNodeResponse, *errcode.AppError) {
	item, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "get node failed", err)
	}
	if item == nil {
		return nil, errcode.New(errcode.NotFound, "node not found")
	}
	if item.Status != model.NodeStatusPending {
		return nil, errcode.New(errcode.InvalidArgument, "only pending nodes can show install commands")
	}

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)
	plainToken, err := randomHex(32)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate enrollment token failed", err)
	}
	tokenHash := hashEnrollmentToken(plainToken)

	if err := s.nodes.SetPendingEnrollmentToken(ctx, nodeID, tokenHash, &expiresAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errcode.New(errcode.InvalidArgument, "only pending nodes can show install commands")
		}
		return nil, errcode.Wrap(errcode.Internal, "update enrollment token failed", err)
	}

	detailJSON := "{}"
	if b, err := json.Marshal(map[string]string{"name": item.Name}); err == nil {
		detailJSON = string(b)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "user",
		ActorID:      actorUserID,
		ActorName:    actorUsername,
		Action:       "node.regenerate_enrollment",
		ResourceType: "node",
		ResourceID:   strconv.FormatInt(nodeID, 10),
		Detail:       detailJSON,
	})

	return &CreatePendingNodeResponse{
		ID:                  nodeID,
		Name:                item.Name,
		Status:              model.NodeStatusPending,
		EnrollmentToken:     plainToken,
		EnrollmentExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func hashEnrollmentToken(plain string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(plain)))
	return hex.EncodeToString(sum[:])
}

// List returns all nodes with latest runtime state.
func (s *Service) List(ctx context.Context) (*ListResponse, *errcode.AppError) {
	items, err := s.nodes.List(ctx)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "list nodes failed", err)
	}

	result := &ListResponse{Items: make([]*ListItem, 0, len(items))}
	for _, item := range items {
		runtimeState, err := s.runtime.GetByNodeID(ctx, item.ID)
		if err != nil {
			return nil, errcode.Wrap(errcode.Internal, "query runtime state failed", err)
		}
		result.Items = append(result.Items, toListItem(item, runtimeState))
	}
	return result, nil
}

// GetDetail returns node detail by ID.
func (s *Service) GetDetail(ctx context.Context, nodeID int64) (*DetailResponse, *errcode.AppError) {
	item, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "get node failed", err)
	}
	if item == nil {
		return nil, errcode.New(errcode.NotFound, "node not found")
	}
	runtimeState, err := s.runtime.GetByNodeID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "get runtime state failed", err)
	}
	return &DetailResponse{
		ID:               item.ID,
		Name:             item.Name,
		Status:           item.Status,
		Hostname:         item.Hostname,
		Platform:         item.Platform,
		PlatformVersion:  item.PlatformVersion,
		Arch:             item.Arch,
		AgentVersion:     item.AgentVersion,
		NodeKey:          item.NodeKey,
		LastHeartbeatAt:  item.LastHeartbeatAt.Format(time.RFC3339),
		LastOnlineAt:     item.LastOnlineAt.Format(time.RFC3339),
		Labels:           []string{},
		RuntimeState:     runtimeState,
		Services:         []ServiceItem{},
		RecentTasks:      []TaskItem{},
		Alerts:           []AlertItem{},
		ShortTermMetrics: []MetricPoint{},
	}, nil
}

// Delete removes a node by ID, related MySQL runtime row (CASCADE), and Redis metric keys.
func (s *Service) Delete(ctx context.Context, nodeID int64, actorUserID, actorUsername string) *errcode.AppError {
	item, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return errcode.Wrap(errcode.Internal, "get node failed", err)
	}
	if item == nil {
		return errcode.New(errcode.NotFound, "node not found")
	}

	if err := s.nodes.DeleteByID(ctx, nodeID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errcode.New(errcode.NotFound, "node not found")
		}
		return errcode.Wrap(errcode.Internal, "delete node failed", err)
	}

	_ = s.runtime.DeleteForNode(ctx, nodeID)

	detailJSON := "{}"
	if b, err := json.Marshal(map[string]string{"name": item.Name, "node_key": item.NodeKey}); err == nil {
		detailJSON = string(b)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "user",
		ActorID:      actorUserID,
		ActorName:    actorUsername,
		Action:       "node.delete",
		ResourceType: "node",
		ResourceID:   strconv.FormatInt(nodeID, 10),
		Detail:       detailJSON,
	})

	return nil
}

// AuthenticateAgent authenticates an agent by long-lived credentials.
func (s *Service) AuthenticateAgent(ctx context.Context, agentID, agentSecret string) (*model.Node, *errcode.AppError) {
	item, err := s.nodes.GetByAgentCredential(ctx, agentID, agentSecret)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "query node credential failed", err)
	}
	if item == nil {
		return nil, errcode.New(errcode.AgentUnauthorized, "invalid agent credential")
	}
	return item, nil
}

func randomHex(byteLen int) (string, error) {
	buf := make([]byte, byteLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func toListItem(item *model.Node, runtimeState *model.NodeRuntimeState) *ListItem {
	rs := runtimeState
	if rs == nil {
		rs = &model.NodeRuntimeState{NodeID: item.ID}
	}
	return &ListItem{
		ID:              item.ID,
		Name:            item.Name,
		Status:          item.Status,
		Hostname:        item.Hostname,
		Platform:        item.Platform,
		Arch:            item.Arch,
		AgentVersion:    item.AgentVersion,
		LastHeartbeatAt: item.LastHeartbeatAt.Format(time.RFC3339),
		Labels:          []string{},
		RuntimeState:    rs,
	}
}
