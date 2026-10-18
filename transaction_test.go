package logseq_test

import (
	"context"
	"gotest.tools/v3/golden"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/andreoliwa/logseq-go/v2"
	"github.com/andreoliwa/logseq-go/v2/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Transaction", func() {
	var (
		tempDir string
		graph   *logseq.Graph
	)

	BeforeEach(func() {
		var err error
		tempDir, err = os.MkdirTemp("", "logseq-test")
		Expect(err).NotTo(HaveOccurred())

		// Create a minimal graph structure
		err = os.MkdirAll(filepath.Join(tempDir, "logseq"), 0755)
		Expect(err).NotTo(HaveOccurred())

		err = os.MkdirAll(filepath.Join(tempDir, "pages"), 0755)
		Expect(err).NotTo(HaveOccurred())

		err = os.MkdirAll(filepath.Join(tempDir, "journals"), 0755)
		Expect(err).NotTo(HaveOccurred())

		// Create minimal config file
		configPath := filepath.Join(tempDir, "logseq", "config.edn")
		err = os.WriteFile(configPath, []byte("{}"), 0644)
		Expect(err).NotTo(HaveOccurred())

		ctx := context.Background()
		graph, err = logseq.Open(ctx, tempDir, logseq.WithPreserveOutput())
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if tempDir != "" {
			os.RemoveAll(tempDir)
		}
	})

	Describe("Save", func() {
		It("should succeed when no pages are opened", func() {
			transaction := graph.NewTransaction()

			err := transaction.Save()
			Expect(err).NotTo(HaveOccurred())
		})

		// Dynamically create test cases for all .md files in testdata directory
		Context("should save Markdown examples correctly", func() {
			// Read all .md files from testdata directory and create tests dynamically
			files, err := os.ReadDir("testdata")
			if err != nil {
				panic("Failed to read testdata directory: " + err.Error())
			}

			mapFileTestName := make(map[string]string)
			for _, file := range files {
				if !file.IsDir() && strings.HasSuffix(file.Name(), ".md") {
					fileContents, err := os.ReadFile(filepath.Join("testdata", file.Name()))
					if err != nil {
						panic("Failed to read file: " + file.Name() + " - " + err.Error())
					}

					lines := strings.Split(string(fileContents), "\n")
					if len(lines) <= 1 {
						panic("The first line of the Markdown file should have a test name: " + file.Name())
					}

					firstLine := strings.TrimPrefix(lines[0], "- ")
					firstLine = strings.TrimSpace(firstLine)
					if firstLine == "" {
						firstLine = "process " + strings.TrimSuffix(file.Name(), ".md")
					}
					mapFileTestName[file.Name()] = firstLine
				}
			}

			if len(mapFileTestName) == 0 {
				panic("No .md files found in testdata directory")
			}

			// Create individual test cases for each markdown file in the testdata directory
			for filename, testName := range mapFileTestName {
				It("should "+testName+" ("+filename+")", func() {
					if strings.HasPrefix(filename, "fix_") {
						Skip("This fix is not implemented yet")
					}

					transaction := graph.NewTransaction()

					// Extract base name without extension for page name
					baseName := strings.TrimSuffix(filename, ".md")
					pageName := "test-page-" + baseName
					destFileName := pageName + ".md"
					goldenFileName := pageName + ".md.golden"

					destPath := copyMarkdownExample(tempDir, filename, destFileName)
					goldenPath := copyMarkdownExample(tempDir, filename, goldenFileName)

					// Create the golden file by reading the original content and appending the expected content
					// This ensures consistent line endings (LF) to match what the logseq library produces
					originalContent, err := os.ReadFile(goldenPath)
					Expect(err).NotTo(HaveOccurred())

					// Convert any CRLF to LF to match logseq library output
					originalContentStr := strings.ReplaceAll(string(originalContent), "\r\n", "\n")
					expectedContent := originalContentStr + "- Hello world"
					if strings.HasSuffix(originalContentStr, "\n") {
						expectedContent += "\n"
					}

					err = os.WriteFile(goldenPath, []byte(expectedContent), 0644)
					Expect(err).NotTo(HaveOccurred())

					// Add a block to the page using the DOM
					page, err := transaction.OpenPage(pageName)
					Expect(err).NotTo(HaveOccurred())
					page.AddBlock(content.NewBlock(
						content.NewParagraph(
							content.NewText("Hello world"),
						),
					))

					// Save the page with the transaction
					err = transaction.Save()
					Expect(err).NotTo(HaveOccurred())

					// Compare the saved file with the golden file
					savedContent, err := os.ReadFile(destPath)
					Expect(err).NotTo(HaveOccurred())

					// Use golden instead of asserting manually, because it works also on Windows.
					// Search for GOTESTTOOLS_GOLDEN_NormalizeCRLFToLF in this repo
					golden.Assert(GinkgoT(), string(savedContent), goldenPath)
				})
			}
		})
	})
})

