package customer

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/alipay"
	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/jwt"
	"github.com/redis/go-redis/v9"
)

const defaultNewCustomerNickname = "德贤茶友"

type Service struct {
	repo   *Repository
	rdb    *redis.Client
	cfg    *config.Config
	alipay *alipay.Client
}

func NewService(repo *Repository, rdb *redis.Client, cfg *config.Config, alipayClient *alipay.Client) *Service {
	return &Service{
		repo:   repo,
		rdb:    rdb,
		cfg:    cfg,
		alipay: alipayClient,
	}
}

func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	if s.repo.db == nil {
		return nil, ErrDBDisabled
	}

	var openID string
	var err error
	switch req.Platform {
	case PlatformAlipay:
		openID, err = s.exchangeAlipayOpenID(ctx, req.Code)
	case PlatformWeixin:
		return nil, bizerror.New(CodePlatformNotSupported)
	default:
		return nil, bizerror.New(CodePlatformInvalid)
	}
	if err != nil {
		return nil, err
	}

	cust, err := s.repo.FindByAlipayOpenID(ctx, openID)
	if err != nil {
		return nil, err
	}
	if cust == nil {
		cust, err = s.repo.Create(ctx, &Customer{
			Name:         defaultNewCustomerNickname,
			AlipayOpenID: openID,
			Status:       1,
			Platform:     req.Platform.DBValue(),
		})
		if err != nil {
			return nil, err
		}
	}

	token, err := jwt.Generate(s.cfg.JWTSecret, jwt.Claims{
		UID:      cust.ID,
		Platform: string(req.Platform),
	}, s.cfg.JWTExpireSeconds())
	if err != nil {
		return nil, bizerror.New(bizerror.CodeInternal, "生成登录令牌失败")
	}

	return &LoginResponse{
		Token:    token,
		Customer: cust,
	}, nil
}

func (s *Service) exchangeAlipayOpenID(ctx context.Context, code string) (string, error) {
	if s.alipay == nil {
		return "", bizerror.New(CodeAlipayNotConfigured)
	}
	openID, err := s.alipay.ExchangeCode(ctx, code)
	if err != nil {
		return "", bizerror.New(CodeAlipayAuthFailed, err.Error())
	}
	return openID, nil
}

var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)
