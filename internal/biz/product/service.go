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
