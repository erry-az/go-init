package usecase

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/erry-az/go-init/internal/domain"
	"github.com/erry-az/go-init/internal/repository/sqlc"
	"github.com/erry-az/go-init/proto/api/v1"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type productUsecase struct {
	db        sqlc.Querier
	publisher *cqrs.EventBus
}

// NewProductUsecase creates a new product usecase instance
func NewProductUsecase(db sqlc.Querier, publisher *cqrs.EventBus) ProductUsecase {
	return &productUsecase{
		db:        db,
		publisher: publisher,
	}
}

func (p *productUsecase) CreateProduct(ctx context.Context, name, price string) (*domain.Product, error) {
	// Create domain entity
	product, err := domain.NewProductFromString(name, price)
	if err != nil {
		return nil, err
	}

	// Convert decimal to pgtype.Numeric for database
	var dbPrice pgtype.Numeric
	if err := dbPrice.Scan(product.Price.String()); err != nil {
		return nil, domain.NewValidationError(fmt.Sprintf("invalid price conversion: %v", err))
	}

	params := sqlc.CreateProductParams{
		ID:    product.ID,
		Name:  product.Name,
		Price: dbPrice,
	}

	dbProduct, err := p.db.CreateProduct(ctx, params)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to create product: %v", err))
	}

	createdProduct := p.mapDBProductToDomain(dbProduct)

	// Publish product created event
	if err := p.publishProductCreatedEvent(ctx, createdProduct); err != nil {
		fmt.Printf("Failed to publish product created event: %v\n", err)
	}

	return createdProduct, nil
}

func (p *productUsecase) GetProduct(ctx context.Context, productID string) (*domain.Product, error) {
	id, err := uuid.Parse(productID)
	if err != nil {
		return nil, domain.NewValidationError(fmt.Sprintf("invalid product ID: %v", err))
	}

	dbProduct, err := p.db.GetProductByID(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.NewNotFoundError("product not found")
		}
		return nil, domain.NewInternalError(fmt.Sprintf("failed to get product: %v", err))
	}

	return p.mapDBProductToDomain(dbProduct), nil
}

func (p *productUsecase) UpdateProduct(ctx context.Context, productID, name, price string) (*domain.Product, error) {
	// Get existing product for price change detection
	existingProduct, err := p.GetProduct(ctx, productID)
	if err != nil {
		return nil, err
	}

	// Store old price for event
	oldPrice := existingProduct.Price.String()

	// Update domain entity
	if err := existingProduct.UpdateDetailsFromString(name, price); err != nil {
		return nil, err
	}

	// Convert decimal to pgtype.Numeric for database
	var dbPrice pgtype.Numeric
	if err := dbPrice.Scan(existingProduct.Price.String()); err != nil {
		return nil, domain.NewValidationError(fmt.Sprintf("invalid price conversion: %v", err))
	}

	params := sqlc.UpdateProductParams{
		ID:    existingProduct.ID,
		Name:  existingProduct.Name,
		Price: dbPrice,
	}

	dbProduct, err := p.db.UpdateProduct(ctx, params)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to update product: %v", err))
	}

	updatedProduct := p.mapDBProductToDomain(dbProduct)

	// Publish product updated event
	if err := p.publishProductUpdatedEvent(ctx, updatedProduct); err != nil {
		fmt.Printf("Failed to publish product updated event: %v\n", err)
	}

	// If price changed, also publish price change event
	newPrice := updatedProduct.Price.String()
	if oldPrice != newPrice {
		if err := p.publishProductPriceChangedEvent(ctx, updatedProduct, oldPrice, newPrice); err != nil {
			fmt.Printf("Failed to publish product price changed event: %v\n", err)
		}
	}

	return updatedProduct, nil
}

func (p *productUsecase) DeleteProduct(ctx context.Context, productID string) error {
	// Get product before deletion for event
	product, err := p.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	if err := p.db.DeleteProduct(ctx, product.ID); err != nil {
		return domain.NewInternalError(fmt.Sprintf("failed to delete product: %v", err))
	}

	// Publish product deleted event
	if err := p.publishProductDeletedEvent(ctx, product); err != nil {
		fmt.Printf("Failed to publish product deleted event: %v\n", err)
	}

	return nil
}

