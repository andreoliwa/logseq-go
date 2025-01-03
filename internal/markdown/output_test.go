package markdown_test

import (
	"strings"

	"github.com/andreoliwa/logseq-go/content"
	"github.com/andreoliwa/logseq-go/internal/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Output", func() {
	var writer *markdown.Output
	var buf *strings.Builder

	BeforeEach(func() {
		buf = &strings.Builder{}
		writer = markdown.NewWriter(buf)
	})

	Describe("Text", func() {
		It("can write text", func() {
			err := writer.Write(content.NewText("abc"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc"))
		})

		It("can write text with soft breaks", func() {
			err := writer.Write(content.NewText("abc").WithSoftLineBreak())
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n"))
		})

		It("can write text with hard breaks", func() {
			err := writer.Write(content.NewText("abc").WithHardLineBreak())
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\\\n"))
		})

		It("can write text with characters that should be escaped", func() {
			err := writer.Write(content.NewText("abc*"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\\*"))
		})

		It("can write multiple text nodes", func() {
			err := writer.Write(content.NewText("abc"))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewText("def"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abcdef"))
		})

		It("can write text with multiple lines", func() {
			err := writer.Write(content.NewText("abc\ndef"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\ndef"))
		})

		It("can write multiple text nodes with soft breaks", func() {
			err := writer.Write(content.NewText("abc").WithSoftLineBreak())
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewText("def"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\ndef"))
		})
	})

	Describe("Text formatting", func() {
		It("can write emphasis", func() {
			err := writer.Write(content.NewEmphasis(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("*abc*"))
		})

		It("can write emphasis + emphasis", func() {
			err := writer.Write(content.NewParagraph(
				content.NewEmphasis(content.NewText("abc")),
				content.NewEmphasis(content.NewText("def")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("*abc* *def*"))
		})

		It("can write emphasis wrapping multiple text nodes", func() {
			err := writer.Write(content.NewEmphasis(content.NewText("abc"), content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("*abcdef*"))
		})

		It("can write strong", func() {
			err := writer.Write(content.NewStrong(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("**abc**"))
		})

		It("can write strong wrapping multiple text nodes", func() {
			err := writer.Write(content.NewStrong(content.NewText("abc"), content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("**abcdef**"))
		})

		It("can write strong + strong", func() {
			err := writer.Write(content.NewParagraph(
				content.NewStrong(content.NewText("abc")),
				content.NewStrong(content.NewText("def")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("**abc** **def**"))
		})

		It("can write strong & emphasis", func() {
			err := writer.Write(content.NewStrong(content.NewEmphasis(content.NewText("abc"))))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("***abc***"))
		})

		It("can write strong + emphasis", func() {
			err := writer.Write(content.NewParagraph(
				content.NewStrong(content.NewText("abc")),
				content.NewEmphasis(content.NewText("def")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("**abc***def*"))
		})

		It("can write strikethrough", func() {
			err := writer.Write(content.NewStrikethrough(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("~~abc~~"))
		})

		It("can write strikethrough that contains ~~", func() {
			err := writer.Write(content.NewStrikethrough(content.NewText("abc~~def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("~~abc~\\~def~~"))
		})

		It("can write code", func() {
			err := writer.Write(content.NewCodeSpan("abc"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("`abc`"))
		})

		It("can write code with backtick in it", func() {
			err := writer.Write(content.NewCodeSpan("abc`def"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("``abc`def``"))
		})

		It("can write code with double backtick in it", func() {
			err := writer.Write(content.NewCodeSpan("abc``def"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("```abc``def```"))
		})
	})

	Describe("Links", func() {
		Describe("Normal links", func() {
			It("can write a link", func() {
				err := writer.Write(content.NewLink("https://example.com", content.NewText("abc")))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("[abc](https://example.com)"))
			})

			It("can write a link with multiple text nodes", func() {
				err := writer.Write(content.NewLink("https://example.com", content.NewText("abc"), content.NewText("def")))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("[abcdef](https://example.com)"))
			})

			It("can write a link with a title", func() {
				err := writer.Write(content.NewLink("https://example.com", content.NewText("abc")).WithTitle("title"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("[abc](https://example.com 'title')"))
			})

			It("can write a link with a title that needs escaping", func() {
				err := writer.Write(content.NewLink("https://example.com", content.NewText("abc")).WithTitle("title)"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("[abc](https://example.com 'title\\)')"))
			})

		})

		Describe("Auto links", func() {
			It("can write an auto link", func() {
				err := writer.Write(content.NewAutoLink("https://example.com"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("https://example.com"))
			})

			It("can write a non-automatic auto link", func() {
				err := writer.Write(content.NewAutoLink("www.example.com"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("<www.example.com>"))
			})
		})

		Describe("Page links", func() {
			It("can write a page link", func() {
				err := writer.Write(content.NewPageLink("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("[[abc]]"))
			})

			It("can write a page link with spaces", func() {
				err := writer.Write(content.NewPageLink("abc def"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("[[abc def]]"))
			})
		})

		Describe("Hashtag", func() {
			It("can write a hashtag", func() {
				err := writer.Write(content.NewHashtag("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("#abc"))
			})

			It("can write a hashtag with spaces", func() {
				err := writer.Write(content.NewHashtag("abc def"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("#[[abc def]]"))
			})
		})

		Describe("Block references", func() {
			It("can write a block reference", func() {
				err := writer.Write(content.NewBlockRef("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("((abc))"))
			})
		})
	})

	Describe("Images", func() {
		It("can write an image", func() {
			err := writer.Write(content.NewImage("https://example.com"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("![](https://example.com)"))
		})

		It("can write image with child nodes", func() {
			err := writer.Write(content.NewImage("https://example.com", content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("![abc](https://example.com)"))
		})

		It("can write image with title", func() {
			err := writer.Write(content.NewImage("https://example.com", content.NewText("abc")).WithTitle("title"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("![abc](https://example.com 'title')"))
		})

		It("can write image with title that needs escaping", func() {
			err := writer.Write(content.NewImage("https://example.com", content.NewText("abc")).WithTitle("title)"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("![abc](https://example.com 'title\\)')"))
		})
	})

	Describe("Macros", func() {
		It("can write a macro", func() {
			err := writer.Write(content.NewMacro("abc"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("{{abc}}"))
		})

		It("can write a macro with arguments", func() {
			err := writer.Write(content.NewMacro("abc", "def"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("{{abc def}}"))
		})

		It("can write a macro with multiple arguments", func() {
			err := writer.Write(content.NewMacro("abc", "def", "ghi"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("{{abc def, ghi}}"))
		})

		It("can write a macro with argument that contains spaces", func() {
			err := writer.Write(content.NewMacro("abc", "def ghi"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("{{abc def ghi}}"))
		})

		It("can write a macro with argument that contains commas", func() {
			err := writer.Write(content.NewMacro("abc", "def, ghi"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("{{abc \"def, ghi\"}}"))
		})

		Describe("Query", func() {
			It("can write a query", func() {
				err := writer.Write(content.NewQuery("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("{{query abc}}"))
			})
		})

		Describe("Page embed", func() {
			It("can write a page embed", func() {
				err := writer.Write(content.NewPageEmbed("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("{{embed [[abc]]}}"))
			})
		})

		Describe("Block embed", func() {
			It("can write a block embed", func() {
				err := writer.Write(content.NewBlockEmbed("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("{{embed ((abc))}}"))
			})
		})

		Describe("Cloze", func() {
			It("can write a cloze", func() {
				err := writer.Write(content.NewCloze("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("{{cloze abc}}"))
			})

			It("can write cloze with cue", func() {
				err := writer.Write(content.NewClozeWithCue("abc", "def"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("{{cloze abc \\ def}}"))
			})
		})
	})

	Describe("Headings", func() {
		It("can write a heading", func() {
			err := writer.Write(content.NewHeading(1, content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("# abc"))
		})

		It("can write a heading with multiple text nodes", func() {
			err := writer.Write(content.NewHeading(1, content.NewText("abc"), content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("# abcdef"))
		})

		It("can write a heading with a link", func() {
			err := writer.Write(content.NewHeading(1, content.NewLink("https://example.com", content.NewText("abc"))))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("# [abc](https://example.com)"))
		})

		It("can write a heading after a paragraph", func() {
			err := writer.Write(content.NewParagraph(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewHeading(1, content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n\n# def"))
		})

		It("can write multiple headings", func() {
			err := writer.Write(content.NewHeading(1, content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewHeading(1, content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("# abc\n\n# def"))
		})

		It("can write headings of different levels", func() {
			err := writer.Write(content.NewHeading(1, content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewHeading(2, content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("# abc\n\n## def"))
		})
	})

	Describe("Paragraphs", func() {
		It("can write a paragraph", func() {
			err := writer.Write(content.NewParagraph(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc"))
		})

		It("can write multiple paragraphs", func() {
			err := writer.Write(content.NewParagraph(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewParagraph(content.NewText("def")))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n\ndef"))
		})
	})

	Describe("Lists", func() {
		Describe("Unordered", func() {
			It("can write an unordered list", func() {
				err := writer.Write(content.NewUnorderedList(
					content.NewListItem(
						content.NewParagraph(content.NewText("abc")),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("* abc"))
			})

			It("can write an unordered list with + marker", func() {
				err := writer.Write(content.NewListFromMarker('+',
					content.NewListItem(
						content.NewParagraph(content.NewText("abc")),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("+ abc"))
			})

			It("can write multiple unordered lists", func() {
				err := writer.Write(content.NewUnorderedList(content.NewListItem(content.NewParagraph(content.NewText("abc")))))
				Expect(err).ToNot(HaveOccurred())

				err = writer.Write(content.NewUnorderedList(content.NewListItem(content.NewParagraph(content.NewText("def")))))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("* abc\n\n* def"))
			})

			It("can write an unordered list with multiple items", func() {
				err := writer.Write(content.NewUnorderedList(
					content.NewListItem(content.NewParagraph(content.NewText("abc"))),
					content.NewListItem(content.NewParagraph(content.NewText("def"))),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("* abc\n* def"))
			})

			It("can write an unordered list with multiple paragraphs", func() {
				err := writer.Write(content.NewUnorderedList(
					content.NewListItem(
						content.NewParagraph(content.NewText("abc")),
						content.NewParagraph(content.NewText("def")),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("* abc\n\n  def"))
			})
		})

		Describe("Ordered", func() {
			It("can write an ordered list", func() {
				err := writer.Write(content.NewOrderedList(
					content.NewListItem(
						content.NewParagraph(content.NewText("abc")),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("1. abc"))
			})

			It("can write multiple ordered lists", func() {
				err := writer.Write(content.NewOrderedList(content.NewListItem(content.NewParagraph(content.NewText("abc")))))
				Expect(err).ToNot(HaveOccurred())

				err = writer.Write(content.NewOrderedList(content.NewListItem(content.NewParagraph(content.NewText("def")))))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("1. abc\n\n1. def"))
			})

			It("can write an ordered list with multiple items", func() {
				err := writer.Write(content.NewOrderedList(
					content.NewListItem(content.NewParagraph(content.NewText("abc"))),
					content.NewListItem(content.NewParagraph(content.NewText("def"))),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("1. abc\n2. def"))
			})

			It("can write an ordered list with multiple paragraphs", func() {
				err := writer.Write(content.NewOrderedList(
					content.NewListItem(
						content.NewParagraph(content.NewText("abc")),
						content.NewParagraph(content.NewText("def")),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("1. abc\n\n   def"))
			})
		})

		Describe("Nested", func() {
			It("can write a nested list", func() {
				err := writer.Write(content.NewUnorderedList(
					content.NewListItem(
						content.NewOrderedList(
							content.NewListItem(
								content.NewParagraph(content.NewText("def")),
							),
						),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("* 1. def"))
			})

			It("can write a nested list with multiple items", func() {
				err := writer.Write(content.NewUnorderedList(
					content.NewListItem(
						content.NewOrderedList(
							content.NewListItem(content.NewParagraph(content.NewText("abc"))),
							content.NewListItem(content.NewParagraph(content.NewText("def"))),
						),
					),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("* 1. abc\n  2. def"))
			})
		})
	})

	Describe("Blockquotes", func() {
		It("can write a blockquote", func() {
			err := writer.Write(content.NewBlockquote(
				content.NewParagraph(content.NewText("abc")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("> abc"))
		})

		It("can write multiple blockquotes", func() {
			err := writer.Write(content.NewBlockquote(content.NewParagraph(content.NewText("abc"))))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewBlockquote(content.NewParagraph(content.NewText("def"))))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("> abc\n\n> def"))
		})

		It("can write a blockquote with multiple paragraphs", func() {
			err := writer.Write(content.NewBlockquote(
				content.NewParagraph(content.NewText("abc")),
				content.NewParagraph(content.NewText("def")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("> abc\n>\n> def"))
		})

		It("can write block quote as first node in list item", func() {
			err := writer.Write(content.NewUnorderedList(
				content.NewListItem(
					content.NewBlockquote(
						content.NewParagraph(content.NewText("abc")),
					),
				),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("* > abc"))
		})
	})

	Describe("Code blocks", func() {
		It("can write a code block", func() {
			err := writer.Write(content.NewCodeBlock("package main"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("```\npackage main\n```"))
		})

		It("can write code block with language", func() {
			err := writer.Write(content.NewCodeBlock("package main").WithLanguage("go"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("```go\npackage main\n```"))
		})

		It("can write code block after paragraph", func() {
			err := writer.Write(content.NewParagraph(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewCodeBlock("package main"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n\n```\npackage main\n```"))
		})

		It("ending newline in code block is ignored", func() {
			err := writer.Write(content.NewCodeBlock("package main\n"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("```\npackage main\n```"))
		})
	})

	Describe("HTML", func() {
		It("can write inline HTML", func() {
			err := writer.Write(content.NewRawHTML("<b>"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("<b>"))
		})

		It("can write block HTML", func() {
			err := writer.Write(content.NewRawHTMLBlock("<p>Testing</p>"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("<p>Testing</p>"))
		})

		It("can write block HTML after paragraph", func() {
			err := writer.Write(content.NewParagraph(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewRawHTMLBlock("<p>Testing</p>"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n\n<p>Testing</p>"))
		})
	})

	Describe("Thematic breaks", func() {
		It("can write a thematic break", func() {
			err := writer.Write(content.NewThematicBreak())
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("---"))
		})

		It("can write thematic break in blockquote", func() {
			err := writer.Write(content.NewBlockquote(content.NewThematicBreak()))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("> ---"))
		})

		It("can write thematic break after paragraph", func() {
			err := writer.Write(content.NewParagraph(content.NewText("abc")))
			Expect(err).ToNot(HaveOccurred())

			err = writer.Write(content.NewThematicBreak())
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n\n---"))
		})
	})

	Describe("Blocks", func() {
		It("can write block with only content", func() {
			err := writer.Write(content.NewBlock(
				content.NewParagraph(content.NewText("abc")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc"))
		})

		It("can write block with no-content and sub-blocks", func() {
			err := writer.Write(content.NewBlock(
				content.NewBlock(content.NewParagraph(content.NewText("abc"))),
				content.NewBlock(content.NewParagraph(content.NewText("def"))),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("- abc\n- def"))
		})

		It("can write block with content and sub-blocks", func() {
			err := writer.Write(content.NewBlock(
				content.NewParagraph(content.NewText("abc")),
				content.NewBlock(content.NewParagraph(content.NewText("block 1"))),
				content.NewBlock(content.NewParagraph(content.NewText("block 2"))),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n- block 1\n- block 2"))
		})

		It("can write nested blocks", func() {
			err := writer.Write(content.NewBlock(
				content.NewParagraph(content.NewText("abc")),
				content.NewBlock(
					content.NewParagraph(content.NewText("block 1")),
					content.NewBlock(content.NewParagraph(content.NewText("block 2"))),
				),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\n- block 1\n\t- block 2"))
		})

		It("can write block with content containing new lines with sub-blocks", func() {
			err := writer.Write(content.NewBlock(
				content.NewParagraph(
					content.NewText("abc").WithSoftLineBreak(),
					content.NewText("def"),
				),
				content.NewBlock(content.NewParagraph(content.NewText("block 1"))),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\ndef\n- block 1"))
		})

		It("can write block with sub-blocks that contain new lines", func() {
			err := writer.Write(content.NewBlock(
				content.NewBlock(
					content.NewParagraph(content.NewText("abc")),
				),
				content.NewBlock(content.NewParagraph(
					content.NewText("def").WithSoftLineBreak(),
					content.NewText("continued"),
				)),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("- abc\n- def\n  continued"))
		})
	})

	Describe("Properties", func() {
		It("can write properties", func() {
			err := writer.Write(content.NewProperties(
				content.NewProperty("key", content.NewText("value")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("key:: value"))
		})

		It("can write properties with non-text value", func() {
			err := writer.Write(content.NewProperties(
				content.NewProperty("key", content.NewAutoLink("https://example.com")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("key:: https://example.com"))
		})

		It("can write paragraph with properties at the end", func() {
			err := writer.Write(content.NewBlock(
				content.NewParagraph(content.NewText("abc")),
				content.NewProperties(
					content.NewProperty("key", content.NewText("value")),
				),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\nkey:: value"))
		})

		It("can write paragraph with properties at the beginning", func() {
			err := writer.Write(content.NewBlock(
				content.NewProperties(
					content.NewProperty("key1", content.NewText("value1")),
					content.NewProperty("key2", content.NewHashtag("value2")),
				),
				content.NewParagraph(content.NewText("abc")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("key1:: value1\nkey2:: #value2\nabc"))
		})

		It("can write paragraph with properties in the middle", func() {
			err := writer.Write(content.NewBlock(
				content.NewParagraph(content.NewText("abc")),
				content.NewProperties(
					content.NewProperty("key", content.NewText("value")),
				),
				content.NewParagraph(content.NewText("def")),
			))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("abc\nkey:: value\ndef"))
		})
	})

	Describe("Advanced commands", func() {
		It("can write", func() {
			err := writer.Write(content.NewAdvancedCommand("ABC", "def"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("#+BEGIN_ABC\ndef\n#+END_ABC"))
		})

		It("can write multiple lines", func() {
			err := writer.Write(content.NewAdvancedCommand("ABC", "def\nghi"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("#+BEGIN_ABC\ndef\nghi\n#+END_ABC"))
		})

		It("can write and indent is kept", func() {
			err := writer.Write(content.NewAdvancedCommand("ABC", "def\n  ghi"))
			Expect(err).ToNot(HaveOccurred())

			Expect(buf.String()).To(Equal("#+BEGIN_ABC\ndef\n  ghi\n#+END_ABC"))
		})

		Describe("Query", func() {
			It("can write query", func() {
				err := writer.Write(content.NewQueryCommand("abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("#+BEGIN_QUERY\nabc\n#+END_QUERY"))
			})

			It("can write query ending in newline", func() {
				err := writer.Write(content.NewQueryCommand("abc\n"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("#+BEGIN_QUERY\nabc\n#+END_QUERY"))
			})
		})

		Describe("Quote", func() {
			It("can write quote", func() {
				err := writer.Write(content.NewAdvancedCommand("QUOTE", "abc"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("#+BEGIN_QUOTE\nabc\n#+END_QUOTE"))
			})

			It("can write quote ending in newline", func() {
				err := writer.Write(content.NewAdvancedCommand("QUOTE", "abc\n"))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("#+BEGIN_QUOTE\nabc\n#+END_QUOTE"))
			})
		})
	})

	Describe("Tasks", func() {
		Describe("Markers", func() {
			It("can write a TODO", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusTodo),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("TODO Task"))
			})

			It("can write DOING", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusDoing),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("DOING Task"))
			})

			It("can write DONE", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusDone),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("DONE Task"))
			})

			It("can write LATER", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusLater),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("LATER Task"))
			})

			It("can write NOW", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusNow),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("NOW Task"))
			})

			It("can write CANCELLED", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusCancelled),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("CANCELLED Task"))
			})

			It("can write CANCELED", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusCanceled),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("CANCELED Task"))
			})

			It("can write IN-PROGRESS", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusInProgress),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("IN-PROGRESS Task"))
			})

			It("can write WAIT", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusWait),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("WAIT Task"))
			})

			It("can write WAITING", func() {
				err := writer.Write(content.NewParagraph(
					content.NewTaskMarker(content.TaskStatusWaiting),
					content.NewText("Task"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal("WAITING Task"))
			})
		})

		Describe("Logbooks", func() {
			It("can write empty logbook", func() {
				err := writer.Write(content.NewLogbook())
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal(":LOGBOOK:\n:END:"))
			})

			It("can write logbook with raw entry", func() {
				err := writer.Write(content.NewLogbook(
					content.NewLogbookEntryRaw("abc"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal(":LOGBOOK:\nabc\n:END:"))
			})

			It("can write logbook with several raw entries", func() {
				err := writer.Write(content.NewLogbook(
					content.NewLogbookEntryRaw("abc"),
					content.NewLogbookEntryRaw("def"),
				))
				Expect(err).ToNot(HaveOccurred())

				Expect(buf.String()).To(Equal(":LOGBOOK:\nabc\ndef\n:END:"))
			})
		})
	})
})
