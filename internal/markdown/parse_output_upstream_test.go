package markdown_test

import (
	"github.com/andreoliwa/logseq-go/internal/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func upstreamparseAndOutput(input string, opts ...markdown.Option) string {
	block, err := parseIncomingString(input)
	Expect(err).ToNot(HaveOccurred())
	v, err := markdown.AsString(block, append([]markdown.Option{markdown.WithStrictOutput()}, opts...)...)
	Expect(err).ToNot(HaveOccurred())
	return v
}

func upstreamFullyEqual(name string, input string, opts ...markdown.Option) {
	It(name, func() {
		v := upstreamparseAndOutput(input, opts...)
		Expect(v).To(Equal(input))
	})
}

// FullyEqualWhenReadWith is upstreamFullyEqual for input that is read with the
// settings of a graph.
func FullyEqualWhenReadWith(name string, input string, parseOpts ...markdown.ParseOption) {
	It(name, func() {
		block, err := parseIncomingString(input, parseOpts...)
		Expect(err).ToNot(HaveOccurred())

		v, err := markdown.AsString(block, markdown.WithStrictOutput())
		Expect(err).ToNot(HaveOccurred())
		Expect(v).To(Equal(input))
	})
}

func upstreamVaries(name string, input string, output string, opts ...markdown.Option) {
	It(name, func() {
		v := upstreamparseAndOutput(input, opts...)
		Expect(v).To(Equal(output))
	})
}

