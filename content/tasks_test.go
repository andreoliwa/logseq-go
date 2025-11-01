package content_test

import (
	"time"

	"github.com/andreoliwa/logseq-go/content"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Task status", func() {
	DescribeTable("can be parsed from strings",
		func(input string, expectedStatus content.TaskStatus) {
			taskMarker := content.NewTaskMarkerFromString(input)
			Expect(taskMarker.Status).To(Equal(expectedStatus))
		},
		Entry("TODO", "TODO", content.TaskStatusTodo),
		Entry("DONE", "DONE", content.TaskStatusDone),
		Entry("DOING", "DOING", content.TaskStatusDoing),
		Entry("LATER", "LATER", content.TaskStatusLater),
		Entry("NOW", "NOW", content.TaskStatusNow),
		Entry("CANCELLED", "CANCELLED", content.TaskStatusCancelled),
		Entry("CANCELED", "CANCELED", content.TaskStatusCanceled),
		Entry("IN-PROGRESS", "IN-PROGRESS", content.TaskStatusInProgress),
		Entry("WAIT", "WAIT", content.TaskStatusWait),
		Entry("WAITING", "WAITING", content.TaskStatusWaiting),
	)

	DescribeTable("returns an error for invalid input",
		func(invalidInput string) {
			taskMarker := content.NewTaskMarkerFromString(invalidInput)
			Expect(taskMarker.Status).To(Equal(content.TaskStatusNone))
		},
		Entry("Invalid input", "Invalid"),
		Entry("Empty input", ""))
})

var _ = Describe("Task categories", func() {
	DescribeTable("returns the correct category for each status",
		func(status content.TaskStatus, expectedCategory content.TaskCategory) {
			Expect(status.Category()).To(Equal(expectedCategory))
		},
		// TaskCategoryNone
		Entry("TaskStatusNone", content.TaskStatusNone, content.TaskCategoryNone),
		// TaskCategoryTodo
		Entry("TaskStatusTodo", content.TaskStatusTodo, content.TaskCategoryTodo),
		Entry("TaskStatusLater", content.TaskStatusLater, content.TaskCategoryTodo),
		// TaskCategoryDoing
		Entry("TaskStatusDoing", content.TaskStatusDoing, content.TaskCategoryDoing),
		Entry("TaskStatusNow", content.TaskStatusNow, content.TaskCategoryDoing),
		Entry("TaskStatusInProgress", content.TaskStatusInProgress, content.TaskCategoryDoing),
		// TaskCategoryCancelled
		Entry("TaskStatusCancelled", content.TaskStatusCancelled, content.TaskCategoryCancelled),
		Entry("TaskStatusCanceled", content.TaskStatusCanceled, content.TaskCategoryCancelled),
		// TaskCategoryWaiting
		Entry("TaskStatusWaiting", content.TaskStatusWaiting, content.TaskCategoryWaiting),
		Entry("TaskStatusWait", content.TaskStatusWait, content.TaskCategoryWaiting),
		// TaskCategoryDone
		Entry("TaskStatusDone", content.TaskStatusDone, content.TaskCategoryDone),
	)
})

