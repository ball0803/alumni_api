package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

// Config struct holds all the configurations for the application
type Config struct {
	ServerPort      string
	Neo4jURI        string
	Neo4jUsername   string
	Neo4jPassword   string
	RedisAddress    string
	RedisPassword   string
	KafkaBrokers    []string
	MinREDThreshold int
	MaxREDThreshold int
	MaxREDProb      float64
}

// LoadConfig loads configuration values from environment variables or defaults
func LoadConfig() Config {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	return Config{
		ServerPort:    getEnv("SERVER_PORT", "3000"),
		Neo4jURI:      getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername: getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword: getEnv("NEO4J_PASSWORD", "password"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
