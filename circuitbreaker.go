package circuitbreaker

import (
	"context"
	"errors"
	"time"

	"circuitbreaker/internal/health"
)

// Health ...
type Health interface {
	Healthy() bool
	AddMetric(timestamp time.Time, metricType health.MetricType) error
}

// CircuitOpenError ...
type CircuitOpenError struct {
}

func (e *CircuitOpenError) Error() string {
	return "circuit is open"
}

// Status ...
type Status int64

// circuit breaker states
const (
	Open Status = iota + 1
	HalfOpen
	Closed
)

// Valid determines whether the value of a State is valid
func (s *Status) Valid() bool {
	return *s >= 1 && *s <= 3
}

// State ...
type State struct {
	status  Status
	updated time.Time
}

// Config ...
type Config struct {
	// the length of time in milliseconds to wait before retrying when the circuit is open
	SleepWindowMillisenconds int64
}

// CircuitBreaker ...
type CircuitBreaker struct {
	state     State
	config    Config
	health    Health
	stateChan chan State
	fallback  func() (interface{}, error)
}

// New ...
func New(config Config, health Health, ch chan State, fallback func() (interface{}, error)) *CircuitBreaker {

	if fallback == nil {
		fallback = defaultFallback
	}

	return &CircuitBreaker{
		config:   config,
		health:   health,
		fallback: fallback,
		state: State{
			status:  Closed,
			updated: time.Now(),
		},
		stateChan: ch,
	}
}

// DoWithContext ...
func (c *CircuitBreaker) DoWithContext(ctx context.Context, operation func() (interface{}, error)) (interface{}, error) {

	now := time.Now()

	// fail immediately and call fallback
	if c.Status() == Open && nanoToMilli(now.UnixNano()-c.state.updated.UnixNano()) < c.config.SleepWindowMillisenconds {
		return c.fallback()
	}

	// the sleep window has elapsed
	if c.Status() == Open {
		c.SetStatus(HalfOpen)
	}

	// if the service is now unhealthy, set the status to Open and call the fallback. HalfOpen status is exempt
	if !c.health.Healthy() && c.Status() == Closed {
		c.SetStatus(Open)
		return c.fallback()
	}

	result, err := operation()
	if err != nil {
		c.health.AddMetric(now, health.Error)
		return result, err
	}

	c.health.AddMetric(now, health.Success)
	return result, nil
}

// Status ...
func (c *CircuitBreaker) Status() Status {
	return c.state.status
}

// SetStatus ...
func (c *CircuitBreaker) SetStatus(status Status) error {
	if !status.Valid() {
		return errors.New("invlaid status")
	}
	if status == c.Status() {
		return nil
	}

	c.state.updated = time.Now()
	c.state.status = status

	// send the new state to the channel
	if c.stateChan != nil {
		go func() {
			c.stateChan <- c.state
		}()
	}

	return nil
}

func nanoToMilli(nano int64) int64 {
	return nano / 1e6
}

func defaultFallback() (interface{}, error) {
	return nil, &CircuitOpenError{}
}

// TODO  retry with default and custom backoff
// TODO  Context Cancellation
