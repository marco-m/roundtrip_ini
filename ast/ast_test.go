// Copyright 2022 Marco Molteni and contributors. All rights reserved.
// Use of this source code is governed by the MIT license; see file LICENSE.

// This file tests the low-level parser.

package ast_test

import (
	"strings"
	"testing"

	"gotest.tools/v3/assert"

	"github.com/marco-m/roundtrip_ini/ast"
)

func TestGrammarIsWellFormed(t *testing.T) {
	parser := ast.NewParser()

	assert.Assert(t, parser != nil)
}

// Parse input and return the AST.
func parse(t *testing.T, input string) *ast.AST {
	t.Helper()

	parser := ast.NewParser()

	tree, err := parser.ParseString("", input)
	assert.NilError(t, err)

	return tree
}

// Assert that prop has key k and value v, where v is a string.
func checkKeyString(t *testing.T, prop *ast.Property, k string, v string) {
	t.Helper()
	assert.Equal(t, prop.Key, k)
	value, ok := prop.Value.(ast.String)
	assert.Assert(t, ok)
	assert.Equal(t, value.S, v)
}

// Assert that prop has key k and value v, where v is a float64.
func checkKeyFloat(t *testing.T, prop *ast.Property, k string, v float64) {
	t.Helper()
	assert.Equal(t, prop.Key, k)
	value, ok := prop.Value.(ast.Number)
	assert.Assert(t, ok)
	assert.Equal(t, value.N, v)
}

func TestParseKeyValueWithString(t *testing.T) {
	input := `name = "Johnny Stecchino"`

	tree := parse(t, input)

	checkKeyString(t, tree.Properties[0], "name", "Johnny Stecchino")
}

func TestParseKeyValueWithNumbersMultipleLines(t *testing.T) {
	input := `
age = 21
score = 1.2`

	tree := parse(t, input)

	checkKeyFloat(t, tree.Properties[0], "age", 21)
	checkKeyFloat(t, tree.Properties[1], "score", 1.2)
}

func TestParseSections(t *testing.T) {
	input := `
[address]
city = "Milan"`

	tree := parse(t, input)

	assert.Equal(t, tree.Sections[0].Name, "address")
	checkKeyString(t, tree.Sections[0].Properties[0], "city", "Milan")
}

func TestLookupPropertyFound(t *testing.T) {
	input := `
[address]
city = "Milan"`
	tree := parse(t, input)

	prop := tree.Lookup("address/city")

	assert.Assert(t, prop != nil)
	checkKeyString(t, prop, "city", "Milan")
}

func TestLookupPropertyNotFound(t *testing.T) {
	input := `
[address]
city = "Milan"`
	tree := parse(t, input)

	prop := tree.Lookup("address/town")

	assert.Assert(t, prop == nil)
}

func TestLookupSectionFound(t *testing.T) {
	input := `
[foo]
[address]
city = "Milan"`
	tree := parse(t, input)

	sect := tree.LookupSection("address")

	assert.Assert(t, sect != nil)
	assert.Equal(t, sect.Name, "address")
}

func TestLookupSectionNotFound(t *testing.T) {
	input := `
hello = 2
[address]
city = "Milan"`
	tree := parse(t, input)

	sect := tree.LookupSection("hello")

	assert.Assert(t, sect == nil)
}

func TestParserKeepsPropertyCommentsAndBlanks(t *testing.T) {
	input := `
# comment for name
name = "Bob Smith"
; comment for foo
foo = "bar"
# comment and blank below
fruit = "banana"

another_blank = 2


[address]
# comment line 1
# comment line 2
city = "Venezia"`

	testCases := []struct {
		name           string
		keyPath        string
		wantComments   []string
		wantBlankLines []string
	}{
		{
			name:         "comment # single line",
			keyPath:      "name",
			wantComments: []string{"# comment for name"},
		},
		{
			name:         "comment ; single line",
			keyPath:      "foo",
			wantComments: []string{"; comment for foo"},
		},
		{
			name:         "comment multiple lines",
			keyPath:      "address/city",
			wantComments: []string{"# comment line 1", "# comment line 2"},
		},
		{
			name:           "comment and blank",
			keyPath:        "fruit",
			wantComments:   []string{"# comment and blank below"},
			wantBlankLines: []string{"\n"},
		},
		{
			name:           "multiple blanks are reported correctly",
			keyPath:        "another_blank",
			wantBlankLines: []string{"\n", "\n"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := parse(t, input)

			prop := tree.Lookup(tc.keyPath)
			assert.Assert(t, prop != nil)
			assert.DeepEqual(t, prop.Comments, tc.wantComments)
			assert.DeepEqual(t, prop.BlankLines, tc.wantBlankLines)
		})
	}
}

