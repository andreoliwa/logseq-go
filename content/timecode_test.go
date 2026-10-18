package content_test

import (
	"fmt"
	"strings"

	"github.com/andreoliwa/logseq-go/v2/content"
	"github.com/andreoliwa/logseq-go/v2/internal/markdown"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Leading timecodes", func() {
	It("recognizes every separator, hour width, prefix and supported wrapper without mutation", func() {
		prefixes := []string{"", "[#A] ", "[#B] ", "[#C] "}
		for _, status := range content.TaskStatusStrings() {
			for _, priority := range []string{"", "[#A] ", "[#B] ", "[#C] "} {
				prefixes = append(prefixes, status+" "+priority)
			}
		}
		for _, hour := range []string{"9", "09"} {
			for _, sep := range []string{":", "h", "H", ".", ",", ";"} {
				for _, prefix := range prefixes {
					for _, wrapper := range []string{"%s Meeting", "**%s** Meeting", "*%s* Meeting", "_%s_ Meeting", "**%s Meeting**", "*%s Meeting*", "***%s Meeting***", "**_%s Meeting_**", "%s"} {
						input := prefix + fmt.Sprintf(wrapper, hour+sep+"00")
						block, err := markdown.ParseString(input)
						Expect(err).NotTo(HaveOccurred())
						before, err := markdown.AsString(block)
						Expect(err).NotTo(HaveOccurred())
						value, ok := content.LeadingTime(block)
						Expect(ok).To(BeTrue(), input)
						Expect(value.Hour()).To(Equal(9), input)
						Expect(value.Minute()).To(Equal(0), input)
						Expect(value.Minutes()).To(Equal(540), input)
						after, err := markdown.AsString(block)
						Expect(err).NotTo(HaveOccurred())
						Expect(after).To(Equal(before), input)
					}
				}
			}
		}
	})
	It("validates the day boundaries", func() {
		for _, input := range []string{"00:00", "23:59"} {
			b, err := markdown.ParseString(input)
			Expect(err).NotTo(HaveOccurred())
			value, ok := content.LeadingTime(b)
			Expect(ok).To(BeTrue())
			Expect(value.String()).To(Equal(input))
		}
		_, ok := content.NewTimeOfDay(24, 0)
		Expect(ok).To(BeFalse())
		_, ok = content.NewTimeOfDay(0, 60)
		Expect(ok).To(BeFalse())
		_, ok = content.NewTimeOfDay(-1, 0)
		Expect(ok).To(BeFalse())
	})
	It("rejects malformed tokens and excluded contexts", func() {
		for _, input := range []string{"25:00 x", "12:60 x", "9:5 x", "900", "9", "9am", "9:00Meeting", "09:00:30", "009:00", "9:000", "Meet at 9h00", "`9h00` Meeting", "[9h00](url)", "[[9h00]]", "```\n9h00\n```", "time:: 9h00", ":LOGBOOK:\nCLOCK: [2025-01-01 Wed 09:00:00]\n:END:", "Notes\n9h00 Meeting", "**9h00**Meeting", "~~9h00~~ Meeting"} {
			b, err := markdown.ParseString(input)
			Expect(err).NotTo(HaveOccurred())
			_, ok := content.LeadingTime(b)
			Expect(ok).To(BeFalse(), input)
		}
	})
})

