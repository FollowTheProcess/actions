// Package actions provides a toolkit for writing GitHub Actions in Go.
package actions

import (
	"fmt"
	"os"
	"strings"
)

// Input gets the value of an actions input variable.
//
// It returns the value of the variable (stripped of leading or trailing whitespace)
// and a boolean which indicates whether it was defined.
func Input(name string) (value string, ok bool) {
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
		return false, fmt.Errorf("input variable %s not defined", name)
	}

	switch value {
	case "true", "True", "TRUE":
		return true, nil
	case "false", "False", "FALSE":
		return false, nil
	default:
		return false, fmt.Errorf("input variable %s is invalid bool: %q", name, value)
	}
}
