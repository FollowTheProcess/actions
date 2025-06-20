package actions //nolint: testpackage // See below
// testpackage is off because we need access to envFile and outFile in here because this project
// itself is tested on GitHub Actions, meaning we can't remove or manipulate the real
// $GITHUB_ENV file to test things like what happens if it's not present etc. so for tests
// we use $TEST_GITHUB_ENV, while the real code uses $GITHUB_ENV.

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.followtheprocess.codes/test"
)

const (
	testEnvName        = "TEST_GITHUB_ENV"
	testOutName        = "TEST_GITHUB_OUTPUT"
	testSummaryName    = "TEST_GITHUB_STEP_SUMMARY"
	testGitHubPathName = "TEST_GITHUB_PATH"
	testStateName      = "TEST_GITHUB_STATE"
	testRealPathName   = "TEST_PATH"
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
		tmp, err := os.CreateTemp(t.TempDir(), "TestSetEnv*")
		test.Ok(t, err)
		tmp.Close()

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
		tmp, err := os.CreateTemp(t.TempDir(), "TestSetEnv*")
		test.Ok(t, err)
		tmp.Close()

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
		// Not setting $TEST_GITHUB_ENV
		err := SetEnv("KEY", "value")
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
		tmp, err := os.CreateTemp(t.TempDir(), "TestSetOutput*")
		test.Ok(t, err)
		tmp.Close()

		t.Setenv(outFile, tmp.Name()) // Set $TEST_GITHUB_OUTPUT to the path to our file

		err = SetOutput("SOMETHING", "value")
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		hasOutputSet := bytes.Contains(contents, []byte("SOMETHING=value"))
		test.True(t, hasOutputSet)
	})
	t.Run("multiline", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "TestSetOutput*")
		test.Ok(t, err)
		tmp.Close()

		t.Setenv(outFile, tmp.Name()) // Set $TEST_GITHUB_OUTPUT to the path to our file

		value := "some\nlines\nhere"

		err = SetOutput("MULTILINE", value)
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		test.True(t, bytes.Contains(contents, []byte("MULTILINE<<")))
		test.True(t, bytes.Contains(contents, []byte("ghadelimiter_")))
		test.True(t, bytes.Contains(contents, []byte(value)))
	})
	t.Run("unset", func(t *testing.T) {
		// Not setting $TEST_GITHUB_OUTPUT
		err := SetOutput("KEY", "value")
		test.Err(t, err)
	})
	t.Run("set but no file", func(t *testing.T) {
		t.Setenv(envFile, "missing") // Set $TEST_GITHUB_OUTPUT to the path of a missing file

		err := SetOutput("CONFIG", "yes")
		test.Err(t, err)
	})
}

func TestGetState(t *testing.T) {
	tests := []struct {
		name string            // Name of the test case
		env  map[string]string // Env vars to set for the test
		key  string            // The key of the state var to get
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
				"STATE_SOMETHING": "here",
			},
			key:  "OTHER",
			want: "",
			ok:   false,
		},
		{
			name: "present but empty",
			env: map[string]string{
				"STATE_EMPTY": "",
			},
			key:  "EMPTY",
			want: "",
			ok:   true,
		},
		{
			name: "present and set",
			env: map[string]string{
				"STATE_FULL": "here",
			},
			key:  "FULL",
			want: "here",
			ok:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for key, val := range tt.env {
				t.Setenv(key, val)
			}

			got, ok := GetState(tt.key)
			test.Equal(t, ok, tt.ok)
			test.Equal(t, got, tt.want)
		})
	}
}

