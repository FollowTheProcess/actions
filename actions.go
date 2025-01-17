// Package actions provides a toolkit for writing GitHub Actions in Go.
package actions

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// TODO(@FollowTheProcess): This might be better as `package input` with exported functions
// like `Get`, `Bool`, `Lines` etc.?
//
// I'd also like to support the common pattern of lists being either multiline or comma
// separated strings and being able to seamlessly handle either.
//
// If it was `package input`, this would have nice symmetry with `package output` in charge
// of setting outputs from steps, but this might be *too* granular.

// Input gets the value of an actions input variable.
//
// It returns the value of the variable (stripped of leading or trailing whitespace)
// and a boolean which indicates whether it was defined.
func Input(name string) (value string, ok bool) {
	if name == "" {
		return "", false
	}

	cleaned := strings.ReplaceAll(name, " ", "_")
	envName := fmt.Sprintf("INPUT_%s", strings.ToUpper(cleaned))

	value, ok = os.LookupEnv(envName)
	return strings.TrimSpace(value), ok
}

// InputBool gets the boolean value of an actions input variable.
//
// Specifically InputBool supports:
//
//	true | True | TRUE | false | False | FALSE
//
// If the variable is not defined, or if the value is not in the
// supported list, an error is returned.
func InputBool(name string) (bool, error) {
	value, ok := Input(name)
	if !ok {
		return false, fmt.Errorf("input variable %q not defined", name)
	}

	switch value {
	case "true", "True", "TRUE":
		return true, nil
	case "false", "False", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("input variable %q is invalid bool: %q", name, value)
	}
}

// InputLines gets the values of a multiline actions input variable.
//
// Each line is stripped of any leading or trailing whitespace prior
// to returning.
//
// If the variable is not defined, or the input is malformed, an error will be returned.
func InputLines(name string) ([]string, error) {
	value, ok := Input(name)
	if !ok {
		return nil, fmt.Errorf("input variable %q not defined", name)
	}

	scanner := bufio.NewScanner(strings.NewReader(value))
	scanner.Split(bufio.ScanLines)

	var lines []string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		lines = append(lines, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("malformed input: %w", err)
	}

	return lines, nil
}
