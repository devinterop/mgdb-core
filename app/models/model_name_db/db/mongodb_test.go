package db

import (
	"testing"
	"time"
)

func TestNormalizeConnectionConfigDefaults(t *testing.T) {
	config := normalizeConnectionConfig(ConnectionConfig{})

	if config.MaxPoolSize != DefaultMaxPoolSize {
		t.Fatalf("MaxPoolSize = %d, want %d", config.MaxPoolSize, DefaultMaxPoolSize)
	}
	if config.MinPoolSize != DefaultMinPoolSize {
		t.Fatalf("MinPoolSize = %d, want %d", config.MinPoolSize, DefaultMinPoolSize)
	}
	if config.MaxConnIdleTime != DefaultMaxConnIdleTime {
		t.Fatalf("MaxConnIdleTime = %s, want %s", config.MaxConnIdleTime, DefaultMaxConnIdleTime)
	}
}

func TestNormalizeConnectionConfigPreservesConfiguredValues(t *testing.T) {
	want := ConnectionConfig{
		AppName:         "appointmentapi",
		MaxPoolSize:     250,
		MinPoolSize:     10,
		MaxConnIdleTime: 15 * time.Minute,
	}

	if got := normalizeConnectionConfig(want); got != want {
		t.Fatalf("normalizeConnectionConfig() = %#v, want %#v", got, want)
	}
}

func TestNormalizeConnectionConfigDefaultsEachMissingValueIndependently(t *testing.T) {
	tests := []struct {
		name  string
		input ConnectionConfig
		want  ConnectionConfig
	}{
		{
			name: "missing max pool size",
			input: ConnectionConfig{
				MinPoolSize:     10,
				MaxConnIdleTime: 15 * time.Minute,
			},
			want: ConnectionConfig{
				MaxPoolSize:     DefaultMaxPoolSize,
				MinPoolSize:     10,
				MaxConnIdleTime: 15 * time.Minute,
			},
		},
		{
			name: "missing min pool size",
			input: ConnectionConfig{
				MaxPoolSize:     250,
				MaxConnIdleTime: 15 * time.Minute,
			},
			want: ConnectionConfig{
				MaxPoolSize:     250,
				MinPoolSize:     DefaultMinPoolSize,
				MaxConnIdleTime: 15 * time.Minute,
			},
		},
		{
			name: "missing max connection idle time",
			input: ConnectionConfig{
				MaxPoolSize: 250,
				MinPoolSize: 10,
			},
			want: ConnectionConfig{
				MaxPoolSize:     250,
				MinPoolSize:     10,
				MaxConnIdleTime: DefaultMaxConnIdleTime,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeConnectionConfig(tt.input); got != tt.want {
				t.Fatalf("normalizeConnectionConfig() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
