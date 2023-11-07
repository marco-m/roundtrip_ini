// Copyright 2022 Marco Molteni and contributors. All rights reserved.
// Use of this source code is governed by the MIT license; see file LICENSE.

package ast

import (
	"path"
	"strings"
)

// Lookup returns the [Property] associated with keyPath, where keyPath has the
// format
// "section/key". For example:
//   - "foo"     will look for key "foo" in the global section
//   - "bar/foo" will look for key "foo" in section "bar"
//
// If keyPath doesn't exist, Lookup returns nil.
func (tree *AST) Lookup(keyPath string) *Property {
	section, key := path.Split(keyPath)
	section = strings.TrimSuffix(section, "/")
	// Search in the global section.
	if section == "" {
		if i := index(tree.Properties, key); i != -1 {
			return tree.Properties[i]
		}
		return nil
	}
	// Search in the named sections.
	for _, sec := range tree.Sections {
		if sec.Name == section {
			if i := index(sec.Properties, key); i != -1 {
				return sec.Properties[i]
			}
			return nil
		}
	}
	return nil
}

// LookupSection returns the [Section] secName.
// If the section doesn't exist, LookupSection returns nil.
func (tree *AST) LookupSection(secName string) *Section {
	if i := index(tree.Sections, secName); i != -1 {
		return tree.Sections[i]
	}
	return nil
}

// Remove deletes keyPath, where keyPath has the format "section/key".
//
// If keyPath does not exist, Remove does nothing.
func (tree *AST) Remove(keyPath string) {
	section, key := path.Split(keyPath)
	section = strings.TrimSuffix(section, "/")
	// Search in the global section.
	if section == "" {
		if i := index(tree.Properties, key); i != -1 {
			tree.Properties = removeFromSlice(tree.Properties, i)
		}
		return
	}
	// Search in the named sections.
	for _, sec := range tree.Sections {
		if sec.Name == section {
			if i := index(sec.Properties, key); i != -1 {
				sec.Properties = removeFromSlice(sec.Properties, i)
			}
			return
		}
	}
}

// RemoveSection deletes secName and all its properties.
//
// If secName does not exist, RemoveSection does nothing.
func (tree *AST) RemoveSection(secName string) {
	if i := index(tree.Sections, secName); i != -1 {
		tree.Sections = removeFromSlice(tree.Sections, i)
	}
}

// Add replaces the value of keyPath with newVal, where keyPath has the format
// "section/key".
//
// If keyPath does not exist, Add appends the key pair at the end of the
// section. Note that the type of newVal can be different from the previous
// type.
//
// Use [Lookup] beforehand if you need to ensure the presence of keyPath.
func (tree *AST) Add(keyPath string, newVal Value) {
	section, key := path.Split(keyPath)
	section = strings.TrimSuffix(section, "/")

	// Add in the global section.
	if section == "" {
		add(&tree.Properties, key, newVal)
		return
	}

	// Add in the named sections.
	for _, sect := range tree.Sections {
		if sect.Name == section {
			add(&sect.Properties, key, newVal)
			return
		}
	}

	// The section doesn't exist. Create it and add the pair there.
	tree.Sections = append(tree.Sections, &Section{
		Name: section,
		Properties: []*Property{{
			Key:   key,
			Value: newVal,
		}},
	})
}

func add(properties *[]*Property, key string, newVal Value) {
	if i := index(*properties, key); i != -1 {
		// replace
		(*properties)[i].Value = newVal
		return
	}
	// append
	*properties = append(*properties, &Property{
		Key:   key,
		Value: newVal,
	})
	return
}

// index returns the first element of a that matches name.
// If no match, index returns -1.
func index[S ~[]E, E namer](a S, name string) int {
	for i := range a {
		if a[i].name() == name {
			return i
		}
	}
	return -1
}

// Remove the element at index i from slice a. No bounds checks.
// Use it like append:
//
//	x = removeFromSlice(x, 42)
//
// Will not cause memory leaks, safe for slice of pointers
// https://yourbasic.org/golang/delete-element-slice/
// https://github.com/golang/go/wiki/SliceTricks#delete
func removeFromSlice[S ~[]E, E any](a S, i int) S {
	copy(a[i:], a[i+1:])  // Shift a[i+1:] left one index.
	a[len(a)-1] = *new(E) // Erase last element (write zero value).
	a = a[:len(a)-1]      // Truncate slice.
	return a
}
