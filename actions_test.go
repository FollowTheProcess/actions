package actions_test

import (
	"testing"

	"github.com/FollowTheProcess/actions"
	"github.com/FollowTheProcess/test"
)

func TestInput(t *testing.T) {
	tests := []struct {
		env   map[string]string // Env vars to set for the test
		name  string            // Name of the test case
		input string            // The name of the input variable to get
		want  string            // Expected return value
		ok    bool              // Expected ok value
	}{
		{
			name:  "empty",
			env:   map[string]string{},
			input: "something",
			want:  "",
			ok:    false,
		},
		{
			name: "missing",
			env: map[string]string{
				"INPUT_SOMETHING": "here",
			},
			input: "something_else", // Would be looking for INPUT_SOMETHING_ELSE
			want:  "",
			ok:    false,
		},
		{
			name: "found",
			env: map[string]string{
				"INPUT_SOMETHING": "here",
			},
			input: "something",
			want:  "here",
			ok:    true,
		},
		{
			name: "found with space",
			env: map[string]string{
				"INPUT_SOMETHING_ELSE": "here",
			},
			input: "something else",
			want:  "here",
			ok:    true,
		},
		{
			name: "return trimmed",
			env: map[string]string{
				"INPUT_TRIM_ME": "  okay   ",
			},
			input: "trim_me",
			want:  "okay",
			ok:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, ok := actions.Input(tt.input)
			test.Equal(t, ok, tt.ok, test.Context("ok return value did not match"))
			test.Equal(t, got, tt.want, test.Context("returned input variable was wrong"))
		})
	}
}

func TestInputBool(t *testing.T) {
	tests := []struct {
		env     map[string]string // Env vars to set for the test
		name    string            // Name of the test case
		input   string            // Name of the input variable to get
		want    bool              // Expected return value
		wantErr bool              // Whether we wanted an error
	}{
		{
			name:    "empty",
			env:     map[string]string{},
			input:   "something",
			want:    false,
			wantErr: true,
		},
		{
			name: "missing",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "true",
			},
			input:   "other",
			want:    false,
			wantErr: true,
		},
		{
			name: "valid true lower",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "true",
			},
			input:   "do_something",
			want:    true,
			wantErr: false,
		},
		{
			name: "valid true title",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "True",
			},
			input:   "do_something",
			want:    true,
			wantErr: false,
		},
		{
			name: "valid true upper",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "TRUE",
			},
			input:   "do_something",
			want:    true,
			wantErr: false,
		},
		{
			name: "valid false lower",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "false",
			},
			input:   "do_something",
			want:    false,
			wantErr: false,
		},
		{
			name: "valid false title",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "False",
			},
			input:   "do_something",
			want:    false,
			wantErr: false,
		},
		{
			name: "valid false upper",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "FALSE",
			},
			input:   "do_something",
			want:    false,
			wantErr: false,
		},
		{
			name: "invalid true",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "yes",
			},
			input:   "do_something",
			want:    false,
			wantErr: true,
		},
		{
			name: "invalid false",
			env: map[string]string{
				"INPUT_DO_SOMETHING": "no",
			},
			input:   "do_something",
			want:    false,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, err := actions.InputBool(tt.input)
			test.WantErr(t, err, tt.wantErr)
			test.Equal(t, got, tt.want)
		})
	}
}
