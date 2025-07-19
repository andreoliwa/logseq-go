package logseq_test

import (
	"context"
	"gotest.tools/v3/golden"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestTransaction(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Transaction Suite")
}

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
		graph, err = logseq.Open(ctx, tempDir)
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
