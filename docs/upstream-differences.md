# Intentional Differences from Upstream

This fork follows upstream APIs and defaults except for the corrections and intentional behavior differences listed below. The corrections are suitable for contribution to upstream through focused pull requests; they remain maintained here independently of upstream adoption.

The comparison below uses upstream commit [`1cc097e`](https://github.com/aholstenson/logseq-go/tree/1cc097e4d1e36d5f28a02d24d4c834c005e3d194). It describes that snapshot, not all future upstream versions.

## Default parsing differences

These behaviors are enabled by default. They are not formatting-preservation options.

| Input                                       | Upstream behavior                                                      | Fork behavior and rationale                                                                                                                                       |
| ------------------------------------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `- - nested`                                | Creates an empty parent block with a nested block containing `nested`. | Keeps one block containing literal `- nested`, retaining extra dash markers rather than changing the hierarchy. This is an intentional interpretation difference. |
| A standalone `---` between two root bullets | Attaches a thematic break to the preceding block.                      | Keeps the divider as a separate root block containing `---`, rather than attaching it to the preceding task. This is an intentional interpretation difference.    |
| `before\q after`                            | Drops the backslash from the text value.                               | Keeps the backslash because `q` is not an escapable Markdown punctuation character.                                                                               |
| `[text #tag](https://example.com)`          | Produces text and a tag whose target includes the closing link syntax. | Parses a Markdown link with literal `#tag` in its label, avoiding a broken link and unintended page reference.                                                    |
| `#tag] suffix`                              | Includes `]` in the tag target.                                        | Ends the tag at `]`, leaving the closing bracket as text.                                                                                                         |
| `~~Bas~\~ic~~ content`                      | Produces nested strikethrough nodes and trailing tilde text.           | Produces one strikethrough node containing `Bas~~ic`, respecting the escaped tilde.                                                                               |

## Regression coverage

The parsing cases live in `internal/markdown/parsing_test.go` and the round-trip cases in `internal/markdown/parse_output_test.go`. Extra-dash fixtures also live in `testdata/extra_dash.md`. These tests retain the fork's intended behavior independently of upstream adoption.

Formatting choices, such as emphasis delimiters, list numbering, escaping and whitespace, are separate from these semantic parsing differences. Preserving formatting must not hide edits to the parsed content.
