package spanprocessor

import (
	"go.opentelemetry.io/collector/config"

	"github.com/GANGAV08/filter_config/filterconfig"
)

type Config struct {
	config.ProcessorSettings `mapstructure:",squash"`
	filterconfig.MatchConfig `mapstructure:",squash"`

	Rename Name `mapstructure:"name"`
}

type Name struct {
	FromAttributes []string `mapstructure:"from_attributes"`

	Separator string `mapstructure:"separator"`

	ToAttributes *ToAttributes `mapstructure:"to_attributes"`
}

type ToAttributes struct {
	Rules []string `mapstructure:"rules"`

	BreakAfterMatch bool `mapstructure:"break_after_match"`
}

var _ config.Processor = (*Config)(nil)

func (cfg *Config) Validate() error {
	return nil
}
