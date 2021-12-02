package main

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	typeStr = "span"
)

var processorCapabilities = consumer.Capabilities{MutatesData: true}

var errMissingRequiredField = errors.New("error creating \"span\" processor: either \"from_attributes\" or \"to_attributes\" must be specified in \"name:\"")

func NewFactory() component.ProcessorFactory {
	return processorhelper.NewFactory(
		typeStr,
		createDefaultConfig,
		processorhelper.WithTraces(createTracesProcessor))
}

func createDefaultConfig() config.Processor {
	return &Config{
		ProcessorSettings: config.NewProcessorSettings(config.NewComponentID(typeStr)),
	}
}

func createTracesProcessor(
	_ context.Context,
	_ component.ProcessorCreateSettings,
	cfg config.Processor,
	nextConsumer consumer.Traces,
) (component.TracesProcessor, error) {

	oCfg := cfg.(*Config)
	if len(oCfg.Rename.FromAttributes) == 0 &&
		(oCfg.Rename.ToAttributes == nil || len(oCfg.Rename.ToAttributes.Rules) == 0) {
		return nil, errMissingRequiredField
	}

	sp, err := newSpanProcessor(*oCfg)
	if err != nil {
		return nil, err
	}
	return processorhelper.NewTracesProcessor(
		cfg,
		nextConsumer,
		sp.processTraces,
		processorhelper.WithCapabilities(processorCapabilities))
}
