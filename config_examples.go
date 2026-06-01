package main

import (
	"context"
	"fmt"

	"marketpay/internal/store"
	"marketpay/internal/ussd"
)

// ConfigExample shows different configuration options for the MarketPay USSD service

// Config represents the service configuration
type Config struct {
	Port           int
	LogLevel       string
	StateStore     ussd.StateStore
	RequestTimeout int
}

// NewConfig creates a new configuration with default values
func NewConfig() *Config {
	return &Config{
		Port:           8080,
		LogLevel:       "info",
		StateStore:     store.NewInMemoryStateStore(),
		RequestTimeout: 30,
	}
}

// ConfigExample1SimpleInMemory shows a simple in-memory configuration
func ConfigExample1SimpleInMemory() {
	fmt.Println("Example 1: Simple In-Memory Configuration")
	fmt.Println("==========================================\n")

	config := NewConfig()
	fmt.Printf("Port: %d\n", config.Port)
	fmt.Printf("Log Level: %s\n", config.LogLevel)
	fmt.Printf("State Store: In-Memory\n")
	fmt.Printf("Request Timeout: %d seconds\n\n", config.RequestTimeout)

	// Create flow service
	flowService := ussd.NewMarketPayFlowService(config.StateStore)
	ctx := context.Background()

	// Simulate a simple flow
	result, _ := flowService.Advance(ctx, ussd.AdvanceFlowInput{
		SessionID:   "test-001",
		CurrentPage: ussd.PageSelectService,
		Values:      map[string]string{},
	})

	fmt.Printf("Service started successfully with message:\n%s\n\n", result.Message)
}

// ConfigExample2MultipleStores shows how to support multiple store types
func ConfigExample2MultipleStores() {
	fmt.Println("Example 2: Multiple Store Implementations")
	fmt.Println("=========================================\n")

	configs := map[string]*Config{
		"In-Memory": {
			Port:       8080,
			LogLevel:   "info",
			StateStore: store.NewInMemoryStateStore(),
		},
		// PostgreSQL store would be implemented similar to InMemoryStateStore
		// "PostgreSQL": {
		//     Port:       8080,
		//     LogLevel:   "info",
		//     StateStore: store.NewPostgresStateStore(dbConn),
		// },
		// Redis store would be implemented similar to InMemoryStateStore
		// "Redis": {
		//     Port:       8080,
		//     LogLevel:   "info",
		//     StateStore: store.NewRedisStateStore(redisClient),
		// },
	}

	for name, config := range configs {
		fmt.Printf("%s Store Configuration:\n", name)
		fmt.Printf("  - Port: %d\n", config.Port)
		fmt.Printf("  - Log Level: %s\n\n", config.LogLevel)
	}
}

// ConfigExample3EnvironmentBased shows how to configure from environment
func ConfigExample3EnvironmentBased() {
	fmt.Println("Example 3: Environment-Based Configuration")
	fmt.Println("==========================================\n")

	// In a real application, these would come from environment variables
	configs := map[string]string{
		"PORT":            "8080",
		"LOG_LEVEL":       "debug",
		"STATE_STORE":     "memory",
		"REQUEST_TIMEOUT": "30",
		"DB_HOST":         "localhost",
		"DB_PORT":         "5432",
		"REDIS_URL":       "redis://localhost:6379",
	}

	fmt.Println("Environment Variables:")
	for key, value := range configs {
		fmt.Printf("  %s=%s\n", key, value)
	}
	fmt.Println()
}

// ConfigExample4ServiceOptions shows different service configurations
func ConfigExample4ServiceOptions() {
	fmt.Println("Example 4: Service Configuration Options")
	fmt.Println("========================================\n")

	serviceConfigs := []struct {
		name    string
		config  *Config
		purpose string
	}{
		{
			name: "Development",
			config: &Config{
				Port:           8080,
				LogLevel:       "debug",
				StateStore:     store.NewInMemoryStateStore(),
				RequestTimeout: 30,
			},
			purpose: "Local development with verbose logging",
		},
		{
			name: "Testing",
			config: &Config{
				Port:           8081,
				LogLevel:       "warn",
				StateStore:     store.NewInMemoryStateStore(),
				RequestTimeout: 10,
			},
			purpose: "Testing environment with short timeout",
		},
		{
			name: "Production",
			config: &Config{
				Port:           8080,
				LogLevel:       "error",
				StateStore:     store.NewInMemoryStateStore(), // Would be PostgreSQL or Redis
				RequestTimeout: 15,
			},
			purpose: "Production with persistent state store",
		},
	}

	for _, sc := range serviceConfigs {
		fmt.Printf("%s Configuration:\n", sc.name)
		fmt.Printf("  Purpose: %s\n", sc.purpose)
		fmt.Printf("  Port: %d\n", sc.config.Port)
		fmt.Printf("  Log Level: %s\n", sc.config.LogLevel)
		fmt.Printf("  Request Timeout: %d seconds\n\n", sc.config.RequestTimeout)
	}
}

