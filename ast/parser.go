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
	"strconv"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// NewParser returns a participle parser for parsing INI files.
//
// NewParse can panic, but only in case the roundtrip_ini grammar definition
// is wrong. This means that the panic is 100% deterministic: if you have only
// one test that calls NewParser successfully, then newParser will not panic
// in production.
func NewParser() *participle.Parser[AST] {
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

	return participle.MustBuild[AST](
		participle.Lexer(iniLexer),
		participle.Unquote("String"),
		participle.Union[Value](String{}, Number{}),
		participle.UseLookahead(4), // to associate comments with the correct node
	)
}

// AST is the root struct created by the parser.
type AST struct {
	Pos        lexer.Position
	BlankLines []string    `parser:"@NewLine*"`
	Properties []*Property `parser:"@@*"`
	Sections   []*Section  `parser:"@@*"`
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

type namer interface {
	name() string
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
