// Copyright 2014 Canonical Ltd.
// Licensed under the AGPLv3, see LICENCE file for details.

package charm_test

import (
	"sort"
	"strings"

	gc "gopkg.in/check.v1"

	"gopkg.in/juju/charm.v4"
)

// Keys returns a list of all defined metrics keys.
func Keys(m *charm.Metrics) []string {
	result := make([]string, 0, len(m.Metrics))

	for name := range m.Metrics {
		result = append(result, name)

	}
	sort.Strings(result)
	return result
}

type MetricsSuite struct{}

var _ = gc.Suite(&MetricsSuite{})

func (s *MetricsSuite) TestReadEmpty(c *gc.C) {
	metrics, err := charm.ReadMetrics(strings.NewReader(""))
	c.Assert(err, gc.IsNil)
	c.Assert(metrics, gc.NotNil)
}

func (s *MetricsSuite) TestReadAlmostEmpty(c *gc.C) {
	metrics, err := charm.ReadMetrics(strings.NewReader(`
metrics:
`))
	c.Assert(err, gc.IsNil)
	c.Assert(metrics, gc.NotNil)
}

func (s *MetricsSuite) TestNoDescription(c *gc.C) {
	metrics, err := charm.ReadMetrics(strings.NewReader(`
metrics:
  some-metric:
    type: gauge
`))
	c.Assert(err, gc.ErrorMatches, "invalid metrics declaration: metric \"some-metric\" lacks description")
	c.Assert(metrics, gc.IsNil)
}

func (s *MetricsSuite) TestIncorrectType(c *gc.C) {
	metrics, err := charm.ReadMetrics(strings.NewReader(`
metrics:
  some-metric:
    type: not-a-type
    description: Some description.
`))
	c.Assert(err, gc.ErrorMatches, "invalid metrics declaration: metric \"some-metric\" has unknown type \"not-a-type\"")
	c.Assert(metrics, gc.IsNil)
}

func (s *MetricsSuite) TestMultipleDefinition(c *gc.C) {
	metrics, err := charm.ReadMetrics(strings.NewReader(`
metrics:
  some-metric:
    type: gauge
    description: Some description.
  some-metric:
    type: absolute
    description: Some other description.

`))
	c.Assert(err, gc.IsNil)
	c.Assert(metrics.Metrics, gc.HasLen, 1)
	c.Assert(metrics.Metrics["some-metric"].Type, gc.Equals, charm.MetricTypeAbsolute)
}

func (s *MetricsSuite) TestValidYaml(c *gc.C) {
	metrics, err := charm.ReadMetrics(strings.NewReader(`
metrics:
  blips:
    type: absolute
    description: An absolute metric.
  blops:
    type: gauge
    description: A gauge metric.
  juju-unit-time:
    type: gauge
    description: Unit time.
`))
	c.Assert(err, gc.IsNil)
	c.Assert(metrics, gc.NotNil)
	c.Assert(Keys(metrics), gc.DeepEquals, []string{"blips", "blops", "juju-unit-time"})

	testCases := []struct {
		about string
		name  string
		value string
		err   string
	}{{
		about: "valid gauge metric",
		name:  "blops",
		value: "1",
		err:   "",
	}, {
		about: "valid absolute metric",
		name:  "blips",
		value: "0",
		err:   "",
	}, {
		about: "valid gauge metric, float value",
		name:  "blops",
		value: "0.15",
		err:   "",
	}, {
		about: "valid absolute metric, float value",
		name:  "blips",
		value: "6.015e15",
		err:   "",
	}, {
		about: "undeclared metric",
		name:  "undeclared",
		value: "6.015e15",
		err:   "metric \"undeclared\" not defined",
	}, {
		about: "invalid type for gauge metric",
		name:  "blops",
		value: "true",
		err:   "invalid value type: expected float, got \"true\"",
	}, {
		about: "metric value too large",
		name:  "blips",
		value: "1111111111111111111111111111111",
		err:   "metric value is too large",
	},
	}

	for i, t := range testCases {
		c.Logf("test %d: %s", i, t.about)
		err := metrics.ValidateMetric(t.name, t.value)
		if t.err == "" {
			c.Check(err, gc.IsNil)
		} else {
			c.Check(err, gc.ErrorMatches, t.err)
		}
	}

}
