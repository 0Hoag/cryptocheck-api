package httpserver

import (
	"github.com/0Hoag/cryptocheck-api/internal/adapters/dexscreener"
	"github.com/0Hoag/cryptocheck-api/internal/adapters/etherscan"
	"github.com/0Hoag/cryptocheck-api/internal/core/scanner"
	scanDomain "github.com/0Hoag/cryptocheck-api/internal/scanner"
	pkgCrt "github.com/0Hoag/cryptocheck-api/pkg/encrypter"
	pkgLog "github.com/0Hoag/cryptocheck-api/pkg/log"
	pkgMongo "github.com/0Hoag/cryptocheck-api/pkg/mongo"
	"github.com/0Hoag/cryptocheck-api/pkg/rabbitmq"
	"github.com/gin-gonic/gin"
)

const productionMode = "production"

var ginMode = gin.DebugMode

type HTTPServer struct {
	gin          *gin.Engine
	l            pkgLog.Logger
	port         int
	db           pkgMongo.Database
	amqpConn     rabbitmq.Connection
	jwtSecretKey string
	mode         string
	hoagConfig   HoagConfig
	internalKey  string
	encrypter    pkgCrt.Encrypter
	secretConfig SecretConfig

	// Usecase
	scannerUC  scanDomain.UseCase
	scanEngine *scanner.Engine
	dexClient  *dexscreener.Client
	ethClient  *etherscan.Client
}

type Config struct {
	Port         int
	JWTSecretKey string
	DB           pkgMongo.Database
	AMQPConn     rabbitmq.Connection
	Mode         string
	HoagConfig   HoagConfig
	InternalKey  string
	Encrypter    pkgCrt.Encrypter
	SecretConfig SecretConfig

	// Dependency Injection
	ScanEngine *scanner.Engine
	DexClient  *dexscreener.Client
	EthClient  *etherscan.Client

	// Pre-built UC (Optional)
	ScannerUC scanDomain.UseCase
}

type HoagConfig struct {
	AdminDomain string
}

type SecretConfig struct {
	SecretKey string
}

func New(l pkgLog.Logger, cfg Config) *HTTPServer {
	if cfg.Mode == productionMode {
		ginMode = gin.ReleaseMode
	}

	gin.SetMode(ginMode)

	engine := gin.Default()

	// Simple CORS middleware
	engine.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	return &HTTPServer{
		l:            l,
		gin:          engine,
		port:         cfg.Port,
		db:           cfg.DB,
		amqpConn:     cfg.AMQPConn,
		jwtSecretKey: cfg.JWTSecretKey,
		mode:         cfg.Mode,
		hoagConfig:   cfg.HoagConfig,
		internalKey:  cfg.InternalKey,
		encrypter:    cfg.Encrypter,
		secretConfig: cfg.SecretConfig,

		scannerUC:  cfg.ScannerUC,
		scanEngine: cfg.ScanEngine,
		dexClient:  cfg.DexClient,
		ethClient:  cfg.EthClient,
	}
}
