package ipacl

import (
	"net/netip"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/phuslu/iploc"
)

type Config struct {
	// when returned true, our middleware is skipped
	Filter func(c *fiber.Ctx) bool

	// function to run when there is error decoding jwt
	Unauthorized fiber.Handler

	// function to decode our jwt token
	CheckIp func(c *fiber.Ctx, config Config) bool

	// set jwt expiry in seconds
	Country string
}

var ConfigDefault = Config{
	Filter:       Filter,
	CheckIp:      nil,
	Unauthorized: nil,
	Country:      "EE",
}

// var config *Config

func configDefault(conf ...Config) Config {
	// Return default config if nothing provided
	if len(conf) < 1 {
		return ConfigDefault
	}

	// Override default config
	cfg := conf[0]

	// Set default values if not passed
	if cfg.Filter == nil {
		cfg.Filter = ConfigDefault.Filter
	}

	// Set default expiry if not passed
	if cfg.Country == "" {
		cfg.Country = ConfigDefault.Country
	}

	// this is the main jwt decode function of our middleware
	if cfg.CheckIp == nil {
		// Set default Decode function if not passed
		cfg.CheckIp = CheckIp
	}

	// Set default Unauthorized if not passed
	if cfg.Unauthorized == nil {
		cfg.Unauthorized = func(c *fiber.Ctx) error {
			return c.SendStatus(fiber.StatusGatewayTimeout)
		}
	}
	return cfg
}

func CheckIp(c *fiber.Ctx, config Config) bool {
	return iploc.IPCountry(netip.MustParseAddr(c.IP())) == config.Country
}

func Filter(c *fiber.Ctx) bool {
	return c.IP() == "127.0.0.1"
}

func New(c Config) fiber.Handler {

	// For setting default config
	cfg := configDefault(c)

	return func(c *fiber.Ctx) error {
		// Don't execute middleware if Filter returns true
		if cfg.Filter != nil && cfg.Filter(c) {
			log.Info("IP is allowed by filter: ", c.IP())
			return c.Next()
		}

		if !CheckIp(c, cfg) {
			log.Info(c.IP(), "is not from allowed IP country: ", iploc.IPCountry(netip.MustParseAddr(c.IP())))
			return cfg.Unauthorized(c)
		}
		log.Info("IP ALLOWED: ", c.IP(), " from country: ", iploc.IPCountry(netip.MustParseAddr(c.IP())))
		return c.Next()

	}
}
