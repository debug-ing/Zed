package server

import (
	"bytes"
	"testing"
)

func TestDetectProtocol(t *testing.T) {
	tests := []struct {
		name string
		buf  []byte
		n    int
		want Protocol
	}{
		{"tcp", []byte{0x01}, 1, ProtocolTCP},
		{"udp", []byte{0x02}, 1, ProtocolUDP},
		{"unknown byte", []byte{0x03}, 1, ProtocolUnknown},
		{"empty", []byte{}, 0, ProtocolUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectProtocol(tt.buf, tt.n)
			if got != tt.want {
				t.Errorf("expected %v, got %v", tt.want, got)
			}
		})
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		name        string
		input       []byte
		wantDest    string
		wantPayload []byte
		wantErr     bool
	}{
		{"valid", []byte("8.8.8.8:53\x00hello"), "8.8.8.8:53", []byte("hello"), false},
		{"missing delimiter", []byte("invalidheader"), "", nil, true},
		{"empty payload", []byte("example.com:80\x00"), "example.com:80", []byte{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest, payload, err := parseHeader(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("expected err=%v got %v", tt.wantErr, err)
			}
			if dest != tt.wantDest {
				t.Errorf("expected dest=%q got %q", tt.wantDest, dest)
			}
			if !bytes.Equal(payload, tt.wantPayload) {
				t.Errorf("expected payload=%q got %q", tt.wantPayload, payload)
			}
		})
	}
}
