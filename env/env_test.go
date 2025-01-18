package env //nolint: testpackage
// testpackage is off because we need access to envFile in here because this project
// itself is tested on GitHub Actions, meaning we can't remove or manipulate the real
// $GITHUB_ENV file to test things like what happens if it's not present etc. so for tests
// we use $TEST_GITHUB_ENV, while the real code uses $GITHUB_ENV.

import (
	"bytes"
	"os"
	"testing"

	"github.com/FollowTheProcess/test"
)

const testEnvName = "TEST_GITHUB_ENV"

func TestGet(t *testing.T) {
	tests := []struct {
		name string            // Name of the test case
		env  map[string]string // Env vars to set for the test
		key  string            // The key of the env var to get
		want string            // Expected value
		ok   bool              // Expected ok
	}{
		{
			name: "empty",
			env:  map[string]string{},
			key:  "",
			want: "",
			ok:   false,
		},
		{
			name: "missing",
			env: map[string]string{
				"SOMETHING": "here",
			},
			key:  "OTHER",
			want: "",
			ok:   false,
		},
		{
			name: "present but empty",
			env: map[string]string{
				"SOMETHING": "",
			},
			key:  "SOMETHING",
			want: "",
			ok:   true,
		},
		{
			name: "present and set",
			env: map[string]string{
				"SOMETHING": "here",
			},
			key:  "SOMETHING",
			want: "here",
			ok:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, ok := Get(tt.key)
			test.Equal(t, ok, tt.ok)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string // Name of the test case
		key     string // Key to set
		value   string // Value to set
		errMsg  string // If we wanted an error, what should it say
		wantErr bool   // Whether we want an error
	}{
		{
			name:    "empty key",
			key:     "",
			value:   "something",
			wantErr: true,
			errMsg:  "key cannot be empty",
		},
		{
			name:    "empty val",
			key:     "something",
			value:   "",
			wantErr: true,
			errMsg:  "value cannot be empty",
		},
		{
			name:    "key ci",
			key:     "CI",
			value:   "anything",
			wantErr: true,
			errMsg:  "setting $CI is disallowed",
		},
		{
			name:    "key node options",
			key:     "NODE_OPTIONS",
			value:   "anything",
			wantErr: true,
			errMsg:  "setting $NODE_OPTIONS is disallowed",
		},
		{
			name:    "key github",
			key:     "GITHUB_ANYTHING",
			value:   "value",
			wantErr: true,
			errMsg:  "setting $GITHUB_ANYTHING is disallowed",
		},
		{
			name:    "key runner",
			key:     "RUNNER_ANYTHING",
			value:   "value",
			wantErr: true,
			errMsg:  "setting $RUNNER_ANYTHING is disallowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			envFile = testEnvName
			err := Set(tt.key, tt.value)
			test.WantErr(t, err, tt.wantErr)
			if err != nil {
				test.Equal(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestSetFile(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		envFile = testEnvName
		tmp, err := os.CreateTemp("", "TestSetFile*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		t.Setenv(envFile, tmp.Name()) // Set $TEST_GITHUB_ENV to the path to our file

		err = Set("SOMETHING", "value")
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		hasEnvVarSet := bytes.Contains(contents, []byte("SOMETHING=value"))
		test.True(t, hasEnvVarSet)
	})
	t.Run("unset", func(t *testing.T) {
		envFile = testEnvName
		tmp, err := os.CreateTemp("", "TestSetFile*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		// Not setting $TEST_GITHUB_ENV
		err = Set("KEY", "value")
		test.Err(t, err)
	})
	t.Run("set but no file", func(t *testing.T) {
		envFile = testEnvName
		t.Setenv(envFile, "missing") // Set $TEST_GITHUB_ENV to the path of a missing file

		err := Set("CONFIG", "yes")
		test.Err(t, err)
	})
}