var _ = Describe("Parsing then outputting", func() {
	Describe("Paragraphs", func() {
		upstreamFullyEqual("Paragraph", "Basic content")
		upstreamFullyEqual("Paragraph with soft newline", "Basic\ncontent")
		upstreamFullyEqual("Paragraph with hard newline via backslash", "Basic\\\ncontent")
		upstreamVaries("Paragraph with hard newline via two spaces", "Basic  \ncontent", "Basic\\\ncontent")

		upstreamFullyEqual("Multiple paragraphs", "Basic content\n\nMore content")
	})

	Describe("Inline formatting", func() {
		upstreamFullyEqual("Bold text", "**Basic** content")
		upstreamFullyEqual("Bold text with newline", "**Basic\ncontent**")
		upstreamFullyEqual("Bold text with hard newline", "**Basic\\\ncontent**")
		upstreamVaries("Bold text with hard newline via two spaces", "**Basic  \ncontent**", "**Basic\\\ncontent**")

		upstreamFullyEqual("Italic text", "*Basic* content")
		upstreamFullyEqual("Italic text with newline", "*Basic\ncontent*")
		upstreamFullyEqual("Italic text with hard newline", "*Basic\\\ncontent*")
		upstreamVaries("Italic text with hard newline via two spaces", "*Basic  \ncontent*", "*Basic\\\ncontent*")

		upstreamFullyEqual("Strikethrough text", "~~Basic~~ content")
		upstreamFullyEqual("Strikethrough text with newline", "~~Basic\ncontent~~")
		upstreamFullyEqual("Strikethrough text with hard newline", "~~Basic\\\ncontent~~")
		upstreamVaries("Strikethrough text with hard newline via two spaces", "~~Basic  \ncontent~~", "~~Basic\\\ncontent~~")
		upstreamFullyEqual("Strikethrough text containing escaped ~~", "~~Bas~\\~ic~~ content")

		upstreamFullyEqual("Highlighted text", "^^Basic^^ content")
		upstreamFullyEqual("Highlighted text with newline", "^^Basic\ncontent^^")
		upstreamFullyEqual("Highlighted text with hard newline", "^^Basic\\\ncontent^^")
		upstreamFullyEqual("Highlighted text containing escaped ^^", "^^Bas^\\^ic^^ content")
		upstreamFullyEqual("Text with a single ^", "2^10 content")

		// Code text maintains spaces and newlines
		upstreamFullyEqual("Code text", "`Basic` content")
		upstreamFullyEqual("Code text maintains newline", "`Basic\ncontent`")
		upstreamFullyEqual("Code text maintains spaces before 'hard' newline", "`Basic  \ncontent`")

		upstreamFullyEqual("Code text containing a backtick", "``Basic ` content``")
		upstreamFullyEqual("Code text starting with a backtick", "`` `Basic ``")
		upstreamFullyEqual("Code text ending with a backtick", "`` Basic` ``")
		upstreamFullyEqual("Code text of a single backtick", "`` ` ``")
		// Padding is only kept where it is needed to read the value back.
		upstreamVaries("Code text containing double backticks", "``` Basic ``content ```", "```Basic ``content```")
		upstreamFullyEqual("Code text surrounded by spaces", "`  Basic  `")
		upstreamFullyEqual("Code text of only spaces", "`  `")
	})

	Describe("Tags", func() {
		upstreamFullyEqual("Tag", "#tag and content")
		upstreamFullyEqual("Tag at end of line", "content #tag")
		upstreamFullyEqual("Tag with spaces", "#[[tag with spaces]] and content")
		upstreamFullyEqual("Hash followed by a space", "issue # 5")
		upstreamFullyEqual("Hash at the end of a word", "C# is fine")
		upstreamFullyEqual("Hash at the end of a line", "content #")
	})

	Describe("Heading", func() {
		upstreamFullyEqual("Heading 1", "# Heading")
		upstreamFullyEqual("Heading 2", "## Heading")
		upstreamFullyEqual("Heading 3", "### Heading")
		upstreamFullyEqual("Heading 4", "#### Heading")
		upstreamFullyEqual("Heading 5", "##### Heading")
		upstreamFullyEqual("Heading 6", "###### Heading")
	})

	Describe("Heading combined with paragraph", func() {
		upstreamFullyEqual("Heading 1", "# Heading\n\nParagraph")
		upstreamFullyEqual("Heading 2", "## Heading\n\nParagraph")
		upstreamFullyEqual("Heading 3", "### Heading\n\nParagraph")
		upstreamFullyEqual("Heading 4", "#### Heading\n\nParagraph")
		upstreamFullyEqual("Heading 5", "##### Heading\n\nParagraph")
		upstreamFullyEqual("Heading 6", "###### Heading\n\nParagraph")
	})

	Describe("Code blocks", func() {
		upstreamFullyEqual("Code block", "```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```")
		upstreamFullyEqual("Code block with newline", "```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```\n\nParagraph")

		upstreamFullyEqual("Code block after paragraph", "Paragraph\n\n```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```")
		upstreamFullyEqual("Code block interrupting paragraph", "Paragraph\n```go\nfunc main() {\n\tfmt.Println(\"Hello world\")\n}\n```")
	})

	Describe("Footnotes", func() {
		upstreamFullyEqual("Footnote reference", "Task with a footnote[^1] in it")
		upstreamFullyEqual("Footnote reference with a named label", "Text[^note]")
		upstreamFullyEqual("Footnote definition", "[^1]: The footnote")
		upstreamFullyEqual("Footnote definition with formatting", "[^1]: The **footnote** with [[Page]]")
		upstreamFullyEqual("Footnote definition after a paragraph", "Text[^1]\n\n[^1]: The footnote")
		upstreamFullyEqual("Footnote definition in a block", "- Text[^1]\n  [^1]: The footnote")
		upstreamFullyEqual("Footnote definition over several lines", "[^1]: The\n    footnote")

		upstreamFullyEqual("Link with a caret stays a link", "[^1](https://example.com)")
		upstreamFullyEqual("Brackets with a space stay text", "Text \\[^not a label\\] more")
	})

	Describe("Math", func() {
		upstreamFullyEqual("Inline math", "The formula $E = mc^2$ is famous")
		upstreamFullyEqual("Displayed math within a line", "The formula $$E = mc^2$$ is famous")
		upstreamFullyEqual("Math block", "$$\nE = mc^2\n$$")
		upstreamFullyEqual("Math block with several lines", "$$\na = b\nc = d\n$$")
		upstreamFullyEqual("Math block after a paragraph", "Content\n\n$$\nE = mc^2\n$$")
		upstreamFullyEqual("Math block in a block", "- Formula\n  $$\n  E = mc^2\n  $$")

		upstreamFullyEqual("Amounts are not math", "It costs $5 and $10")
		upstreamFullyEqual("Dollar sign followed by a space is not math", "Costs $ 5 or $ 10")
		upstreamFullyEqual("Escaped math stays text", "Not math: \\$x\\$")
	})

	Describe("Tables", func() {
		upstreamFullyEqual("Table", "| Name | Value |\n| ---- | ----- |\n| a    | 1     |")
		upstreamFullyEqual("Table with only a header", "| Name |\n| ---- |")
		upstreamFullyEqual("Table with alignments", "| a   | b   | c   | d   |\n| :-- | --: | :-: | --- |")
		upstreamFullyEqual("Table with formatting in a cell", "| Name               |\n| ------------------ |\n| **a** and [[Page]] |")
		upstreamFullyEqual("Table with an escaped pipe in a cell", "| Name   |\n| ------ |\n| a \\| b |")
		upstreamFullyEqual("Table after a paragraph", "Content\n\n| a   |\n| --- |")
		upstreamFullyEqual("Table in a block", "- | a   |\n  | --- |")

		// The cells are padded to the width of the column they are in, so a
		// table that is written any other way lines up on the way out.
		upstreamVaries("Table without padding", "|Name|Value|\n|-|-|\n|a|1|", "| Name | Value |\n| ---- | ----- |\n| a    | 1     |")
		upstreamVaries("Table with extra padding", "| a      |\n| ------ |", "| a   |\n| --- |")

		// A row is read with the columns the header has, so cells that no
		// column belongs to are dropped and missing ones are filled in.
		upstreamVaries("Table with a row that is too long", "| a   |\n| --- |\n| b   | c |", "| a   |\n| --- |\n| b   |")
		upstreamVaries("Table with a row that is too short", "| a   | b   |\n| --- | --- |\n| c   |", "| a   | b   |\n| --- | --- |\n| c   |     |")
	})

	Describe("Macros", func() {
		upstreamFullyEqual("Macro with no arguments", "{{poem}}")
		upstreamFullyEqual("Macro with one argument", "{{poem red}}")
		upstreamFullyEqual("Macro with two arguments", "{{poem red, blue}}")
		upstreamVaries("Macro with two arguments, no space", "{{poem red,blue}}", "{{poem red, blue}}")
		upstreamFullyEqual("Macro with one argument and spaces", "{{poem red blue}}")
		upstreamFullyEqual("Macro with quoted argument with comma", "{{poem \"red, blue\"}}")

		Describe("Invalid", func() {
			upstreamFullyEqual("Macro without end", "{{poem red blue")
			upstreamFullyEqual("Macro with three curly braces missing the last closing brace", "{{{poem}}")
		})
	})

	Describe("Properties", func() {
		upstreamFullyEqual("Single property", "key:: value")
		upstreamFullyEqual("Property with a value of several words", "key:: a longer value")
		upstreamFullyEqual("Property with several values", "alias:: First, Second Name")
		upstreamFullyEqual("Property with several values as page links", "alias:: [[First]], [[Second Name]]")
		upstreamFullyEqual("Property with values that mix page links and text", "tags:: [[First]], Second, #Third")
		upstreamFullyEqual("Property with values separated by a full width comma", "tags:: First，Second")
		upstreamFullyEqual("Property with extra space around its values", "tags::  First ,  Second")
		upstreamFullyEqual("Property with values that are escaped", "tags:: First, Second \\[escaped\\]")
		FullyEqualWhenReadWith("Property with values of a configured list property",
			"key:: First, Second Name",
			markdown.WithPropertiesSeparatedByCommas("key"),
		)
		FullyEqualWhenReadWith("Property whose references are ignored",
			"author:: [[Someone]], Another",
			markdown.WithIgnoredPageReferences("author"),
		)
		upstreamFullyEqual("Multiple properties", "key:: value\nkey2:: value2")
		upstreamFullyEqual("Properties followed by trailing paragraph", "key:: value\nParagraph")
		upstreamFullyEqual("Paragraphs interrupted by properties", "Paragraph\nkey:: value")
		upstreamFullyEqual("Paragraphs interrupted by properties followed by more paragraph", "Paragraph\nkey:: value\nParagraph")
		upstreamFullyEqual("Paragraph followed by properties", "Paragraph\n\nkey:: value")
	})

	Describe("Tasks", func() {
		Describe("Markers", func() {
			upstreamFullyEqual("TODO Task", "TODO Task")
			upstreamFullyEqual("DOING Task", "DOING Task")
			upstreamFullyEqual("DONE Task", "DONE Task")
			upstreamFullyEqual("LATER Task", "LATER Task")
			upstreamFullyEqual("NOW Task", "NOW Task")
			upstreamFullyEqual("CANCELLED Task", "CANCELLED Task")
			upstreamFullyEqual("CANCELED Task", "CANCELED Task")
			upstreamFullyEqual("IN-PROGRESS Task", "IN-PROGRESS Task")
			upstreamFullyEqual("WAIT Task", "WAIT Task")
			upstreamFullyEqual("WAITING Task", "WAITING Task")

			upstreamVaries("Task with leading space", " TODO Task", "TODO Task")
		})

		Describe("Markers without text", func() {
			upstreamFullyEqual("TODO on its own", "TODO")
			upstreamFullyEqual("DONE on its own", "DONE")
			upstreamFullyEqual("TODO with sub block", "- TODO\n\t- child")
			upstreamVaries("TODO with trailing space", "TODO ", "TODO")

			// The marker takes up the whole first line, so what follows it moves
			// up to sit after the marker.
			upstreamVaries("TODO with text on the next line", "TODO\nmore", "TODO more")
			upstreamVaries("TODO followed by formatting", "TODO**bold**", "TODO **bold**")
		})

		Describe("Priorities", func() {
			upstreamFullyEqual("Priority A", "TODO [#A] Task")
			upstreamFullyEqual("Priority B", "TODO [#B] Task")
			upstreamFullyEqual("Priority C", "TODO [#C] Task")
			upstreamFullyEqual("Priority without a marker", "[#A] Task")
			upstreamFullyEqual("Priority without content after it", "TODO [#A]")
			upstreamFullyEqual("Priority in sub blocks", "- TODO [#A] Task\n- TODO [#B] Other")
			upstreamFullyEqual("Priority followed by a tag and a page link", "TODO [#A] #tag and [[page]]")
			upstreamFullyEqual("Priority with a scheduled date", "TODO [#A] Task\nSCHEDULED: <2024-01-15 Mon>")

			// Brackets that Logseq does not read as a priority stay as text.
			upstreamFullyEqual("Priority that is not at the start", "Task [#A] more")
			upstreamFullyEqual("Unknown priority", "TODO [#D] Task")
			upstreamFullyEqual("Priority not followed by a space", "Task [#A]more")
		})

		Describe("Logbooks", func() {
			upstreamFullyEqual("Clock", "TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25:56]--[2023-06-26 Mon 17:25:58] =>  00:00:02\n:END:")
			upstreamFullyEqual("Clock that is still running", "TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25:56]\n:END:")
			upstreamFullyEqual("Clock spanning more than a day", "TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 09:00:00]--[2023-06-27 Tue 11:30:45] =>  26:30:45\n:END:")
			upstreamFullyEqual("State change", "TODO Task\n:LOGBOOK:\n* State \"DONE\" from \"TODO\" [2023-06-26 Mon 17:25:56]\n:END:")
			upstreamFullyEqual("State change without a previous status", "TODO Task\n:LOGBOOK:\n* State \"DONE\" [2023-06-26 Mon 17:25:56]\n:END:")
			upstreamFullyEqual("State change followed by a clock", "TODO Task\n:LOGBOOK:\n* State \"DONE\" from \"TODO\" [2023-06-26 Mon 17:25:56]\nCLOCK: [2023-06-26 Mon 17:25:56]--[2023-06-26 Mon 17:25:58] =>  00:00:02\n:END:")

			// The day name is regenerated from the date, so a wrong one is
			// corrected on the way out.
			upstreamVaries(
				"Clock with the wrong day name",
				"TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Fri 17:25:56]\n:END:",
				"TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25:56]\n:END:",
			)

			// Entries that are not modelled are kept as they were found.
			upstreamFullyEqual("Unknown entry", "TODO Task\n:LOGBOOK:\nsomething else\n:END:")
			upstreamFullyEqual("State change with an unknown status", "TODO Task\n:LOGBOOK:\n* State \"SOMEDAY\" [2023-06-26 Mon 17:25:56]\n:END:")

			// Entries are written with the precision the graph is set up for,
			// so one written with the other precision is converted.
			upstreamVaries(
				"Clock without seconds",
				"TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25]--[2023-06-26 Mon 17:27] =>  00:02\n:END:",
				"TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25:00]--[2023-06-26 Mon 17:27:00] =>  00:02:00\n:END:",
			)
		})

		Describe("Logbooks without second support", func() {
			withoutSeconds := markdown.WithLogbookSeconds(false)

			upstreamFullyEqual("Clock", "TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25]--[2023-06-26 Mon 17:27] =>  00:02\n:END:", withoutSeconds)
			upstreamFullyEqual("Clock that is still running", "TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25]\n:END:", withoutSeconds)
			upstreamFullyEqual("Clock spanning more than a day", "TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 09:00]--[2023-06-27 Tue 11:30] =>  26:30\n:END:", withoutSeconds)
			upstreamFullyEqual("State change", "TODO Task\n:LOGBOOK:\n* State \"DONE\" from \"TODO\" [2023-06-26 Mon 17:25]\n:END:", withoutSeconds)

			upstreamVaries(
				"Clock with seconds",
				"TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25:56]--[2023-06-26 Mon 17:27:58] =>  00:02:02\n:END:",
				"TODO Task\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25]--[2023-06-26 Mon 17:27] =>  00:02\n:END:",
				withoutSeconds,
			)
		})

		Describe("Scheduled and deadline", func() {
			upstreamFullyEqual("Scheduled", "TODO Task\nSCHEDULED: <2024-01-15 Mon>")
			upstreamFullyEqual("Deadline", "TODO Task\nDEADLINE: <2024-01-15 Mon>")
			upstreamFullyEqual("Scheduled with time", "TODO Task\nSCHEDULED: <2024-01-15 Mon 09:30>")
			upstreamFullyEqual("Scheduled with repeater", "TODO Task\nSCHEDULED: <2024-01-15 Mon .+3d>")
			upstreamFullyEqual("Scheduled with time and repeater", "TODO Task\nSCHEDULED: <2024-01-15 Mon 09:30 ++1w>")
			upstreamFullyEqual("Scheduled and deadline", "TODO Task\nSCHEDULED: <2024-01-15 Mon>\nDEADLINE: <2024-01-20 Sat>")
			upstreamFullyEqual("Scheduled followed by logbook", "TODO Task\nSCHEDULED: <2024-01-15 Mon>\n:LOGBOOK:\nCLOCK: [2023-06-26 Mon 17:25:56]\n:END:")
			upstreamFullyEqual("Scheduled in sub blocks", "- TODO Task\n  SCHEDULED: <2024-01-15 Mon>\n- TODO Other")
			upstreamFullyEqual("Scheduled after a blank line", "TODO Task\n\nSCHEDULED: <2024-01-15 Mon>")

			// The day name is regenerated from the date, so a wrong one is
			// corrected on the way out.
			upstreamVaries("Scheduled with the wrong day name", "TODO Task\nSCHEDULED: <2024-01-15 Fri>", "TODO Task\nSCHEDULED: <2024-01-15 Mon>")

			// Syntax that is not modelled stays as text so nothing is lost.
			upstreamFullyEqual("Scheduled with a time range", "TODO Task\nSCHEDULED: <2024-01-15 Mon 10:00-11:00>")
			upstreamFullyEqual("Scheduled with a plain date", "TODO Task\nSCHEDULED: 2024-01-15")
		})
	})
})
