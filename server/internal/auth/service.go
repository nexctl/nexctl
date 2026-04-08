package auth

import (
	"context"
	"encoding/json"
	"time"

	"github.com/nexctl/nexctl/server/internal/audit"
	"github.com/nexctl/nexctl/server/internal/config"
	"github.com/nexctl/nexctl/server/internal/repository"
	"github.com/nexctl/nexctl/server/pkg/errcode"
	"github.com/nexctl/nexctl/server/pkg/jwtutil"
	"golang.org/x/crypto/bcrypt"
)

// Service implements operator authentication.
type Service struct {
	cfg   config.AuthConfig
	users repository.UserRepository
	audit *audit.Service
}

// NewService creates an auth service.
func NewService(cfg config.AuthConfig, users repository.UserRepository, audit *audit.Service) *Service {
	return &Service{cfg: cfg, users: users, audit: audit}
}

// Login authenticates a user and returns a JWT.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, *errcode.AppError) {
	user, roleCode, err := s.users.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "query user failed", err)
	}
	if user == nil || user.Status != "active" {
		return nil, errcode.New(errcode.Unauthorized, "invalid username or password")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errcode.New(errcode.Unauthorized, "invalid username or password")
	}

	expire := time.Duration(s.cfg.JWTExpireHours) * time.Hour
	token, err := jwtutil.Sign(s.cfg.JWTSecret, user.ID, user.Username, roleCode, expire)
	if err != nil {
		return nil, errcode.Wrap(errcode.Internal, "sign token failed", err)
	}

	loginDetail := "{}"
	if b, err := json.Marshal(map[string]string{"result": "success"}); err == nil {
		loginDetail = string(b)
	}
	_ = s.audit.Record(ctx, audit.RecordInput{
		ActorType:    "user",
		ActorID:      user.Username,
		ActorName:    user.DisplayName,
		Action:       "auth.login",
		ResourceType: "session",
		ResourceID:   user.Username,
		Detail:       loginDetail,
	})

	return &LoginResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(expire.Seconds()),
	}, nil
}
