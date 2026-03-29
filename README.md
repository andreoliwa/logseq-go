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
