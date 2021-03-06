// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package charm

import (
	"fmt"
	"io"
	"io/ioutil"
	"strconv"

	goyaml "gopkg.in/yaml.v1"
)

// MetricType is used to identify metric types supported by juju.
type MetricType string

const (
	// Supported metric types.
	MetricTypeGauge    MetricType = "gauge"
	MetricTypeAbsolute MetricType = "absolute"
)

// validateValue checks if the supplied metric value fits the requirements
// of its expected type.
func (m MetricType) validateValue(value string) error {
	switch m {
	case MetricTypeGauge, MetricTypeAbsolute:
		// The largest number of digits that can be returned by strconv.FormatFloat is 24, so
		// choose an arbitrary limit somewhat higher than that.
		if len(value) > 30 {
			return fmt.Errorf("metric value is too large")
		}
		_, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return fmt.Errorf("invalid value type: expected float, got %q", value)
		}
	default:
		return fmt.Errorf("unknown metric type %q", m)
	}
	return nil
}

// Metric represents a single metric definition
type Metric struct {
	Type        MetricType
	Description string
}

// Metrics contains the metrics declarations encoded in the metrics.yaml
// file.
type Metrics struct {
	Metrics map[string]Metric
}

// ReadMetrics reads a MetricsDeclaration in YAML format.
func ReadMetrics(r io.Reader) (*Metrics, error) {
	data, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var metrics Metrics
	if err := goyaml.Unmarshal(data, &metrics); err != nil {
		return nil, err
	}
	if metrics.Metrics == nil {
		return &metrics, nil
	}
	for name, metric := range metrics.Metrics {
		switch metric.Type {
		case MetricTypeGauge, MetricTypeAbsolute:
		default:
			return nil, fmt.Errorf("invalid metrics declaration: metric %q has unknown type %q", name, metric.Type)
		}
		if metric.Description == "" {
			return nil, fmt.Errorf("invalid metrics declaration: metric %q lacks description", name)
		}
	}
	return &metrics, nil
}

// ValidateMetric validates the supplied metric name and value against the loaded
// metric definitions.
func (m Metrics) ValidateMetric(name, value string) error {
	metric, exists := m.Metrics[name]
	if !exists {
		return fmt.Errorf("metric %q not defined", name)
	}
	return metric.Type.validateValue(value)
}
