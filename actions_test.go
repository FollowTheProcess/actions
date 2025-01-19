package actions //nolint: testpackage
// testpackage is off because we need access to envFile and outFile in here because this project
// itself is tested on GitHub Actions, meaning we can't remove or manipulate the real
// $GITHUB_ENV file to test things like what happens if it's not present etc. so for tests
// we use $TEST_GITHUB_ENV, while the real code uses $GITHUB_ENV.

import (
	"bytes"
	"os"
	"testing"

	"github.com/FollowTheProcess/test"
)

const (
	testEnvName = "TEST_GITHUB_ENV"
	testOutName = "TEST_GITHUB_OUTPUT"
)

func TestGetEnv(t *testing.T) {
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

			got, ok := GetEnv(tt.key)
			test.Equal(t, ok, tt.ok)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestSetEnvValidation(t *testing.T) {
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
			old := envFile
			envFile = testEnvName
			t.Cleanup(func() { envFile = old })

			err := SetEnv(tt.key, tt.value)
			test.WantErr(t, err, tt.wantErr)
			if err != nil {
				test.Equal(t, err.Error(), tt.errMsg)
			}
		})
	}
}

func TestSetEnv(t *testing.T) {
	old := envFile
	envFile = testEnvName
	t.Cleanup(func() { envFile = old })

	t.Run("exists", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "TestSetEnv*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		t.Setenv(envFile, tmp.Name()) // Set $TEST_GITHUB_ENV to the path to our file

		err = SetEnv("SOMETHING", "value")
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		hasEnvVarSet := bytes.Contains(contents, []byte("SOMETHING=value"))
		test.True(t, hasEnvVarSet)

		// The real environment should also have it set
		test.Equal(t, os.Getenv("SOMETHING"), "value")
	})
	t.Run("multiline", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "TestSetEnv*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		t.Setenv(envFile, tmp.Name()) // Set $TEST_GITHUB_ENV to the path to our file

		value := "values\nacross\nmultiple\nlines"

		err = SetEnv("MULTILINE", value)
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		test.True(t, bytes.Contains(contents, []byte("MULTILINE<<")))
		test.True(t, bytes.Contains(contents, []byte("ghadelimiter_")))
		test.True(t, bytes.Contains(contents, []byte(value)))

		// The real environment should also have it set
		test.Equal(t, os.Getenv("MULTILINE"), value)
	})
	t.Run("unset", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "TestSetEnv*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		// Not setting $TEST_GITHUB_ENV
		err = SetEnv("KEY", "value")
		test.Err(t, err)
	})
	t.Run("set but no file", func(t *testing.T) {
		t.Setenv(envFile, "missing") // Set $TEST_GITHUB_ENV to the path of a missing file

		err := SetEnv("CONFIG", "yes")
		test.Err(t, err)
	})
}

func TestSetOutput(t *testing.T) {
	old := outFile
	outFile = testOutName
	t.Cleanup(func() { outFile = old })

	t.Run("exists", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "TestSetOutput*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		t.Setenv(outFile, tmp.Name()) // Set $TEST_GITHUB_OUTPUT to the path to our file

		err = SetOutput("SOMETHING", "value")
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		hasOutputSet := bytes.Contains(contents, []byte("SOMETHING=value"))
		test.True(t, hasOutputSet)
	})
	t.Run("multiline", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "TestSetOutput*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		t.Setenv(outFile, tmp.Name()) // Set $TEST_GITHUB_OUTPUT to the path to our file

		value := "values\nacross\nmultiple\nlines"

		err = SetOutput("MULTILINE", value)
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		test.True(t, bytes.Contains(contents, []byte("MULTILINE<<")))
		test.True(t, bytes.Contains(contents, []byte("ghadelimiter_")))
		test.True(t, bytes.Contains(contents, []byte(value)))
	})
	t.Run("unset", func(t *testing.T) {
		tmp, err := os.CreateTemp("", "TestSetOutput*")
		test.Ok(t, err)
		t.Cleanup(func() { os.RemoveAll(tmp.Name()) })

		// Not setting $TEST_GITHUB_OUTPUT
		err = SetOutput("KEY", "value")
		test.Err(t, err)
	})
	t.Run("set but no file", func(t *testing.T) {
		t.Setenv(envFile, "missing") // Set $TEST_GITHUB_OUTPUT to the path of a missing file

		err := SetOutput("CONFIG", "yes")
		test.Err(t, err)
	})
}
