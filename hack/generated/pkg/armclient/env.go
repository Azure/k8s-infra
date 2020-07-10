/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package armclient

import (
	"os"
)

type (
	Enver interface {
		GetEnv(key string) string
	}

	stdEnv struct{}
)

// GetEnv will return os.GetEnv for a given key
func (*stdEnv) GetEnv(key string) string {
	return os.Getenv(key)
}