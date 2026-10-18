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
> You can download it from [Go Packages](https://pkg.go.dev/github.com/andreoliwa/logseq-go/v2) with `go get github.com/andreoliwa/logseq-go/v2`.

See [Intentional Differences from Upstream](docs/upstream-differences.md) for parsing corrections and behavior differences maintained by this fork.

## Features

- Read and write journals and pages
- Rename and delete pages, with references to a renamed page updated across the
  graph
- Search, linked references and lookup of blocks by their id
- Aliases and namespaces, such as `Parent/Child`
- Rich content model
  - Blocks
  - Formatting via headings, paragraphs, lists, code blocks, etc.
  - Highlights via `^^Example^^`
  - Tables in the GitHub Flavored Markdown syntax
  - LaTeX via `$formula$` and math blocks
  - Footnotes via `[^1]` and `[^1]: The footnote`
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

In a graph with indexing enabled, opening a page by one of the aliases in its
`alias::` property opens the page the alias belongs to, instead of creating a
second page for the same content.

The properties that hold a list of pages, such as `alias::` and `tags::`, have
their values read as references to those pages, so `tags:: Book, Reading`
points at two pages without either being written as a link. The graph decides
which other properties are read that way via `:property/separated-by-commas`,
and which properties never point at a page via
`:ignored-page-references-keywords`.

The blocks that reference a page, what Logseq shows as its linked references,
are available from the page itself:

```go
references, err := page.LinkedReferences(ctx)

for _, reference := range references.Results() {
  block, referencingPage, err := reference.Open()
  // ...
}
```

Pages with a `/` in their title are part of a namespace, and the pages in a
namespace can be found from the page it belongs to:

```go
page, err := graph.OpenPage("Parent/Child")

page.Namespace() // Parent

children, err := page.NamespaceChildren(ctx)
```

Blocks that have an id, which is what block references such as `((id))` point
at, can be opened directly in a graph that has indexing enabled:

```go
block, page, err := graph.OpenBlock(ctx, "65a1b2c3-d4e5-6789-abcd-ef0123456789")
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

Pages can be renamed and deleted in a transaction as well. Renaming points all
the references to the page, such as `[[Old]]` and `#Old`, at the new title,
which requires the graph to be opened with indexing enabled:

```go
tx := graph.NewTransaction()

err = tx.RenamePage(ctx, "Old", "New")
err = tx.DeletePage("Unwanted")

err = tx.Save()
```

Pages in the namespace of a renamed page keep their titles, unless the rename
asks for them to come along:

```go
// Renames Old, Old/Child and Old/Child/Grandchild
err = tx.RenamePage(ctx, "Old", "New", logseq.WithNamespaceChildren())
```

Deleted pages are removed from the graph, unless the graph is opened with
`logseq.WithRecycleDeletedPages()`, in which case they are moved to the
`logseq/.recycle` directory that Logseq recovers deleted pages from.

## Limitations

This library is limited to working with Markdown files. As the library provides
an AST for the content there might be some issues with formatting that comes
out wrong after having been read and saved again.

If this happens to you, please do open an issue with an example of content
that is causing the issue.

## Releases

See [release procedures](docs/releases.md) for local validation, reviewed draft
release PRs and publication approval.

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

## Task status changes

`marker.WithStatus(status)` is a chainable setter with no clock or property
side effects. Use `marker.TransitionTo(status)` when returning a task to TODO
should stop a running clock or remove completed/cancelled metadata. It returns
`(*content.TaskMarker, error)` and retains the fork's transition rules.

## Compatible parsing and page navigation

The existing `Page` interface and parsing function signatures remain unchanged.
Pages opened by the graph also implement the optional `logseq.PageNavigation`
interface. Use a type assertion to access aliases, namespaces, linked references,
`ContentBlocks()` (including the unbulleted pre-block) and `PageProperties()`
(page-level properties before the first bullet). `Page.Blocks()` still returns
bullet blocks, and `Page.Properties()` still accesses the first bullet's properties.

Priorities use position-sensitive A/B/C syntax, including lowercase letters.
Middle and unknown-letter markers remain literal text. Default parsing still
keeps raw logbooks and the existing alias representation. Opt into structured
logbook entries or built-in list properties through `logseq.ParseBlockWithOptions`:

```go
block, err := logseq.ParseBlockWithOptions(text,
  logseq.WithTypedLogbooks(),
  logseq.WithBuiltInPropertyLists(),
)
```

`content.Priority` is the enum and `content.TaskPriority` is its node.
`PriorityHigh`, `PriorityMedium` and `PriorityLow` alias A, B and C.

Pass the same parser options to `logseq.WithParseOptions(...)` when opening a
graph to enable those representations for loaded pages, indexing and graph
parsing. Graph property configuration applies after these parser options.

Default output retains the existing Markdown escaping and save behavior. Use
`logseq.AsMarkdownWithOptions(node, logseq.WithStrictOutput())` for stricter
literal escaping. Graphs accept `logseq.WithStrictMarkdownOutput()` for that
formatting when saving, and `logseq.WithFinalNewline()` to end saved files with
a newline. Neither option changes the default behavior.