func (p *productUsecase) ListProducts(ctx context.Context, req *ListProductsRequest) (*ListProductsResponse, error) {
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := int32(0)
	if req.PageToken != "" {
		decodedOffset, err := base64.StdEncoding.DecodeString(req.PageToken)
		if err != nil {
			return nil, domain.NewValidationError("invalid page token")
		}
		if _, err := fmt.Sscanf(string(decodedOffset), "%d", &offset); err != nil {
			return nil, domain.NewValidationError("invalid page token format")
		}
	}

	var dbProducts []sqlc.Product
	var err error

	// Handle different query types based on request parameters
	if req.SearchQuery != "" && req.PriceRange != nil {
		var minPrice, maxPrice pgtype.Numeric
		if err := minPrice.Scan(req.PriceRange.MinPrice); err != nil {
			return nil, domain.NewValidationError(fmt.Sprintf("invalid min price: %v", err))
		}
		if err := maxPrice.Scan(req.PriceRange.MaxPrice); err != nil {
			return nil, domain.NewValidationError(fmt.Sprintf("invalid max price: %v", err))
		}

		params := sqlc.SearchProductsWithPriceRangeParams{
			Limit:       pageSize + 1,
			Offset:      offset,
			SearchQuery: "%" + req.SearchQuery + "%",
			MinPrice:    minPrice,
			MaxPrice:    maxPrice,
		}
		dbProducts, err = p.db.SearchProductsWithPriceRange(ctx, params)
	} else if req.SearchQuery != "" {
		params := sqlc.SearchProductsParams{
			Limit:       pageSize + 1,
			Offset:      offset,
			SearchQuery: "%" + req.SearchQuery + "%",
		}
		dbProducts, err = p.db.SearchProducts(ctx, params)
	} else if req.PriceRange != nil {
		var minPrice, maxPrice pgtype.Numeric
		if err := minPrice.Scan(req.PriceRange.MinPrice); err != nil {
			return nil, domain.NewValidationError(fmt.Sprintf("invalid min price: %v", err))
		}
		if err := maxPrice.Scan(req.PriceRange.MaxPrice); err != nil {
			return nil, domain.NewValidationError(fmt.Sprintf("invalid max price: %v", err))
		}

		params := sqlc.ListProductsByPriceRangeParams{
			Limit:    pageSize + 1,
			Offset:   offset,
			MinPrice: minPrice,
			MaxPrice: maxPrice,
		}
		dbProducts, err = p.db.ListProductsByPriceRange(ctx, params)
	} else {
		params := sqlc.ListProductsParams{
			Limit:  pageSize + 1,
			Offset: offset,
		}
		dbProducts, err = p.db.ListProducts(ctx, params)
	}

	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to list products: %v", err))
	}

	// Check if there are more pages
	hasNextPage := len(dbProducts) > int(pageSize)
	if hasNextPage {
		dbProducts = dbProducts[:pageSize]
	}

	products := make([]*domain.Product, len(dbProducts))
	for i, dbProduct := range dbProducts {
		products[i] = p.mapDBProductToDomain(dbProduct)
	}

	var nextPageToken string
	if hasNextPage {
		nextOffset := offset + pageSize
		nextPageToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", nextOffset)))
	}

	// Get total count
	totalCount, err := p.db.CountProducts(ctx)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to count products: %v", err))
	}

	return &ListProductsResponse{
		Products:      products,
		NextPageToken: nextPageToken,
		TotalCount:    int32(totalCount),
	}, nil
}

func (p *productUsecase) BulkUpdatePrices(ctx context.Context, updates []BulkPriceUpdate) (*BulkUpdatePricesResponse, error) {
	var updatedProducts []*domain.Product
	var failedIDs []string

	for _, update := range updates {
		// Get the current product to preserve name
		product, err := p.GetProduct(ctx, update.ID)
		if err != nil {
			failedIDs = append(failedIDs, update.ID)
			continue
		}

		updatedProduct, err := p.UpdateProduct(ctx, update.ID, product.Name, update.Price)
		if err != nil {
			failedIDs = append(failedIDs, update.ID)
			continue
		}

		updatedProducts = append(updatedProducts, updatedProduct)
	}

	return &BulkUpdatePricesResponse{
		UpdatedProducts: updatedProducts,
		FailedIDs:       failedIDs,
	}, nil
}

