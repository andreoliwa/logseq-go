package content_test

import (
	"github.com/andreoliwa/logseq-go/v2/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Priority API", func() {
	It("uses an enum value in a TaskPriority node", func() {
		var value content.Priority = content.PriorityA
		var node *content.TaskPriority = content.NewTaskPriority(value)
		Expect(node.Priority).To(Equal(value))
		Expect(node.WithPriority(content.PriorityB)).To(BeIdenticalTo(node))
		Expect(node.Priority).To(Equal(content.PriorityB))
		Expect(content.PriorityHigh).To(Equal(content.PriorityA))
		Expect(content.PriorityMedium).To(Equal(content.PriorityB))
		Expect(content.PriorityLow).To(Equal(content.PriorityC))
	})
	It("keeps convenience constructors on the same node type", func() {
		var node *content.TaskPriority = content.NewPriority(content.PriorityC)
		Expect(node.Priority).To(Equal(content.PriorityC))
		node = content.NewPriorityFromString("b")
		Expect(node.Priority).To(Equal(content.PriorityB))
	})
})
