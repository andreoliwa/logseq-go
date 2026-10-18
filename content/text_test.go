package content_test

import (
	"github.com/andreoliwa/logseq-go/v2/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Text source spelling", func() {
	It("does not replay old spelling after direct edits", func() {
		text := content.NewText("value").CaptureSource("source")
		source, valid := text.PreservedSource()
		Expect(source).To(Equal("source"))
		Expect(valid).To(BeTrue())
		text.Value = "Changed"
		_, valid = text.PreservedSource()
		Expect(valid).To(BeFalse())
	})
})
