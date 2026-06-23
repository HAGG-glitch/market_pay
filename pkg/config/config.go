package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all application configuration.
type Config struct {
	App          AppConfig          `mapstructure:"app"`
	Database     DatabaseConfig     `mapstructure:"database"`
	Redis        RedisConfig        `mapstructure:"redis"`
	JWT          JWTConfig          `mapstructure:"jwt"`
	Monime       MonimeConfig       `mapstructure:"monime"`
	USSD         USSDConfig         `mapstructure:"ussd"`
	Notifications NotificationConfig `mapstructure:"notifications"`
	CreditScore  CreditScoreConfig  `mapstructure:"credit_score"`
	Loans        LoanProductsConfig `mapstructure:"loans"`
	Repayment    RepaymentConfig    `mapstructure:"repayment"`
	Payment      PaymentConfig      `mapstructure:"payment"`
	Outbox       OutboxConfig       `mapstructure:"outbox"`
	CORS         CORSConfig         `mapstructure:"cors"`
	RateLimit    RateLimitConfig    `mapstructure:"rate_limit"`
	Logging      LoggingConfig      `mapstructure:"logging"`
}

type AppConfig struct {
	Name      string `mapstructure:"name"`
	Version   string `mapstructure:"version"`
	Env       string `mapstructure:"env"`
	Port      int    `mapstructure:"port"`
	GRPCPort  int    `mapstructure:"grpc_port"`
	PublicURL string `mapstructure:"public_url"`
	Debug     bool   `mapstructure:"debug"`
}

type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Name            string        `mapstructure:"name"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
	)
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	AccessSecret  string        `mapstructure:"access_secret"`
	RefreshSecret string        `mapstructure:"refresh_secret"`
	AccessExpiry  time.Duration `mapstructure:"access_expiry"`
	RefreshExpiry time.Duration `mapstructure:"refresh_expiry"`
}

type MonimeConfig struct {
	BaseURL          string              `mapstructure:"base_url"`
	APIKey           string              `mapstructure:"api_key"`
	WebhookSecret    string              `mapstructure:"webhook_secret"`
	RSAPrivateKeyPEM string              `mapstructure:"rsa_private_key_pem"`
	Timeout          time.Duration       `mapstructure:"timeout"`
	Payout           MonimePayoutConfig  `mapstructure:"payout"`
}

type ProviderMapping struct {
	Prefix     string `mapstructure:"prefix"`
	ProviderID string `mapstructure:"provider_id"`
}

type MonimePayoutConfig struct {
	BaseURL            string            `mapstructure:"base_url"`
	APIKey             string            `mapstructure:"api_key"`
	SpaceID            string            `mapstructure:"space_id"`
	FinancialAccountID string            `mapstructure:"financial_account_id"`
	ProviderID         string            `mapstructure:"provider_id"`
	ProviderMappings   []ProviderMapping `mapstructure:"provider_mappings"`
	Timeout            time.Duration     `mapstructure:"timeout"`
}

func (c MonimePayoutConfig) ProviderForPhone(phone string) string {
	for _, m := range c.ProviderMappings {
		if len(phone) >= len(m.Prefix) && phone[:len(m.Prefix)] == m.Prefix {
			return m.ProviderID
		}
	}
	return c.ProviderID
}

type USSDConfig struct {
	SessionTimeout     time.Duration `mapstructure:"session_timeout"`
	MaxSessionsPerHour int           `mapstructure:"max_sessions_per_hour"`
}

type NotificationConfig struct {
	SMS      SMSConfig      `mapstructure:"sms"`
	WhatsApp WhatsAppConfig `mapstructure:"whatsapp"`
	Email    EmailConfig    `mapstructure:"email"`
}

type SMSConfig struct {
	Provider string `mapstructure:"provider"`
	APIKey   string `mapstructure:"api_key"`
	Username string `mapstructure:"username"`
}

type WhatsAppConfig struct {
	Provider   string `mapstructure:"provider"`
	AccountSID string `mapstructure:"account_sid"`
	AuthToken  string `mapstructure:"auth_token"`
	From       string `mapstructure:"from"`
}

type EmailConfig struct {
	Provider string `mapstructure:"provider"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type CreditScoreConfig struct {
	TransactionVolumeWeight     int     `mapstructure:"transaction_volume_weight"`
	TransactionConsistencyWeight int     `mapstructure:"transaction_consistency_weight"`
	RepaymentHistoryWeight      int     `mapstructure:"repayment_history_weight"`
	MarketAssociationWeight     int     `mapstructure:"market_association_weight"`
	KYCCompletenessWeight       int     `mapstructure:"kyc_completeness_weight"`
	GroupBonus                  int     `mapstructure:"group_bonus"`
	MinScore                    float64 `mapstructure:"min_score"`
	AutoApproveScore            float64 `mapstructure:"auto_approve_score"`
}

type LoanProductConfig struct {
	MinAmount       float64 `mapstructure:"min_amount"`
	MaxAmount       float64 `mapstructure:"max_amount"`
	TermWeeks       int     `mapstructure:"term_weeks"`
	TermWeeksMin    int     `mapstructure:"term_weeks_min"`
	TermWeeksMax    int     `mapstructure:"term_weeks_max"`
	InterestRate    float64 `mapstructure:"interest_rate"`
	AutoApprove     bool    `mapstructure:"auto_approve"`
	DecliningBalance bool   `mapstructure:"declining_balance"`
}

type LoanProductsConfig struct {
	EmergencyAdvance LoanProductConfig `mapstructure:"emergency_advance"`
	StarterLoan      LoanProductConfig `mapstructure:"starter_loan"`
	GrowthLoan       LoanProductConfig `mapstructure:"growth_loan"`
}

type RepaymentConfig struct {
	GracePeriodDays    int     `mapstructure:"grace_period_days"`
	DefaultPenaltyRate float64 `mapstructure:"default_penalty_rate"`
}

type PaymentConfig struct {
	FeeRate float64 `mapstructure:"fee_rate"`
}

type OutboxConfig struct {
	WorkerInterval time.Duration   `mapstructure:"worker_interval"`
	MaxRetries     int             `mapstructure:"max_retries"`
	RetryIntervals []time.Duration `mapstructure:"retry_intervals"`
}

type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

type RateLimitConfig struct {
	RequestsPerMinute int `mapstructure:"requests_per_minute"`
	Burst             int `mapstructure:"burst"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
	Output string `mapstructure:"output"`
}

// Load reads configuration from file and environment variables.
func Load(path string) (*Config, error) {
	v := viper.New()

	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// Allow env overrides: APP_PORT overrides app.port
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &cfg, nil
}

// IsProduction returns true if the app is running in production mode.
func (c *Config) IsProduction() bool {
	return c.App.Env == "production"
}
