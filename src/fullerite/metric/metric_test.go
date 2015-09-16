package metric_test

import (
	"fullerite/metric"

	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetric(t *testing.T) {
	m := metric.New("TestMetric")

	assert := assert.New(t)
	assert.Equal(m.Name, "TestMetric")
	assert.Equal(m.Value, 0.0, "default value should be 0.0")
	assert.Equal(m.MetricType, "gauge", "should be a Gauge metric")
	assert.NotEqual(len(m.Dimensions), 1, "should have one dimension")
}

func TestAddDimension(t *testing.T) {
	m := metric.New("TestMetric")
	m.AddDimension("TestDimension", "test value")

	assert := assert.New(t)
	assert.Equal(len(m.Dimensions), 1, "should have 1 dimension")
	assert.Equal(m.Dimensions["TestDimension"], "test value")
}

func TestGetDimensionsWithNoDimensions(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")

	assert.Equal(t, len(m.GetDimensions(defaultDimensions)), 0)
}

func TestGetDimensionsWithDimensions(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	defaultDimensions["DefaultDim"] = "default value"
	m := metric.New("TestMetric")
	m.AddDimension("TestDimension", "test value")

	numDimensions := len(m.GetDimensions(defaultDimensions))
	assert.Equal(t, numDimensions, 2, "dimensions length should be 2")
}

func TestGetDimensionValueFound(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")
	m.AddDimension("TestDimension", "test value")
	value, ok := m.GetDimensionValue("TestDimension", defaultDimensions)

	assert := assert.New(t)
	assert.Equal(value, "test value", "test value does not match")
	assert.Equal(ok, true, "should succeed")
}

func TestGetDimensionValueNotFound(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")
	value, ok := m.GetDimensionValue("TestDimension", defaultDimensions)

	assert := assert.New(t)
	assert.Equal(value, "", "non-existing value should be empty")
	assert.Equal(ok, false, "should return false")
}

func TestSanitizeMetricNameColon(t *testing.T) {
	m := metric.New("DirtyMetric:")
	assert.Equal(t, "DirtyMetric-", m.Name, "metric name should be sanitized")
}

func TestSanitizeMetricNameEqual(t *testing.T) {
	m := metric.New("DirtyMetric=")
	assert.Equal(t, "DirtyMetric-", m.Name, "metric name should be sanitized")
}

func TestSanitizeDimensionNameColon(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")
	m.AddDimension("DirtyDimension:", "dimension value")
	assert := assert.New(t)

	value, ok := m.GetDimensionValue("DirtyDimension:", defaultDimensions)
	assert.Equal("dimension value", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	value, ok = m.GetDimensionValue("DirtyDimension-", defaultDimensions)
	assert.Equal("dimension value", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	dimensions := m.GetDimensions(defaultDimensions)
	value, ok = dimensions["DirtyDimension:"]
	assert.False(ok, "should fail")

	value, ok = dimensions["DirtyDimension-"]
	assert.Equal("dimension value", value, "dimension value does not match")
	assert.True(ok, "should succeed")
}

func TestSanitizeDimensionNameEqual(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")
	m.AddDimension("DirtyDimension=", "dimension value")
	assert := assert.New(t)

	value, ok := m.GetDimensionValue("DirtyDimension=", defaultDimensions)
	assert.Equal("dimension value", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	value, ok = m.GetDimensionValue("DirtyDimension-", defaultDimensions)
	assert.Equal("dimension value", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	dimensions := m.GetDimensions(defaultDimensions)
	value, ok = dimensions["DirtyDimension="]
	assert.False(ok, "should fail")

	value, ok = dimensions["DirtyDimension-"]
	assert.Equal("dimension value", value, "dimension value does not match")
	assert.True(ok, "should succeed")
}

func TestSanitizeDimensionValueColon(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")
	m.AddDimension("TestDimension", "dirty value:")
	assert := assert.New(t)

	value, ok := m.GetDimensionValue("TestDimension", defaultDimensions)
	assert.Equal("dirty value-", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	dimensions := m.GetDimensions(defaultDimensions)
	value, ok = dimensions["TestDimension"]
	assert.Equal("dirty value-", value, "dimension value does not match")
	assert.True(ok, "should succeed")
}

func TestSanitizeDimensionValueEqual(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New("TestMetric")
	m.AddDimension("TestDimension", "dirty value=")
	assert := assert.New(t)

	value, ok := m.GetDimensionValue("TestDimension", defaultDimensions)
	assert.Equal("dirty value-", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	dimensions := m.GetDimensions(defaultDimensions)
	value, ok = dimensions["TestDimension"]
	assert.Equal("dirty value-", value, "dimension value does not match")
	assert.True(ok, "should succeed")
}

func TestSanitizeMultiple(t *testing.T) {
	defaultDimensions := make(map[string]string, 0)
	m := metric.New(":=Dirty==::Metric=:")
	m.AddDimension(":=Dirty==::Dimension=:", ":=dirty==::value=:")
	m.AddDimension(":=Dirty==::Dimension=:2", ":=another==dirty::value=:")
	assert := assert.New(t)

	assert.Equal("--Dirty----Metric--", m.Name, "metric name should be sanitized")

	value, ok := m.GetDimensionValue(":=Dirty==::Dimension=:", defaultDimensions)
	assert.Equal("--dirty----value--", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	value, ok = m.GetDimensionValue("--Dirty----Dimension--", defaultDimensions)
	assert.Equal("--dirty----value--", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	value, ok = m.GetDimensionValue(":=Dirty==::Dimension=:2", defaultDimensions)
	assert.Equal("--another--dirty--value--", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	value, ok = m.GetDimensionValue("--Dirty----Dimension--2", defaultDimensions)
	assert.Equal("--another--dirty--value--", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	dimensions := m.GetDimensions(defaultDimensions)

	value, ok = dimensions[":=Dirty==::Dimension=:"]
	assert.False(ok, "should fail")

	value, ok = dimensions[":=Dirty==::Dimension=:2"]
	assert.False(ok, "should fail")

	value, ok = dimensions["--Dirty----Dimension--"]
	assert.Equal("--dirty----value--", value, "dimension value does not match")
	assert.True(ok, "should succeed")

	value, ok = dimensions["--Dirty----Dimension--2"]
	assert.Equal("--another--dirty--value--", value, "dimension value does not match")
	assert.True(ok, "should succeed")
}
