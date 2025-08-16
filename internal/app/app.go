package app

import (
	"database/sql"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/cqrs"
	"github.com/erry-az/go-sample/config"
	handlergrpc "github.com/erry-az/go-sample/internal/handler/grpc"
	"github.com/erry-az/go-sample/internal/repository/sqlc"
	"github.com/erry-az/go-sample/internal/usecase"
	"github.com/erry-az/go-sample/pkg/watmil"
	"github.com/jackc/pgx/v5/pgxpool"
)

// App represents the application with all dependencies
type App struct {
	UserUsecase    usecase.UserUsecase
	ProductUsecase usecase.ProductUsecase
	UserService    *handlergrpc.UserService
	ProductService *handlergrpc.ProductService
	Publisher      *cqrs.EventBus
}

// Dependencies holds external dependencies
type Dependencies struct {
	DBPool *pgxpool.Pool
	SqlDB  *sql.DB
	Logger watermill.LoggerAdapter
	Config *config.Config
}

// New creates a new application with all dependencies wired
func New(deps Dependencies) (*App, error) {
	// Create Watermill publisher
	publisher, err := watmil.NewPublisher(deps.SqlDB, deps.Logger)
	if err != nil {
		return nil, err
	}

	// Create SQLC querier
	querier := sqlc.New(deps.DBPool)

	// Create usecases
	userUsecase := usecase.NewUserUsecase(querier, publisher)
	productUsecase := usecase.NewProductUsecase(querier, publisher)

	// Create services
	userService := handlergrpc.NewUserService(userUsecase)
	productService := handlergrpc.NewProductService(productUsecase)

	return &App{
		UserUsecase:    userUsecase,
		ProductUsecase: productUsecase,
		UserService:    userService,
		ProductService: productService,
		Publisher:      publisher,
	}, nil
}
