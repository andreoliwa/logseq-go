# logseq-go

logseq-go is a Go library to work with a [Logseq](https://logseq.com) graph,
with support for reading and modifying journals and pages.

⚠️ **Note:** This library is still in early development, it may destroy your data
when pages are modified. Please open issues if you find any bugs.

[![codecov](https://codecov.io/github/andreoliwa/logseq-go/graph/badge.svg?token=57MKPZ2UZD)](https://codecov.io/github/andreoliwa/logseq-go)

> [!NOTE]
> **This is an active fork of [aholstenson/logseq-go](https://github.com/aholstenson/logseq-go).**
>
> [Some pull requests](https://github.com/aholstenson/logseq-go/pulls?q=sort%3Aupdated-desc+is%3Apr+is%3Aopen+author%3Aandreoliwa)
> were opened in December 2024, and a follow-up email was sent in July 2025, but neither received a
> response.
>
> This fork continues development independently and includes all features from the original
> plus additional improvements.
> You can download it from [Go Packages](https://pkg.go.dev/github.com/andreoliwa/logseq-go) with `go get github.com/andreoliwa/logseq-go`.

## Features

- Read and write journals and pages
- Rich content model
  - Blocks
  - Formatting via headings, paragraphs, lists, code blocks, etc.
  - Page links via `[[Example]]`
  - Tags via `#Example` and `#[[Example with space]]`
  - Macros via `{{macro param1 param2}}`
  - Block references via `((block-id))`

## Usage

Open a graph to access its content:

```go
graph, err := logseq.Open(ctx, "path/to/graph")
```

Content can be opened read only:

```go
journalPage, err := graph.OpenJournal(time.Now())
page, err := graph.OpenPage("Example")

for _, block := range page.Blocks() {
  // ...
}
```

Content can also be opened for writing, by creating a transaction:

```go
tx := graph.NewTransaction()

today, err := tx.OpenJournal(time.Now())

today.AddBlock(content.NewBlock(
  content.NewText("Hello!")
))

// Save all the changes made
err = tx.Save()
```

## Limitations

This library is limited to working with Markdown files. As the library provides
an AST for the content there might be some issues with formatting that comes
out wrong after having been read and saved again.

If this happens to you, please do open an issue with an example of content
that is causing the issue.

## License

This project is licensed under the MIT license, see [LICENSE](LICENSE).

## Leading times

Use the parsed content API to compare first-line times without changing their
source spelling or formatting:

```go
block, err := logseq.ParseBlock("TODO [#B] **9h00** Meeting")
if err != nil {
  return err
}
if clock, ok := content.LeadingTime(block); ok {
  fmt.Println(clock.Hour(), clock.Minute(), clock.Minutes()) // 9 0 540
}
original, err := logseq.AsString(block) // TODO [#B] **9h00** Meeting

changed := content.NormalizeLeadingTime(block) // true
normalized, err := logseq.AsString(block) // TODO [#B] 09:00 Meeting
changed = content.NormalizeLeadingTime(block) // false
```

`content.NewTimeOfDay(hour, minute)` validates values; `String()` renders
`hh:mm`. Recognition accepts one or two hour digits, two minute digits, and
`:`, `h`, `H`, `.`, `,`, or `;`. Values must be in 00:00 through 23:59 and end
at whitespace or the end of the first line. Optional task and priority markers
may precede the time. Bold and italic wrappers around the time or the time and
description are supported.

Normalization is explicit. For example, `**9h00 Meeting**` becomes
`09:00 **Meeting**`. It preserves formatting outside the time, metadata,
children and continuation lines. Neither operation searches sentence interiors,
links, code, properties, logbooks or child blocks. Invalid, absent and already
plain canonical times cause no normalization change. Ordinary parsing and output
do not normalize times.

## Raw outline source ranges

`logseq.ParseSource(source)` returns the parsed root block, a slice of
`logseq.BlockSourceRange`, an `outlined` flag, and an error. Each range has
`Block`, `Start`, and `End` fields. `Block` is the actual parsed root child;
`source[Start:End]` retains its original bytes, metadata, continuation lines,
children, fences, whitespace, and line endings. Ranges end at the next root
block's line start or the end of the source. They do not serialize or normalize
content.

Only dash lists are Logseq outlines, matching `ParseBlock`. Other list markers
remain ordinary Markdown lists. `outlined` is true when the source starts with
an explicit outline, allowing whitespace and page properties before it. Plain
prose before an outline makes it false, but ranges still identify later root
blocks. Unbulleted input has no ranges. Preserve `source[:ranges[0].Start]`
separately as the page prefix when ranges are present. Source-tracking failures
return an error rather than uncertain offsets. On range errors, the parsed
root and `outlined` flag remain available with nil ranges, so single-block
consumers can retain ordinary parsing behavior when `outlined` is false.