func TestParserKeepsSectionCommentsAndBlanks(t *testing.T) {
	input := `
# comment for name
name = "Bob Smith"
# comment for address
[address]
; comment for plants
[plants]
# comment line 1
# comment line 2
[fruits]

bananas = 42
[algae]


x = 1
`

	testCases := []struct {
		name           string
		secName        string
		wantComments   []string
		wantBlankLines []string
	}{
		{
			name:         "comment # single line",
			secName:      "address",
			wantComments: []string{"# comment for address"},
		},
		{
			name:         "comment ; single line",
			secName:      "plants",
			wantComments: []string{"; comment for plants"},
		},
		{
			name:           "comment multiple lines and blank",
			secName:        "fruits",
			wantComments:   []string{"# comment line 1", "# comment line 2"},
			wantBlankLines: []string{"\n"},
		},
		{
			name:           "multiple blanks are correctly reported",
			secName:        "algae",
			wantBlankLines: []string{"\n", "\n"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := parse(t, input)

			sect := tree.LookupSection(tc.secName)
			assert.Assert(t, sect != nil)
			assert.DeepEqual(t, sect.Comments, tc.wantComments)
			assert.DeepEqual(t, sect.BlankLines, tc.wantBlankLines)
		})
	}
}

// Useful to preprocess tc.want
// Remove leading newline and add trailing newline.
func normalizeEnds(s string) string {
	return strings.TrimSpace(s) + "\n"
}

func TestRoundTripNoEdits(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name: "no comments, no blank lines",
			input: `
a = 1
x = "A"
[s1]
b = 2
y = "B"
[s2]
c = 3
z = "C"`,
		},
		{
			name: "single line comments",
			input: `
a = 1
# property comment
x = "A"
# section comment
[s1]
b = 2`,
		},
		{
			name: "multiline comments",
			input: `
a = 1
# property comment line 1
# property comment line 2
x = "A"
# section comment line 1
# section comment line 2
[s1]
b = 2`,
		},
		{
			name: "property single blank lines",
			input: `
a = 1

b = 2`,
		},
		{
			name: "section single blank lines",
			input: `
[fruits]

bananas = 42`,
		},
		{
			name: "property multiple blank lines",
			input: `
a = 1


b = 2`,
		},
		{
			name: "section multiple blank lines",
			input: `
[fruits]


bananas = 42`,
		},
		{
			name: "mix",
			input: `
# property comment line 1
x = "A"

# section comment line 1
# section comment line 2
[s1]

b = 2`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.input = normalizeEnds(tc.input)
			tree := parse(t, tc.input)

			have := tree.String()

			assert.Equal(t, have, tc.input)
		})
	}
}

// Document the current behavior; this might change because I am not sure what is
// the best to do :-(
// Maybe should have a flag like "NormalizeLeadingTrailing" ?
func TestRoundTripCornerCases(t *testing.T) {
	type testCase struct {
		name  string
		input string
		want  string
	}

	run := func(t *testing.T, tc testCase) {
		tree := parse(t, tc.input)
		have := tree.String()

		assert.Equal(t, have, tc.want)
	}

	testCases := []testCase{
		{
			name:  "global section, missing trailing newline is added to output",
			input: "a = 1",
			want:  "a = 1" + "\n",
		},
		{
			name:  "global section, spurious leading newline is removed",
			input: "\n" + "a = 1\n",
			want:  "a = 1\n",
		},
		{
			name:  "named section, missing trailing newline is added to output",
			input: "[s1]",
			want:  "[s1]" + "\n",
		},
		{
			name:  "named section, spurious leading newline is removed",
			input: "\n" + "[s1]\n",
			want:  "[s1]\n",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) { run(t, tc) })
	}
}

func TestRoundTripPrettyPrint(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		{
			input: `
b=2`,
			want: `
b = 2`,
		},
		{
			input: `
[ s1 ]
b =2`,
			want: `
[s1]
b = 2`,
		},
	}

	for _, tc := range testCases {
		t.Run("", func(t *testing.T) {
			tree := parse(t, tc.input)
			tc.want = normalizeEnds(tc.want)

			have := tree.String()
			assert.Equal(t, have, tc.want)
		})
	}
}

