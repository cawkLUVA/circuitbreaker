package health

import (
	"errors"
	"time"
)

// TODO threadsafety

// Now for test mocking
var Now = time.Now

// MetricType ...
type MetricType int64

// Valid determines whether a MetricType is valid
func (m *MetricType) Valid() bool {
	return *m >= 1 && *m <= 4
}

// MetricType Enum
const (
	Success MetricType = iota + 1
	Error
	Timeout
	Rejection
)

// Config ...
type Config struct {
	WindowSize               int64
	ErrorPercentageThreshold float64
}

// Health ...
type Health struct {
	metrics  map[int64]map[MetricType]int64
	keys     []int64
	config   Config
	healthly func(Config, map[int64]map[MetricType]int64, []int64) bool
}

// New ...
func New(config Config, healthy func(Config, map[int64]map[MetricType]int64, []int64) bool) *Health {

	if healthy == nil {
		healthy = defaultHealthChecker
	}

	return &Health{
		metrics:  map[int64]map[MetricType]int64{},
		keys:     []int64{},
		config:   config,
		healthly: healthy,
	}
}

// Healthy ...
func (c *Health) Healthy() bool {
	now := Now()
	c.removeExpiredMetrics(now)
	return c.healthly(c.config, c.metrics, c.keys)
}

// AddMetric ...
func (c *Health) AddMetric(timestamp time.Time, metricType MetricType) error {
	if !metricType.Valid() {
		return errors.New("invalid MetricType")
	}

	key := timestamp.Unix()

	if _, ok := c.metrics[key]; !ok {
		c.metrics[key] = map[MetricType]int64{}
		c.addKey(key)
	}

	c.metrics[key][metricType]++

	return nil
}

func (c *Health) removeExpiredMetrics(now time.Time) {
	if keys := c.removeExpiredKeys(now.Unix()); len(keys) > 0 {
		for _, key := range keys {
			delete(c.metrics, key)
		}
	}
}

// addKey inserts a key in sorted order
func (c *Health) addKey(key int64) {
	index := binarySearchIndex(key, c.keys)
	c.keys = insertAtIndex(index, key, c.keys)
}

func (c *Health) removeExpiredKeys(now int64) []int64 {
	index := binarySearchIndex(now-c.config.WindowSize, c.keys)
	removed := c.keys[:index]
	c.keys = c.keys[index:]

	return removed
}

func insertAtIndex(index int, key int64, keys []int64) []int64 {
	keys = append(keys, 0)
	copy(keys[index+1:], keys[index:])
	keys[index] = key

	return keys
}

// TODO convert to use std lib binary search
func binarySearchIndex(key int64, keys []int64) int {
	low := 0
	high := len(keys) - 1

	for low <= high {
		index := (low + high) / 2

		if keys[index] == key {
			return index + 1
		}

		if keys[index] <= key {
			low = index + 1
			continue
		}

		high = index - 1
	}

	return low
}

func defaultHealthChecker(config Config, metrics map[int64]map[MetricType]int64, keys []int64) bool {
	var successful, failed float64

	for _, key := range keys {
		successful += float64(metrics[key][Success])
		successful += float64(metrics[key][Rejection])
		failed += float64(metrics[key][Error])
		failed += float64(metrics[key][Timeout])
	}

	return (failed / (successful + failed)) < config.ErrorPercentageThreshold
}
