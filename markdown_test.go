package logseq_test

import (
	"strings"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("AsMarkdown", func() {
	It("serializes a simple text block", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewText("Hello, world!"),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("Hello, world!"))
	})

	It("serializes bold text", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewStrong(content.NewText("bold")),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("**bold**"))
	})

	It("serializes italic text", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewEmphasis(content.NewText("italic")),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("*italic*"))
	})

	It("serializes a page link", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewPageLink("My Page"),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("[[My Page]]"))
	})

	It("serializes a task marker", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewTaskMarker(content.TaskStatusTodo),
				content.NewText("Buy groceries"),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("TODO Buy groceries"))
	})

	It("serializes a code block", func() {
		block := content.NewBlock(
			content.NewCodeBlock("fmt.Println(\"hi\")\n").WithLanguage("go"),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("```go\nfmt.Println(\"hi\")\n```"))
	})

	It("serializes inline code", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewText("use "),
				content.NewCodeSpan("foo()"),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("use `foo()`"))
	})

	It("serializes a hashtag", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewHashtag("project"),
			),
		)

		result, err := logseq.AsMarkdown(block)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("#project"))
	})

	It("serializes nested blocks", func() {
		parent := content.NewBlock(
			content.NewParagraph(content.NewText("parent")),
		)
		child := content.NewBlock(
			content.NewParagraph(content.NewText("child")),
		)
		parent.AddChild(child)

		result, err := logseq.AsMarkdown(parent)
		Expect(err).ToNot(HaveOccurred())
		Expect(result).To(Equal("parent\n- child"))
	})

	Describe("Tasks with priorities", func() {
		It("serializes a TODO with priority A", func() {
			block := content.NewBlock(
				content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusTodo),
					content.NewPriority(content.PriorityHigh),
					content.NewText("Important task"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("TODO [#A] Important task"))
		})

		It("serializes a DOING with priority B", func() {
			block := content.NewBlock(
				content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusDoing),
					content.NewPriority(content.PriorityMedium),
					content.NewText("Medium task"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("DOING [#B] Medium task"))
		})

		It("serializes a LATER with priority C", func() {
			block := content.NewBlock(
				content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusLater),
					content.NewPriority(content.PriorityLow),
					content.NewText("Low priority task"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("LATER [#C] Low priority task"))
		})
	})

	Describe("Block with properties, task, and priority", func() {
		It("serializes a block with properties followed by a task with priority", func() {
			block := content.NewBlock(
				content.NewProperties(
					content.NewProperty("id", content.NewText("abc-123")),
				),
				content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusTodo),
					content.NewPriority(content.PriorityMedium),
					content.NewText("Review PR"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("id:: abc-123\nTODO [#B] Review PR"))
		})
	})

	Describe("Deep nesting", func() {
		It("serializes three levels of nested blocks", func() {
			block := content.NewBlock(
				content.NewParagraph(content.NewText("parent")),
				content.NewBlock(
					content.NewParagraph(content.NewText("child")),
					content.NewBlock(
						content.NewParagraph(content.NewText("grandchild")),
					),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("parent\n- child\n\t- grandchild"))
		})
	})

	Describe("Bullet lists inside blocks", func() {
		It("serializes a block with an unordered list", func() {
			block := content.NewBlock(
				content.NewParagraph(content.NewText("Shopping list:")),
				content.NewUnorderedList(
					content.NewListItem(content.NewParagraph(content.NewText("Milk"))),
					content.NewListItem(content.NewParagraph(content.NewText("Eggs"))),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("Shopping list:\n\n* Milk\n* Eggs"))
		})

		It("serializes a block with an ordered list", func() {
			block := content.NewBlock(
				content.NewParagraph(content.NewText("Steps:")),
				content.NewOrderedList(
					content.NewListItem(content.NewParagraph(content.NewText("First"))),
					content.NewListItem(content.NewParagraph(content.NewText("Second"))),
					content.NewListItem(content.NewParagraph(content.NewText("Third"))),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("Steps:\n\n1. First\n2. Second\n3. Third"))
		})
	})

	Describe("Mixed inline formatting", func() {
		It("serializes bold, italic, code, page link, and hashtag in one paragraph", func() {
			block := content.NewBlock(
				content.NewParagraph(
					content.NewStrong(content.NewText("Bold")),
					content.NewText(" and "),
					content.NewEmphasis(content.NewText("italic")),
					content.NewText(" with "),
					content.NewCodeSpan("code"),
					content.NewText(" and "),
					content.NewPageLink("Page Link"),
					content.NewText(" "),
					content.NewHashtag("tag"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("**Bold** and *italic* with `code` and [[Page Link]] #tag"))
		})
	})

	Describe("Logbook", func() {
		It("serializes a DONE task with logbook entries", func() {
			block := content.NewBlock(
				content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusDone),
					content.NewText("Clean up"),
				),
				content.NewLogbook(
					content.NewLogbookEntryRaw("CLOCK: [2024-01-15 Mon 10:00]--[2024-01-15 Mon 11:00] =>  01:00"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("DONE Clean up\n\n:LOGBOOK:\nCLOCK: [2024-01-15 Mon 10:00]--[2024-01-15 Mon 11:00] =>  01:00\n:END:"))
		})
	})

	Describe("Formatted link text", func() {
		It("serializes a link with bold text", func() {
			block := content.NewBlock(
				content.NewParagraph(
					content.NewLink("https://example.com",
						content.NewStrong(content.NewText("bold link")),
					),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("[**bold link**](https://example.com)"))
		})
	})

	Describe("Block with logbook and properties", func() {
		It("serializes a block with properties, task, and logbook", func() {
			block := content.NewBlock(
				content.NewProperties(
					content.NewProperty("id", content.NewText("task-456")),
					content.NewProperty("priority", content.NewText("high")),
				),
				content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusDone),
					content.NewText("Deploy to production"),
				),
				content.NewLogbook(
					content.NewLogbookEntryRaw("CLOCK: [2024-03-01 Fri 14:00]--[2024-03-01 Fri 15:30] =>  01:30"),
				),
			)

			result, err := logseq.AsMarkdown(block)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("id:: task-456\npriority:: high\nDONE Deploy to production\n\n:LOGBOOK:\nCLOCK: [2024-03-01 Fri 14:00]--[2024-03-01 Fri 15:30] =>  01:30\n:END:"))
		})
	})
})

var _ = Describe("WriteMarkdown", func() {
	It("writes to an io.Writer", func() {
		block := content.NewBlock(
			content.NewParagraph(
				content.NewText("Hello!"),
			),
		)

		var buf strings.Builder
		err := logseq.WriteMarkdown(block, &buf)
		Expect(err).ToNot(HaveOccurred())
		Expect(buf.String()).To(Equal("Hello!"))
	})
})
