package config

import (
	"fmt"
	"planets-server/internal/shared/utils"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Redis     RedisConfig
	Auth      AuthConfig
	OAuth     OAuthConfig
	Frontend  FrontendConfig
	Logging   LoggingConfig
	RateLimit RateLimitConfig
	Universe  UniverseConfig
	Admin     AdminConfig
}

type RedisConfig struct {
	Enabled  bool
	URL      string
	Host     string
	Port     string
	Password string
	DB       int
}

type ServerConfig struct {
	Port         string
	URL          string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type DatabaseConfig struct {
	Host            string
	Port            string
	User            string
	Password        string
	Name            string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	MigrationsPath  string
}

type AuthConfig struct {
	JWTSecret       string
	TokenExpiration time.Duration
	CookieSecure    bool
	CookieSameSite  string
}

type OAuthConfig struct {
	Google  GoogleOAuthConfig
	GitHub  GitHubOAuthConfig
	Discord DiscordOAuthConfig
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

type DiscordOAuthConfig struct {
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

type RateLimitConfig struct {
	Enabled           bool
	RequestsPerSecond float64
	BurstSize         int
}

type UniverseConfig struct {
	SectorCount         int
	SystemsPerSector    int
	MinPlanetsPerSystem int
	MaxPlanetsPerSystem int
	DefaultGalaxyName   string
}

type AdminConfig struct {
	Email       string
	Username    string
	DisplayName string
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
		Server:    loadServerConfig(),
		Database:  loadDatabaseConfig(),
		Redis:     loadRedisConfig(),
		Auth:      loadAuthConfig(),
		OAuth:     loadOAuthConfig(),
		Frontend:  loadFrontendConfig(),
		Logging:   loadLoggingConfig(),
		RateLimit: loadRateLimitConfig(),
		Universe:  loadUniverseConfig(),
		Admin:     loadAdminConfig(),
	}

	return config, nil
}

func loadRedisConfig() RedisConfig {
	enabled := utils.GetEnv("REDIS_ENABLED", "true") == "true"
	redisURL := utils.GetEnv("REDIS_URL", "")

	db, _ := strconv.Atoi(utils.GetEnv("REDIS_DB", "0"))

	return RedisConfig{
		Enabled:  enabled,
		URL:      redisURL,
		Host:     utils.GetEnv("REDIS_HOST", "localhost"),
		Port:     utils.GetEnv("REDIS_PORT", "6379"),
		Password: utils.GetEnv("REDIS_PASSWORD", ""),
		DB:       db,
	}
}

func loadServerConfig() ServerConfig {
	readTimeout, _ := strconv.Atoi(utils.GetEnv("SERVER_READ_TIMEOUT_SECONDS", "15"))
	writeTimeout, _ := strconv.Atoi(utils.GetEnv("SERVER_WRITE_TIMEOUT_SECONDS", "15"))
	idleTimeout, _ := strconv.Atoi(utils.GetEnv("SERVER_IDLE_TIMEOUT_SECONDS", "60"))

	return ServerConfig{
		Port:         utils.GetEnv("SERVER_PORT", "8080"),
		URL:          utils.GetEnv("SERVER_URL", "http://localhost:8080"),
		Environment:  utils.GetEnv("ENVIRONMENT", "development"),
		ReadTimeout:  time.Duration(readTimeout) * time.Second,
		WriteTimeout: time.Duration(writeTimeout) * time.Second,
		IdleTimeout:  time.Duration(idleTimeout) * time.Second,
	}
}

func loadDatabaseConfig() DatabaseConfig {
	maxOpenConns, _ := strconv.Atoi(utils.GetEnv("DB_MAX_OPEN_CONNS", "25"))
	maxIdleConns, _ := strconv.Atoi(utils.GetEnv("DB_MAX_IDLE_CONNS", "5"))
	connMaxLifetime, _ := strconv.Atoi(utils.GetEnv("DB_CONN_MAX_LIFETIME_MINUTES", "5"))

	return DatabaseConfig{
		Host:            utils.GetEnv("DB_HOST", "localhost"),
		Port:            utils.GetEnv("DB_PORT", "5432"),
		User:            utils.GetEnv("DB_USER", "postgres"),
		Password:        utils.GetEnv("DB_PASSWORD", "postgres"),
		Name:            utils.GetEnv("DB_NAME", "planets"),
		SSLMode:         utils.GetEnv("DB_SSLMODE", "disable"),
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: time.Duration(connMaxLifetime) * time.Minute,
		MigrationsPath:  utils.GetEnv("DB_MIGRATIONS_PATH", "migrations"),
	}
}

func loadAuthConfig() AuthConfig {
	tokenExpiration, _ := strconv.Atoi(utils.GetEnv("JWT_EXPIRATION_HOURS", "24"))

	environment := utils.GetEnv("ENVIRONMENT", "development")
	cookieSecure := environment == "production"

	return AuthConfig{
		JWTSecret:       utils.GetEnv("JWT_SECRET", ""),
		TokenExpiration: time.Duration(tokenExpiration) * time.Hour,
		CookieSecure:    cookieSecure,
		CookieSameSite:  utils.GetEnv("COOKIE_SAME_SITE", "lax"),
	}
}

func loadOAuthConfig() OAuthConfig {
	serverURL := utils.GetEnv("SERVER_URL", "http://localhost:8080")

	return OAuthConfig{
		Google: GoogleOAuthConfig{
			ClientID:     utils.GetEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: utils.GetEnv("GOOGLE_CLIENT_SECRET", ""),
			RedirectURL:  serverURL + "/auth/google/callback",
			Scopes:       []string{"openid", "profile", "email"},
		},
		GitHub: GitHubOAuthConfig{
			ClientID:     utils.GetEnv("GITHUB_CLIENT_ID", ""),
			ClientSecret: utils.GetEnv("GITHUB_CLIENT_SECRET", ""),
			RedirectURL:  serverURL + "/auth/github/callback",
			Scopes:       []string{"user:email"},
		},
		Discord: DiscordOAuthConfig{
			ClientID:     utils.GetEnv("DISCORD_CLIENT_ID", ""),
			ClientSecret: utils.GetEnv("DISCORD_CLIENT_SECRET", ""),
			RedirectURL:  serverURL + "/auth/discord/callback",
			Scopes:       []string{"identify", "email"},
		},
	}
}

func loadFrontendConfig() FrontendConfig {
	corsDebug := utils.GetEnv("CORS_DEBUG", "") == "true"

	return FrontendConfig{
		URL:       utils.GetEnv("FRONTEND_URL", "http://localhost:3000"),
		CORSDebug: corsDebug,
	}
}

func loadLoggingConfig() LoggingConfig {
	environment := utils.GetEnv("ENVIRONMENT", "development")
	jsonFormat := environment == "production"

	return LoggingConfig{
		Level:      utils.GetEnv("LOG_LEVEL", "debug"),
		Format:     utils.GetEnv("LOG_FORMAT", "text"),
		JSONFormat: jsonFormat,
	}
}

func loadRateLimitConfig() RateLimitConfig {
	enabled := utils.GetEnv("RATE_LIMIT_ENABLED", "true") == "true"
	requestsPerSecond, _ := strconv.ParseFloat(utils.GetEnv("RATE_LIMIT_REQUESTS_PER_SECOND", "10"), 64)
	burstSize, _ := strconv.Atoi(utils.GetEnv("RATE_LIMIT_BURST_SIZE", "20"))

	return RateLimitConfig{
		Enabled:           enabled,
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         burstSize,
	}
}

func loadUniverseConfig() UniverseConfig {
	sectorCount, _ := strconv.Atoi(utils.GetEnv("UNIVERSE_SECTOR_COUNT", "16"))
	systemsPerSector, _ := strconv.Atoi(utils.GetEnv("UNIVERSE_SYSTEMS_PER_SECTOR", "16"))
	minPlanets, _ := strconv.Atoi(utils.GetEnv("UNIVERSE_MIN_PLANETS_PER_SYSTEM", "3"))
	maxPlanets, _ := strconv.Atoi(utils.GetEnv("UNIVERSE_MAX_PLANETS_PER_SYSTEM", "12"))

	return UniverseConfig{
		SectorCount:         sectorCount,
		SystemsPerSector:    systemsPerSector,
		MinPlanetsPerSystem: minPlanets,
		MaxPlanetsPerSystem: maxPlanets,
		DefaultGalaxyName:   utils.GetEnv("UNIVERSE_DEFAULT_GALAXY_NAME", "Milky Way"),
	}
}

func loadAdminConfig() AdminConfig {
	return AdminConfig{
		Email:       utils.GetEnv("ADMIN_EMAIL", "admin@localhost"),
		Username:    utils.GetEnv("ADMIN_USERNAME", "admin"),
		DisplayName: utils.GetEnv("ADMIN_DISPLAY_NAME", "Admin"),
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
		return fmt.Errorf("SERVER_PORT is required")
	}

	if c.Database.Host == "" {
		return fmt.Errorf("DB_HOST is required")
	}

	if c.Database.Name == "" {
		return fmt.Errorf("DB_NAME is required")
	}

	if c.Server.URL == "" {
		return fmt.Errorf("SERVER_URL is required")
	}

	return nil
}

func (c *Config) GoogleOAuthConfigured() bool {
	return c.OAuth.Google.ClientID != "" && c.OAuth.Google.ClientSecret != ""
}

func (c *Config) GitHubOAuthConfigured() bool {
	return c.OAuth.GitHub.ClientID != "" && c.OAuth.GitHub.ClientSecret != ""
}

func (c *Config) DiscordOAuthConfigured() bool {
	return c.OAuth.Discord.ClientID != "" && c.OAuth.Discord.ClientSecret != ""
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
