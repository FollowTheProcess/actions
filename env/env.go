// Package env provides functionality for interacting with GitHub [Environment files].
//
// By writing to and reading from $GITHUB_ENV, steps in a GitHub Actions Workflow may pass simple information
// between steps, or set global configuration.
//
// For passing specific information between workflow steps and/or jobs prefer the output package.
//
// [Environment files]: https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#environment-files
package env

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"strings"
)

// default file permissions in case of creation.
const filePermissions = 0o644

// envFile is the name of the env var containing the filepath of the special GitHub file
// to which env vars should be written.
var envFile = "GITHUB_ENV"

// Get retrieves the value of a named environment variable written
// to $GITHUB_ENV.
//
// If the variable is present in the environment the value (which may be empty) is returned and
// the boolean is true. Otherwise the returned value will be empty and the boolean will be false.
func Get(key string) (value string, ok bool) {
	return os.LookupEnv(key)
}

// TODO(@FollowTheProcess): Multiline stuff as per https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#multiline-strings

// Set sets an environment variable by writing it to $GITHUB_ENV.
//
// If the value contains newlines, Set will use the "EOF" pattern to
// correctly set multiline env vars with the delimiter being a randomly
// generated string, minimising the chance of collision with the contents.
//
// See https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/workflow-commands-for-github-actions#multiline-strings.
//
// Attempting to set $GITHUB_*, $RUNNER_*, $CI or $NODE_OPTIONS is not allowed and will
// return an error.
func Set(key, value string) error {
	if key == "" {
		return errors.New("key cannot be empty")
	}
	if value == "" {
		return errors.New("value cannot be empty")
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)

	if key == "CI" || key == "NODE_OPTIONS" || strings.HasPrefix(key, "GITHUB_") || strings.HasPrefix(key, "RUNNER_") {
		return fmt.Errorf("setting $%s is disallowed", key)
	}
	path, ok := os.LookupEnv(envFile)
	if !ok {
		return fmt.Errorf("$%s is not set", envFile)
	}

	// Append to the file
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, filePermissions)
	if err != nil {
		return fmt.Errorf("could not open $%s file: %w", envFile, err)
	}
	defer file.Close()

	// If the value is multi-line, do the whole EOF delimiter thing, but with
	// a random string to make pretty sure it never collides with file content
	if strings.Contains(value, "\n") {
		delimiter := fmt.Sprintf("ghadelimiter_%s", randString())
		fmt.Fprintf(file, "%s<<%s\n%s\n%s", key, delimiter, value, delimiter)
	} else {
		fmt.Fprintf(file, "%s=%s\n", key, value)
	}

	os.Setenv(key, value) // Set it in the actual environment too
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
		b[i] = charset[rand.Intn(len(charset))]
	}

	return string(b)
}