var _ = Describe("Task status transitions to TODO", func() {
	var (
		frozenTime time.Time
		block      *content.Block
		paragraph  *content.Paragraph
		taskMarker *content.TaskMarker
	)

	BeforeEach(func() {
		frozenTime = time.Date(2025, 4, 5, 3, 0, 0, 0, time.Local)
	})

	createTaskWithStatus := func(status content.TaskStatus) *content.TaskMarker {
		block = content.NewBlock()
		paragraph = content.NewParagraph(content.NewText("Task text"))
		block.AddChild(paragraph)

		taskMarker = content.NewTaskMarker(status)
		taskMarker.SetParentReferences(paragraph, block)
		// Override time provider for testing
		taskMarker.SetTimeNow(func() time.Time { return frozenTime })
		paragraph.PrependChild(taskMarker)

		return taskMarker
	}

	// Test transitions from categories that require no side effects
	Describe("from TaskCategoryNone, TaskCategoryTodo, or TaskCategoryWaiting", func() {
		DescribeTable("should change status without side effects",
			func(fromStatus content.TaskStatus) {
				task := createTaskWithStatus(fromStatus)
				result, err := task.WithStatus(content.TaskStatusTodo)

				Expect(err).ToNot(HaveOccurred())
				Expect(result.Status).To(Equal(content.TaskStatusTodo))
			},
			Entry("from TaskStatusNone", content.TaskStatusNone),
			Entry("from TaskStatusTodo", content.TaskStatusTodo),
			Entry("from TaskStatusLater", content.TaskStatusLater),
			Entry("from TaskStatusWait", content.TaskStatusWait),
			Entry("from TaskStatusWaiting", content.TaskStatusWaiting),
		)
	})

	// Test transitions from TaskCategoryDoing (DOING, NOW, IN-PROGRESS)
	Describe("from TaskCategoryDoing", func() {
		Context("when there is no logbook", func() {
			DescribeTable("should change status without errors",
				func(fromStatus content.TaskStatus) {
					task := createTaskWithStatus(fromStatus)
					result, err := task.WithStatus(content.TaskStatusTodo)

					Expect(err).ToNot(HaveOccurred())
					Expect(result.Status).To(Equal(content.TaskStatusTodo))
				},
				Entry("from TaskStatusDoing", content.TaskStatusDoing),
				Entry("from TaskStatusNow", content.TaskStatusNow),
				Entry("from TaskStatusInProgress", content.TaskStatusInProgress),
			)
		})

		Context("when there is a logbook with no running CLOCK", func() {
			It("should change status without modifying the logbook", func() {
				task := createTaskWithStatus(content.TaskStatusDoing)

				logbook := content.NewLogbook(
					content.NewLogbookEntryRaw("CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 02:00:00] =>  01:00:00"),
				)
				block.AddChild(logbook)

				result, err := task.WithStatus(content.TaskStatusTodo)

				Expect(err).ToNot(HaveOccurred())
				Expect(result.Status).To(Equal(content.TaskStatusTodo))

				// Verify logbook entry is unchanged
				entry := logbook.FirstChild().(*content.LogbookEntryRaw)
				Expect(entry.Value).To(Equal("CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 02:00:00] =>  01:00:00"))
			})
		})

		Context("when there is a logbook with a running CLOCK", func() {
			DescribeTable("should stop the running CLOCK",
				func(fromStatus content.TaskStatus, clockStart string, expectedClock string) {
					task := createTaskWithStatus(fromStatus)

					logbook := content.NewLogbook(
						content.NewLogbookEntryRaw(clockStart),
					)
					block.AddChild(logbook)

					result, err := task.WithStatus(content.TaskStatusTodo)

					Expect(err).ToNot(HaveOccurred())
					Expect(result.Status).To(Equal(content.TaskStatusTodo))

					// Verify CLOCK was stopped with correct end time and duration
					entry := logbook.FirstChild().(*content.LogbookEntryRaw)
					Expect(entry.Value).To(Equal(expectedClock))
				},
				Entry("from TaskStatusDoing",
					content.TaskStatusDoing,
					"CLOCK: [2025-04-05 Sat 01:00:00]",
					"CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00"),
				Entry("from TaskStatusNow",
					content.TaskStatusNow,
					"CLOCK: [2025-04-05 Sat 02:30:00]",
					"CLOCK: [2025-04-05 Sat 02:30:00]--[2025-04-05 Sat 03:00:00] =>  00:30:00"),
				Entry("from TaskStatusInProgress",
					content.TaskStatusInProgress,
					"CLOCK: [2025-04-05 Sat 02:45:30]",
					"CLOCK: [2025-04-05 Sat 02:45:30]--[2025-04-05 Sat 03:00:00] =>  00:14:30"),
			)
		})
	})

	// Test transitions from TaskCategoryDone
	Describe("from TaskCategoryDone", func() {
		It("should remove the completed property", func() {
			task := createTaskWithStatus(content.TaskStatusDone)

			// Add completed property
			properties := block.Properties()
			properties.Set("completed", content.NewText("2025-04-05"))

			result, err := task.WithStatus(content.TaskStatusTodo)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.Status).To(Equal(content.TaskStatusTodo))

			// Verify completed property was removed
			Expect(properties.Get("completed")).To(BeEmpty())
		})

		It("should not error if completed property doesn't exist", func() {
			task := createTaskWithStatus(content.TaskStatusDone)

			result, err := task.WithStatus(content.TaskStatusTodo)

			Expect(err).ToNot(HaveOccurred())
			Expect(result.Status).To(Equal(content.TaskStatusTodo))
		})
	})

	// Test transitions from TaskCategoryCancelled (CANCELLED, CANCELED)
	Describe("from TaskCategoryCancelled", func() {
		DescribeTable("should remove the cancelled property",
			func(fromStatus content.TaskStatus) {
				task := createTaskWithStatus(fromStatus)

				// Add cancelled property
				properties := block.Properties()
				properties.Set("cancelled", content.NewText("2025-04-05"))

				result, err := task.WithStatus(content.TaskStatusTodo)

				Expect(err).ToNot(HaveOccurred())
				Expect(result.Status).To(Equal(content.TaskStatusTodo))

				// Verify cancelled property was removed
				Expect(properties.Get("cancelled")).To(BeEmpty())
			},
			Entry("from TaskStatusCancelled", content.TaskStatusCancelled),
			Entry("from TaskStatusCanceled", content.TaskStatusCanceled),
		)
	})
})