func (p *productUsecase) GetProductAnalytics(ctx context.Context) (*ProductAnalyticsResponse, error) {
	// Get total count
	totalCount, err := p.db.CountProducts(ctx)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to get product count: %v", err))
	}

	// Get price analytics
	avgPriceInterface, err := p.db.GetAveragePrice(ctx)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to get average price: %v", err))
	}

	minPriceInterface, err := p.db.GetMinPrice(ctx)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to get min price: %v", err))
	}

	maxPriceInterface, err := p.db.GetMaxPrice(ctx)
	if err != nil {
		return nil, domain.NewInternalError(fmt.Sprintf("failed to get max price: %v", err))
	}

	// Convert interface{} to string (they should be pgtype.Numeric)
	avgPriceStr := "0"
	minPriceStr := "0"
	maxPriceStr := "0"

	if avgPrice, ok := avgPriceInterface.(pgtype.Numeric); ok && avgPrice.Valid {
		avgPriceStr = p.numericToString(avgPrice)
	}
	if minPrice, ok := minPriceInterface.(pgtype.Numeric); ok && minPrice.Valid {
		minPriceStr = p.numericToString(minPrice)
	}
	if maxPrice, ok := maxPriceInterface.(pgtype.Numeric); ok && maxPrice.Valid {
		maxPriceStr = p.numericToString(maxPrice)
	}

	return &ProductAnalyticsResponse{
		TotalProducts: int32(totalCount),
		AveragePrice:  avgPriceStr,
		HighestPrice:  maxPriceStr,
		LowestPrice:   minPriceStr,
		CategoryStats: []*CategoryStats{}, // Placeholder
	}, nil
}

// Helper methods
func (p *productUsecase) mapDBProductToDomain(dbProduct sqlc.Product) *domain.Product {
	priceStr := p.numericToString(dbProduct.Price)
	price, _ := decimal.NewFromString(priceStr) // Safe since we control the conversion

	return &domain.Product{
		ID:        dbProduct.ID,
		Name:      dbProduct.Name,
		Price:     price,
		CreatedAt: dbProduct.CreatedAt.Time,
		UpdatedAt: dbProduct.UpdatedAt.Time,
	}
}

func (p *productUsecase) numericToString(n pgtype.Numeric) string {
	if !n.Valid || n.NaN {
		return "0"
	}

	val, err := n.Value()
	if err != nil {
		return "0"
	}

	if str, ok := val.(string); ok {
		return str
	}

	return "0"
}

func (p *productUsecase) publishProductCreatedEvent(ctx context.Context, product *domain.Product) error {
	event := &eventv1.ProductCreatedEvent{
		EventId:       uuid.New().String(),
		Product:       p.domainProductToProto(product),
		EventTime:     timestamppb.Now(),
		CorrelationId: p.getCorrelationID(ctx),
		Data: &eventv1.ProductCreatedEventData{
			Source: "product-service",
			Metadata: map[string]string{
				"operation": "create_product",
				"version":   "v1",
			},
		},
	}
	return p.publisher.Publish(ctx, event)
}

func (p *productUsecase) publishProductUpdatedEvent(ctx context.Context, product *domain.Product) error {
	event := &eventv1.ProductUpdatedEvent{
		EventId:       uuid.New().String(),
		Product:       p.domainProductToProto(product),
		EventTime:     timestamppb.Now(),
		CorrelationId: p.getCorrelationID(ctx),
		Data: &eventv1.ProductUpdatedEventData{
			Source:        "product-service",
			ChangedFields: []string{"name", "price"},
			Metadata: map[string]string{
				"operation": "update_product",
				"version":   "v1",
			},
		},
	}
	return p.publisher.Publish(ctx, event)
}

func (p *productUsecase) publishProductPriceChangedEvent(ctx context.Context, product *domain.Product, oldPrice, newPrice string) error {
	event := &eventv1.ProductPriceChangedEvent{
		EventId:       uuid.New().String(),
		Product:       p.domainProductToProto(product),
		EventTime:     timestamppb.Now(),
		CorrelationId: p.getCorrelationID(ctx),
		Data: &eventv1.ProductPriceChangedEventData{
			Source:        "product-service",
			PreviousPrice: oldPrice,
			NewPrice:      newPrice,
			Metadata: map[string]string{
				"operation": "price_change",
				"version":   "v1",
			},
		},
	}
	return p.publisher.Publish(ctx, event)
}

func (p *productUsecase) publishProductDeletedEvent(ctx context.Context, product *domain.Product) error {
	event := &eventv1.ProductDeletedEvent{
		EventId:       uuid.New().String(),
		Product:       p.domainProductToProto(product),
		EventTime:     timestamppb.Now(),
		CorrelationId: p.getCorrelationID(ctx),
		Data: &eventv1.ProductDeletedEventData{
			Source: "product-service",
			Reason: "manual_deletion",
			Metadata: map[string]string{
				"operation": "delete_product",
				"version":   "v1",
			},
		},
	}
	return p.publisher.Publish(ctx, event)
}

func (p *productUsecase) domainProductToProto(product *domain.Product) *v1.Product {
	return &v1.Product{
		Id:        product.ID.String(),
		Name:      product.Name,
		Price:     product.GetPriceString(),
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}
}

func (p *productUsecase) getCorrelationID(ctx context.Context) string {
	return uuid.New().String()
}