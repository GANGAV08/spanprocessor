package main

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"go.opentelemetry.io/collector/model/pdata"

	"github.com/GANGAV08/filterspan/filterspan"
)

type spanProcessor struct {
	config           Config
	toAttributeRules []toAttributeRule
	include          filterspan.Matcher
	exclude          filterspan.Matcher
}

type toAttributeRule struct {
	re *regexp.Regexp

	attrNames []string
}

func newSpanProcessor(config Config) (*spanProcessor, error) {
	include, err := filterspan.NewMatcher(config.Include)
	if err != nil {
		return nil, err
	}
	exclude, err := filterspan.NewMatcher(config.Exclude)
	if err != nil {
		return nil, err
	}

	sp := &spanProcessor{
		config:  config,
		include: include,
		exclude: exclude,
	}

	if config.Rename.ToAttributes != nil {
		for _, pattern := range config.Rename.ToAttributes.Rules {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return nil, fmt.Errorf("invalid regexp pattern %s", pattern)
			}

			rule := toAttributeRule{
				re: re,

				attrNames: re.SubexpNames(),
			}

			sp.toAttributeRules = append(sp.toAttributeRules, rule)
		}
	}

	return sp, nil
}

func (sp *spanProcessor) processTraces(_ context.Context, td pdata.Traces) (pdata.Traces, error) {
	rss := td.ResourceSpans()
	for i := 0; i < rss.Len(); i++ {
		rs := rss.At(i)
		ilss := rs.InstrumentationLibrarySpans()
		fmt.Println("ilss result: ", ilss)
		resource := rs.Resource()
		for j := 0; j < ilss.Len(); j++ {
			ils := ilss.At(j)
			spans := ils.Spans()
			fmt.Println("spans list is: ", ilss)
			library := ils.InstrumentationLibrary()
			for k := 0; k < spans.Len(); k++ {
				s := spans.At(k)
				fmt.Println("span is __: ", ilss)
				if filterspan.SkipSpan(sp.include, sp.exclude, s, resource, library) {
					continue
				}
				sp.processFromAttributes(s)
				sp.processToAttributes(s)
			}
		}
	}
	return td, nil
}

func (sp *spanProcessor) processFromAttributes(span pdata.Span) {
	if len(sp.config.Rename.FromAttributes) == 0 {

		return
	}

	attrs := span.Attributes()
	if attrs.Len() == 0 {

		return
	}

	var sb strings.Builder
	for i, key := range sp.config.Rename.FromAttributes {
		attr, found := attrs.Get(key)

		if !found {
			return
		}

		if i > 0 && sp.config.Rename.Separator != "" {
			sb.WriteString(sp.config.Rename.Separator)
		}

		switch attr.Type() {
		case pdata.AttributeValueTypeString:
			sb.WriteString(attr.StringVal())
		case pdata.AttributeValueTypeBool:
			sb.WriteString(strconv.FormatBool(attr.BoolVal()))
		case pdata.AttributeValueTypeDouble:
			sb.WriteString(strconv.FormatFloat(attr.DoubleVal(), 'f', -1, 64))
		case pdata.AttributeValueTypeInt:
			sb.WriteString(strconv.FormatInt(attr.IntVal(), 10))
		default:
			sb.WriteString("<unknown-attribute-type>")
		}
	}
	span.SetName(sb.String())
}

func (sp *spanProcessor) processToAttributes(span pdata.Span) {
	if span.Name() == "" {

		return
	}

	if sp.config.Rename.ToAttributes == nil {

		return
	}

	for _, rule := range sp.toAttributeRules {
		re := rule.re
		oldName := span.Name()

		submatches := re.FindStringSubmatch(oldName)
		if submatches == nil {
			continue
		}

		submatchIdxPairs := re.FindStringSubmatchIndex(oldName)

		var sb strings.Builder

		var oldNameIndex = 0

		attrs := span.Attributes()

		for i := 1; i < len(submatches); i++ {
			attrs.UpsertString(rule.attrNames[i], submatches[i])

			matchStartIndex := submatchIdxPairs[i*2] // start of i'th submatch.
			sb.WriteString(oldName[oldNameIndex:matchStartIndex] + "{" + rule.attrNames[i] + "}")

			oldNameIndex = submatchIdxPairs[i*2+1] // end of i'th submatch.
		}
		if oldNameIndex < len(oldName) {

			sb.WriteString(oldName[oldNameIndex:])
		}

		span.SetName(sb.String())

		if sp.config.Rename.ToAttributes.BreakAfterMatch {

			break
		}
	}
}
