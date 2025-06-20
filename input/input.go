// Package input provides mechanisms for fetching type safe input from a GitHub Action.
package input // import "go.followtheprocess.codes/actions/input"

import (
	"bufio"
	"bytes"
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
// If the variable is not defined, or if the value is not a valid
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

// List fetches input given as a list of comma-separated or line-separated values.
//
// Each item is stripped of any leading or trailing whitespace prior
// to returning.
//
// If the variable is not defined, or if the value is malformed, an error is returned.
func List(name string) ([]string, error) {
	value, ok := Get(name)
	if !ok {
		return nil, fmt.Errorf("input variable %q not defined", name)
	}

	scanner := bufio.NewScanner(strings.NewReader(value))
	scanner.Split(scanItems)

	var items []string

	for scanner.Scan() {
		item := strings.TrimSpace(scanner.Text())
		items = append(items, item)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("malformed input: %w", err)
	}

	return items, nil
}

// scanItems is a [bufio.SplitFunc] that scans comma-separated or line-separated text.
//
// It is used in [List].
func scanItems(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := bytes.IndexByte(data, ','); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}
