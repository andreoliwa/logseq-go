package logseq_test

import (
	logseq "github.com/andreoliwa/logseq-go"
	"github.com/andreoliwa/logseq-go/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"time"
)

// Deliberately implement the original interface without embedding library types.
type externalPage struct{}

func (*externalPage) Type() logseq.PageType                            { return logseq.PageTypeDedicated }
func (*externalPage) IsNew() bool                                      { return false }
func (*externalPage) Title() string                                    { return "external" }
func (*externalPage) Date() time.Time                                  { return time.Time{} }
func (*externalPage) LastModified() time.Time                          { return time.Time{} }
func (*externalPage) Properties() *content.Properties                  { return content.NewProperties() }
func (*externalPage) Blocks() content.BlockList                        { return nil }
func (*externalPage) AddBlock(*content.Block)                          {}
func (*externalPage) RemoveBlock(*content.Block)                       {}
func (*externalPage) PrependBlock(*content.Block)                      {}
func (*externalPage) InsertBlockAfter(*content.Block, *content.Block)  {}
func (*externalPage) InsertBlockBefore(*content.Block, *content.Block) {}

var _ logseq.Page = (*externalPage)(nil)
var _ func(string) (*content.Block, error) = logseq.ParseBlock
var _ func(string) (content.NodeList, error) = logseq.ParseNodes

var _ = Describe("Existing fork consumers", func() {
	It("retains legacy priority constructors and values", func() {
		var value content.PriorityValue = content.PriorityHigh
		var priority *content.Priority = content.NewPriority(value).WithPriority(content.PriorityMedium)
		Expect(priority.Priority).To(Equal(content.PriorityMedium))
		Expect(content.NewPriorityFromString("b").Priority).To(Equal(content.PriorityMedium))
	})
	It("keeps CLOCK logbook entries raw by default", func() {
		block, err := logseq.ParseBlock("TODO Task\n:LOGBOOK:\nCLOCK: [2026-10-17 Sat 09:00]\n:END:")
		Expect(err).NotTo(HaveOccurred())
		var found bool
		for _, child := range block.Children() {
			if book, ok := child.(*content.Logbook); ok {
				for _, entry := range book.Children() {
					_, found = entry.(*content.LogbookEntryRaw)
					Expect(found).To(BeTrue())
				}
			}
		}
		Expect(found).To(BeTrue())
	})
	It("recognizes leading time and preserves outline source ranges", func() {
		block, err := logseq.ParseBlock("TODO [#B] **9h00** Meeting")
		Expect(err).NotTo(HaveOccurred())
		_, ok := content.LeadingTime(block)
		Expect(ok).To(BeTrue())
		_, ranges, outlined, err := logseq.ParseSource("- TODO [#B] **9h00** Meeting\n")
		Expect(err).NotTo(HaveOccurred())
		Expect(outlined).To(BeTrue())
		Expect(ranges).To(HaveLen(1))
	})
})
