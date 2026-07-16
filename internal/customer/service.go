package customer

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/bizerror"
	"github.com/dextea-v3/dextea-customer/api/internal/config"
	"github.com/dextea-v3/dextea-customer/api/internal/jwt"
	"github.com/redis/go-redis/v9"
)

const defaultNewCustomerNickname = "德贤茶友"

type Service struct {
	repo    *Repository
	rdb     *redis.Client
	cfg     *config.Config
	alipay  *alipayClient
}

func NewService(repo *Repository, rdb *redis.Client, cfg *config.Config) *Service {
	svc := &Service{repo: repo, rdb: rdb, cfg: cfg}
	if cfg.AlipayAppID != "" && cfg.AlipayPrivateKey != "" {
		if client, err := newAlipayClient(cfg.AlipayAppID, cfg.AlipayPrivateKey, cfg.AlipayPublicKey, cfg.AlipayGateway); err == nil {
			svc.alipay = client
		}
	}
	return svc
}

func (s *Service) Login(ctx context.Context, code string, platform Platform) (*LoginResponse, error) {
	if s.repo.db == nil {
		return nil, ErrDBDisabled
	}

	var openID string
	var err error
	switch platform {
	case PlatformAlipay:
		openID, err = s.exchangeAlipayOpenID(ctx, code)
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
		})
		if err != nil {
			return nil, err
		}
	}

	token, err := jwt.Generate(s.cfg.JWTSecret, jwt.Claims{
		UID:      cust.ID,
		Platform: string(platform),
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
	openID, err := s.alipay.exchangeCode(ctx, code)
	if err != nil {
		return "", bizerror.New(CodeAlipayAuthFailed, err.Error())
	}
	return openID, nil
}

var ErrDBDisabled = bizerror.New(bizerror.CodeDBDisabled)