func copyMarkdownExample(tempDir, src, dest string) string {
	sourcePath := filepath.Join("testdata", src)
	destPath := filepath.Join(tempDir, "pages", dest)

	sourceFile, err := os.Open(sourcePath)
	Expect(err).NotTo(HaveOccurred())
	defer sourceFile.Close()

	destFile, err := os.Create(destPath)
	Expect(err).NotTo(HaveOccurred())
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	Expect(err).NotTo(HaveOccurred())

	return destPath
}

var _ = Describe("Page newline policy", func() {
	DescribeTable("selects page endings independently of fragment output", func(preserve, isNew bool, source, want string) {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "logseq"), 0755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pages"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0644)).To(Succeed())
		path := filepath.Join(root, "pages/Example.md")
		if !isNew {
			Expect(os.WriteFile(path, []byte(source), 0644)).To(Succeed())
		}
		var opts []logseq.Option
		if preserve {
			opts = append(opts, logseq.WithPreserveOutput())
		}
		graph, err := logseq.Open(context.Background(), root, opts...)
		Expect(err).NotTo(HaveOccurred())
		defer graph.Close()
		tx := graph.NewTransaction()
		page, err := tx.OpenPage("Example")
		Expect(err).NotTo(HaveOccurred())
		if isNew {
			block, err := logseq.ParseBlock("_word_")
			Expect(err).NotTo(HaveOccurred())
			page.AddBlock(block)
		}
		Expect(tx.Save()).To(Succeed())
		actual, err := os.ReadFile(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(string(actual)).To(Equal(want))
	},
		Entry("default existing page", false, false, "- _word_", "- *word*\n"),
		Entry("preserved existing page without newline", true, false, "- _word_", "- _word_"),
		Entry("preserved existing page with newline", true, false, "- _word_\n", "- _word_\n"),
		Entry("default new page", false, true, "", "- *word*\n"),
		Entry("preserved new page", true, true, "", "- _word_"),
		Entry("default blank page", false, false, "\n", ""),
		Entry("preserved blank page", true, false, "\n", "\n"),
		Entry("preserved empty page", true, false, "", ""),
	)

})

var _ = Describe("Unchanged preserved saves", func() {
	It("does not rewrite an unchanged preserved page", func() {
		root := GinkgoT().TempDir()
		Expect(os.MkdirAll(filepath.Join(root, "logseq"), 0755)).To(Succeed())
		Expect(os.MkdirAll(filepath.Join(root, "pages"), 0755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0644)).To(Succeed())
		path := filepath.Join(root, "pages/Example.md")
		Expect(os.WriteFile(path, []byte("- _word_\n"), 0644)).To(Succeed())
		old := time.Unix(100, 0)
		Expect(os.Chtimes(path, old, old)).To(Succeed())
		graph, err := logseq.Open(context.Background(), root, logseq.WithPreserveOutput())
		Expect(err).NotTo(HaveOccurred())
		defer graph.Close()
		tx := graph.NewTransaction()
		_, err = tx.OpenPage("Example")
		Expect(err).NotTo(HaveOccurred())
		Expect(tx.Save()).To(Succeed())
		info, err := os.Stat(path)
		Expect(err).NotTo(HaveOccurred())
		Expect(info.ModTime()).To(Equal(old))
	})
})

var _ = Describe("Read-time page ending", func() {
	It("uses captured newline presence rather than rereading it at save time", func() {
		for _, ending := range []string{"", "\n"} {
			root := GinkgoT().TempDir()
			Expect(os.MkdirAll(filepath.Join(root, "logseq"), 0755)).To(Succeed())
			Expect(os.MkdirAll(filepath.Join(root, "pages"), 0755)).To(Succeed())
			Expect(os.WriteFile(filepath.Join(root, "logseq/config.edn"), []byte("{}"), 0644)).To(Succeed())
			path := filepath.Join(root, "pages/Example.md")
			Expect(os.WriteFile(path, []byte("- task"+ending), 0644)).To(Succeed())
			graph, err := logseq.Open(context.Background(), root, logseq.WithPreserveOutput())
			Expect(err).NotTo(HaveOccurred())
			tx := graph.NewTransaction()
			page, err := tx.OpenPage("Example")
			Expect(err).NotTo(HaveOccurred())
			replacement := "\n"
			if ending != "" {
				replacement = ""
			}
			Expect(os.WriteFile(path, []byte("- task"+replacement), 0644)).To(Succeed())
			// Keep the original mtime to isolate the captured-ending rule from the
			// transaction's separate modified-since-opened safety check.
			Expect(os.Chtimes(path, page.LastModified(), page.LastModified())).To(Succeed())
			Expect(tx.Save()).To(Succeed())
			actual, err := os.ReadFile(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(actual)).To(Equal("- task" + ending))
			Expect(graph.Close()).To(Succeed())
		}
	})
})
