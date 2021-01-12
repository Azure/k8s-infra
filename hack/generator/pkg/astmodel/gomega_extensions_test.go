/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"fmt"

	"github.com/onsi/gomega/format"
	"github.com/onsi/gomega/types"
)

func MatchErrorString(value string) types.GomegaMatcher {
	return ErrorStringMatcher{value}
}

type ErrorStringMatcher struct {
	value string
}

func (m ErrorStringMatcher) Match(actual interface{}) (bool, error) {
	actualErr, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("ErrorString matcher requires an error.\nGot:%s", format.Object(actual, 1))
	}

	return actualErr.Error() == m.value, nil
}

func (m ErrorStringMatcher) FailureMessage(actual interface{}) string {
	return format.Message(actual, "to be an error with value matching", m.value)
}

func (m ErrorStringMatcher) NegatedFailureMessage(actual interface{}) string {
	return format.Message(actual, "to be an error with value that doesn't match", m.value)
}
