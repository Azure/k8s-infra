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

// \W is all non-word characters (https://golang.org/pkg/regexp/syntax/)
var filterRegex = regexp.MustCompile(`[\W_]`)

type Visibility string

const (
	Exported    = Visibility("exported")
	NotExported = Visibility("notexported")
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
	// wholeRenames is a map of custom transformations to apply to entire identifiers
	wholeRenames map[string]string
	// partRenames is a map of custom transformations to apply to parts of identifiers
	partRenames map[string]string
}

// assert the implementation exists
var _ IdentifierFactory = (*identifierFactory)(nil)

// NewIdentifierFactory creates an IdentifierFactory ready for use
func NewIdentifierFactory() IdentifierFactory {
	return &identifierFactory{
		wholeRenames: createWholeRenames(),
		partRenames:  createPartRenames(),
	}
}

// CreateIdentifier returns a valid Go identifier with the specified visibility
func (factory *identifierFactory) CreateIdentifier(name string, visibility Visibility) string {
	// Apply a whole rename if we have one defined
	if identifier, ok := factory.wholeRenames[name]; ok {
		return factory.adjustVisibility(identifier, visibility)
	}

	// Apply part renaming
	var parts []string
	for _, w := range factory.sliceIntoParts(name) {
		if id, ok := factory.partRenames[w]; ok {
			parts = append(parts, id)
		} else {
			parts = append(parts, w)
		}
	}

	// Divide parts further into words (thisOrThat => this, Or, That)
	var cleanWords []string
	for _, w := range parts {
		cleanWords = append(cleanWords, sliceIntoWords(w)...)
	}

	// Correct the letter case for each word
	var caseCorrectedWords []string
	for i, word := range cleanWords {

		if visibility == NotExported && i == 0 {
			caseCorrectedWords = append(caseCorrectedWords, strings.ToLower(word))
		} else {
			caseCorrectedWords = append(caseCorrectedWords, strings.Title(word))
		}
	}

	return strings.Join(caseCorrectedWords, "")
}

// sliceIntoParts divides the name into parts based on punctuation and other symbols
func (factory *identifierFactory) sliceIntoParts(name string) []string {
	// replace with spaces so title-casing works nicely
	clean := filterRegex.ReplaceAllLiteralString(name, " ")

	// Split into parts and check each one for a rename
	result := strings.Split(clean, " ")
	return result
}

func (factory *identifierFactory) adjustVisibility(identifier string, visibility Visibility) string {
	// Just lowercase the first character according to visibility
	r := []rune(identifier)
	if visibility == NotExported {
		r[0] = unicode.ToLower(r[0])
	} else {
		r[0] = unicode.ToUpper(r[0])
	}

	return string(r)
}

func (factory *identifierFactory) CreateFieldName(fieldName string, visibility Visibility) FieldName {
	id := factory.CreateIdentifier(fieldName, visibility)
	return FieldName(id)
}

func (factory *identifierFactory) CreatePackageNameFromVersion(version string) string {
	return "v" + sanitizePackageName(version)
}

func (factory *identifierFactory) CreateGroupName(group string) string {
	return strings.ToLower(group)
}

func (factory *identifierFactory) CreateEnumIdentifier(namehint string) string {
	return factory.CreateIdentifier(namehint, Exported)
}

// sanitizePackageName removes all non-alphanum characters and converts to lower case
func sanitizePackageName(input string) string {
	var builder []rune

	for _, r := range input {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			builder = append(builder, unicode.ToLower(r))
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
	// Trim any leading and trailing spaces to make our life easier later
	identifier = strings.Trim(identifier, " ")

	var result []string
	chars := []rune(identifier)
	lastStart := 0
	for i := range chars {
		preceedingLower := i > 0 && unicode.IsLower(chars[i-1])
		succeedingLower := i+1 < len(chars) && unicode.IsLower(chars[i+1])
		isSpace := unicode.IsSpace(chars[i])
		foundUpper := unicode.IsUpper(chars[i])
		if i > lastStart && foundUpper && (preceedingLower || succeedingLower) {
			result = append(result, string(chars[lastStart:i]))
			lastStart = i
		} else if isSpace {
			r := string(chars[lastStart:i])
			r = strings.Trim(r, " ")
			// If r is entirely spaces... just don't append anything
			if len(r) != 0 {
				result = append(result, r)
			}
			lastStart = i + 1 // skip the space
		}
	}

	if lastStart < len(chars) {
		result = append(result, string(chars[lastStart:]))
	}

	return result
}

// createWholeRenames returns a map used to customize the transformation of entire identifiers
// Keys should be the identifier exactly as it appears in the JSON schema
// Values should be the Go Identifier to use
// Method positioned at end of file so it's easy to find and amend with new renames as required
func createWholeRenames() map[string]string {
	return map[string]string{
		"$schema": "Schema",
	}
}

// createPartRenames returns a map used to customize the transformation of the parts of identifiers between underscores ('_')
// Keys should be an identifier part that's used in multiple underscore_separated_identifiers.
// Values should be the replacement to substitute
// Method positioned at end of file so it's easy to find and amend with new renames as required
func createPartRenames() map[string]string {
	return map[string]string{
		"batchAccounts": "batchAccount",
	}
}
