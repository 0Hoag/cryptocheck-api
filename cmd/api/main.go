package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/0Hoag/cryptocheck-api/config"
	httpserver "github.com/0Hoag/cryptocheck-api/internal/httpserver"

	"github.com/0Hoag/cryptocheck-api/internal/appconfig/mongo"
	pkgCrt "github.com/0Hoag/cryptocheck-api/pkg/encrypter"
	pkgLog "github.com/0Hoag/cryptocheck-api/pkg/log"
	"github.com/0Hoag/cryptocheck-api/pkg/rabbitmq"
	"github.com/joho/godotenv"

	"github.com/0Hoag/cryptocheck-api/internal/adapters/dexscreener"
	"github.com/0Hoag/cryptocheck-api/internal/adapters/etherscan"
	"github.com/0Hoag/cryptocheck-api/internal/adapters/gemini"
	"github.com/0Hoag/cryptocheck-api/internal/core/scanner"
)

func main() {

	// Load .env file
	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	crp := pkgCrt.NewEncrypter(cfg.Encrypter.Key)
	client, err := mongo.Connect(cfg.Mongo, crp)
	if err != nil {
		panic(err)
	}
	defer mongo.Disconnect(client)

	db := client.Database(cfg.Mongo.Database)

	l := pkgLog.InitializeZapLogger(pkgLog.ZapConfig{
		Level:    cfg.Logger.Level,
		Mode:     cfg.Logger.Mode,
		Encoding: cfg.Logger.Encoding,
	})

	amqpConn, err := rabbitmq.Dial(cfg.RabbitConfig.URL, true)
	if err != nil {
		l.Warnf(context.Background(), "RabbitMQ not connected, running without queue...")
		amqpConn = rabbitmq.Connection{}
	}
	defer amqpConn.Close()

	portStr := os.Getenv("PORT")
	if portStr == "" {
		portStr = strconv.Itoa(cfg.HTTPServer.Port) // fallback when running locally
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		panic(err)
	}

	// ---------------------------------------------------------
	// SCANNER MODULE INIT
	// ---------------------------------------------------------
	ethKey := os.Getenv("ETHERSCAN_API_KEY")
	bscKey := os.Getenv("BSCSCAN_API_KEY")
	baseKey := os.Getenv("BASESCAN_API_KEY")
	arbKey := os.Getenv("ARBISCAN_API_KEY")
	polyKey := os.Getenv("POLYGONSCAN_API_KEY")
	geminiKey := os.Getenv("GEMINI_API_KEY")

	apiKeys := map[string]string{
		etherscan.NetworkETH:      ethKey,
		etherscan.NetworkBSC:      bscKey,
		etherscan.NetworkBase:     baseKey,
		etherscan.NetworkArbitrum: arbKey,
		etherscan.NetworkPolygon:  polyKey,
	}
	ethClient := etherscan.NewClient(apiKeys)

	var geminiClient *gemini.Client
	if geminiKey != "" {
		geminiClient = gemini.NewClient(geminiKey)
		fmt.Println("✅ Gemini AI Integration: ENABLED")
	}

	scanEngine := scanner.NewEngine(geminiClient)
	dexClient := dexscreener.NewClient()
	// ---------------------------------------------------------

	srv := httpserver.New(l, httpserver.Config{
		Port:           port,
		DB:             db,
		AMQPConn:       amqpConn,
		JWTSecretKey:   cfg.JWT.SecretKey,
		Mode:           cfg.HTTPServer.Mode,
		AllowedOrigins: cfg.CORS.AllowedOrigins,
		TrustedProxies: cfg.CORS.TrustedProxies,
		Encrypter:      crp,
		SecretConfig: httpserver.SecretConfig{
			SecretKey: cfg.Encrypter.Key,
		},
		ScanEngine: scanEngine,
		DexClient:  dexClient,
		EthClient:  ethClient,
	})

	if err := srv.Run(); err != nil {
		panic(err)
	}
}
