package grpc

import (
	"context"

	"github.com/erry-az/go-init/internal/domain"
	"github.com/erry-az/go-init/internal/usecase"
	"github.com/erry-az/go-init/proto/api/v1"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserService struct {
	v1.UnimplementedUserServiceServer
	userUsecase usecase.UserUsecase
}

func NewUserService(userUsecase usecase.UserUsecase) *UserService {
	return &UserService{
		userUsecase: userUsecase,
	}
}

func (s *UserService) CreateUser(ctx context.Context, req *v1.CreateUserRequest) (*v1.CreateUserResponse, error) {
	user, err := s.userUsecase.CreateUser(ctx, req.Name, req.Email)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &v1.CreateUserResponse{User: s.domainUserToProto(user)}, nil
}

func (s *UserService) GetUser(ctx context.Context, req *v1.GetUserRequest) (*v1.GetUserResponse, error) {
	user, err := s.userUsecase.GetUser(ctx, req.Id)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &v1.GetUserResponse{User: s.domainUserToProto(user)}, nil
}

func (s *UserService) UpdateUser(ctx context.Context, req *v1.UpdateUserRequest) (*v1.UpdateUserResponse, error) {
	user, err := s.userUsecase.UpdateUser(ctx, req.Id, req.Name, req.Email)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &v1.UpdateUserResponse{User: s.domainUserToProto(user)}, nil
}

func (s *UserService) DeleteUser(ctx context.Context, req *v1.DeleteUserRequest) (*emptypb.Empty, error) {
	err := s.userUsecase.DeleteUser(ctx, req.Id)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	return &emptypb.Empty{}, nil
}

func (s *UserService) ListUsers(ctx context.Context, req *v1.ListUsersRequest) (*v1.ListUsersResponse, error) {
	listReq := &usecase.ListUsersRequest{
		PageSize:    req.PageSize,
		PageToken:   req.PageToken,
		SearchQuery: req.SearchQuery,
	}

	result, err := s.userUsecase.ListUsers(ctx, listReq)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	users := make([]*v1.User, len(result.Users))
	for i, user := range result.Users {
		users[i] = s.domainUserToProto(user)
	}

	return &v1.ListUsersResponse{
		Users:         users,
		NextPageToken: result.NextPageToken,
		TotalCount:    result.TotalCount,
	}, nil
}

func (s *UserService) BulkCreateUsers(ctx context.Context, req *v1.BulkCreateUsersRequest) (*v1.BulkCreateUsersResponse, error) {
	bulkUsers := make([]usecase.BulkCreateUserRequest, len(req.Users))
	for i, userReq := range req.Users {
		bulkUsers[i] = usecase.BulkCreateUserRequest{
			Name:  userReq.Name,
			Email: userReq.Email,
		}
	}

	result, err := s.userUsecase.BulkCreateUsers(ctx, bulkUsers)
	if err != nil {
		if domainErr, ok := err.(*domain.DomainError); ok {
			return nil, domainErr.ToGRPCError()
		}
		return nil, err
	}

	users := make([]*v1.User, len(result.Users))
	for i, user := range result.Users {
		users[i] = s.domainUserToProto(user)
	}

	return &v1.BulkCreateUsersResponse{
		Users:        users,
		FailedEmails: result.FailedEmails,
	}, nil
}

// Helper method to convert domain user to protobuf
func (s *UserService) domainUserToProto(user *domain.User) *v1.User {
	return &v1.User{
		Id:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}
}

