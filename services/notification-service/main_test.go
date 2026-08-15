package main

import (
	"strings"
	"testing"
)

func TestParseWahaBotCommands(t *testing.T) {
	tests := []struct {
		input          string
		expectedSubstr string
	}{
		{"cek order", "BOT DESA"},
		{"status pesanan", "BOT DESA"},
		{"daftar warung", "Warung Bu Ani"},
		{"info tarif", "Rp 10.000"},
		{"halo bantuan", "Selamat datang"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			res := parseWahaBotCommand(tt.input)
			if !strings.Contains(res, tt.expectedSubstr) {
				t.Errorf("Expected response for '%s' to contain '%s', got '%s'", tt.input, tt.expectedSubstr, res)
			}
		})
	}
}

func TestParseWahaBotCommandEmpty(t *testing.T) {
	res := parseWahaBotCommand("random text non bot")
	if res != "" {
		t.Errorf("Expected empty response for non-bot command, got '%s'", res)
	}
}
