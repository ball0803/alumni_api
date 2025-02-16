package db

import (
	"context"
	"fmt"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"go.uber.org/zap"
)

func ConnectToDB(ctx context.Context, uri, username, password string, logger *zap.Logger) (neo4j.DriverWithContext, error) {
	// Create a new Neo4j driver instance
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		logger.Error("Failed to create Neo4j driver", zap.Error(err))
		return nil, fmt.Errorf("could not create driver: %w", err)
	}

	// Verify connectivity to Neo4j
	if err := driver.VerifyConnectivity(ctx); err != nil {
		logger.Error("Neo4j connectivity verification failed", zap.Error(err))
		driver.Close(ctx)
		return nil, fmt.Errorf("connectivity verification failed: %w", err)
	}
	logger.Info("Successfully connected to Neo4j", zap.String("URI", uri))

	return driver, nil
}
