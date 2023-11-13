package ast_test

import (
	"testing"

	"github.com/go-quicktest/qt"

	"github.com/marco-m/roundtrip_ini/ast"
)

var corpus = []string{
	`
name = "Johnny Stecchino"`,
	`
age = 21
score = 1.2`,
	`
[address]
city = "Bologna"`,
	`
top = 0
[section 1]
s1 = 1
[section 2]
s2 = 2
`,
}

// Go has built-in support for fuzzing.
//
// Run as plain test:
//
//	go test -run=FuzzIniParse ./ast
//
// Run as fuzz test until interrupted:
//
//	go test -fuzz=FuzzIniParse ./ast
//
// Run as fuzz test with a time limit (example: in CI environment):
//
//	go test -fuzz=FuzzIniParse -fuzztime=30s ./ast
func FuzzIniParse(f *testing.F) {
	// Fill a seed corpus.
	for _, seed := range corpus {
		// The type of the argument of AddProp must match the type of the argument
		// of the fuzz target (see below).
		f.Add(seed)
	}

	// The fuzz target.
	target := func(t *testing.T, input string) {
		sut := ast.NewParser()

		tree, err := sut.ParseString("", input)
		// Since this is a brute-force fuzz, the only thing we can do is skip
		// on any error :-/
		if err != nil {
			return
		}

		// As any fuzz test, we must find an invariant on which to assert;
		// we cannot assert on a specific expected output as we do for normal
		// tests.
		qt.Assert(t, qt.IsNotNil(tree), qt.Commentf("input: %q", input))

		// FIXME can we find a better invariant??? As-is, I think we are
		//   just wasting CPU time...
		// Ah maybe we could serialize again? Mhh not really, because we also
		// pretty-print so the majority of times we would have a mismatch...

		// Another consequence of this test being too brute force. :-/
		if tree.String() == "" {
			return
		}

		// Assert that the tree is not empty.
		qt.Assert(t, qt.IsTrue(tree.Properties != nil || tree.Sections != nil),
			qt.Commentf("input: %q", input))

		// XOR on the type of each property.
		// Probably a stupid assert because this is guaranteed by the parser
		// implementation of the INI grammar...
		for _, prop := range tree.Properties {
			_, stringOK := prop.Value.(ast.String)
			_, numberOK := prop.Value.(ast.Number)
			qt.Assert(t, qt.IsFalse(stringOK && numberOK),
				qt.Commentf("input: %q", input))
			qt.Assert(t, qt.IsTrue(stringOK || numberOK),
				qt.Commentf("input: %q", input))
		}
	}

	// Let's go!
	f.Fuzz(target)
}
