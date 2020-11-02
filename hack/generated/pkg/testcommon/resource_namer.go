/*
Copyright (c) Microsoft Corporation.
Licensed under the MIT license.
*/

package testcommon

import (
	"hash/fnv"
	"math/rand"
	"strings"
)

type ResourceNameConfig struct {
	runes       []rune
	prefix      string
	randomChars int
	separator   string
}

type ResourceNamer struct {
	ResourceNameConfig
	rand *rand.Rand
}

func (rnc ResourceNameConfig) NewResourceNamer(name string) ResourceNamer {
	hasher := fnv.New64()
	hasher.Write([]byte(name))
	seed := hasher.Sum64()
	return ResourceNamer{
		ResourceNameConfig: rnc,
		// nolint: do not want cryptographic randomness here
		rand: rand.New(rand.NewSource(int64(seed))),
	}
}

func NewResourceNameConfig(prefix string, separator string, randomChars int) *ResourceNameConfig {
	return &ResourceNameConfig{
		runes:       []rune("abcdefghijklmnopqrstuvwxyz"),
		prefix:      prefix,
		randomChars: randomChars,
		separator:   separator,
	}
}

func (n ResourceNameConfig) WithSeparator(separator string) *ResourceNameConfig {
	n.separator = separator
	return &n
}

func (n ResourceNamer) WithSeparator(separator string) ResourceNamer {
	n.separator = separator
	return n
}

func (n ResourceNamer) generateName(prefix string, num int) string {
	result := make([]rune, num)
	for i := 0; i < num; i++ {
		result[i] = n.runes[n.rand.Intn(len(n.runes))]
	}

	var s []string
	if prefix != "" {
		s = []string{n.prefix, prefix, string(result)}
	} else {
		s = []string{n.prefix, string(result)}
	}

	return strings.Join(s, n.separator)
}

func (n ResourceNamer) GenerateName(prefix string) string {
	return n.generateName(prefix, n.randomChars)
}
