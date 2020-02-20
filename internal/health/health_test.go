package health

import (
	"reflect"
	"testing"
	"time"
)

func Test_binarySearchIndex(t *testing.T) {
	type args struct {
		key  int64
		keys []int64
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		{
			name: "finds index",
			args: args{
				key:  1,
				keys: []int64{1},
			},
			want: 1,
		},
		{
			name: "finds index with equal key",
			args: args{
				key:  2,
				keys: []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
			},
			want: 2,
		},
		{
			name: "finds index with non-equal key",
			args: args{
				key:  4,
				keys: []int64{1, 2, 3, 5, 6, 7, 8, 9, 10, 11},
			},
			want: 3,
		},
		{
			name: "finds index at the start",
			args: args{
				key:  1,
				keys: []int64{2, 3, 5, 6, 7, 8, 9, 10, 11},
			},
			want: 0,
		},
		{
			name: "finds index at the end",
			args: args{
				key:  12,
				keys: []int64{2, 3, 5, 6, 7, 8, 9, 10, 11},
			},
			want: 9,
		},
		{
			name: "finds equal index at the end",
			args: args{
				key:  11,
				keys: []int64{2, 3, 5, 6, 7, 8, 9, 10, 11},
			},
			want: 9,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := binarySearchIndex(tt.args.key, tt.args.keys); got != tt.want {
				t.Errorf("findIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHealth_addKey(t *testing.T) {
	type fields struct {
		metrics map[int64]map[MetricType]int64
		keys    []int64
		config  Config
		want    []int64
	}
	type args struct {
		key int64
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []int64
	}{
		{
			name: "can add a key to an empty slice",
			fields: fields{
				keys: []int64{},
			},
			args: args{
				key: 946728000,
			},
			want: []int64{946728000},
		},
		{
			name: "can add a key to the start",
			fields: fields{
				keys: []int64{950000000, 950000001, 950000002},
			},
			args: args{
				key: 946728000,
			},
			want: []int64{946728000, 950000000, 950000001, 950000002},
		},
		{
			name: "can add a key to the end",
			fields: fields{
				keys: []int64{930000000, 930000001, 930000002},
			},
			args: args{
				key: 946728000,
			},
			want: []int64{930000000, 930000001, 930000002, 946728000},
		},
		{
			name: "can add a key in the middle",
			fields: fields{
				keys: []int64{930000000, 950000000},
			},
			args: args{
				key: 946728000,
			},
			want: []int64{930000000, 946728000, 950000000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Health{
				metrics: tt.fields.metrics,
				keys:    tt.fields.keys,
				config:  tt.fields.config,
			}
			c.addKey(tt.args.key)

			if !reflect.DeepEqual(c.keys, tt.want) {
				t.Errorf("addKey() = %v, want %v", c.keys, tt.want)
			}

		})
	}
}

// TODO add windowsize trimming logic
func TestHealth_AddMetric(t *testing.T) {
	type fields struct {
		metrics map[int64]map[MetricType]int64
		keys    []int64
		config  Config
	}
	type args struct {
		timestamp time.Time
		metric    MetricType
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		wantKeys    []int64
		wantMetrics map[int64]map[MetricType]int64
	}{
		{
			name: "can add a success metric to an empty list",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{},
				keys:    []int64{},
				config:  Config{},
			},
			args: args{
				timestamp: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				metric:    Success,
			},
			wantErr:  false,
			wantKeys: []int64{946728000},
			wantMetrics: map[int64]map[MetricType]int64{
				946728000: map[MetricType]int64{
					Success: 1,
				},
			},
		},
		{
			name: "can add a success metric to an existing node in the list",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					946728000: map[MetricType]int64{
						Success: 1,
					},
				},
				keys:   []int64{946728000},
				config: Config{},
			},
			args: args{
				timestamp: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				metric:    Success,
			},
			wantErr:  false,
			wantKeys: []int64{946728000},
			wantMetrics: map[int64]map[MetricType]int64{
				946728000: map[MetricType]int64{
					Success: 2,
				},
			},
		},
		{
			name: "can add an error metric to an existing node in the list",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					946728000: map[MetricType]int64{
						Success: 1,
					},
				},
				keys:   []int64{946728000},
				config: Config{},
			},
			args: args{
				timestamp: time.Date(2000, 1, 1, 12, 0, 0, 0, time.UTC),
				metric:    Error,
			},
			wantErr:  false,
			wantKeys: []int64{946728000},
			wantMetrics: map[int64]map[MetricType]int64{
				946728000: map[MetricType]int64{
					Success: 1,
					Error:   1,
				},
			},
		},
		{
			name: "can add a new metric key to the list",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					946728000: map[MetricType]int64{
						Success: 1,
					},
					946728002: map[MetricType]int64{
						Error: 1,
					},
				},
				keys:   []int64{946728000, 946728002},
				config: Config{},
			},
			args: args{
				timestamp: time.Date(2000, 1, 1, 12, 0, 1, 0, time.UTC),
				metric:    Error,
			},
			wantErr:  false,
			wantKeys: []int64{946728000, 946728001, 946728002},
			wantMetrics: map[int64]map[MetricType]int64{
				946728000: map[MetricType]int64{
					Success: 1,
				},
				946728001: map[MetricType]int64{
					Error: 1,
				},
				946728002: map[MetricType]int64{
					Error: 1,
				},
			},
		},
		{
			name: "returns an error for invalid metric types",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					100: map[MetricType]int64{
						Success: 1,
					},
				},
				keys:   []int64{100},
				config: Config{},
			},
			args: args{
				timestamp: time.Date(2000, 1, 1, 12, 0, 1, 0, time.UTC),
				metric:    -1,
			},
			wantErr:  true,
			wantKeys: []int64{100},
			wantMetrics: map[int64]map[MetricType]int64{
				100: map[MetricType]int64{
					Success: 1,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Health{
				metrics: tt.fields.metrics,
				keys:    tt.fields.keys,
				config:  tt.fields.config,
			}
			if err := c.AddMetric(tt.args.timestamp, tt.args.metric); (err != nil) != tt.wantErr {
				t.Errorf("Health.AddMetric() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(c.keys, tt.wantKeys) {
				t.Errorf("AddMetric() keys = %v, want %v", c.keys, tt.wantKeys)
			}

			if !reflect.DeepEqual(c.metrics, tt.wantMetrics) {
				t.Errorf("AddMetric() metrics = %v, want %v", c.metrics, tt.wantMetrics)
			}

		})
	}
}

