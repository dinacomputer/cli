package cli

import (
	"testing"
)

func TestParseBuildArgs(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr bool
	}{
		{
			name:  "nil input",
			input: nil,
			want:  nil,
		},
		{
			name:  "empty input",
			input: []string{},
			want:  nil,
		},
		{
			name:  "single arg",
			input: []string{"FOO=bar"},
			want:  map[string]string{"FOO": "bar"},
		},
		{
			name:  "multiple args",
			input: []string{"FOO=bar", "BAZ=qux"},
			want:  map[string]string{"FOO": "bar", "BAZ": "qux"},
		},
		{
			name:  "value with equals sign",
			input: []string{"FOO=bar=baz"},
			want:  map[string]string{"FOO": "bar=baz"},
		},
		{
			name:  "empty value",
			input: []string{"FOO="},
			want:  map[string]string{"FOO": ""},
		},
		{
			name:    "missing equals sign",
			input:   []string{"FOO"},
			wantErr: true,
		},
		{
			name:    "one valid one invalid",
			input:   []string{"FOO=bar", "BAD"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseBuildArgs(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.want == nil {
				if got != nil {
					t.Fatalf("expected nil, got %v", got)
				}
				return
			}
			if len(got) != len(tt.want) {
				t.Fatalf("got %d entries, want %d", len(got), len(tt.want))
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("got[%q] = %q, want %q", k, got[k], v)
				}
			}
		})
	}
}
