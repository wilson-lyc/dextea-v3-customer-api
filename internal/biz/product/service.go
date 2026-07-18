package product

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

func (s *Service) GetDetail(ctx context.Context, req GetProductDetailRequest) (*ProductDetailResponse, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.CodeDBDisabled)
	}

	p, err := s.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Status != 1 {
		return nil, bizerror.New(bizerror.CodeNotFound, "商品不存在")
	}

	storeStatus := 0
	if ss, err := s.repo.FindProductStoreStatus(ctx, req.ProductID, req.StoreID); err != nil {
		return nil, err
	} else if ss != nil {
		storeStatus = *ss
	}

	customizations, err := s.repo.FindCustomizations(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	customizationIDs := make([]int64, 0, len(customizations))
	for _, c := range customizations {
		customizationIDs = append(customizationIDs, c.ID)
	}

	options, err := s.repo.FindCustomizationOptions(ctx, customizationIDs)
	if err != nil {
		return nil, err
	}

	optionIDs := make([]int64, 0, len(options))
	for _, o := range options {
		optionIDs = append(optionIDs, o.ID)
	}
	optionStoreStatusMap := make(map[int64]int, len(optionIDs))
	if statuses, err := s.repo.FindCustomizationOptionStoreStatuses(ctx, optionIDs, req.StoreID); err != nil {
		return nil, err
	} else {
		for _, st := range statuses {
			optionStoreStatusMap[st.CustomizationOptionID] = st.Status
		}
	}

	optionsByCustomization := make(map[int64][]CustomizationOption, len(customizations))
	for _, o := range options {
		optionsByCustomization[o.CustomizationID] = append(optionsByCustomization[o.CustomizationID], o)
	}

	customizationItems := make([]CustomizationItem, 0, len(customizations))
	for _, c := range customizations {
		opts := optionsByCustomization[c.ID]
		optionItems := make([]CustomizationOptionItem, 0, len(opts))
		for _, o := range opts {
			status := 0
			if v, ok := optionStoreStatusMap[o.ID]; ok {
				status = v
			}
			optionItems = append(optionItems, CustomizationOptionItem{
				ID:      o.ID,
				Name:    o.Name,
				Price:   o.Price,
				Sort:    o.Sort,
				Status:  status,
			})
		}
		customizationItems = append(customizationItems, CustomizationItem{
			ID:       c.ID,
			Name:     c.Name,
			Sort:     c.Sort,
			Status:   c.Status,
			Options:  optionItems,
		})
	}

	imgRows, err := s.repo.FindProductImages(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	images := make([]ProductImageItem, 0, len(imgRows))
	for _, im := range imgRows {
		images = append(images, ProductImageItem{
			ID:  im.ImageID,
			URL: im.URL,
			Sort:    im.Sort,
			Type:    im.Type,
		})
	}

	return &ProductDetailResponse{
		ID:             p.ID,
		Name:           p.Name,
		Brief:          p.Brief,
		Description:    p.Description,
		Status:         storeStatus,
		Price:          p.Price,
		Images:         images,
		Customizations: customizationItems,
	}, nil
}

// GetStoreStatus 获取指定门店下「全局上架（status=1）」商品的状态列表。
// productId 选填：
//   - 不填：返回全部全局上架商品在该门店的状态列表。
//   - 填写：仅返回指定商品的状态，且该商品必须为全局上架，否则返回业务错误。
//
// 对在 product_store_status 中未配置记录的全局上架商品，默认状态为 0。
func (s *Service) GetStoreStatus(ctx context.Context, req GetProductStoreStatusRequest) ([]ProductStoreStatusItem, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.CodeDBDisabled)
	}

	// 指定了商品 ID：仅返回该商品的状态，且必须为全局上架。
	if req.ProductID > 0 {
		return s.getSingleStoreStatus(ctx, req.ProductID, req.StoreID)
	}

	// 1. 取全局上架商品
	products, err := s.repo.FindActiveProducts(ctx)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return []ProductStoreStatusItem{}, nil
	}

	ids := make([]int64, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.ID)
	}

	// 2. 批量取这些商品在该门店下的专属状态
	statusMap, err := s.repo.FindStoreStatuses(ctx, ids, req.StoreID)
	if err != nil {
		return nil, err
	}

	// 3. 组装结果：全局上架商品全量返回，未配置门店状态则默认 0
	items := make([]ProductStoreStatusItem, 0, len(products))
	for _, p := range products {
		status := 0
		if v, ok := statusMap[p.ID]; ok {
			status = v
		}
		items = append(items, ProductStoreStatusItem{
			ProductID: p.ID,
			Name:      p.Name,
			Status:    status,
		})
	}
	return items, nil
}

// getSingleStoreStatus 返回指定商品在该门店下的状态。
// 商品必须存在且为全局上架（status=1），否则返回业务错误。
func (s *Service) getSingleStoreStatus(ctx context.Context, productID, storeID int64) ([]ProductStoreStatusItem, error) {
	p, err := s.repo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Status != 1 {
		return nil, bizerror.New(bizerror.CodeNotFound, "商品不存在或未上架")
	}

	status := 0
	if ss, err := s.repo.FindProductStoreStatus(ctx, productID, storeID); err != nil {
		return nil, err
	} else if ss != nil {
		status = *ss
	}

	return []ProductStoreStatusItem{
		{
			ProductID: p.ID,
			Name:      p.Name,
			Status:    status,
		},
	}, nil
}