func TestAdd(t *testing.T) {
	testCases := []struct {
		name  string
		input string
		want  string
		key   string
		value ast.Value
	}{
		{
			name:  "update property, global section",
			input: `foo = "bar"`,
			want:  `foo = "zoo"`,
			key:   "foo",
			value: ast.String{S: "zoo"},
		},
		{
			name: "update property, named section",
			input: `
hello = 2
[fruits]
foo = "bar"`,
			want: `
hello = 2
[fruits]
foo = "zoo"`,
			key:   "fruits/foo",
			value: ast.String{S: "zoo"},
		},
		{
			name: "append property, global section",
			input: `
foo = "bar"
[section1]
k = "v"`,
			want: `
foo = "bar"
new = "yes"
[section1]
k = "v"`,
			key:   "new",
			value: ast.String{S: "yes"},
		},
		{
			name: "append property, named section",
			input: `
g1 = 1
[section1]
s1 = 2`,
			want: `
g1 = 1
[section1]
s1 = 2
new = "yes"`,
			key:   "section1/new",
			value: ast.String{S: "yes"},
		},
		{
			name: "create new section and append key there",
			input: `
[s1]
a = 1`,
			want: `
[s1]
a = 1
[s2]
b = 2`,
			key:   "s2/b",
			value: ast.Number{N: 2},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := parse(t, tc.input)
			tc.want = normalizeEnds(tc.want)

			tree.Add(tc.key, tc.value)
			have := tree.String()

			assert.Equal(t, have, tc.want)
		})
	}

}

func TestRemove(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		want   string
		remove string
	}{
		{
			name:   "non-existing is not an error",
			input:  `hello = 1`,
			want:   `hello = 1`,
			remove: "foo",
		},
		{
			name: "does not remove a section with the same name of the key",
			input: `
[s1]
a = 1`,
			want: `
[s1]
a = 1`,
			remove: "s1",
		},
		{
			name: "remove from global section",
			input: `
a = 1
b = 2
c = 3`,
			want: `
a = 1
c = 3`,
			remove: "b",
		},
		{
			name: "remove from global section with comments",
			input: `
# comment for a
a = 1
# comment for b
b = 2
# comment for c
c = 3`,
			want: `
# comment for a
a = 1
# comment for c
c = 3`,
			remove: "b",
		},
		{
			name: "remove from global section with comments and newlines",
			input: `
a = 1

# comment for b
b = 2

c = 3`,
			want: `
a = 1

c = 3`,
			remove: "b",
		},
		{
			name: "remove from named section",
			input: `
a = 1
[s1]
b = 2
c = 3

[s2]
`,
			want: `
a = 1
[s1]
c = 3

[s2]`,
			remove: "s1/b",
		},
		{
			name: "leaves section alone also if last element",
			input: `
[s1]
a = 1`,
			want: `
[s1]`,
			remove: "s1/a",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := parse(t, tc.input)
			tc.want = normalizeEnds(tc.want)

			tree.Remove(tc.remove)
			have := tree.String()

			assert.Equal(t, have, tc.want)
		})
	}
}

func TestRemoveSection(t *testing.T) {
	testCases := []struct {
		name   string
		input  string
		want   string
		remove string
	}{
		{
			name: "non-existing section is not an error",
			input: `
a = 1
[s1]
b = 2`,
			want: `
a = 1
[s1]
b = 2`,
			remove: "s2",
		},
		{
			name: "does not remove a key with the same Name of the section",
			input: `
a = 1
[s1]
b = 2`,
			want: `
a = 1
[s1]
b = 2`,
			remove: "a",
		},
		{
			name: "remove empty section",
			input: `
a = 1
[s1]
[s2]
b = 2`,
			want: `
a = 1
[s2]
b = 2`,
			remove: "s1",
		},
		{
			name: "remove non empty section",
			input: `
a = 1
# comment for s1
[s1]
b = 2
# comment for s2
[s2]
c = 3`,
			want: `
a = 1
# comment for s2
[s2]
c = 3`,
			remove: "s1",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tree := parse(t, tc.input)
			tc.want = normalizeEnds(tc.want)

			tree.RemoveSection(tc.remove)
			have := tree.String()

			assert.Equal(t, have, tc.want)
		})
	}
}

func TestAddCommentViaLookup(t *testing.T) {
	type testCase struct {
		name     string
		input    string
		want     string
		path     string
		value    ast.Value
		comments []string
	}

	test := func(t *testing.T, tc testCase) {
		tree := parse(t, tc.input)
		tc.want = normalizeEnds(tc.want)

		prop := tree.Lookup(tc.path)
		if prop == nil {
			// FIXME Add should return the prop ?
			tree.Add(tc.path, tc.value)
			prop = tree.Lookup(tc.path)
		}
		assert.Assert(t, prop != nil)
		prop.Comments = tc.comments

		have := tree.String()
		assert.Equal(t, have, tc.want)
	}

	testCases := []testCase{
		{
			name: "add comment to existing property",
			input: `
a = 1`,
			want: `
# comment for a
a = 1`,
			path:     "a",
			comments: []string{"# comment for a"},
		},
		{
			name: "add new property and multi-line comment",
			input: `
a = 1`,
			want: `
a = 1
# line 1
# line 2
b = 2`,
			path:     "b",
			value:    ast.Number{N: 2},
			comments: []string{"# line 1", "# line 2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			test(t, tc)
		})
	}
}