var _ = Describe("CLOCK entry parsing", func() {
	Describe("ParseClockEntry", func() {
		It("should parse a running CLOCK entry", func() {
			raw := "CLOCK: [2025-04-05 Sat 01:00:00]"
			entry, err := content.ParseClockEntry(raw)

			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
			Expect(entry.StartTime).To(Equal(time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local)))
			Expect(entry.EndTime).To(BeNil())
			Expect(entry.Duration).To(BeNil())
		})

		It("should parse a completed CLOCK entry", func() {
			raw := "CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00"
			entry, err := content.ParseClockEntry(raw)

			Expect(err).ToNot(HaveOccurred())
			Expect(entry).ToNot(BeNil())
			Expect(entry.StartTime).To(Equal(time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local)))
			Expect(*entry.EndTime).To(Equal(time.Date(2025, 4, 5, 3, 0, 0, 0, time.Local)))
			Expect(*entry.Duration).To(Equal(2 * time.Hour))
		})

		It("should return error for invalid format", func() {
			raw := "Invalid CLOCK entry"
			_, err := content.ParseClockEntry(raw)

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("FormatClockEntry", func() {
		It("should format a running CLOCK entry", func() {
			entry := &content.ClockEntryData{
				StartTime: time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local),
			}
			formatted := content.FormatClockEntry(entry)

			Expect(formatted).To(Equal("CLOCK: [2025-04-05 Sat 01:00:00]"))
		})

		It("should format a completed CLOCK entry", func() {
			endTime := time.Date(2025, 4, 5, 3, 0, 0, 0, time.Local)
			duration := 2 * time.Hour
			entry := &content.ClockEntryData{
				StartTime: time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local),
				EndTime:   &endTime,
				Duration:  &duration,
			}
			formatted := content.FormatClockEntry(entry)

			Expect(formatted).To(Equal("CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00"))
		})
	})

	Describe("IsRunningClock", func() {
		It("should return true for running CLOCK", func() {
			raw := "CLOCK: [2025-04-05 Sat 01:00:00]"
			Expect(content.IsRunningClock(raw)).To(BeTrue())
		})

		It("should return false for completed CLOCK", func() {
			raw := "CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00"
			Expect(content.IsRunningClock(raw)).To(BeFalse())
		})

		It("should return false for invalid CLOCK", func() {
			raw := "Invalid CLOCK"
			Expect(content.IsRunningClock(raw)).To(BeFalse())
		})
	})

	Describe("StopClock", func() {
		It("should stop a running CLOCK", func() {
			raw := "CLOCK: [2025-04-05 Sat 01:00:00]"
			now := time.Date(2025, 4, 5, 3, 0, 0, 0, time.Local)

			stopped, err := content.StopClock(raw, now)

			Expect(err).ToNot(HaveOccurred())
			Expect(stopped).To(Equal("CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00"))
		})

		It("should return error for already stopped CLOCK", func() {
			raw := "CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00"
			now := time.Date(2025, 4, 5, 4, 0, 0, 0, time.Local)

			_, err := content.StopClock(raw, now)

			Expect(err).To(HaveOccurred())
		})
	})
})

