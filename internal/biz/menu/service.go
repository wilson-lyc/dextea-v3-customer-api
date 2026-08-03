package menu

import (
	"context"

	"github.com/dextea-v3/dextea-customer/api/internal/common/bizerror"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetStoreMenu(ctx context.Context, storeID int64) (*StoreMenuResponse, error) {
	// 检查数据库是否启用
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.ErrMysqlDisabled)
	}

	// 查询门店绑定的菜单ID
	menuID, err := s.repo.FindStoreMenuID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if menuID == 0 {
		return nil, bizerror.New(&bizerror.BizError{Code: 40400, Message: "资源不存在"}, "门店未配置菜单")
	}

	// 查询菜单基础信息
	m, err := s.repo.FindMenu(ctx, menuID)
	if err != nil {
		return nil, err
	}

	// 查询菜单分组信息
	groups, err := s.repo.FindGroupsByMenu(ctx, menuID)
	if err != nil {
		return nil, err
	}

	resp := &StoreMenuResponse{
		Menu: MenuInfo{
			ID:          m.ID,
			Name:        m.Name,
			Description: m.Description,
		},
		Groups: make([]GroupInfo, 0, len(groups)),
	}
	if len(groups) == 0 {
		return resp, nil
	}

	groupIDs := make([]int64, 0, len(groups))
	for _, g := range groups {
		groupIDs = append(groupIDs, g.ID)
	}

	rows, err := s.repo.FindGroupProducts(ctx, groupIDs, storeID)
	if err != nil {
		return nil, err
	}

	productIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		productIDs = append(productIDs, row.ProductID)
	}

	imgs, err := s.repo.FindProductImages(ctx, productIDs)
	if err != nil {
		return nil, err
	}
	imgMap := make(map[int64]string, len(imgs))
	for _, im := range imgs {
		imgMap[im.ProductID] = im.URL
	}

	productsByGroup := make(map[int64][]ProductInfo, len(groups))
	for _, row := range rows {
		status := 0
		if row.StoreStatus != nil {
			status = *row.StoreStatus
		}
		productsByGroup[row.GroupID] = append(productsByGroup[row.GroupID], ProductInfo{
			ID:     row.ProductID,
			Name:   row.Name,
			Brief:  row.Brief,
			Price:  row.Price,
			Sort:   row.ProductSort,
			Status: status,
			Cover:  imgMap[row.ProductID],
		})
	}

	for _, g := range groups {
		products := productsByGroup[g.ID]
		if products == nil {
			products = []ProductInfo{}
		}
		resp.Groups = append(resp.Groups, GroupInfo{
			ID:       g.ID,
			Name:     g.Name,
			Sort:     g.Sort,
			Products: products,
		})
	}
	return resp, nil
}
