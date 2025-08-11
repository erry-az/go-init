package grpc

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	sqlc2 "github.com/erry-az/go-init/internal/repository/sqlc"
	"github.com/erry-az/go-init/pkg/watermill"
	"github.com/erry-az/go-init/proto/api/v1"
	eventv1 "github.com/erry-az/go-init/proto/event/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductService struct {
	v1.UnimplementedProductServiceServer
	db        sqlc2.Querier
	publisher *watermill.Publisher
}

func NewProductService(db sqlc2.Querier, publisher *watermill.Publisher) *ProductService {
	return &ProductService{
		db:        db,
		publisher: publisher,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	productID := uuid.New()

	// Parse and validate price
	priceDecimal, err := decimal.NewFromString(req.Price)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid price format: %v", err)
	}

	// Convert decimal to pgtype.Numeric
	var price pgtype.Numeric
	if err := price.Scan(priceDecimal.String()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid price conversion: %v", err)
	}

	params := sqlc2.CreateProductParams{
		ID:    productID,
		Name:  req.Name,
		Price: price,
	}

	dbProduct, err := s.db.CreateProduct(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create product: %v", err)
	}

	product := s.mapDBProductToProto(dbProduct)

	// Publish product created event
	event := &eventv1.ProductCreatedEvent{
		EventId:       uuid.New().String(),
		Product:       product,
		EventTime:     timestamppb.Now(),
		CorrelationId: s.getCorrelationID(ctx),
		Data: &eventv1.ProductCreatedEventData{
			Source: "product-service",
			Metadata: map[string]string{
				"operation": "create_product",
				"version":   "v1",
			},
		},
	}

	event.Publish(ctx, s.publisher)
	if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicProductCreated, event); err != nil {
		fmt.Printf("Failed to publish product created event: %v\n", err)
	}

	return &v1.CreateProductResponse{Product: product}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.GetProductResponse, error) {
	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID: %v", err)
	}

	dbProduct, err := s.db.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	product := s.mapDBProductToProto(dbProduct)
	return &v1.GetProductResponse{Product: product}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductResponse, error) {
	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID: %v", err)
	}

	// Get existing product for price change detection
	existingProduct, err := s.db.GetProductByID(ctx, productID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get existing product: %v", err)
	}

	// Parse and validate price
	newPriceDecimal, err := decimal.NewFromString(req.Price)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid price format: %v", err)
	}

	// Convert decimal to pgtype.Numeric
	var newPrice pgtype.Numeric
	if err := newPrice.Scan(newPriceDecimal.String()); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid price conversion: %v", err)
	}

	params := sqlc2.UpdateProductParams{
		ID:    productID,
		Name:  req.Name,
		Price: newPrice,
	}

	dbProduct, err := s.db.UpdateProduct(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update product: %v", err)
	}

	product := s.mapDBProductToProto(dbProduct)

	// Compare prices for change detection
	existingPriceStr := s.numericToString(existingProduct.Price)
	newPriceStr := s.numericToString(newPrice)

	// Publish product updated event
	updateEvent := &eventv1.ProductUpdatedEvent{
		EventId:       uuid.New().String(),
		Product:       product,
		EventTime:     timestamppb.Now(),
		CorrelationId: s.getCorrelationID(ctx),
		Data: &eventv1.ProductUpdatedEventData{
			Source:        "product-service",
			ChangedFields: []string{"name", "price"}, // Could be more dynamic
			Metadata: map[string]string{
				"operation": "update_product",
				"version":   "v1",
			},
		},
	}

	if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicProductUpdated, updateEvent); err != nil {
		fmt.Printf("Failed to publish product updated event: %v\n", err)
	}

	// If price changed, also publish price change event
	if existingPriceStr != newPriceStr {
		priceChangeEvent := &eventv1.ProductPriceChangedEvent{
			EventId:       uuid.New().String(),
			Product:       product,
			EventTime:     timestamppb.Now(),
			CorrelationId: s.getCorrelationID(ctx),
			Data: &eventv1.ProductPriceChangedEventData{
				Source:        "product-service",
				PreviousPrice: existingPriceStr,
				NewPrice:      newPriceStr,
				Metadata: map[string]string{
					"operation": "price_change",
					"version":   "v1",
				},
			},
		}

		if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicProductPriceChanged, priceChangeEvent); err != nil {
			fmt.Printf("Failed to publish product price changed event: %v\n", err)
		}
	}

	return &v1.UpdateProductResponse{Product: product}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	productID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid product ID: %v", err)
	}

	// Get product before deletion for event
	dbProduct, err := s.db.GetProductByID(ctx, productID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "product not found")
		}
		return nil, status.Errorf(codes.Internal, "failed to get product: %v", err)
	}

	err = s.db.DeleteProduct(ctx, productID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete product: %v", err)
	}

	product := s.mapDBProductToProto(dbProduct)

	// Publish product deleted event
	event := &eventv1.ProductDeletedEvent{
		EventId:       uuid.New().String(),
		Product:       product,
		EventTime:     timestamppb.Now(),
		CorrelationId: s.getCorrelationID(ctx),
		Data: &eventv1.ProductDeletedEventData{
			Source: "product-service",
			Reason: "manual_deletion",
			Metadata: map[string]string{
				"operation": "delete_product",
				"version":   "v1",
			},
		},
	}

	if err := s.publisher.PublishProtoMessage(ctx, watermill.TopicProductDeleted, event); err != nil {
		fmt.Printf("Failed to publish product deleted event: %v\n", err)
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
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
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token")
		}
		if _, err := fmt.Sscanf(string(decodedOffset), "%d", &offset); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid page token format")
		}
	}

	var dbProducts []sqlc2.Product
	var err error

	// Handle different query types based on request parameters
	if req.SearchQuery != "" && req.PriceRange != nil {
		var minPrice, maxPrice pgtype.Numeric
		if err := minPrice.Scan(req.PriceRange.MinPrice); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid min price: %v", err)
		}
		if err := maxPrice.Scan(req.PriceRange.MaxPrice); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid max price: %v", err)
		}

		params := sqlc2.SearchProductsWithPriceRangeParams{
			Limit:       pageSize + 1,
			Offset:      offset,
			SearchQuery: "%" + req.SearchQuery + "%",
			MinPrice:    minPrice,
			MaxPrice:    maxPrice,
		}
		dbProducts, err = s.db.SearchProductsWithPriceRange(ctx, params)
	} else if req.SearchQuery != "" {
		params := sqlc2.SearchProductsParams{
			Limit:       pageSize + 1,
			Offset:      offset,
			SearchQuery: "%" + req.SearchQuery + "%",
		}
		dbProducts, err = s.db.SearchProducts(ctx, params)
	} else if req.PriceRange != nil {
		var minPrice, maxPrice pgtype.Numeric
		if err := minPrice.Scan(req.PriceRange.MinPrice); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid min price: %v", err)
		}
		if err := maxPrice.Scan(req.PriceRange.MaxPrice); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid max price: %v", err)
		}

		params := sqlc2.ListProductsByPriceRangeParams{
			Limit:    pageSize + 1,
			Offset:   offset,
			MinPrice: minPrice,
			MaxPrice: maxPrice,
		}
		dbProducts, err = s.db.ListProductsByPriceRange(ctx, params)
	} else {
		params := sqlc2.ListProductsParams{
			Limit:  pageSize + 1,
			Offset: offset,
		}
		dbProducts, err = s.db.ListProducts(ctx, params)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list products: %v", err)
	}

	// Check if there are more pages
	hasNextPage := len(dbProducts) > int(pageSize)
	if hasNextPage {
		dbProducts = dbProducts[:pageSize]
	}

	products := make([]*v1.Product, len(dbProducts))
	for i, dbProduct := range dbProducts {
		products[i] = s.mapDBProductToProto(dbProduct)
	}

	var nextPageToken string
	if hasNextPage {
		nextOffset := offset + pageSize
		nextPageToken = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", nextOffset)))
	}

	// Get total count
	totalCount, err := s.db.CountProducts(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to count products: %v", err)
	}

	return &v1.ListProductsResponse{
		Products:      products,
		NextPageToken: nextPageToken,
		TotalCount:    int32(totalCount),
	}, nil
}

