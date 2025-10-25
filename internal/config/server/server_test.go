package server

import (
	"flag"
	"os"
	"testing"
)

func TestNewServerParams(t *testing.T) {
	// Сохраняем оригинальные значения флагов и окружения
	originalArgs := os.Args
	originalEnvVars := make(map[string]string)
	envVars := []string{"RUN_ADDRESS", "DATABASE_URI", "ACCRUAL_SYSTEM_ADDRESS", "KEY"}
	for _, env := range envVars {
		if val, ok := os.LookupEnv(env); ok {
			originalEnvVars[env] = val
		}
	}
	defer func() {
		// Восстанавливаем оригинальные значения
		os.Args = originalArgs
		flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

		for _, env := range envVars {
			if val, exists := originalEnvVars[env]; exists {
				os.Setenv(env, val)
			} else {
				os.Unsetenv(env)
			}
		}
	}()

	tests := []struct {
		name       string
		args       []string
		envVars    map[string]string
		wantConfig ServerParams
		wantErr    bool
	}{
		{
			name: "default values",
			args: []string{"test"},
			wantConfig: ServerParams{
				AdrHost:              "localhost:8080",
				AccrualSystemAddress: "http://localhost:8081",
				DatabaseDSN:          "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
				HashKey:              "",
				ConsumerCountWorkers: consumerCountWorkers,
			},
		},
		{
			name: "command line flags",
			args: []string{
				"test",
				"-a", "127.0.0.1:9090",
				"-r", "accrual.example.com:8082",
				"-d", "postgres://user:pass@localhost:5432/testdb",
				"-k", "secret-key",
			},
			wantConfig: ServerParams{
				AdrHost:              "127.0.0.1:9090",
				AccrualSystemAddress: "accrual.example.com:8082",
				DatabaseDSN:          "postgres://user:pass@localhost:5432/testdb",
				HashKey:              "secret-key",
				ConsumerCountWorkers: consumerCountWorkers,
			},
		},
		{
			name: "environment variables override flags",
			args: []string{
				"test",
				"-a", "flag-host:8080",
				"-r", "flag-accrual:8081",
				"-d", "flag-dsn",
				"-k", "flag-key",
			},
			envVars: map[string]string{
				"RUN_ADDRESS":            "env-host:9090",
				"ACCRUAL_SYSTEM_ADDRESS": "env-accrual:9091",
				"DATABASE_URI":           "env-dsn",
				"KEY":                    "env-key",
			},
			wantConfig: ServerParams{
				AdrHost:              "env-host:9090",
				AccrualSystemAddress: "env-accrual:9091",
				DatabaseDSN:          "env-dsn",
				HashKey:              "env-key",
				ConsumerCountWorkers: consumerCountWorkers,
			},
		},
		{
			name: "partial environment variables",
			args: []string{"test"},
			envVars: map[string]string{
				"RUN_ADDRESS": "custom-host:8080",
				"KEY":         "custom-key",
			},
			wantConfig: ServerParams{
				AdrHost:              "custom-host:8080",
				AccrualSystemAddress: "http://localhost:8081",
				DatabaseDSN:          "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable",
				HashKey:              "custom-key",
				ConsumerCountWorkers: consumerCountWorkers,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Очищаем окружение
			for _, env := range envVars {
				os.Unsetenv(env)
			}

			// Устанавливаем переменные окружения для теста
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Устанавливаем аргументы командной строки
			os.Args = tt.args
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)

			// Получаем конфигурацию
			got := NewServerParams()

			// Проверяем результаты
			if got.AdrHost != tt.wantConfig.AdrHost {
				t.Errorf("AdrHost = %v, want %v", got.AdrHost, tt.wantConfig.AdrHost)
			}
			if got.AccrualSystemAddress != tt.wantConfig.AccrualSystemAddress {
				t.Errorf("AccrualSystemAddress = %v, want %v", got.AccrualSystemAddress, tt.wantConfig.AccrualSystemAddress)
			}
			if got.DatabaseDSN != tt.wantConfig.DatabaseDSN {
				t.Errorf("DatabaseDSN = %v, want %v", got.DatabaseDSN, tt.wantConfig.DatabaseDSN)
			}
			if got.HashKey != tt.wantConfig.HashKey {
				t.Errorf("HashKey = %v, want %v", got.HashKey, tt.wantConfig.HashKey)
			}
			if got.ConsumerCountWorkers != tt.wantConfig.ConsumerCountWorkers {
				t.Errorf("ConsumerCountWorkers = %v, want %v", got.ConsumerCountWorkers, tt.wantConfig.ConsumerCountWorkers)
			}
		})
	}
}
