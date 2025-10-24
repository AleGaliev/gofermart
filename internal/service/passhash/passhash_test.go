package passhash

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name      string
		password  string
		wantError bool
	}{
		{
			name:      "valid password",
			password:  "mySecurePassword123",
			wantError: false,
		},
		{
			name:      "empty password",
			password:  "",
			wantError: false,
		},
		{
			name:      "long password",
			password:  "veryLongPasswordWithSpecialChars!@#$%^&*()",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.wantError {
				t.Errorf("HashPassword() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError && hash == "" {
				t.Error("HashPassword() returned empty hash for valid password")
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "testPassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to create hash for testing: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongPassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "invalid hash",
			password: password,
			hash:     "invalidHashString",
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPasswordHash(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
