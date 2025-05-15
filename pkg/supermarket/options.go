package supermarket

import "time"

type Config struct {
	UserAgent    string
	AppVersion   string
	RefreshToken string
	ClientID     string
	ApiKey       string
	Timeout      time.Duration
	Debug        bool
	StoreID      string
}

type Option func(*Config)

// WithCredentials sets the client id and refresh token.
func WithCredentials(clientId, refreshToken string) Option {
	return func(c *Config) {
		c.ClientID = clientId
		c.RefreshToken = refreshToken
	}
}

// WithApiKey sets the API key.
func WithApiKey(apiKey string) Option { return func(c *Config) { c.ApiKey = apiKey } }

// WithStoreID sets the store id for promotions.
func WithStoreID(storeID string) Option { return func(c *Config) { c.StoreID = storeID } }

// WithUserAgent sets the user agent string.
func WithUserAgent(userAgent string) Option { return func(c *Config) { c.UserAgent = userAgent } }

// WithAppVersion sets the app version to emulate.
func WithAppVersion(appVersion string) Option { return func(c *Config) { c.AppVersion = appVersion } }

// WithTimeout sets the request timeout in seconds.
func WithTimeout(timeout time.Duration) Option { return func(c *Config) { c.Timeout = timeout } }

// WithDebug enables debug logging.
func WithDebug(debug bool) Option { return func(c *Config) { c.Debug = debug } }
