/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

import (
	"regexp"
	"strings"
	"unicode"
)

var filterRegex = regexp.MustCompile(`[\W_]`)

type Visibility string

const (
	Public   = Visibility("public")
	Internal = Visibility("internal")
)

// IdentifierFactory is a factory for creating Go identifiers from Json schema names
type IdentifierFactory interface {
	CreateIdentifier(name string, visibility Visibility) string
	CreateFieldName(fieldName string, visibility Visibility) FieldName
	CreatePackageNameFromVersion(version string) string
	CreateGroupName(name string) string
	// CreateEnumIdentifier generates the canonical name for an enumeration
	CreateEnumIdentifier(namehint string) string
}

// identifierFactory is an implementation of the IdentifierFactory interface
type identifierFactory struct {
	renames map[string]string
}

// assert the implementation exists
var _ IdentifierFactory = (*identifierFactory)(nil)

// NewIdentifierFactory creates an IdentifierFactory ready for use
func NewIdentifierFactory() IdentifierFactory {
	return &identifierFactory{
		renames: createRenames(),
	}
}

// CreateIdentifier returns a valid Go public identifier
func (factory *identifierFactory) CreateIdentifier(name string, visibility Visibility) string {
	if identifier, ok := factory.renames[name]; ok {
		name = identifier
	}

	// replace with spaces so titlecasing works nicely
	clean := filterRegex.ReplaceAllLiteralString(name, " ")

	result := strings.Title(clean)

	if visibility == Internal {
		// TODO: This is a hack, as there are cases (acronyms, etc) where
		// TODO: this doesn't work...
		// Lowercase the first rune if visibility is internal
		done := false
		result = strings.Map(
			func(r rune) rune {
				if !done {
					done = true
					return unicode.ToLower(r)
				}
				return r
			},
			result)
	}
	result = strings.ReplaceAll(result, " ", "")
	return result
}

func (factory *identifierFactory) CreateFieldName(fieldName string, visibility Visibility) FieldName {
	id := factory.CreateIdentifier(fieldName, visibility)
	return FieldName(id)
}

func createRenames() map[string]string {
	return map[string]string{
		"$schema": "Schema",
	}
}

func (factory *identifierFactory) CreatePackageNameFromVersion(version string) string {
	return "v" + sanitizePackageName(version)
}

func (factory *identifierFactory) CreateGroupName(group string) string {
	return strings.ToLower(group)
}

func (factory *identifierFactory) CreateEnumIdentifier(namehint string) string {
	return factory.CreateIdentifier(namehint, Public)
}

// sanitizePackageName removes all non-alphanum characters and converts to lower case
func sanitizePackageName(input string) string {
	var builder []rune

	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder = append(builder, unicode.ToLower(rune(r)))
		}
	}

	return string(builder)
}

// transformToSnakeCase transforms a string LikeThis to a snake-case string like_this
func transformToSnakeCase(input string) string {
	words := sliceIntoWords(input)

	// my kingdom for LINQ
	var lowerWords []string
	for _, word := range words {
		lowerWords = append(lowerWords, strings.ToLower(word))
	}

	return strings.Join(lowerWords, "_")
}

func simplifyName(context string, name string) string {
	contextWords := sliceIntoWords(context)
	nameWords := sliceIntoWords(name)

	var result []string
	for _, w := range nameWords {
		found := false
		for i, c := range contextWords {
			if c == w {
				found = true
				contextWords[i] = ""
				break
			}
		}
		if !found {
			result = append(result, w)
		}
	}

	if len(result) == 0 {
		return name
	}

	return strings.Join(result, "")
}

func sliceIntoWords(identifier string) []string {
	var result []string
	chars := []rune(identifier)
	lastStart := 0
	for i := range chars {
		preceedingLower := i > 0 && unicode.IsLower(chars[i-1])
		succeedingLower := i+1 < len(chars) && unicode.IsLower(chars[i+1])
		foundUpper := unicode.IsUpper(chars[i])
		if i > lastStart && foundUpper && (preceedingLower || succeedingLower) {
			result = append(result, string(chars[lastStart:i]))
			lastStart = i
		}
	}

	if lastStart < len(chars) {
		result = append(result, string(chars[lastStart:]))
	}

	return result
}
