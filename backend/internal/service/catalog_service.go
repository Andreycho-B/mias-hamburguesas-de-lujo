package service

import (
	"context"

	"mias-backend/internal/database/db"
)

type CatalogService struct {
	queries db.Querier
}

func NewCatalogService(queries db.Querier) *CatalogService {
	return &CatalogService{queries: queries}
}

type CatalogResponse struct {
	Categories []db.Category `json:"categories"`
	Products   []ProductDTO  `json:"products"`
}

type ProductDTO struct {
	ID           int32  `json:"id"`
	CategoryID   int32  `json:"category_id"`
	CategoryName string `json:"category_name"`
	CategorySlug string `json:"category_slug"`
	Slug         string `json:"slug"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	Price        int32  `json:"price"`
	BreadType    string `json:"bread_type,omitempty"`
	ImageURL     string `json:"image_url,omitempty"`
	IsBestseller bool   `json:"is_bestseller"`
	IsChefChoice bool   `json:"is_chef_choice"`
	IsSeason     bool   `json:"is_season"`
	IsInStock    bool   `json:"is_in_stock"`
}

func (s *CatalogService) GetPublicMenu(ctx context.Context) (*CatalogResponse, error) {
	categories, err := s.queries.ListActiveCategories(ctx)
	if err != nil {
		return nil, err
	}

	productsRaw, err := s.queries.ListActiveProducts(ctx)
	if err != nil {
		return nil, err
	}

	var products []ProductDTO
	for _, p := range productsRaw {
		products = append(products, ProductDTO{
			ID:           p.ID,
			CategoryID:   p.CategoryID,
			CategoryName: p.CategoryName,
			CategorySlug: p.CategorySlug,
			Slug:         p.Slug,
			Name:         p.Name,
			Description:  p.Description,
			Price:        p.Price,
			BreadType:    p.BreadType.String,
			ImageURL:     p.ImageUrl.String,
			IsBestseller: p.IsBestseller,
			IsChefChoice: p.IsChefChoice,
			IsSeason:     p.IsSeason,
			IsInStock:    p.IsInStock,
		})
	}

	return &CatalogResponse{
		Categories: categories,
		Products:   products,
	}, nil
}

func (s *CatalogService) ToggleStock(ctx context.Context, productID int32, inStock bool) error {
	_, err := s.queries.ToggleProductStock(ctx, db.ToggleProductStockParams{
		ID:        productID,
		IsInStock: inStock,
	})
	return err
}
