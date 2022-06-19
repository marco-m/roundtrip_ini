// Copyright 2022 Marco Molteni and contributors. All rights reserved.
// Use of this source code is governed by the MIT license; see file LICENSE.

// Package ast performs low-level decoding, editing and encoding of the INI
// format, preserving comments and blank lines.
//
// The decoding is based on the participle [INI example].
//
// [INI example]: https://github.com/alecthomas/participle/tree/master/_examples/ini
package ast

import (
	"fmt"
	"path"
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// NewParser returns a participle parser for parsing INI files.
//
// NewParse can panic, but only in case the roundtripini grammar definition is wrong:
// This means that the panic is 100% deterministic: if you have only one test that
// calls NewParser successfully, then newParser will not panic in production.
func NewParser() *participle.Parser {
	rules := []lexer.SimpleRule{
		{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`},
		{`String`, `"(?:\\.|[^"])*"`},
		{`Float`, `\d+(?:\.\d+)?`},
		{`Punct`, `[][=]`},
		{"Comment", `[#;][^\n]*`},
		{"NewLine", `\n`},
		{"whitespace", `[\t ]+`},
	}

	iniLexer := lexer.MustSimple(rules)

	return participle.MustBuild(&AST{},
		participle.Lexer(iniLexer),
		participle.Unquote("String"),
		participle.Union[Value](String{}, Number{}),
		participle.UseLookahead(4), // to associate comments with the correct node
	)
}

// AST is the struct created by the parser.
type AST struct {
	Pos        lexer.Position
	BlankLines []string    `parser:"@NewLine*"`
	Properties []*Property `parser:"@@*"`
	Sections   []*Section  `parser:"@@*"`
}

// Lookup returns the [Property] associated with keyPath, where keyPath has the format
// "section/key". For example:
//  - "foo"     will look for key "foo" in the global section
//  - "bar/foo" will look for key "foo" in section "bar"
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
// If keyPath does not exist, Add appends the key pair at the end of the section.
// Note that the type of newVal can be different from the previous type.
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

// String encodes the AST to the INI format.
func (tree *AST) String() string {
	var bld strings.Builder

	for _, prop := range tree.Properties {
		fmt.Fprint(&bld, prop.String())
	}

	for _, sec := range tree.Sections {
		fmt.Fprint(&bld, sec.String())
	}

	return bld.String()
}

// Property is a key/value pair, with optional metadata for encoding fidelity
// (comment and blank lines).
type Property struct {
	Comments   []string `parser:"(@Comment NewLine)*"`
	Key        string   `parser:"@Ident '='"`
	Value      Value    `parser:"@@ NewLine?"`
	BlankLines []string `parser:"@NewLine*"`
}

// String encodes the Property to the INI format.
func (prop *Property) String() string {
	var bld strings.Builder

	for _, cmt := range prop.Comments {
		fmt.Fprintln(&bld, cmt)
	}

	fmt.Fprintf(&bld, "%s = %s\n", prop.Key, prop.Value)

	for range prop.BlankLines {
		fmt.Fprintln(&bld)
	}

	return bld.String()
}

func (prop *Property) name() string {
	return prop.Key
}

// Value is the value of a INI key.
// Note that it is a union type implemented as a [sealed interface].
//
// Concrete types are associated with participle.Union() in [NewParser].
//
// [sealed interface]: https://github.com/alecthomas/participle#union-types
type Value interface{ value() }

// String is one of the possible types for a Value.
type String struct {
	S string `parser:"@String"`
}

func (s String) value() {} // sealed

func (s String) String() string {
	return fmt.Sprintf("%q", s.S)
}

// Number is one of the possible types for a Value.
type Number struct {
	N float64 `parser:"@Float"`
}

func (nu Number) value() {} // sealed

func (nu Number) String() string {
	return strconv.FormatFloat(nu.N, 'f', -1, 64)
}

// Section is a INI file section, with optional metadata for encoding fidelity
// (comment and blank lines).
type Section struct {
	Comments   []string    `parser:"(@Comment NewLine)*"`
	Name       string      `parser:"'[' @Ident ']' NewLine?"`
	BlankLines []string    `parser:"@NewLine*"`
	Properties []*Property `parser:"@@*"`
}

// String encodes the Section to the INI format.
func (sec *Section) String() string {
	var bld strings.Builder

	for _, cmt := range sec.Comments {
		fmt.Fprintln(&bld, cmt)
	}

	fmt.Fprintf(&bld, "[%s]\n", sec.Name)

	for range sec.BlankLines {
		fmt.Fprintln(&bld)
	}

	for _, prop := range sec.Properties {
		fmt.Fprint(&bld, prop.String())
	}

	return bld.String()
}

func (sec *Section) name() string {
	return sec.Name
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

type namer interface {
	name() string
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

// Remove the element at index i from a. No bounds checks.
// Use it like append:
//
//     x = removeFromSlice(x, 42)
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
