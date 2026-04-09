package node

import (
	"context"
	"crypto/rand"
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

// Service implements node provisioning and query business logic.
type Service struct {
	cfg         config.NodeConfig
	nodes       repository.NodeRepository
	runtime     repository.RuntimeStateRepository
	audit       *audit.Service
	externalURL string
}

// NewService creates a node service.
func NewService(cfg config.NodeConfig, nodes repository.NodeRepository, runtime repository.RuntimeStateRepository, audit *audit.Service, externalURL string) *Service {
	return &Service{
		cfg:         cfg,
		nodes:       nodes,
		runtime:     runtime,
		audit:       audit,
		externalURL: strings.TrimRight(externalURL, "/"),
	}
}

// CreatePendingNode 在控制台创建节点并生成固定的 agent_id / agent_secret / node_key，Agent 凭此直接与控制面通信，无需再调用注册接口兑换凭证。
func (s *Service) CreatePendingNode(ctx context.Context, req CreatePendingNodeRequest, actorUserID, actorUsername string) (*CreatePendingNodeResponse, *errcode.AppError) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errcode.New(errcode.InvalidArgument, "name is required")
	}

	agentID, err := randomHex(12)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate agent_id failed", err)
	}
	agentSecret, err := randomHex(24)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate agent_secret failed", err)
	}
	nodeKey, err := randomHex(16)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "generate node_key failed", err)
	}

	now := time.Now().UTC()
	record := &model.Node{
		AgentID:         agentID,
		AgentSecret:     agentSecret,
		NodeKey:         nodeKey,
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

	if err := s.nodes.Create(ctx, record); err != nil {
		return nil, errcode.Wrap(errcode.Internal, "create node failed", err)
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
		ID:          record.ID,
		Name:        name,
		Status:      model.NodeStatusPending,
		AgentID:     agentID,
		AgentSecret: agentSecret,
		NodeKey:     nodeKey,
		WSURL:       fmt.Sprintf("%s/api/v1/agents/ws", s.externalURL),
	}, nil
}

// GetNodeAgentCredentials 返回节点固定凭据（供控制台「安装」展示；需已登录且有权限）。
func (s *Service) GetNodeAgentCredentials(ctx context.Context, nodeID int64) (*AgentCredentialsResponse, *errcode.AppError) {
	item, err := s.nodes.GetByID(ctx, nodeID)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "get node failed", err)
	}
	if item == nil {
		return nil, errcode.New(errcode.NotFound, "node not found")
	}
	return &AgentCredentialsResponse{
		AgentID:     item.AgentID,
		AgentSecret: item.AgentSecret,
		NodeKey:     item.NodeKey,
		WSURL:       fmt.Sprintf("%s/api/v1/agents/ws", s.externalURL),
	}, nil
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
