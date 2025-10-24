package server

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestNewServerConfig(t *testing.T) {
	tests := []struct {
		name    string
		want    ServerConfig
		env     map[string]string
		wantErr bool
	}{
		{
			name: "default config",
			want: ServerConfig{
				AdrHost:         "localhost:8080",
				StoreInterval:   2,
				FileStoragePath: "storage.json",
				DatabaseDSN:     "",
				Restore:         true,
			},
			env:     map[string]string{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for envKey, envValue := range tt.env {
				os.Setenv(envKey, envValue)
			}
			flag.Parse()
			fmt.Println(os.Getenv("ADDRESS"))
			got, err := NewServerConfig()
			if (err != nil) != tt.wantErr {
				t.Errorf("NewServerConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewServerConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}
