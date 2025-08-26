package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
	OAuth    OAuthConfig
	Frontend FrontendConfig
	Logging  LoggingConfig
}

type ServerConfig struct {
	Port         string
	BaseURL      string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host                string
	Port                string
	User                string
	Password            string
	Name                string
	SSLMode             string
	MaxOpenConns        int
	MaxIdleConns        int
	ConnMaxLifetime     time.Duration
	MigrationsPath      string
}

type AuthConfig struct {
	JWTSecret       string
	TokenExpiration time.Duration
	CookieSecure    bool
	CookieSameSite  string
}

type OAuthConfig struct {
	Google GoogleOAuthConfig
	GitHub GitHubOAuthConfig
}

type GoogleOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type GitHubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type FrontendConfig struct {
	URL       string
	CORSDebug bool
}

type LoggingConfig struct {
	Level      string
	Format     string
	JSONFormat bool
}

var GlobalConfig *Config

func Init() error {
	if err := godotenv.Load(); err != nil {
		fmt.Println("No .env file found, using system environment variables")
	}

	config, err := load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	if err := config.validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	GlobalConfig = config
	return nil
}

func load() (*Config, error) {
	config := &Config{
		Server:   loadServerConfig(),
		Database: loadDatabaseConfig(),
		Auth:     loadAuthConfig(),
		OAuth:    loadOAuthConfig(),
		Frontend: loadFrontendConfig(),
		Logging:  loadLoggingConfig(),
	}

	return config, nil
}

func loadServerConfig() ServerConfig {
	readTimeout, _ := strconv.Atoi(getEnv("SERVER_READ_TIMEOUT_SECONDS", "15"))
	writeTimeout, _ := strconv.Atoi(getEnv("SERVER_WRITE_TIMEOUT_SECONDS", "15"))
	idleTimeout, _ := strconv.Atoi(getEnv("SERVER_IDLE_TIMEOUT_SECONDS", "60"))

	return ServerConfig{
		Port:         getEnv("PORT", "8080"),
		BaseURL:      getEnv("BASE_URL", "http://localhost:8080"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}
}

func loadDatabaseConfig() DatabaseConfig {
	maxOpenConns, _ := strconv.Atoi(getEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(getEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := strconv.Atoi(getEnv("DB_CONN_MAX_LIFETIME_MINUTES", "5"))

	return DatabaseConfig{
		Host:                getEnv("DB_HOST", "localhost"),
		Port:                getEnv("DB_PORT", "5432"),
		User:                getEnv("DB_USER", "postgres"),
		Password:            getEnv("DB_PASSWORD", "postgres"),
		Name:                getEnv("DB_NAME", "planets"),
		SSLMode:             getEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:        maxOpenConns,
		MaxIdleConns:        maxIdleConns,
		ConnMaxLifetime:     time.Duration(connMaxLifetime) * time.Minute,
		MigrationsPath:      getEnv("DB_MIGRATIONS_PATH", "migrations"),
	}
}

func loadAuthConfig() AuthConfig {
	tokenExpiration, _ := strconv.Atoi(getEnv("JWT_EXPIRATION_HOURS", "24"))
	
	environment := getEnv("ENVIRONMENT", "development")
	cookieSecure := environment == "production"
	
	return AuthConfig{
		JWTSecret:       getEnv("JWT_SECRET", ""),
		TokenExpiration: time.Duration(tokenExpiration) * time.Hour,
		CookieSecure:    cookieSecure,
		CookieSameSite:  getEnv("COOKIE_SAME_SITE", "lax"),
	}
}

func loadOAuthConfig() OAuthConfig {
	baseURL := getEnv("BASE_URL", "http://localhost:8080")
	
	return OAuthConfig{
		Google: GoogleOAuthConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  baseURL + "/auth/google/callback",
			Scopes:       []string{"openid", "profile", "email"},
		},
		GitHub: GitHubOAuthConfig{
			ClientID:     getEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: getEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  baseURL + "/auth/github/callback",
			Scopes:       []string{"user:email"},
		},
	}
}

func loadFrontendConfig() FrontendConfig {
	corsDebug := getEnv("CORS_DEBUG", "") == "true"
	
	return FrontendConfig{
		URL:       getEnv("FRONTEND_URL", "http://localhost:3000"),
		CORSDebug: corsDebug,
	}
}

func loadLoggingConfig() LoggingConfig {
	environment := getEnv("ENVIRONMENT", "development")
	jsonFormat := environment == "production"
	
	return LoggingConfig{
		Level:      getEnv("LOG_LEVEL", "debug"),
		Format:     getEnv("LOG_FORMAT", "text"),
		JSONFormat: jsonFormat,
	}
}

func (c *Config) validate() error {
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required")
	}
	
	if len(c.Auth.JWTSecret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters long")
	}
	
	if c.Server.Port == "" {
		return fmt.Errorf("PORT is required")
	}
	
	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}
	
	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}
	
	if c.Server.BaseURL == "" {
		return fmt.Errorf("BASE_URL is required")
	}
	
	return nil
}

func (c *Config) IsProduction() bool {
	return c.Server.Environment == "production"
}

func (c *Config) IsDevelopment() bool {
	return c.Server.Environment == "development"
}

func (c *Config) GoogleOAuthConfigured() bool {
	return c.OAuth.Google.ClientID != "" && c.OAuth.Google.ClientSecret != ""
}

func (c *Config) GitHubOAuthConfigured() bool {
	return c.OAuth.GitHub.ClientID != "" && c.OAuth.GitHub.ClientSecret != ""
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