func (s *ProductService) BulkUpdatePrices(ctx context.Context, req *v1.BulkUpdatePricesRequest) (*v1.BulkUpdatePricesResponse, error) {
	var updatedProducts []*v1.Product
	var failedIDs []string

	for _, update := range req.Updates {
		updateReq := &v1.UpdateProductRequest{
			Id:    update.Id,
			Price: update.Price,
		}

		// We need to get the current product to preserve name
		productID, err := uuid.Parse(update.Id)
		if err != nil {
			failedIDs = append(failedIDs, update.Id)
			continue
		}

		existingProduct, err := s.db.GetProductByID(ctx, productID)
		if err != nil {
			failedIDs = append(failedIDs, update.Id)
			continue
		}

		updateReq.Name = existingProduct.Name

		updateResp, err := s.UpdateProduct(ctx, updateReq)
		if err != nil {
			failedIDs = append(failedIDs, update.Id)
			continue
		}

		updatedProducts = append(updatedProducts, updateResp.Product)
	}

	return &v1.BulkUpdatePricesResponse{
		UpdatedProducts: updatedProducts,
		FailedIds:       failedIDs,
	}, nil
}

func (s *ProductService) GetProductAnalytics(ctx context.Context, req *v1.ProductAnalyticsRequest) (*v1.ProductAnalyticsResponse, error) {
	// This is a placeholder implementation - in a real system you'd have dedicated analytics queries
	totalCount, err := s.db.CountProducts(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get product count: %v", err)
	}

	// Get average, min, max prices (these return interface{} due to COALESCE)
	avgPriceInterface, err := s.db.GetAveragePrice(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get average price: %v", err)
	}

	minPriceInterface, err := s.db.GetMinPrice(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get min price: %v", err)
	}

	maxPriceInterface, err := s.db.GetMaxPrice(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get max price: %v", err)
	}

	// Convert interface{} to string (they should be pgtype.Numeric)
	avgPriceStr := "0"
	minPriceStr := "0"
	maxPriceStr := "0"

	if avgPrice, ok := avgPriceInterface.(pgtype.Numeric); ok && avgPrice.Valid {
		avgPriceStr = s.numericToString(avgPrice)
	}
	if minPrice, ok := minPriceInterface.(pgtype.Numeric); ok && minPrice.Valid {
		minPriceStr = s.numericToString(minPrice)
	}
	if maxPrice, ok := maxPriceInterface.(pgtype.Numeric); ok && maxPrice.Valid {
		maxPriceStr = s.numericToString(maxPrice)
	}

	return &v1.ProductAnalyticsResponse{
		TotalProducts: int32(totalCount),
		AveragePrice:  avgPriceStr,
		HighestPrice:  maxPriceStr,
		LowestPrice:   minPriceStr,
		CategoryStats: []*v1.ProductCategoryStats{}, // Placeholder - would need category field in schema
	}, nil
}

func (s *ProductService) mapDBProductToProto(dbProduct sqlc2.Product) *v1.Product {
	return &v1.Product{
		Id:        dbProduct.ID.String(),
		Name:      dbProduct.Name,
		Price:     s.numericToString(dbProduct.Price),
		CreatedAt: timestamppb.New(dbProduct.CreatedAt.Time),
		UpdatedAt: timestamppb.New(dbProduct.UpdatedAt.Time),
	}
}

func (s *ProductService) numericToString(n pgtype.Numeric) string {
	if !n.Valid || n.NaN {
		return "0"
	}

	// Use Value() method to get the string representation
	val, err := n.Value()
	if err != nil {
		return "0"
	}

	if str, ok := val.(string); ok {
		return str
	}

	return "0"
}

func (s *ProductService) getCorrelationID(ctx context.Context) string {
	return uuid.New().String()
}
