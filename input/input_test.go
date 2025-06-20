package input_test

import (
	"slices"
	"testing"

	"go.followtheprocess.codes/actions/input"
	"go.followtheprocess.codes/test"
)

func TestGet(t *testing.T) {
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
			name:  "empty name",
			env:   map[string]string{},
			input: "",
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
			name: "case insensitive",
			env: map[string]string{
				"INPUT_SOMETHING": "here",
			},
			input: "SoMEThIng",
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

			got, ok := input.Get(tt.input)
			test.Equal(t, ok, tt.ok, test.Context("ok return value did not match"))
			test.Equal(t, got, tt.want, test.Context("returned input variable was wrong"))
		})
	}
}

func TestBool(t *testing.T) {
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

			got, err := input.Bool(tt.input)
			test.WantErr(t, err, tt.wantErr)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestLines(t *testing.T) {
	tests := []struct {
		env     map[string]string // Env vars to set for the test
		name    string            // Name of the test case
		input   string            // Name of the input variable to get
		want    []string          // Expected return value
		wantErr bool              // Whether we wanted an error
	}{
		{
			name:    "empty",
			env:     map[string]string{},
			input:   "hello",
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello",
			},
			input:   "something_else",
			want:    nil,
			wantErr: true,
		},
		{
			name: "found",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello\nthere\ngeneral\nkenobi",
			},
			input:   "hello_there",
			want:    []string{"hello", "there", "general", "kenobi"},
			wantErr: false,
		},
		{
			name: "found and trimmed",
			env: map[string]string{
				"INPUT_HELLO_THERE": "\n\n hello\t \n\t there  \n    general\n\t\t kenobi\n\n",
			},
			input:   "hello_there",
			want:    []string{"hello", "there", "general", "kenobi"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, err := input.Lines(tt.input)
			test.WantErr(t, err, tt.wantErr)
			test.EqualFunc(t, got, tt.want, slices.Equal)
		})
	}
}

func TestInt(t *testing.T) {
	tests := []struct {
		env     map[string]string // Env vars to set for the test
		name    string            // Name of the test case
		input   string            // Name of the input variable to get
		want    int               // Expected return value
		wantErr bool              // Whether we wanted an error
	}{
		{
			name:    "empty",
			env:     map[string]string{},
			input:   "hello",
			want:    0,
			wantErr: true,
		},
		{
			name: "missing",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello",
			},
			input:   "something_else",
			want:    0,
			wantErr: true,
		},
		{
			name: "found",
			env: map[string]string{
				"INPUT_NUM_THINGS": "42",
			},
			input:   "num_things",
			want:    42,
			wantErr: false,
		},
		{
			name: "found invalid",
			env: map[string]string{
				"INPUT_NUM_THINGS": "42cheese",
			},
			input:   "num_things",
			want:    0,
			wantErr: true,
		},
		{
			name: "found and trimmed",
			env: map[string]string{
				"INPUT_NUM_THINGS": "\n\n 347 \n\t",
			},
			input:   "num_things",
			want:    347,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, err := input.Int(tt.input)
			test.WantErr(t, err, tt.wantErr)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestFloat(t *testing.T) {
	tests := []struct {
		env     map[string]string // Env vars to set for the test
		name    string            // Name of the test case
		input   string            // Name of the input variable to get
		want    float64           // Expected return value
		wantErr bool              // Whether we wanted an error
	}{
		{
			name:    "empty",
			env:     map[string]string{},
			input:   "hello",
			want:    0,
			wantErr: true,
		},
		{
			name: "missing",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello",
			},
			input:   "something_else",
			want:    0,
			wantErr: true,
		},
		{
			name: "found",
			env: map[string]string{
				"INPUT_PI": "3.14159",
			},
			input:   "pi",
			want:    3.14159,
			wantErr: false,
		},
		{
			name: "found invalid",
			env: map[string]string{
				"INPUT_PI": "3.cheese",
			},
			input:   "pi",
			want:    0,
			wantErr: true,
		},
		{
			name: "found and trimmed",
			env: map[string]string{
				"INPUT_PI": "\n\n 3.14159 \n\t",
			},
			input:   "pi",
			want:    3.14159,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, err := input.Float(tt.input)
			test.WantErr(t, err, tt.wantErr)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestList(t *testing.T) {
	tests := []struct {
		env     map[string]string // Env vars to set for the test
		name    string            // Name of the test case
		input   string            // Name of the input variable to get
		want    []string          // Expected return value
		wantErr bool              // Whether we wanted an error
	}{
		{
			name:    "empty",
			env:     nil,
			input:   "hello",
			want:    nil,
			wantErr: true,
		},
		{
			name: "missing",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello",
			},
			input:   "something_else",
			want:    nil,
			wantErr: true,
		},
		{
			name: "found line separated",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello\nthere\ngeneral\nkenobi",
			},
			input:   "hello_there",
			want:    []string{"hello", "there", "general", "kenobi"},
			wantErr: false,
		},
		{
			name: "found comma separated",
			env: map[string]string{
				"INPUT_HELLO_THERE": "hello,there,general,kenobi",
			},
			input:   "hello_there",
			want:    []string{"hello", "there", "general", "kenobi"},
			wantErr: false,
		},
		{
			name: "found and trimmed line",
			env: map[string]string{
				"INPUT_HELLO_THERE": "\n\n hello\t \n\t there  \n    general\n\t\t kenobi\n\n",
			},
			input:   "hello_there",
			want:    []string{"hello", "there", "general", "kenobi"},
			wantErr: false,
		},
		{
			name: "found and trimmed comma",
			env: map[string]string{
				"INPUT_HELLO_THERE": "\n\n hello\t , \t there  ,  general,\t\t kenobi   ",
			},
			input:   "hello_there",
			want:    []string{"hello", "there", "general", "kenobi"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		for key, val := range tt.env {
			t.Setenv(key, val)
		}

		got, err := input.List(tt.input)
		test.WantErr(t, err, tt.wantErr)
		test.EqualFunc(t, got, tt.want, slices.Equal)
	}
}