func TestSetState(t *testing.T) {
	old := stateFile
	stateFile = testStateName

	t.Cleanup(func() { stateFile = old })

	t.Run("exists", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "TestSetState*")
		test.Ok(t, err)
		tmp.Close()

		t.Setenv(stateFile, tmp.Name()) // Set $TEST_GITHUB_STATE to the path to our file

		err = SetState("SOMETHING", "value")
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		hasOutputSet := bytes.Contains(contents, []byte("SOMETHING=value"))
		test.True(t, hasOutputSet)
	})
	t.Run("multiline", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "TestSetState*")
		test.Ok(t, err)
		tmp.Close()

		t.Setenv(stateFile, tmp.Name()) // Set $TEST_GITHUB_STATE to the path to our file

		value := "more\nlines\nhere\nwoo"

		err = SetState("MULTILINE", value)
		test.Ok(t, err)

		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		test.True(t, bytes.Contains(contents, []byte("MULTILINE<<")))
		test.True(t, bytes.Contains(contents, []byte("ghadelimiter_")))
		test.True(t, bytes.Contains(contents, []byte(value)))
	})
	t.Run("unset", func(t *testing.T) {
		// Not setting $TEST_GITHUB_STATE
		err := SetState("KEY", "value")
		test.Err(t, err)
	})
	t.Run("set but no file", func(t *testing.T) {
		t.Setenv(envFile, "missing") // Set $TEST_GITHUB_STATE to the path of a missing file

		err := SetState("CONFIG", "yes")
		test.Err(t, err)
	})
}

func TestAddPath(t *testing.T) {
	oldpathFile := pathFile
	oldRealPath := realPath
	pathFile = testGitHubPathName
	realPath = testRealPathName

	t.Cleanup(func() {
		pathFile = oldpathFile
		realPath = oldRealPath
	})
	t.Run("empty path", func(t *testing.T) {
		err := AddPath("")
		test.Err(t, err)
		test.Equal(t, err.Error(), "cannot set an empty path")
	})

	t.Run("unset env", func(t *testing.T) {
		err := AddPath("something")
		test.Err(t, err)
		test.Equal(t, err.Error(), "$GITHUB_PATH is not set or is empty")
	})

	t.Run("missing file", func(t *testing.T) {
		// Set the env var to a file that doesn't exist
		t.Setenv(pathFile, "missing.txt")

		err := AddPath("something")
		test.Err(t, err)
		test.True(t, strings.Contains(err.Error(), "could not open $GITHUB_PATH file missing.txt"))
	})

	t.Run("valid", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "TestAddPath*")
		test.Ok(t, err)
		tmp.Close()

		// Set the env var to our now existing file
		t.Setenv(pathFile, tmp.Name())

		err = AddPath("something")
		test.Ok(t, err)

		// The file should now contain "something"
		contents, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)
		test.DiffBytes(t, contents, []byte("something\n"))

		// Our test $PATH env var should have something on the front of it
		path := os.Getenv(realPath)
		test.True(t, strings.HasPrefix(path, "something"+string(os.PathListSeparator)))
	})
}

func TestSummary(t *testing.T) {
	old := summaryFile
	summaryFile = testSummaryName

	t.Cleanup(func() { summaryFile = old })

	t.Run("unset", func(t *testing.T) {
		err := Summary("# Markdown!\n")
		test.Err(t, err) // $TEST_GITHUB_STEP_SUMMARY is not set
	})

	t.Run("exists but empty", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "TestSummary*")
		test.Ok(t, err)
		tmp.Close()

		t.Setenv(summaryFile, tmp.Name()) // Set $TEST_GITHUB_STEP_SUMMARY to our temp file

		contents := "### Hello world! :rocket:\n"

		err = Summary(contents)
		test.Ok(t, err)

		written, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		test.Equal(t, string(written), contents)
	})
	t.Run("overwrite existing contents", func(t *testing.T) {
		tmp, err := os.CreateTemp(t.TempDir(), "TestSummary*")
		test.Ok(t, err)
		tmp.Close()

		t.Setenv(summaryFile, tmp.Name()) // Set $TEST_GITHUB_STEP_SUMMARY to our temp file

		err = os.WriteFile(tmp.Name(), []byte("original contents"), filePermissions)
		test.Ok(t, err)

		contents := "# Only Content\n\nShould be nothing else here\n"

		err = Summary(contents)
		test.Ok(t, err)

		written, err := os.ReadFile(tmp.Name())
		test.Ok(t, err)

		test.Equal(t, string(written), contents)
	})
	t.Run("create if not exists", func(t *testing.T) {
		tmpDir := t.TempDir()

		path := filepath.Join(tmpDir, "createme")

		t.Setenv(summaryFile, path) // Set $TEST_GITHUB_STEP_SUMMARY to tempdir/<file that doesn't exist yet>

		contents := "# Markdown\n\nYeah!\n"

		err := Summary(contents)
		test.Ok(t, err)

		written, err := os.ReadFile(path)
		test.Ok(t, err)

		test.Equal(t, string(written), contents)
	})
}
