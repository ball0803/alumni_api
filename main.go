package main

import (
	"alumni_api/config"
	"alumni_api/internal/db"
	"alumni_api/internal/logger"
	"alumni_api/internal/middlewares"
	"alumni_api/internal/routes"
	"alumni_api/internal/validators"
	"context"
	"crypto/tls"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.uber.org/zap"
)

func main() {
	logger, err := logger.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	ctx := context.Background()
	cfg := config.LoadConfig()
	validators.Init()

	driver, err := db.ConnectToDB(ctx, cfg.Neo4jURI, cfg.Neo4jUsername, cfg.Neo4jPassword, logger)
	if err != nil {
		logger.Fatal("Could not connect to Neo4j", zap.Error(err))
	}
	defer driver.Close(ctx)

	// Set up Fiber app
	app := fiber.New()
	api := app.Group("/v1")
	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE",
		AllowHeaders:     "Content-Type,Authorization",
		AllowCredentials: true,
	}))

	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")
	if err != nil {
		logger.Fatal("Error loading certificate: ", zap.Error(err))
	}

	tlsConfig := &tls.Config{
		Certificates:     []tls.Certificate{cert},
		MinVersion:       tls.VersionTLS12,
		CurvePreferences: []tls.CurveID{tls.X25519, tls.CurveP256},
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	api.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})
	app.Use(middlewares.RequestLogger(logger))

	routes.UserRoutes(api, driver, logger)

	routes.AuthRoutes(api, driver, logger)

	routes.PostRoutes(api, driver, logger)

	routes.UploadRoutes(api, driver, logger)

	routes.MessageRoutes(api, driver, logger)

	routes.StatRoutes(api, driver, logger)

	// Start the server
	ln, err := tls.Listen("tcp", cfg.ServerPort, tlsConfig)
	if err != nil {
		logger.Fatal("Listener error:", zap.Error(err))
	}

	// Start with TLS
	if err := app.Listener(ln); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