// ConfigExample5FeatureFlags shows how to add feature flags
func ConfigExample5FeatureFlags() {
	fmt.Println("Example 5: Feature Flags Configuration")
	fmt.Println("======================================\n")

	type FeatureFlags struct {
		EnableLoanFeature         bool
		EnablePaymentVerification bool
		EnableSMSReceipts         bool
		EnableTransactionHistory  bool
		EnableBalanceCache        bool
	}

	features := map[string]FeatureFlags{
		"Development": {
			EnableLoanFeature:         true,
			EnablePaymentVerification: true,
			EnableSMSReceipts:         true,
			EnableTransactionHistory:  true,
			EnableBalanceCache:        true,
		},
		"Production": {
			EnableLoanFeature:         true,
			EnablePaymentVerification: true,
			EnableSMSReceipts:         true,
			EnableTransactionHistory:  true,
			EnableBalanceCache:        true,
		},
		"Beta": {
			EnableLoanFeature:         false,
			EnablePaymentVerification: true,
			EnableSMSReceipts:         true,
			EnableTransactionHistory:  true,
			EnableBalanceCache:        false,
		},
	}

	for env, flags := range features {
		fmt.Printf("%s Feature Flags:\n", env)
		fmt.Printf("  Loan Feature: %v\n", flags.EnableLoanFeature)
		fmt.Printf("  Payment Verification: %v\n", flags.EnablePaymentVerification)
		fmt.Printf("  SMS Receipts: %v\n", flags.EnableSMSReceipts)
		fmt.Printf("  Transaction History: %v\n", flags.EnableTransactionHistory)
		fmt.Printf("  Balance Cache: %v\n\n", flags.EnableBalanceCache)
	}
}

// ConfigExample6RateLimiting shows rate limiting configuration
func ConfigExample6RateLimiting() {
	fmt.Println("Example 6: Rate Limiting Configuration")
	fmt.Println("======================================\n")

	type RateLimitConfig struct {
		Enabled           bool
		RequestsPerSecond int
		BurstSize         int
		BlockDuration     int // seconds
	}

	configs := map[string]RateLimitConfig{
		"Normal": {
			Enabled:           true,
			RequestsPerSecond: 100,
			BurstSize:         200,
			BlockDuration:     60,
		},
		"Strict": {
			Enabled:           true,
			RequestsPerSecond: 10,
			BurstSize:         20,
			BlockDuration:     300,
		},
		"Disabled": {
			Enabled: false,
		},
	}

	for name, config := range configs {
		fmt.Printf("%s Rate Limiting:\n", name)
		fmt.Printf("  Enabled: %v\n", config.Enabled)
		if config.Enabled {
			fmt.Printf("  Requests/Second: %d\n", config.RequestsPerSecond)
			fmt.Printf("  Burst Size: %d\n", config.BurstSize)
			fmt.Printf("  Block Duration: %d seconds\n", config.BlockDuration)
		}
		fmt.Println()
	}
}

// ConfigExample7APIVersioning shows how to manage API versions
func ConfigExample7APIVersioning() {
	fmt.Println("Example 7: API Versioning Configuration")
	fmt.Println("=======================================\n")

	type APIVersion struct {
		Version        string
		Endpoints      []string
		Deprecated     bool
		SupportedUntil string
	}

	versions := []APIVersion{
		{
			Version: "v1",
			Endpoints: []string{
				"POST /api/ussd/advance",
				"GET /health",
			},
			Deprecated:     true,
			SupportedUntil: "2026-12-31",
		},
		{
			Version: "v2",
			Endpoints: []string{
				"POST /api/v2/ussd/advance",
				"POST /api/v2/ussd/validate",
				"GET /api/v2/health",
				"GET /api/v2/sessions",
			},
			Deprecated:     false,
			SupportedUntil: "2027-12-31",
		},
	}

	for _, v := range versions {
		fmt.Printf("API %s:\n", v.Version)
		fmt.Printf("  Deprecated: %v\n", v.Deprecated)
		fmt.Printf("  Supported Until: %s\n", v.SupportedUntil)
		fmt.Printf("  Endpoints:\n")
		for _, endpoint := range v.Endpoints {
			fmt.Printf("    - %s\n", endpoint)
		}
		fmt.Println()
	}
}
