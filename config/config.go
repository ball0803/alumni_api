package config

import (
	"fmt"
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
)

// Config struct holds all the configurations for the application
type Config struct {
	DBEnv            string
	ServerPort       string
	Neo4jURI         string
	Neo4jUsername    string
	Neo4jPassword    string
	RedisAddress     string
	RedisPassword    string
	KafkaBrokers     []string
	MinREDThreshold  int
	MaxREDThreshold  int
	MaxREDProb       float64
	AESEncryptionKey []byte
}

// LoadConfig loads configuration values from environment variables or defaults
func LoadConfig() Config {
	// Load .env file if available
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables or defaults")
	}

	dbEnv := GetEnv("DB_ENV", "local")

	config := Config{
		DBEnv:            dbEnv,
		ServerPort:       fmt.Sprintf(":%s", GetEnv("PORT", "3000")),
		AESEncryptionKey: []byte(GetEnv("AES_ENCRYPTION_KEY", "thisis32byteslongkeyforaes256!")),
	}

	if dbEnv == "aura" {
		config.Neo4jURI = GetEnv("NEO4J_AURA_URI", "neo4j+ssc://your_aura_uri:7687")
		config.Neo4jUsername = GetEnv("NEO4J_AURA_USERNAME", "aura_user")
		config.Neo4jPassword = GetEnv("NEO4J_AURA_PASSWORD", "aura_password")
	} else {
		config.Neo4jURI = GetEnv("NEO4J_LOCAL_URI", "neo4j://localhost:7687")
		config.Neo4jUsername = GetEnv("NEO4J_LOCAL_USERNAME", "local_user")
		config.Neo4jPassword = GetEnv("NEO4J_LOCAL_PASSWORD", "local_password")
	}

	return config
}

func GetEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := GetEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsFloat(key string, defaultValue float64) float64 {
	valueStr := GetEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}
