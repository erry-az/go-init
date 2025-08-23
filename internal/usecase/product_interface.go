package usecase

import (
	"context"

	"github.com/erry-az/go-init/internal/domain"
)

// ProductUsecase defines the business logic interface for product operations
type ProductUsecase interface {
	CreateProduct(ctx context.Context, name, price string) (*domain.Product, error)
	GetProduct(ctx context.Context, productID string) (*domain.Product, error)
	UpdateProduct(ctx context.Context, productID, name, price string) (*domain.Product, error)
	DeleteProduct(ctx context.Context, productID string) error
	ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error)
	BulkUpdatePrices(ctx context.Context, updates []BulkPriceUpdate) (*BulkUpdatePricesResponse, error)
	GetProductAnalytics(ctx context.Context) (*ProductAnalyticsResponse, error)
}

// Request/Response types for Product operations
type ListProductsRequest struct {
	PageSize    int32
	PageToken   string
	SearchQuery string
	PriceRange  *PriceRange
}

type PriceRange struct {
	MinPrice string
	MaxPrice string
}

type ListProductsResponse struct {
	Products      []*domain.Product
	NextPageToken string
	TotalCount    int32
}

type BulkPriceUpdate struct {
	ID    string
	Price string
}

type BulkUpdatePricesResponse struct {
	UpdatedProducts []*domain.Product
	FailedIDs       []string
}

type ProductAnalyticsResponse struct {
	TotalProducts int32
	AveragePrice  string
	HighestPrice  string
	LowestPrice   string
	CategoryStats []*CategoryStats
}

type CategoryStats struct {
	Category string
	Count    int32
}
