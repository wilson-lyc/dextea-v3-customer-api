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

// GetStoreMenu 获取门店菜单。
// 链路：① 根据门店 ID 取其绑定的第一个菜单作为当前菜单；
// ② 查询该菜单下的分组；③ 查询各分组绑定的商品。
// 仅返回商品全局状态为 1（在售）的商品；返回的商品状态取自门店商品状态表
// （product_store_status.status）。该表为懒加载：门店未配置该商品状态时表中无记录，
// 此时默认状态为 0，代表门店售罄。
func (s *Service) GetStoreMenu(ctx context.Context, storeID int64) (*StoreMenuResponse, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.CodeDBDisabled)
	}

	menuID, err := s.repo.FindStoreMenuID(ctx, storeID)
	if err != nil {
		return nil, err
	}
	if menuID == 0 {
		return nil, bizerror.New(bizerror.CodeNotFound, "门店未配置菜单")
	}

	m, err := s.repo.FindMenu(ctx, menuID)
	if err != nil {
		return nil, err
	}

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
	// 每个商品只取 type=1 中 sort 最小的第一张作为主图 URL。
	imgMap := make(map[int64]string, len(imgs))
	seen := make(map[int64]bool, len(imgs))
	for _, im := range imgs {
		if seen[im.ProductID] {
			continue
		}
		seen[im.ProductID] = true
		imgMap[im.ProductID] = im.URL
	}

	// 按 group_id 聚合商品，并按查询返回的顺序（group_id, product_sort）保持分组内排序
	productsByGroup := make(map[int64][]ProductInfo, len(groups))
	for _, row := range rows {
		// 门店商品状态表为懒加载：无记录即视为门店售罄，状态默认补 0。
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
			Image:  imgMap[row.ProductID],
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