var _ = Describe("CalculateClockDuration and FormatClockEntry", func() {
	DescribeTable("should calculate and format duration correctly",
		func(start, end time.Time, expectedDuration time.Duration, expectedFormatted string) {
			duration := content.CalculateClockDuration(start, end)
			Expect(duration).To(Equal(expectedDuration))

			entry := &content.ClockEntryData{
				StartTime: start,
				EndTime:   &end,
				Duration:  &duration,
			}
			formatted := content.FormatClockEntry(entry)
			Expect(formatted).To(Equal(expectedFormatted))
		},
		Entry("9 seconds",
			time.Date(2025, 11, 1, 18, 51, 36, 0, time.Local),
			time.Date(2025, 11, 1, 18, 51, 45, 0, time.Local),
			9*time.Second,
			"CLOCK: [2025-11-01 Sat 18:51:36]--[2025-11-01 Sat 18:51:45] =>  00:00:09",
		),
		Entry("1 minute",
			time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local),
			time.Date(2025, 4, 5, 1, 1, 0, 0, time.Local),
			1*time.Minute,
			"CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 01:01:00] =>  00:01:00",
		),
		Entry("1 hour",
			time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local),
			time.Date(2025, 4, 5, 2, 0, 0, 0, time.Local),
			1*time.Hour,
			"CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 02:00:00] =>  01:00:00",
		),
		Entry("2 hours",
			time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local),
			time.Date(2025, 4, 5, 3, 0, 0, 0, time.Local),
			2*time.Hour,
			"CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 03:00:00] =>  02:00:00",
		),
		Entry("1 hour 30 minutes 45 seconds",
			time.Date(2025, 4, 5, 1, 0, 0, 0, time.Local),
			time.Date(2025, 4, 5, 2, 30, 45, 0, time.Local),
			1*time.Hour+30*time.Minute+45*time.Second,
			"CLOCK: [2025-04-05 Sat 01:00:00]--[2025-04-05 Sat 02:30:45] =>  01:30:45",
		),
		Entry("23 hours 59 minutes 59 seconds",
			time.Date(2025, 4, 5, 0, 0, 0, 0, time.Local),
			time.Date(2025, 4, 5, 23, 59, 59, 0, time.Local),
			23*time.Hour+59*time.Minute+59*time.Second,
			"CLOCK: [2025-04-05 Sat 00:00:00]--[2025-04-05 Sat 23:59:59] =>  23:59:59",
		),
		Entry("1 day",
			time.Date(2025, 11, 1, 19, 0, 0, 0, time.Local),
			time.Date(2025, 11, 2, 19, 0, 0, 0, time.Local),
			24*time.Hour,
			"CLOCK: [2025-11-01 Sat 19:00:00]--[2025-11-02 Sun 19:00:00] =>  24:00:00",
		),
		Entry("1 week",
			time.Date(2025, 11, 1, 19, 0, 0, 0, time.Local),
			time.Date(2025, 11, 8, 19, 0, 0, 0, time.Local),
			7*24*time.Hour,
			"CLOCK: [2025-11-01 Sat 19:00:00]--[2025-11-08 Sat 19:00:00] =>  168:00:00",
		),
		Entry("1 month (31 days)",
			time.Date(2025, 10, 1, 19, 1, 35, 0, time.UTC),
			time.Date(2025, 11, 1, 19, 1, 52, 0, time.UTC),
			744*time.Hour+17*time.Second, // 31 days × 24 hours
			"CLOCK: [2025-10-01 Wed 19:01:35]--[2025-11-01 Sat 19:01:52] =>  744:00:17",
		),
		Entry("2 years",
			time.Date(2023, 11, 1, 19, 3, 0, 0, time.Local),
			time.Date(2025, 11, 1, 19, 3, 15, 0, time.Local),
			17544*time.Hour+15*time.Second,
			"CLOCK: [2023-11-01 Wed 19:03:00]--[2025-11-01 Sat 19:03:15] =>  17544:00:15",
		),
	)
})