var _ = Describe("Explicit leading-time normalization", func() {
	It("normalizes only the time and preserves description wrappers", func() {
		cases := map[string]string{
			"9:00 Meeting":                             "09:00 Meeting",
			"9h00 Meeting":                             "09:00 Meeting",
			"9H00 Meeting":                             "09:00 Meeting",
			"9.00 Meeting":                             "09:00 Meeting",
			"9,00 Meeting":                             "09:00 Meeting",
			"9;00 Meeting":                             "09:00 Meeting",
			"TODO [#B] **9h00** Meeting":               "TODO [#B] 09:00 Meeting",
			"TODO *9h00* **Meeting**":                  "TODO 09:00 **Meeting**",
			"**9h00 Meeting**":                         "09:00 **Meeting**",
			"*9h00 Meeting*":                           "09:00 *Meeting*",
			"_9h00 Meeting_":                           "09:00 _Meeting_",
			"***9h00 Meeting***":                       "09:00 ***Meeting***",
			"**_9h00 Meeting_**":                       "09:00 **_Meeting_**",
			"**9h00**":                                 "09:00",
			"*9h00*\n10h00 Follow-up":                  "09:00\n10h00 Follow-up",
			"**9h00 Meeting\n10h00 Follow-up**":        "09:00 **Meeting\n10h00 Follow-up**",
			"9h00 Meeting [link](https://example.com)": "09:00 Meeting [link](https://example.com)",
		}
		for input, output := range cases {
			b, err := markdown.ParseString(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(content.NormalizeLeadingTime(b)).To(BeTrue(), input)
			actual, err := markdown.AsString(b)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(output), input)
			Expect(content.NormalizeLeadingTime(b)).To(BeFalse(), input)
		}
	})
	It("does not rewrite canonical, absent or invalid times", func() {
		for _, input := range []string{"TODO 09:00 Meeting", "25:00 Meeting", "9:5 Meeting", "Meet at 9h00", "Notes\n9h00 Meeting", "`9h00` Meeting"} {
			b, err := markdown.ParseString(input)
			Expect(err).NotTo(HaveOccurred())
			before, err := markdown.AsString(b)
			Expect(err).NotTo(HaveOccurred())
			Expect(content.NormalizeLeadingTime(b)).To(BeFalse(), input)
			after, err := markdown.AsString(b)
			Expect(err).NotTo(HaveOccurred())
			Expect(after).To(Equal(before))
		}
	})
	It("preserves child blocks, properties and logbook nodes", func() {
		input := "- TODO **9h00** Meeting\n  id:: 12345678-1234-1234-1234-123456789abc\n  custom:: keep\n  :LOGBOOK:\n  CLOCK: [2025-01-01 Wed 09:00:00]--[2025-01-01 Wed 10:00:00] =>  01:00:00\n  :END:\n  - 10h00 [Child](https://example.com)"
		root, err := markdown.ParseString(input)
		Expect(err).NotTo(HaveOccurred())
		b := root.Blocks()[0]
		before, err := markdown.AsString(root)
		Expect(err).NotTo(HaveOccurred())
		child := b.Blocks()[0]
		Expect(content.NormalizeLeadingTime(b)).To(BeTrue())
		after, err := markdown.AsString(root)
		Expect(err).NotTo(HaveOccurred())
		Expect(after).To(Equal(strings.Replace(before, "**9h00**", "09:00", 1)))
		Expect(b.Blocks()[0]).To(BeIdenticalTo(child))
	})
})

var _ = Describe("Timecode whitespace and line boundaries", func() {
	It("preserves a continuation immediately after a formatted time", func() {
		for input, output := range map[string]string{
			"**9h00\nMeeting**":     "09:00\n**Meeting**",
			"*9h00\nMeeting*":       "09:00\n*Meeting*",
			"9h00\u00a0Meeting":     "09:00\u00a0Meeting",
			"**9h00\u00a0Meeting**": "09:00\u00a0**Meeting**",
		} {
			b, err := markdown.ParseString(input)
			Expect(err).NotTo(HaveOccurred())
			Expect(content.NormalizeLeadingTime(b)).To(BeTrue(), input)
			actual, err := markdown.AsString(b)
			Expect(err).NotTo(HaveOccurred())
			Expect(actual).To(Equal(output), input)
			Expect(content.NormalizeLeadingTime(b)).To(BeFalse())
		}
	})
})

var _ = Describe("Physical first-line time boundaries", func() {
	It("does not skip a property line to normalize continuation content", func() {
		for _, input := range []string{
			"time:: 9h00\n9h00 Meeting",
			"time:: value\nTODO 9h00 Meeting",
			"time:: value\n\n9h00 Meeting",
		} {
			block, err := markdown.ParseString(input)
			Expect(err).NotTo(HaveOccurred())
			_, ok := content.LeadingTime(block)
			Expect(ok).To(BeFalse(), input)
			Expect(content.NormalizeLeadingTime(block)).To(BeFalse(), input)
		}
	})
	It("allows an empty programmatically added properties container", func() {
		block, err := markdown.ParseString("9h00 Meeting")
		Expect(err).NotTo(HaveOccurred())
		block.Properties()
		clock, ok := content.LeadingTime(block)
		Expect(ok).To(BeTrue())
		Expect(clock.Minutes()).To(Equal(540))
	})
})
