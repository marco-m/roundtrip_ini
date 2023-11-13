// Copyright 2022 Marco Molteni and contributors. All rights reserved.
// Use of this source code is governed by the MIT license; see file LICENSE.

package ast_test

import (
	"fmt"
	"os"

	"github.com/marco-m/roundtrip_ini/ast"
)

func Example_add_properties() {
	if err := exampleAddProperties(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Output:
	// a = 1
	// c = "x"
	// [s1]
	// b = 2
	// d = 1
	// [s2]
	// e = 7
}

func exampleAddProperties() error {
	parser := ast.NewParser()

	input := `
a = 1
[s1]
b = 2`

	tree, err := parser.ParseString("", input)
	if err != nil {
		return err
	}

	// Add property to global section.
	tree.Add("c", ast.String{S: "x"})

	// Add property to existing named section.
	tree.Add("s1/d", ast.Number{N: 1})

	// Add property to non-existing (yet) named section.
	tree.Add("s2/e", ast.Number{N: 7})

	// Encode
	fmt.Println(tree)
	return nil
}

func Example_add_comments() {
	if err := exampleAddComments(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Output:
	// # comment for a
	// # line 2 for a
	// a = 1
	// [s1]
	// # comment for b
	// b = 2
}

func exampleAddComments() error {
	parser := ast.NewParser()

	input := `
# comment for a
a = 1
[s1]
b = 2`

	tree, err := parser.ParseString("", input)
	if err != nil {
		return err
	}

	// Add a comment
	prop1 := tree.Lookup("s1/b")
	prop1.Comments = []string{"# comment for b"}

	// Add another line to comment.
	prop2 := tree.Lookup("a")
	prop2.Comments = append(prop2.Comments, "# line 2 for a")

	// Encode
	fmt.Println(tree)
	return nil
}

func Example_remove_section() {
	if err := exampleRemoveSection(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Output:
	// # comment for a
	// a = 1
	// [s2]
}

func exampleRemoveSection() error {
	parser := ast.NewParser()

	input := `
# comment for a
a = 1
[s1]
b = 2
[s2]`

	tree, err := parser.ParseString("", input)
	if err != nil {
		return err
	}

	// Remove section and its contents.
	tree.RemoveSection("s1")

	// Encode
	fmt.Println(tree)
	return nil
}
