package config

import (
	"context"

	"github.com/owncloud/ocis/v2/ocis-pkg/shared"
)

// Config defines the root config structure
type Config struct {
	Commons *shared.Commons `yaml:"-"` // don't use this directly as configuration for a service
	Service Service         `yaml:"-"`
	Tracing *Tracing        `yaml:"tracing"`
	Log     *Log            `yaml:"log"`
	Debug   Debug           `yaml:"debug"`

	GRPC GRPCConfig `yaml:"grpc"`

	TokenManager *TokenManager `yaml:"token_manager"`
	Reva         *shared.Reva  `yaml:"reva"`

	SkipUserGroupsInToken bool `yaml:"skip_user_groups_in_token" env:"AUTH_APP_SKIP_USER_GROUPS_IN_TOKEN" desc:"Disables the encoding of the user's group memberships in the access token. This reduces the token size, especially when users are members of a large number of groups." introductionVersion:"%%NEXT%%"`

	MachineAuthAPIKey string `yaml:"machine_auth_api_key" env:"OCIS_MACHINE_AUTH_API_KEY;AUTH_APP_MACHINE_AUTH_API_KEY" desc:"The machine auth API key used to validate internal requests necessary to access resources from other services." introductionVersion:"%%NEXT%%"`

	Supervised bool            `yaml:"-"`
	Context    context.Context `yaml:"-"`
}

// Log defines the loging configuration
type Log struct {
	Level  string `yaml:"level" env:"OCIS_LOG_LEVEL;AUTH_APP_LOG_LEVEL" desc:"The log level. Valid values are: 'panic', 'fatal', 'error', 'warn', 'info', 'debug', 'trace'." introductionVersion:"%%NEXT%%"`
	Pretty bool   `yaml:"pretty" env:"OCIS_LOG_PRETTY;AUTH_APP_LOG_PRETTY" desc:"Activates pretty log output." introductionVersion:"%%NEXT%%"`
	Color  bool   `yaml:"color" env:"OCIS_LOG_COLOR;AUTH_APP_LOG_COLOR" desc:"Activates colorized log output." introductionVersion:"%%NEXT%%"`
	File   string `yaml:"file" env:"OCIS_LOG_FILE;AUTH_APP_LOG_FILE" desc:"The path to the log file. Activates logging to this file if set." introductionVersion:"%%NEXT%%"`
}

// Service defines the service configuration
type Service struct {
	Name string `yaml:"-"`
}

// Debug defines the debug configuration
type Debug struct {
	Addr   string `yaml:"addr" env:"AUTH_APP_DEBUG_ADDR" desc:"Bind address of the debug server, where metrics, health, config and debug endpoints will be exposed." introductionVersion:"%%NEXT%%"`
	Token  string `yaml:"token" env:"AUTH_APP_DEBUG_TOKEN" desc:"Token to secure the metrics endpoint." introductionVersion:"%%NEXT%%"`
	Pprof  bool   `yaml:"pprof" env:"AUTH_APP_DEBUG_PPROF" desc:"Enables pprof, which can be used for profiling." introductionVersion:"%%NEXT%%"`
	Zpages bool   `yaml:"zpages" env:"AUTH_APP_DEBUG_ZPAGES" desc:"Enables zpages, which can  be used for collecting and viewing traces in-memory." introductionVersion:"%%NEXT%%"`
}

// GRPCConfig defines the GRPC configuration
type GRPCConfig struct {
	Addr      string                 `yaml:"addr" env:"AUTH_APP_GRPC_ADDR" desc:"The bind address of the GRPC service." introductionVersion:"%%NEXT%%"`
	TLS       *shared.GRPCServiceTLS `yaml:"tls"`
	Namespace string                 `yaml:"-"`
	Protocol  string                 `yaml:"protocol" env:"AUTH_APP_GRPC_PROTOCOL" desc:"The transport protocol of the GRPC service." introductionVersion:"%%NEXT%%"`
}
