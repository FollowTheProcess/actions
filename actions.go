// Package actions provides a toolkit for writing GitHub Actions in Go.
package actions

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"strings"
)

// default file permissions in case of creation, we shouldn't ever need to create
// either $GITHUB_ENV or $GITHUB_OUTPUT, but we have to pass something to OpenFile so.
const filePermissions = 0o644

var (
	// We override both of these in tests because we need to test things like what happens if
	// the file doesn't exist, we can't write to it etc. and we have limited ability to do that
	// in actual GitHub Actions workflows, so must use our own test variable.

	// envFile is the name of the env var containing the filepath of the special GitHub file
	// to which env vars should be written.
	envFile = "GITHUB_ENV"

	// outFile is the name of the env var containing the filepath to the special GitHub file
	// to which output variables should be written.
	outFile = "GITHUB_OUTPUT"

	// summaryFile is the name of the env var containing the filepath to the special GitHub file
	// to which step summaries should be written.
	summaryFile = "GITHUB_STEP_SUMMARY"

	// pathFile is the name of the env var containing the filepath to the special GitHub file
	// that holds the value of `$PATH`. By writing to it, you can prepend programs to `$PATH`,
	// installing them.
	pathFile = "GITHUB_PATH"

	// realPath is the $PATH env var, we can't really mess with this in tests so this is here
	// so we have something to manipulate without breaking things.
	realPath = "PATH"
)

// GetEnv retrieves the value of a named environment variable written to $GITHUB_ENV.
//
// If the variable is present in the environment the value (which may be empty) is returned and
// the boolean is true. Otherwise the returned value will be empty and the boolean will be false.
func GetEnv(key string) (value string, ok bool) {
	return os.LookupEnv(key)
}

// SetEnv sets an environment variable by writing it to $GITHUB_ENV.
//
// If the value contains newlines, SetEnv will use the "EOF" pattern to correctly
// set multiline values with the delimiter being a randomly generated string,
// minimising the chance of collision with the contents.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#multiline-strings.
//
// Attempting to set $GITHUB_*, $RUNNER_*, $CI or $NODE_OPTIONS is not allowed and will
// return an error.
func SetEnv(key, value string) error {
	return setVarFile(envFile, key, value)
}

// SetOutput sets an output variable by writing it to $GITHUB_OUTPUT.
//
// If the value contains newlines, SetOutput will use the "EOF" pattern to correctly
// set multiline values with the delimiter being a randomly generated string,
// minimising the chance of collision with the contents.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#multiline-strings.
func SetOutput(key, value string) error {
	return setVarFile(outFile, key, value)
}

// AddPath prepends path to $GITHUB_PATH and does the same with the actual $PATH variable.
//
// This is how things are installed into github actions runners, the binaries are placed
// on $PATH (and $GITHUB_PATH).
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#adding-a-system-path
func AddPath(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return errors.New("cannot set an empty path")
	}

	githubPathFile := os.Getenv(pathFile)
	if githubPathFile == "" {
		return errors.New("$GITHUB_PATH is not set or is empty")
	}

	file, err := os.OpenFile(githubPathFile, os.O_APPEND|os.O_WRONLY, filePermissions)
	if err != nil {
		return fmt.Errorf("could not open $GITHUB_PATH file %s: %w", githubPathFile, err)
	}
	defer file.Close()

	fmt.Fprintf(file, "%s\n", path)

	// Set $PATH
	newPath := fmt.Sprintf("%s%s%s", path, string(os.PathListSeparator), os.Getenv(realPath))
	if err := os.Setenv(realPath, newPath); err != nil {
		return fmt.Errorf("could not update $PATH: %w", err)
	}

	return nil
}

// Summary writes a step summary to $GITHUB_STEP_SUMMARY, creating the backing file if necessary.
//
// GitHub flavoured markdown content is supported. Subsequent calls to Summary overwrite the contents.
//
// For writing complex markdown or html summaries, consider using [html/template] to format your content
// as desired, then passing the rendered template to Summary.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#adding-a-job-summary
//
// [html/template]: https://pkg.go.dev/html/template
func Summary(contents string) error {
	path := os.Getenv(summaryFile)
	if path == "" {
		return fmt.Errorf("$%s is not set or is empty", summaryFile)
	}

	// Write the contents to the file, creating it if necessary, overwriting it if
	// called again
	if err := os.WriteFile(path, []byte(contents), filePermissions); err != nil {
		return fmt.Errorf("could not write to $%s at path %s: %w", summaryFile, path, err)
	}

	return nil
}

// setVarFile sets either a GITHUB_ENV or GITHUB_OUTPUT file variable. The process
// is largely the same for each.
func setVarFile(name, key, value string) error {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "" {
		return errors.New("key cannot be empty")
	}

	if value == "" {
		return errors.New("value cannot be empty")
	}

	if name == envFile {
		// If it's GITHUB_ENV, we aren't allowed to play with these env vars
		if key == "CI" || key == "NODE_OPTIONS" || strings.HasPrefix(key, "GITHUB_") ||
			strings.HasPrefix(key, "RUNNER_") {
			return fmt.Errorf("setting $%s is disallowed", key)
		}
	}

	path := os.Getenv(name)
	if path == "" {
		return fmt.Errorf("$%s is not set or is empty", name)
	}

	// Append to the file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, filePermissions)
	if err != nil {
		return fmt.Errorf("could not open $%s file: %w", outFile, err)
	}
	defer file.Close()

	// If the value is multi-line, do the whole EOF delimiter thing, but with
	// a random string to make pretty sure it never collides with file content
	if strings.Contains(value, "\n") {
		delimiter := "ghadelimiter_" + randString()
		fmt.Fprintf(file, "%s<<%s\n%s\n%s", key, delimiter, value, delimiter)
	} else {
		fmt.Fprintf(file, "%s=%s\n", key, value)
	}

	// If it's an env var, let's export the actual env var too
	if name == envFile {
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set $%s: %w", key, err)
		}
	}

	return nil
}

// randString produces a random string of 16 characters.
func randString() string {
	const (
		charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		size    = 16
	)

	b := make([]byte, size)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))] //nolint: gosec // This is not for security purposes
	}

	return string(b)
}
