package main

import (
	"path"
	"testing"

	"github.com/GANGAV08/filter_config/filterconfig"
	"github.com/GANGAV08/filterset/filterset"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/component/componenttest"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/config/configtest"
)

func TestLoadConfig(t *testing.T) {
	factories, err := componenttest.NopFactories()
	assert.NoError(t, err)

	factory := NewFactory()
	factories.Processors[typeStr] = factory

	cfg, err := configtest.LoadConfigAndValidate(path.Join(".", "testdata", "config.yaml"), factories)

	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	p0 := cfg.Processors[config.NewComponentIDWithName("span", "custom")]
	assert.Equal(t, p0, &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName("span", "custom")),
		Rename: Name{
			FromAttributes: []string{"db.svc", "operation", "id"},
			Separator:      "::",
		},
	})

	p1 := cfg.Processors[config.NewComponentIDWithName("span", "no-separator")]
	assert.Equal(t, p1, &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName("span", "no-separator")),
		Rename: Name{
			FromAttributes: []string{"db.svc", "operation", "id"},
			Separator:      "",
		},
	})

	p2 := cfg.Processors[config.NewComponentIDWithName("span", "to_attributes")]
	assert.Equal(t, p2, &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName("span", "to_attributes")),
		Rename: Name{
			ToAttributes: &ToAttributes{
				Rules: []string{`^\/api\/v1\/document\/(?P<documentId>.*)\/update$`},
			},
		},
	})

	p3 := cfg.Processors[config.NewComponentIDWithName("span", "includeexclude")]
	assert.Equal(t, p3, &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentIDWithName("span", "includeexclude")),
		MatchConfig: filterconfig.MatchConfig{
			Include: &filterconfig.MatchProperties{
				Config:    *createMatchConfig(filterset.Regexp),
				Services:  []string{`banks`},
				SpanNames: []string{"^(.*?)/(.*?)$"},
			},
			Exclude: &filterconfig.MatchProperties{
				Config:    *createMatchConfig(filterset.Strict),
				SpanNames: []string{`donot/change`},
			},
		},
		Rename: Name{
			ToAttributes: &ToAttributes{
				Rules: []string{`(?P<operation_website>.*?)$`},
			},
		},
	})
}

func createMatchConfig(matchType filterset.MatchType) *filterset.Config {
	return &filterset.Config{
		MatchType: matchType,
	}
}
