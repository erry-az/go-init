package grpc

import (
	"context"
	"errors"

	"github.com/erry-az/go-sample/internal/domain"
	"github.com/erry-az/go-sample/internal/usecase"
	"github.com/erry-az/go-sample/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProductService struct {
	v1.UnimplementedProductServiceServer
	productUsecase usecase.ProductUsecase
}

func NewProductService(productUsecase usecase.ProductUsecase) *ProductService {
	return &ProductService{
		productUsecase: productUsecase,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, req *v1.CreateProductRequest) (*v1.CreateProductResponse, error) {
	product, err := s.productUsecase.CreateProduct(ctx, req.Name, req.Price)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &v1.CreateProductResponse{Product: s.domainProductToProto(product)}, nil
}

func (s *ProductService) GetProduct(ctx context.Context, req *v1.GetProductRequest) (*v1.GetProductResponse, error) {
	product, err := s.productUsecase.GetProduct(ctx, req.Id)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &v1.GetProductResponse{Product: s.domainProductToProto(product)}, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, req *v1.UpdateProductRequest) (*v1.UpdateProductResponse, error) {
	product, err := s.productUsecase.UpdateProduct(ctx, req.Id, req.Name, req.Price)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &v1.UpdateProductResponse{Product: s.domainProductToProto(product)}, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, req *v1.DeleteProductRequest) (*emptypb.Empty, error) {
	err := s.productUsecase.DeleteProduct(ctx, req.Id)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *ProductService) ListProducts(ctx context.Context, req *v1.ListProductsRequest) (*v1.ListProductsResponse, error) {
	listReq := &usecase.ListProductsRequest{
		PageSize:    req.PageSize,
		PageToken:   req.PageToken,
		SearchQuery: req.SearchQuery,
	}

	// Convert price range if provided
	if req.PriceRange != nil {
		listReq.PriceRange = &usecase.PriceRange{
			MinPrice: req.PriceRange.MinPrice,
			MaxPrice: req.PriceRange.MaxPrice,
		}
	}

	result, err := s.productUsecase.ListProducts(ctx, listReq)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	products := make([]*v1.Product, len(result.Products))
	for i, product := range result.Products {
		products[i] = s.domainProductToProto(product)
	}

	return &v1.ListProductsResponse{
		Products:      products,
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalCount,
	}, nil
}

func (s *ProductService) BulkUpdatePrices(ctx context.Context, req *v1.BulkUpdatePricesRequest) (*v1.BulkUpdatePricesResponse, error) {
	updates := make([]usecase.BulkPriceUpdate, len(req.Updates))
	for i, update := range req.Updates {
		updates[i] = usecase.BulkPriceUpdate{
			ID:    update.Id,
			Price: update.Price,
		}
	}

	result, err := s.productUsecase.BulkUpdatePrices(ctx, updates)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	updatedProducts := make([]*v1.Product, len(result.UpdatedProducts))
	for i, product := range result.UpdatedProducts {
		updatedProducts[i] = s.domainProductToProto(product)
	}

	return &v1.BulkUpdatePricesResponse{
		UpdatedProducts: updatedProducts,
		FailedIds:       result.FailedIDs,
	}, nil
}

func (s *ProductService) GetProductAnalytics(ctx context.Context, req *v1.ProductAnalyticsRequest) (*v1.ProductAnalyticsResponse, error) {
	result, err := s.productUsecase.GetProductAnalytics(ctx)
	if err != nil {
		var domainErr *domain.DomainError
		if errors.As(err, &domainErr) {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	categoryStats := make([]*v1.ProductCategoryStats, len(result.CategoryStats))
	for i, stat := range result.CategoryStats {
		categoryStats[i] = &v1.ProductCategoryStats{
			Category: stat.Category,
			Count:    stat.Count,
		}
	}

	return &v1.ProductAnalyticsResponse{
		TotalProducts: result.TotalProducts,
		AveragePrice:  result.AveragePrice,
		HighestPrice:  result.HighestPrice,
		LowestPrice:   result.LowestPrice,
		CategoryStats: categoryStats,
	}, nil
}

// Helper method to convert domain product to protobuf
func (s *ProductService) domainProductToProto(product *domain.Product) *v1.Product {
	return &v1.Product{
		Id:        product.ID.String(),
		Name:      product.Name,
		Price:     product.GetPriceString(),
		CreatedAt: timestamppb.New(product.CreatedAt),
		UpdatedAt: timestamppb.New(product.UpdatedAt),
	}
}
