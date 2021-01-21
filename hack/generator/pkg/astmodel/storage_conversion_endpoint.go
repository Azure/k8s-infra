package astmodel

import (
	"github.com/dave/dst"
	"github.com/gobuffalo/flect"
	"strconv"
	"strings"
	"unicode"
)

type KnownLocalsSet map[string]struct{}

// StorageConversionEndpoint represents either a source or a destination for a storage conversion
type StorageConversionEndpoint struct {
	// theType is the Type of the value accessible via this endpoint
	theType Type
	// name is the name of the underlying property, used to generate useful local identifiers
	name string
	// knownLocals is a shared map of locals that have already been created, to prevent duplicates
	knownLocals KnownLocalsSet
}

func NewStorageConversionEndpoint(theType Type, name string, knownLocals KnownLocalsSet) *StorageConversionEndpoint {
	return &StorageConversionEndpoint{
		theType:     theType,
		name:        strings.ToLower(name),
		knownLocals: knownLocals,
	}
}

// Expr returns a clone of the expression for this endpoint
// (returning a clone avoids issues with reuse of fragments within the dst)
func (endpoint *StorageConversionEndpoint) Type() Type {
	return endpoint.theType
}

// Create an identifier for a local variable that represents a single item
// Each call will return a unique identifier
func (endpoint *StorageConversionEndpoint) CreateSingularLocal() *dst.Ident {
	singular := flect.Singularize(endpoint.name)
	id := endpoint.knownLocals.createLocal(singular)
	return dst.NewIdent(id)
}

// CreatePluralLocal creates an identifier for a local variable that represents multiple items
// Each call will return a unique identifier
func (endpoint *StorageConversionEndpoint) CreatePluralLocal(suffix string) *dst.Ident {
	id := endpoint.knownLocals.createLocal(endpoint.name + suffix)
	return dst.NewIdent(id)
}

// WithType creates a new endpoint with a different type
func (endpoint *StorageConversionEndpoint) WithType(theType Type) *StorageConversionEndpoint {
	return NewStorageConversionEndpoint(
		theType,
		endpoint.name,
		endpoint.knownLocals)
}

// createLocal creates a new unique local with the specified suffix
func (locals KnownLocalsSet) createLocal(nameHint string) string {
	name := locals.toPrivate(nameHint)
	id := name
	_, found := locals[id]

	index := 0
	for found {
		index++
		id = name + strconv.Itoa(index)
		_, found = locals[id]
	}

	locals[id] = struct{}{}

	return id
}

func (locals KnownLocalsSet) Add(local string) {
	name := locals.toPrivate(local)
	locals[name] = struct{}{}
}

func (locals KnownLocalsSet) toPrivate(s string) string {
	// Just lowercase the first character
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
