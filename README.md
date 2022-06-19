# roundtripini

[![PkgGoDev](https://pkg.go.dev/github.com/marco-m/roundtripini)](https://pkg.go.dev/github.com/marco-m/roundtripini)

Decode, structural edit (replace, insert, delete) and encode the [INI file] format, preserving comments and blank lines.

Humans care about comments and blank lines and are confused when comments disappear or lines change place for no reason.

Roundtripini is useful when you need to do round trip editing and

* the diff is meant to be presented to a human for review.
* the output is meant to be read or written again by a human.

At the same time, roundtripini also acts as a formatter / pretty-printer, introducing uniformity. Like gofmt, the formatting is not tunable.

## Status

**Unstable, in development, version 0.x, the API will very likely change in a breaking way**.

## Conformity and completeness

The [INI file] is an informal format and reaching conformity or completeness is _not_ a goal of this project.

## Usage and examples

The idea is to have two packages:

* Package **roundtripini** performs **high-level** decoding, editing and encoding of the INI format, preserving comments and blank lines.
* Package **ast** performs **low-level** decoding, editing and encoding of the INI format, preserving comments and blank lines.

One would normally use package roundtripini, reserving package ast to special cases, but currently only package ast is implemented.

* See the examples in ast/example_test.go.
* See the tests.
* See the [godoc-ast] API documentation for package ast, which has also the runnable examples.

## Formatting

* Spurious leading newlines are removed.
* A trailing newline is added if missing.
* Section names are written as `[s1]` (no spaces between parenthesis).
* Properties are written as `foo = 42` (one space around the equal sign).

See `TestRoundTripCornerCases` and `TestRoundTripPrettyPrint` for details.

## Comments and blank lines

Comments and blank lines are preserved.

* A comment is just above a section title or above a property. A comment can be multi-line.
* Blank lines are just below a section title or below a property.

See the tests for details.

## Contributing

**Please, before working on a PR, open an issue to discuss your use case**.

This helps to better understand the pros and cons of a change and not to waste your time (and mine) developing a feature that for some reason doesn't fit well with the spirit of the project or could be implemented differently.

This is in the spirit of [minimalism] and [talk, then code].

## Credits

Thanks to [participle], a parser and lexer generator.

## License

This code is released under the MIT license, see file [LICENSE](LICENSE).

[godoc-ast]: https://pkg.go.dev/github.com/marco-m/roundtripini/ast
[participle]: https://github.com/alecthomas/participle
[INI file]: https://en.wikipedia.org/wiki/INI_file
[talk, then code]: https://dave.cheney.net/2019/02/18/talk-then-code
[minimalism]: https://www.britannica.com/art/Minimalism

