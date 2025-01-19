// Package input provides mechanisms for fetching type safe input from a GitHub Action.
package input

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Get gets the value of an actions input variable.
//
// It returns the string value of the variable (stripped of any leading and trailing whitespace)
// and a boolean which indicates whether it was defined.
func Get(name string) (value string, ok bool) {
	if name == "" {
		return "", false
	}

	cleaned := strings.ReplaceAll(name, " ", "_")
	envName := "INPUT_" + strings.ToUpper(cleaned)

	value, ok = os.LookupEnv(envName)
	if !ok {
		return "", false
	}

	return strings.TrimSpace(value), true
}

// Bool gets the boolean value of an actions input variable.
//
// Specifically Bool supports:
//
//	true | True | TRUE | false | False | FALSE
//
// If the variable is not defined, or if the value is not in the
// supported list, an error is returned.
func Bool(name string) (bool, error) {
	value, ok := Get(name)
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

// Lines gets the values of a multiline actions input variable.
//
// Each line is stripped of any leading or trailing whitespace prior
// to returning.
//
// If the variable is not defined, or the input is malformed, an error will be returned.
func Lines(name string) ([]string, error) {
	value, ok := Get(name)
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

// Int gets the integer value of an actions input variable.
//
// If the variable is not defined, or if the value is not a valid
// integer, an error is returned.
func Int(name string) (int, error) {
	value, ok := Get(name)
	if !ok {
		return 0, fmt.Errorf("input variable %q not defined", name)
	}

	val, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("input variable %q is invalid integer: %q", name, value)
	}

	return val, nil
}

// Float gets the float value of an actions input variable.
//
// if the variable is not defined, or if the value is not a valid
// float, an error is returned.
func Float(name string) (float64, error) {
	value, ok := Get(name)
	if !ok {
		return 0, fmt.Errorf("input variable %q not defined", name)
	}

	val, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("input variable %q is invalid float: %q", name, value)
	}

	return val, nil
}

// TODO(@FollowTheProcess): Support comma separated or line separated lists
