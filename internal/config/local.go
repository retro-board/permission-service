package config

import (
	"fmt"

	"github.com/caarlos0/env/v6"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
)

type UserService struct {
	Key     string `env:"USER_SERVICE_KEY"`
	Address string `env:"USER_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/user"`
}
type CompanyService struct {
	Key     string `env:"COMPANY_SERVICE_KEY"`
	Address string `env:"COMPANY_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/company"`
}
type TimerService struct {
	Key     string `env:"TIMER_SERVICE_KEY"`
	Address string `env:"TIMER_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/key"`
}
type RetroService struct {
	Key     string `env:"RETRO_SERVICE_KEY"`
	Address string `env:"RETRO_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/retro"`
}
type BillingService struct {
	Key     string `env:"BILLING_SERVICE_KEY"`
	Address string `env:"BILLING_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/billing"`
}
type KeyService struct {
	Key     string `env:"KEY_SERVICE_KEY"`
	Address string `env:"KEY_SERVICE_ADDRESS" envDefault:"https://api.retro-board.it/v1/key"`
}

type Services struct {
	UserService
	CompanyService
	TimerService
	RetroService
	BillingService
	KeyService
}

type Local struct {
	KeepLocal   bool `env:"LOCAL_ONLY" envDefault:"false"`
	Development bool `env:"DEVELOPMENT" envDefault:"false"`
	HTTPPort    int  `env:"HTTP_PORT" envDefault:"3000"`
	GRPCPort    int  `env:"GRPC_PORT" envDefault:"8001"`

	OnePasswordKey  string `env:"ONE_PASSWORD_KEY"`
	OnePasswordPath string `env:"ONE_PASSWORD_PATH"`

	Services
}

func BuildLocal(cfg *Config) error {
	local := &Local{}
	if err := env.Parse(local); err != nil {
		return err
	}
	cfg.Local = *local

	if err := BuildServiceKeys(cfg); err != nil {
		return bugLog.Errorf("failed to build service keys: %s", err.Error())
	}
	if err := BuildServiceKey(cfg); err != nil {
		return bugLog.Errorf("failed to build service key: %s", err.Error())
	}

	return nil
}

func BuildServiceKey(cfg *Config) error {
	if cfg.Local.OnePasswordKey != "" {
		return nil
	}

	onePasswordKeyData, err := cfg.getVaultSecrets(cfg.Local.OnePasswordPath)
	if err != nil {
		return err
	}

	for ik, iv := range onePasswordKeyData {
		if ik == "password" {
			cfg.Local.OnePasswordKey = iv.(string)
		}
	}
	return nil
}

func BuildServiceKeys(cfg *Config) error {
	// skip vault if development
	if cfg.Local.Development {
		return nil
	}

	vaultSecrets, err := cfg.getVaultSecrets("kv/data/retro-board/api-keys")
	if err != nil {
		return err
	}

	if vaultSecrets == nil {
		return fmt.Errorf("api keys not found in vault")
	}

	secrets, err := ParseKVSecrets(vaultSecrets)
	if err != nil {
		return err
	}

	for _, secret := range secrets {
		switch secret.Key {
		case "retro":
			cfg.Local.Services.RetroService.Key = secret.Value
		case "user":
			cfg.Local.Services.UserService.Key = secret.Value
		case "company":
			cfg.Local.Services.CompanyService.Key = secret.Value
		case "timer":
			cfg.Local.Services.TimerService.Key = secret.Value
		case "billing":
			cfg.Local.Services.BillingService.Key = secret.Value
		case "key":
			cfg.Local.Services.KeyService.Key = secret.Value
		}
	}

	return nil
}
