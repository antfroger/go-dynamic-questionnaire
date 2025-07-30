package go_dynamic_questionnaire

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Util", func() {
	Describe("contains", func() {
		When("the slice is empty", func() {
			It("should return false for any item", func() {
				slice := []string{}
				Expect(contains(slice, "any")).To(BeFalse())
				Expect(contains(slice, "")).To(BeFalse())
			})
		})

		When("the slice contains the item", func() {
			It("should return true for exact matches", func() {
				slice := []string{"apple", "banana", "cherry"}
				Expect(contains(slice, "apple")).To(BeTrue())
				Expect(contains(slice, "banana")).To(BeTrue())
				Expect(contains(slice, "cherry")).To(BeTrue())
			})

			It("should return true for empty string when slice contains empty string", func() {
				slice := []string{"apple", "", "cherry"}
				Expect(contains(slice, "")).To(BeTrue())
			})

			It("should return true for duplicate items", func() {
				slice := []string{"apple", "banana", "apple", "cherry"}
				Expect(contains(slice, "apple")).To(BeTrue())
			})
		})

		When("the slice does not contain the item", func() {
			It("should return false for non-existent items", func() {
				slice := []string{"apple", "banana", "cherry"}
				Expect(contains(slice, "orange")).To(BeFalse())
				Expect(contains(slice, "Apple")).To(BeFalse())  // case sensitive
				Expect(contains(slice, "apple ")).To(BeFalse()) // trailing space
				Expect(contains(slice, " apple")).To(BeFalse()) // leading space
				Expect(contains(slice, "")).To(BeFalse())
			})
		})

		When("the slice contains special characters", func() {
			It("should handle special characters correctly", func() {
				slice := []string{"hello world", "test@example.com", "file/path", "line\nbreak"}
				Expect(contains(slice, "hello world")).To(BeTrue())
				Expect(contains(slice, "test@example.com")).To(BeTrue())
				Expect(contains(slice, "file/path")).To(BeTrue())
				Expect(contains(slice, "line\nbreak")).To(BeTrue())
			})
		})
	})
})
