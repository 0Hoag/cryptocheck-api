package httpserver

import (
	"context"
	"fmt"

	"github.com/0Hoag/cryptocheck-api/config"
	_ "github.com/0Hoag/cryptocheck-api/docs"
	prod "github.com/0Hoag/cryptocheck-api/internal/delivery/rabbitmq/producer"
	"github.com/0Hoag/cryptocheck-api/pkg/jwt"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/0Hoag/cryptocheck-api/internal/middleware"
	postHTTP "github.com/0Hoag/cryptocheck-api/internal/post/delivery/http"
	postMongo "github.com/0Hoag/cryptocheck-api/internal/post/repository/mongo"
	postUC "github.com/0Hoag/cryptocheck-api/internal/post/usecase"

	followHTTP "github.com/0Hoag/cryptocheck-api/internal/follow/delivery/http"
	followMongo "github.com/0Hoag/cryptocheck-api/internal/follow/repository/mongo"
	followUC "github.com/0Hoag/cryptocheck-api/internal/follow/usecase"

	commentHTTP "github.com/0Hoag/cryptocheck-api/internal/comment/delivery/http"
	commentMongo "github.com/0Hoag/cryptocheck-api/internal/comment/repository/mongo"
	commentUC "github.com/0Hoag/cryptocheck-api/internal/comment/usecase"

	userHTTP "github.com/0Hoag/cryptocheck-api/internal/users/delivery/http"
	userMongo "github.com/0Hoag/cryptocheck-api/internal/users/repository/mongo"
	userUC "github.com/0Hoag/cryptocheck-api/internal/users/usecase"

	authHTTP "github.com/0Hoag/cryptocheck-api/internal/auth/delivery/http"
	authUC "github.com/0Hoag/cryptocheck-api/internal/auth/usecase"

	prelaunchHTTP "github.com/0Hoag/cryptocheck-api/internal/prelaunch/delivery/http"
	scanHTTP "github.com/0Hoag/cryptocheck-api/internal/scanner/delivery/http"
	scanUC "github.com/0Hoag/cryptocheck-api/internal/scanner/usecase"

	// Import this to execute the init function in docs.go which setups the Swagger docs.
	_ "github.com/0Hoag/cryptocheck-api/docs"
)

func (srv HTTPServer) mapHandlers() error {
	srv.gin.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load application configuration: %w", err)
	}
	jwtManager := jwt.NewManager(srv.jwtSecretKey)

	// RabbitMQ is optional in local development. Avoid opening a channel on an
	// empty AMQP connection when the broker is unavailable.
	postProd := prod.NewNoop()
	if srv.amqpConn.IsReady() {
		postProd = prod.New(srv.l, srv.amqpConn)
		if err := postProd.Run(); err != nil {
			return err
		}
	} else {
		srv.l.Warnf(context.Background(), "RabbitMQ unavailable; asynchronous notifications are disabled")
	}

	// Repositories
	userRepo := userMongo.New(srv.l, srv.db)
	postRepo := postMongo.New(srv.l, srv.db)
	followRepo := followMongo.New(srv.l, srv.db)
	commentRepo := commentMongo.New(srv.l, srv.db)

	// Usecases
	userUC := userUC.New(srv.l, userRepo)
	postUC := postUC.New(srv.l, postProd, userUC, postRepo)
	followUC := followUC.New(srv.l, userUC, followRepo)
	commentUC := commentUC.New(srv.l, postUC, commentRepo)
	authUC := authUC.New(srv.l, cfg, userUC)
	scanUsecase := srv.scannerUC
	if scanUsecase == nil {
		if srv.scanEngine == nil || srv.dexClient == nil || srv.ethClient == nil {
			return fmt.Errorf("scanner dependencies are not configured")
		}
		scanUsecase = scanUC.New(srv.l, srv.scanEngine, srv.dexClient, srv.ethClient)
	}

	// Handlers
	userH := userHTTP.New(srv.l, userUC)
	postH := postHTTP.New(srv.l, postUC)
	followH := followHTTP.New(srv.l, followUC)
	commentH := commentHTTP.New(srv.l, commentUC)
	authH := authHTTP.New(srv.l, authUC)
	scanH := scanHTTP.New(srv.l, scanUsecase)

	// Middlewares
	mw := middleware.New(srv.l, userUC, jwtManager, srv.encrypter, srv.internalKey)

	// Public routes
	srv.gin.Use(mw.Locale())
	api := srv.gin.Group("/api/v1")

	newsFeedGroup := api.Group("/news-feed")
	userHTTP.MapRoutes(newsFeedGroup.Group("/users"), userH, mw)
	authHTTP.MapRoutes(newsFeedGroup.Group("/auth"), authH, mw)
	postHTTP.MapRoutes(newsFeedGroup.Group("/posts"), postH, mw)
	followHTTP.MapRoutes(newsFeedGroup.Group("/follow"), followH, mw)
	commentHTTP.MapRoutes(newsFeedGroup.Group("/comment"), commentH, mw)
	scanHTTP.MapRoutes(newsFeedGroup.Group("/scanner"), scanH, mw)
	prelaunchHTTP.MapRoutes(newsFeedGroup.Group("/prelaunch-projects"), srv.db, mw)

	return nil
}
