/*
 * Copyright (c) Microsoft Corporation.
 * Licensed under the MIT license.
 */

package astmodel

// Interface defines an interface implementation
type Interface struct {
	name      TypeName
	functions map[string]Function
}

// NewInterface creates a new interface implementation with the given name and set of functions
func NewInterface(name TypeName, functions map[string]Function) *Interface {
	return &Interface{name: name, functions: functions}
}

// Name returns the name of the interface
func (iface *Interface) Name() TypeName {
	return iface.name
}

// RequiredImports returns a list of packages required by this
func (iface *Interface) RequiredImports() []PackageReference {
	var result []PackageReference

	for _, f := range iface.functions {
		result = append(result, f.RequiredImports()...)
	}

	return result
}

// References indicates whether this type includes any direct references to the given type
func (iface *Interface) References() TypeNameSet {
	var result TypeNameSet

	for _, f := range iface.functions {
		result = SetUnion(result, f.References())
	}

	return result
}

// Equals determines if this interface is equal to another interface
func (iface *Interface) Equals(other *Interface) bool {
	if len(iface.functions) != len(other.functions) {
		return false
	}

	for name, f := range iface.functions {
		otherF, ok := other.functions[name]
		if !ok {
			return false
		}

		if !f.Equals(otherF) {
			return false
		}
	}

	return true
}
