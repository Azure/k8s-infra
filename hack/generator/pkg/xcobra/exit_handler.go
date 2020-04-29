/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

//+build !noexit

package xcobra

import (
	"os"
)

func exitWithCode(err error) {
	if err == nil {
		return
	}

	if e, ok := err.(ErrorWithCode); ok {
		os.Exit(e.Code)
	}
	os.Exit(1)
}
