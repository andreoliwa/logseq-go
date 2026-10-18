package content_test

import (
	"github.com/andreoliwa/logseq-go/v2/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Reference source", func() {
	It("invalidates captured spelling after a direct target edit", func() {
		ref := content.NewPageRefText("target").CaptureSource("source")
		source, valid := ref.PreservedSource()
		Expect(source).To(Equal("source"))
		Expect(valid).To(BeTrue())
		ref.To = "changed"
		_, valid = ref.PreservedSource()
		Expect(valid).To(BeFalse())
	})
})
