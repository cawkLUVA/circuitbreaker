package circuitbreaker

import (
	"circuitbreaker/internal/health"
	"context"
	"reflect"
	"testing"
	"time"
)

type HealthMock struct {
	err      error
	healthly bool
}

func (h *HealthMock) Healthy() bool {
	return h.healthly
}

func (h *HealthMock) AddMetric(timestamp time.Time, metricType health.MetricType) error {
	return h.err
}

func TestCircuitBreaker_DoWithContext(t *testing.T) {
	type fields struct {
		state     State
		config    Config
		stateChan chan State
		fallback  func() (interface{}, error)
		healthy   func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool
	}
	type args struct {
		ctx       context.Context
		operation func() (interface{}, error)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    interface{}
		wantErr bool
	}{
		{
			name: "successfully executes operation given a healthy system and closed circuit",
			fields: fields{
				state: State{
					status: Closed,
				},
				config: Config{
					SleepWindowMillisenconds: 1000,
				},
				healthy: func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool {
					return true
				},
				stateChan: nil,
				fallback:  nil,
			},
			args: args{
				ctx: context.Background(),
				operation: func() (interface{}, error) {
					return 100, nil
				},
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "successfully calls fallback given an open circuit and non expired sleep window",
			fields: fields{
				state: State{
					status:  Open,
					updated: time.Now(),
				},
				config: Config{
					SleepWindowMillisenconds: 100000,
				},
				stateChan: nil,
				healthy: func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool {
					return true
				},
				fallback: func() (interface{}, error) {
					return 5, nil
				},
			},
			args: args{
				ctx: context.Background(),
				operation: func() (interface{}, error) {
					return 100, nil
				},
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "successfully executes operation given an open circuit and expired sleep window",
			fields: fields{
				state: State{
					status:  Open,
					updated: time.Now().Add(-1 * time.Minute),
				},
				config: Config{
					SleepWindowMillisenconds: 1000,
				},
				stateChan: nil,
				healthy: func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool {
					return true
				},
				fallback: func() (interface{}, error) {
					return 5, nil
				},
			},
			args: args{
				ctx: context.Background(),
				operation: func() (interface{}, error) {
					return 100, nil
				},
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "successfully executes operation given an open circuit and expired sleep window",
			fields: fields{
				state: State{
					status:  Open,
					updated: time.Now().Add(-1 * time.Minute),
				},
				config: Config{
					SleepWindowMillisenconds: 1000,
				},
				stateChan: nil,
				healthy: func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool {
					return true
				},
				fallback: func() (interface{}, error) {
					return 5, nil
				},
			},
			args: args{
				ctx: context.Background(),
				operation: func() (interface{}, error) {
					return 100, nil
				},
			},
			want:    100,
			wantErr: false,
		},
		{
			name: "opens the circuit and calls the fallback given an unhealthly system",
			fields: fields{
				state: State{
					status:  Closed,
					updated: time.Now().Add(-1 * time.Minute),
				},
				config: Config{
					SleepWindowMillisenconds: 1000,
				},
				stateChan: nil,
				healthy: func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool {
					return false
				},
				fallback: func() (interface{}, error) {
					return 5, nil
				},
			},
			args: args{
				ctx: context.Background(),
				operation: func() (interface{}, error) {
					return 100, nil
				},
			},
			want:    5,
			wantErr: false,
		},
		{
			name: "uses the default fallback when one is not supplied",
			fields: fields{
				state: State{
					status:  Open,
					updated: time.Now().Add(-1 * time.Minute),
				},
				config: Config{
					SleepWindowMillisenconds: 1000000,
				},

				healthy: func(health.Config, map[int64]map[health.MetricType]int64, []int64) bool {
					return false
				},
				stateChan: nil,
				fallback:  nil,
			},
			args: args{
				ctx: context.Background(),
				operation: func() (interface{}, error) {
					return 100, nil
				},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			c := New(tt.fields.config, tt.fields.stateChan, tt.fields.fallback, tt.fields.healthy)
			c.state = tt.fields.state

			got, err := c.DoWithContext(tt.args.ctx, tt.args.operation)
			if (err != nil) != tt.wantErr {
				t.Errorf("CircuitBreaker.DoWithContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CircuitBreaker.DoWithContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
