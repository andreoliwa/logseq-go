package logseq_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	logseq "github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type preservationErrorWriter struct{ err error }

func (w preservationErrorWriter) Write([]byte) (int, error) { return 0, w.err }

var _ = Describe("Output preservation policy", func() {
	It("escapes literal Markdown by default and preserves it only when requested", func() {
		block := content.NewBlock(content.NewText("[literal] _name_ *"))
		actual, err := logseq.AsMarkdown(block)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(`\[literal\] \_name\_ \*`))
		preserved, err := logseq.AsMarkdownWithOptions(block, logseq.WithPreservedMarkdown())
		Expect(err).NotTo(HaveOccurred())
		Expect(preserved).To(Equal("[literal] _name_ *"))
		var out strings.Builder
		Expect(logseq.WriteMarkdownWithOptions(block, &out, logseq.WithPreservedMarkdown())).To(Succeed())
		Expect(out.String()).To(Equal(preserved))
		failure := errors.New("writer failed")
		Expect(logseq.WriteMarkdownWithOptions(block, preservationErrorWriter{failure}, logseq.WithPreservedMarkdown())).To(MatchError(failure))
	})
	DescribeTable("selects upstream rendering without discarding parsed style", func(source, normal, preserved string) {
		block, err := logseq.ParseBlock(source)
		Expect(err).NotTo(HaveOccurred())
		actual, err := logseq.AsString(block)
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(normal))
		actual, err = logseq.AsMarkdownWithOptions(block, logseq.WithPreservedMarkdown())
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal(preserved))
	},
		Entry("emphasis delimiters", "_word_", "*word*", "_word_"),
		Entry("ordered numbers", "3. third\n7. seventh", "1. third\n2. seventh", "3. third\n7. seventh"),
		Entry("literal brackets", "text [literal]", `text \[literal\]`, "text [literal]"),
		Entry("blockquote prefixes", ">one\n>  two", "> one\n> two", ">one\n>  two"),
		Entry("URL parentheses", "[name](https://example.com/a_(b))", `[name](https://example.com/a_\(b\))`, "[name](https://example.com/a_(b))"),
	)
	It("retains style but reflects direct edits to parsed text", func() {
		block, err := logseq.ParseBlock("_before_")
		Expect(err).NotTo(HaveOccurred())
		paragraph := block.FirstChild().(*content.Paragraph)
		emphasis := paragraph.FirstChild().(*content.Emphasis)
		emphasis.FirstChild().(*content.Text).Value = "after"
		actual, err := logseq.AsMarkdownWithOptions(block, logseq.WithPreservedMarkdown())
		Expect(err).NotTo(HaveOccurred())
		Expect(actual).To(Equal("_after_"))
	})
	It("keeps graph policies independent under concurrent rendering", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "logseq"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0644)).To(Succeed())
		normal, err := logseq.Open(context.Background(), root, logseq.WithPreserveOutput(), logseq.WithStrictMarkdownOutput())
		Expect(err).NotTo(HaveOccurred())
		defer normal.Close()
		preserved, err := logseq.Open(context.Background(), root, logseq.WithStrictMarkdownOutput(), logseq.WithPreserveOutput())
		Expect(err).NotTo(HaveOccurred())
		defer preserved.Close()
		block := content.NewBlock(content.NewText("_literal_"))
		results := make(chan string, 20)
		failures := make(chan error, 20)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			for _, g := range []*logseq.Graph{normal, preserved} {
				wg.Add(1)
				go func(g *logseq.Graph) {
					defer wg.Done()
					value, err := g.AsString(block)
					results <- value
					failures <- err
				}(g)
			}
		}
		wg.Wait()
		close(results)
		close(failures)
		for err := range failures {
			Expect(err).NotTo(HaveOccurred())
		}
		counts := map[string]int{}
		for value := range results {
			counts[value]++
		}
		Expect(counts).To(Equal(map[string]int{`\_literal\_`: 10, "_literal_": 10}))
	})
})

var _ = Describe("Graph parser policy forwarding", func() {
	It("loads typed entries by default and raw entries only on request", func() {
		for _, raw := range []bool{false, true} {
			root := GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(root, "logseq"), 0755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(root, "pages"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0644)).To(Succeed())
			source := "- TODO work\n  :LOGBOOK:\n  CLOCK: [2023-06-26 Mon 09:00]\n  :END:\n"
			Expect(os.WriteFile(filepath.Join(root, "pages/Example.md"), []byte(source), 0644)).To(Succeed())
			options := []logseq.Option{logseq.WithPreserveOutput()}
			if raw {
				options = append(options, logseq.WithParseOptions(logseq.WithRawLogbooks()))
			}
			graph, err := logseq.Open(context.Background(), root, options...)
			Expect(err).NotTo(HaveOccurred())
			page, err := graph.OpenPage("Example")
			Expect(err).NotTo(HaveOccurred())
			entry := page.Blocks()[0].Children().FindDeep(content.IsOfType[*content.Logbook]()).(*content.Logbook).FirstChild()
			if raw {
				Expect(entry).To(BeAssignableToTypeOf(&content.LogbookEntryRaw{}))
			} else {
				Expect(entry).To(BeAssignableToTypeOf(&content.LogbookEntryClock{}))
			}
			parsed, err := graph.ParseBlock(":LOGBOOK:\nCLOCK: [2023-06-26 Mon 09:00]\n:END:")
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed.FirstChild().(*content.Logbook).FirstChild()).To(BeAssignableToTypeOf(entry))
			Expect(graph.Close()).To(Succeed())
		}
	})
})
