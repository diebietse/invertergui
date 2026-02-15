package main

import (
	"os"
	"path/filepath"
	"testing"
)

const testInlineSecret = "inline-secret"

func TestReadPasswordFile(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "plain password",
			content:  "secret",
			expected: "secret",
		},
		{
			name:     "password with trailing newline",
			content:  "secret\n",
			expected: "secret",
		},
		{
			name:     "password with trailing carriage return and newline",
			content:  "secret\r\n",
			expected: "secret",
		},
		{
			name:     "empty file",
			content:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "password")
			if err := os.WriteFile(path, []byte(tt.content), 0o600); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			got, err := readPasswordFile(path)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestReadPasswordFile_NotFound(t *testing.T) {
	_, err := readPasswordFile("/nonexistent/path/password")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestResolvePassword_MutuallyExclusive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(path, []byte("secret"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	conf := &config{}
	conf.MQTT.Password = testInlineSecret
	conf.MQTT.PasswordFile = path

	err := resolvePasswordFile(conf)
	if err == nil {
		t.Fatal("expected error when both mqtt.password and mqtt.password-file are set, got nil")
	}
}

func TestResolvePassword_FromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "password")
	if err := os.WriteFile(path, []byte("file-secret\n"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	conf := &config{}
	conf.MQTT.PasswordFile = path

	if err := resolvePasswordFile(conf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.MQTT.Password != "file-secret" {
		t.Errorf("got %q, want %q", conf.MQTT.Password, "file-secret")
	}
}

func TestResolvePassword_NoFile(t *testing.T) {
	conf := &config{}
	conf.MQTT.Password = testInlineSecret

	if err := resolvePasswordFile(conf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conf.MQTT.Password != testInlineSecret {
		t.Errorf("got %q, want %q", conf.MQTT.Password, testInlineSecret)
	}
}
