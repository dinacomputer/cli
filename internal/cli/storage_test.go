package cli

import "testing"

func TestParseSize(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{input: "0", want: 0},
		{input: "1024", want: 1024},
		{input: "1024B", want: 1024},
		{input: "1KB", want: 1 << 10},
		{input: "500MB", want: 500 << 20},
		{input: "10GB", want: 10 << 30},
		{input: "2TB", want: 2 << 40},
		{input: "10g", want: 10 << 30},
		{input: "10GiB", want: 10 << 30},
		{input: "  10GB  ", want: 10 << 30},
		{input: "1.5GB", want: 1610612736},
		{input: "", wantErr: true},
		{input: "abc", wantErr: true},
		{input: "10PB", wantErr: true},
		{input: "-1GB", wantErr: true},
	}
	for _, tt := range tests {
		got, err := parseSize(tt.input)
		if tt.wantErr {
			if err == nil {
				t.Errorf("parseSize(%q) = %d, want error", tt.input, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parseSize(%q) returned error: %v", tt.input, err)
			continue
		}
		if got != tt.want {
			t.Errorf("parseSize(%q) = %d, want %d", tt.input, got, tt.want)
		}
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		input int64
		want  string
	}{
		{0, "unlimited"},
		{512, "512B"},
		{1 << 10, "1KB"},
		{1536, "1.5KB"},
		{10 << 30, "10GB"},
		{2 << 40, "2TB"},
	}
	for _, tt := range tests {
		if got := formatBytes(tt.input); got != tt.want {
			t.Errorf("formatBytes(%d) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// A quota round-trips through both helpers unchanged for the unit-aligned
// values the flags are meant to take.
func TestParseSizeFormatBytesRoundTrip(t *testing.T) {
	for _, s := range []string{"512MB", "10GB", "2TB", "1KB"} {
		n, err := parseSize(s)
		if err != nil {
			t.Fatalf("parseSize(%q): %v", s, err)
		}
		if got := formatBytes(n); got != s {
			t.Errorf("round trip of %q gave %q", s, got)
		}
	}
}
