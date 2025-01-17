package log_test

import (
	"bytes"
	"testing"

	"github.com/FollowTheProcess/actions/log"
	"github.com/FollowTheProcess/test"
)

func TestIsDebug(t *testing.T) {
	tests := []struct {
		env  map[string]string // Env vars to set for the test
		name string            // Name of the test case
		want bool              // Expected return value
	}{
		{
			name: "unset",
			env:  map[string]string{},
			want: false,
		},
		{
			name: "off",
			env: map[string]string{
				"RUNNER_DEBUG": "0", // Presumably this is what it's set to when off
			},
			want: false,
		},
		{
			name: "on",
			env: map[string]string{
				"RUNNER_DEBUG": "1",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got := log.IsDebug()
			test.Equal(t, got, tt.want)
		})
	}
}

func TestDebug(t *testing.T) {
	tests := []struct {
		name   string // Name of the test case
		format string // Format template
		want   string // Expected output
		args   []any  // Args to be formatted
	}{
		{
			name:   "empty",
			format: "",
			args:   nil,
			want:   "",
		},
		{
			name:   "simple message",
			format: "debug log here",
			args:   nil,
			want:   "::debug::debug log here\n",
		},
		{
			name:   "formatted",
			format: "reading file: %s",
			args:   []any{"some/file.txt"},
			want:   "::debug::reading file: some/file.txt\n",
		},
		{
			name:   "formatted with escaped chars",
			format: "stuff \r\n happening here %d%% complete",
			args:   []any{42},
			want:   "::debug::stuff %0D%0A happening here 42%25 complete\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := log.New(buf)

			logger.Debug(tt.format, tt.args...)

			got := buf.String()
			test.Equal(t, got, tt.want)
		})
	}
}
