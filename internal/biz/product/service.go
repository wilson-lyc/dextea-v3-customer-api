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
		return nil, bizerror.New(bizerror.ErrMysqlDisabled)
	}

	p, err := s.repo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if p == nil || p.Status != 1 {
		return nil, bizerror.New(&bizerror.BizError{Code: 40400, Message: "资源不存在", Kind: bizerror.KindBusiness}, "商品不存在")
	}

	storeStatus := 0
	if ss, err := s.repo.FindProductStoreStatus(ctx, req.ProductID, req.StoreID); err != nil {
		return nil, err
	} else if ss != nil {
		storeStatus = *ss
	}

	items, err := s.repo.FindCustomizationItems(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}

	itemIDs := make([]int64, 0, len(items))
	for _, c := range items {
		itemIDs = append(itemIDs, c.ID)
	}

	options, err := s.repo.FindCustomizationOptions(ctx, itemIDs)
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
			optionStoreStatusMap[st.OptionID] = st.Status
		}
	}

	optionsByItem := make(map[int64][]CustomizationOption, len(items))
	for _, o := range options {
		optionsByItem[o.ItemID] = append(optionsByItem[o.ItemID], o)
	}

	customizationItems := make([]CustomizationItemResponse, 0, len(items))
	for _, c := range items {
		opts := optionsByItem[c.ID]
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
		customizationItems = append(customizationItems, CustomizationItemResponse{
			ID:      c.ID,
			Name:    c.Name,
			Sort:    c.Sort,
			Status:  c.Status,
			Options: optionItems,
		})
	}

	imgRows, err := s.repo.FindProductImages(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	var cover *ProductImageItem
	gallery := make([]ProductImageItem, 0, len(imgRows))
	for _, im := range imgRows {
		item := ProductImageItem{
			ID:   im.ImageID,
			URL:  im.URL,
			Sort: im.Sort,
			Type: im.Type,
		}
		switch im.Type {
		case 1:
			if cover == nil {
				c := item
				cover = &c
			}
		case 2:
			gallery = append(gallery, item)
		}
	}

	return &ProductDetailResponse{
		ID:             p.ID,
		Name:           p.Name,
		Brief:          p.Brief,
		Description:    p.Description,
		Status:         storeStatus,
		Price:          p.Price,
		Cover:          cover,
		Gallery:        gallery,
		Customizations: customizationItems,
	}, nil
}

func (s *Service) GetStoreStatus(ctx context.Context, req GetProductStoreStatusRequest) ([]ProductStoreStatusItem, error) {
	if s.repo.db == nil {
		return nil, bizerror.New(bizerror.ErrMysqlDisabled)
	}

	products, err := s.repo.FindByIDs(ctx, req.ProductIDs)
	if err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return []ProductStoreStatusItem{}, nil
	}

	statusMap, err := s.repo.FindStoreStatuses(ctx, req.ProductIDs, req.StoreID)
	if err != nil {
		return nil, err
	}

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