func TestHealth_Healthy(t *testing.T) {
	type fields struct {
		metrics  map[int64]map[MetricType]int64
		keys     []int64
		config   Config
		healthly func(Config, map[int64]map[MetricType]int64, []int64) bool
		now      time.Time
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "default algorithm can correctly detect a healthy system",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					930000000: map[MetricType]int64{
						Success: 100,
					},
					930000001: map[MetricType]int64{
						Error: 1,
					},
				},
				keys: []int64{930000000, 930000001},
				config: Config{
					windowSize:               999999999,
					errorPercentageThreshold: 0.1,
				},
				healthly: defaultHealthChecker,
				now:      time.Date(2000, 1, 1, 12, 0, 1, 0, time.UTC),
			},
			want: true,
		},
		{
			name: "default algorithm can correctly detect an unhealthy system",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					930000000: map[MetricType]int64{
						Success: 5,
						Error:   2,
					},
					930000001: map[MetricType]int64{
						Success: 2,
						Error:   5,
					},
				},
				keys: []int64{930000000, 930000001},
				config: Config{
					windowSize:               999999999,
					errorPercentageThreshold: 0.5,
				},
				healthly: defaultHealthChecker,
				now:      time.Date(2000, 1, 1, 12, 0, 1, 0, time.UTC),
			},
			want: false,
		},
		{
			name: "correctly removes expired keys",
			fields: fields{
				metrics: map[int64]map[MetricType]int64{
					100: map[MetricType]int64{
						Success: 1,
						Error:   999,
					},
					1580385979: map[MetricType]int64{
						Success: 5,
						Error:   1,
					},
					1580385980: map[MetricType]int64{
						Success: 10,
						Error:   1,
					},
				},
				keys: []int64{100, 1580385979, 1580385980},
				config: Config{
					windowSize:               604800,
					errorPercentageThreshold: 0.5,
				},
				healthly: defaultHealthChecker,
				now:      time.Date(2020, 1, 30, 13, 0, 0, 0, time.UTC),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			Now = func() time.Time {
				return tt.fields.now
			}

			c := &Health{
				metrics:  tt.fields.metrics,
				keys:     tt.fields.keys,
				config:   tt.fields.config,
				healthly: tt.fields.healthly,
			}
			if got := c.Healthy(); got != tt.want {
				t.Errorf("Health.Healthy() = %v, want %v", got, tt.want)
			}
		})
	}
}
