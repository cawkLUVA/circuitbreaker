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
		health    Health
		stateChan chan State
		fallback  func() (interface{}, error)
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
			name: "successfully executes given a healthy system and closed circuit",
			fields: fields{
				state: State{
					status: Closed,
				},
				config: Config{
					SleepWindowMillisenconds: 1000,
				},
				health: &HealthMock{
					err:      nil,
					healthly: true,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CircuitBreaker{
				state:     tt.fields.state,
				config:    tt.fields.config,
				health:    tt.fields.health,
				stateChan: tt.fields.stateChan,
				fallback:  tt.fields.fallback,
			}
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
