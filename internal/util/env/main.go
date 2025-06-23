package env

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type RedisVariables struct {
	URL string
}

type DBVariables struct {
	Port     int
	Name     string
	Username string
	Password string
	Hostname string
}

type JWTVariables struct {
	SecretKey string
}

type EnvVariables struct {
	DB         DBVariables
	ServerPort int
	JWT        JWTVariables
	Redis      RedisVariables
}

func LoadEnv() (*EnvVariables, error) {
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	var envVars EnvVariables
	var (
		serverPortStr string
		dbPortStr     string
		dbName        string
		dbPassword    string
		dbUsername    string
		dbHostname    string
		secretKey     string
		redisURL      string
	)

	serverPortStr, err = variableValidater("SERVER_PORT")
	if err != nil {
		return nil, err
	}
	envVars.ServerPort, err = portValidater(serverPortStr)
	if err != nil {
		return nil, err
	}

	dbPortStr, err = variableValidater("DB_PORT")
	if err != nil {
		return nil, err
	}
	envVars.DB.Port, err = portValidater(dbPortStr)
	if err != nil {
		return nil, err
	}

	dbName, err = variableValidater("DB_NAME")
	if err != nil {
		return nil, err
	}
	envVars.DB.Name = dbName

	dbPassword, err = variableValidater("DB_PASSWORD")
	if err != nil {
		return nil, err
	}
	envVars.DB.Password = dbPassword

	dbUsername, err = variableValidater("DB_USERNAME")
	if err != nil {
		return nil, err
	}
	envVars.DB.Username = dbUsername

	dbHostname, err = variableValidater("DB_HOST")
	if err != nil {
		return nil, err
	}
	envVars.DB.Hostname = dbHostname

	secretKey, err = variableValidater("SECRET_KEY")
	if err != nil {
		return nil, err
	}
	envVars.JWT.SecretKey = secretKey

	redisURL, err = variableValidater("REDIS_URL")

	if err != nil {
		return nil, err
	}

	envVars.Redis.URL = redisURL

	return &envVars, nil
}

func portValidater(portStr string) (int, error) {
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("the Port is not a valid integer: %w", err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("the port %d is out of range", port)
	}
	return port, nil
}

func variableValidater(varStr string) (string, error) {
	variable := os.Getenv(varStr)
	if variable == "" {
		return "", fmt.Errorf("%s environment variable is not set", varStr)
	}
	return variable, nil
}
