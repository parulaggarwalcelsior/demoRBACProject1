package config_test

import (
	"testing"

	"demoRBACProject1/internal/config"
)

func TestConfig(t *testing.T) {
	tests := []struct {
		name        string
		dsn         string
		wantDB      string
		wantRefusal bool
	}{
		{
			name:   "valid shopping cart dsn",
			dsn:    "schema_test:secret@tcp(127.0.0.1:3307)/shopping-cart?parseTime=true&charset=utf8mb4&collation=utf8mb4_0900_bin&loc=UTC",
			wantDB: "shopping-cart",
		},
		{
			name:        "missing dsn refused",
			wantRefusal: true,
		},
		{
			name:        "wrong namespace refused",
			dsn:         "schema_test:secret@tcp(127.0.0.1:3307)/other?parseTime=true",
			wantDB:      "other",
			wantRefusal: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := config.FromDSN(tt.dsn)
			if err != nil {
				t.Fatalf("load config: %v", err)
			}

			if cfg.Database != tt.wantDB {
				t.Fatalf("database = %q, want %q", cfg.Database, tt.wantDB)
			}

			err = config.Validate(cfg)
			if tt.wantRefusal && err == nil {
				t.Fatalf("expected configuration refusal, got nil")
			}
			if !tt.wantRefusal && err != nil {
				t.Fatalf("unexpected configuration refusal: %v", err)
			}
		})
	}
}
