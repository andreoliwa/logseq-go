# Rules for logseq-go

IMPORTANT: before doing any change that has a lot of steps/work, show me your plan with numbered steps and ask for my approval.

## Tests when creating a new feature or fixing a bug

1. Use the Ginkgo framework.
2. Content: if logic changed in @content/, write tests in the corresponding "\_test.go" file.
3. Parsing: if logic changed in @internal/markdown/parsing.go, write tests in @internal/markdown/parsing_test.go.
4. Also evaluate if "parse-and-output" tests are needed in @internal/markdown/parse_output_test.go, and ask my opinion.
5. Saving Markdown files: if logic changed in @transaction.go, write tests in @transaction_test.go.
