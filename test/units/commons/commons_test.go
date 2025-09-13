package commons_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/common/linq"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Commons", func() {
	Context("Linq", func() {
		Context("Map", func() {
			It("Correctly maps items using the given function", func() {
				type TestStruct struct {
					A int
					B string
				}

				value := []TestStruct{{A: 1, B: "2"}, {A: 2, B: "3"}, {A: 3, B: "4"}}
				Expect(linq.Map(value, func(v TestStruct) int {
					return v.A
				})).To(Equal([]int{1, 2, 3}))
			})
		})

		Context("MapNonZero", func() {
			It("Correctly maps non-zero items using the given function", func() {
				type TestStruct struct {
					A int
					B string
				}

				value := []TestStruct{{A: 1, B: "2"}, {A: 0, B: ""}, {A: 2, B: "3"}, {A: 3, B: "4"}}
				Expect(linq.MapNonZero(value, func(v TestStruct) int {
					return v.A
				})).To(Equal([]int{1, 2, 3}))
			})
		})

		Context("Filter", func() {
			It("Returns a correctly filtered slice", func() {
				list := []string{"abc", "abcdef", "bc", "oink"}
				result := linq.Filter(list, func(s string) bool {
					return !strings.Contains(s, "bc")
				})
				Expect(result).To(HaveLen(1))
				Expect(result[0]).To(Equal("oink"))
			})
		})

		Context("FilterNil", func() {
			It("Correctly filters non-nil items using the given function", func() {
				type TestStruct struct {
					A int
					B string
				}

				value := []*TestStruct{{A: 1, B: "2"}, nil, {A: 2, B: "3"}, {A: 3, B: "4"}, nil}
				Expect(linq.FilterNil(value)).To(Equal(
					[]*TestStruct{{A: 1, B: "2"}, {A: 2, B: "3"}, {A: 3, B: "4"}},
				))
			})
		})

		Context("Flatten", func() {
			It("Returns a correctly flattened slice", func() {
				list := [][]int{
					{1, 2},
					{3, 4},
					{5, 6},
				}

				result := linq.Flatten(list)
				Expect(result).To(HaveLen(6))
				Expect(result).To(HaveExactElements(1, 2, 3, 4, 5, 6))
			})
		})

		Context("First", func() {
			It("Returns nil if given nil input", func() {
				result := linq.First(nil, func(a int) bool {
					return a != 0
				})

				Expect(result).To(BeNil())
			})

			It("Returns first match if one exists", func() {
				values := []int{0, 0, 0, 1, 5}
				result := linq.First(values, func(a int) bool {
					return a != 0
				})
				Expect(*result).To(Equal(1))
			})

			It("Returns nil if no item matches given filter function", func() {
				values := []int{0, 0, 0, 1, 5}
				result := linq.First(values, func(a int) bool {
					return a < 0
				})
				Expect(result).To(BeNil())
			})
		})
	})

	Context("Iterables", func() {
		Context("Coalesce", func() {
			It("Returns the first non-nil value of the given arguments", func() {
				expectedPtr := common.Ptr(5)
				result := common.Coalesce(nil, nil, nil, nil, expectedPtr, nil, nil, common.Ptr(1))
				Expect(result).To(Equal(expectedPtr))
			})

			It("Returns a zero value if all args are zero", func() {
				result := common.Coalesce[string]("", "", "", "")
				Expect(result).To(Equal(""))
			})
		})

		Context("ConcatConditional", func() {
			It("Returns a concatenated slice that matches the given filter", func() {
				sliceA := []int{1, 2, 3}
				sliceB := []int{-9, -7, 4, -5, 5, 6, 7, -1}
				results := common.ConcatConditional(sliceA, sliceB, func(item int) bool {
					return item > 0
				})

				Expect(results).To(Equal([]int{1, 2, 3, 4, 5, 6, 7}))
			})
		})
	})

	Context("Maps", func() {
		var testMap map[string]int
		BeforeEach(func() {
			testMap = map[string]int{
				"A": 4,
				"B": 3,
				"C": 2,
				"D": 1,
			}
		})

		Context("MapKeys", func() {
			It("Correctly returns map keys", func() {
				Expect(common.MapKeys(testMap)).To(ContainElements("A", "B", "C", "D"))
			})
		})

		Context("MapValues", func() {
			It("Correctly returns map values", func() {
				Expect(common.MapValues(testMap)).To(ContainElements(4, 3, 2, 1))
			})
		})
	})

	Context("Structs", func() {
		Context("Collector", func() {
			It("Add adds items to the collector", func() {
				var c common.Collector[int]
				c.Add(42)
				Expect(c.Items()).To(Equal([]int{42}))
			})

			It("AddMany adds multiple items", func() {
				var c common.Collector[string]
				c.AddMany([]string{"a", "b", "c"})
				Expect(c.Items()).To(Equal([]string{"a", "b", "c"}))
			})

			It("AddIfNotZero adds only non-zero values", func() {
				var c common.Collector[int]
				c.AddIfNotZero(0)
				c.AddIfNotZero(5)
				Expect(c.Items()).To(Equal([]int{5}))
			})

			It("HasAny returns true if collector has items", func() {
				var c common.Collector[bool]
				Expect(c.HasAny()).To(BeFalse())
				c.Add(true)
				Expect(c.HasAny()).To(BeTrue())
			})

			It("Items returns a reference to collector's internal slice", func() {
				var c common.Collector[int]
				c.Add(1)
				c.Add(2)
				items := c.Items()
				Expect(items).To(Equal([]int{1, 2}))

				items[0] = 99
				Expect(c.Items()[0]).To(Equal(99)) // Underlying value should have changed
			})
		})

		Context("ContextualError", func() {
			It("Returns correct message when empty", func() {
				err := &common.ContextualError{Context: "test"}
				Expect(err.Error()).To(Equal("[test] no errors"))
			})

			It("Includes single-line error messages", func() {
				err := &common.ContextualError{
					Context: "database",
					Errors:  []error{fmt.Errorf("connection failed")},
				}
				Expect(err.Error()).To(ContainSubstring("[database] encountered 1 error"))
				Expect(err.Error()).To(ContainSubstring("- connection failed"))
			})

			It("Handles multi-line errors with indentation", func() {
				multiLine := "first line\nsecond line\nthird line"
				err := &common.ContextualError{
					Context: "multi",
					Errors:  []error{fmt.Errorf(multiLine)},
				}
				msg := err.Error()
				Expect(msg).To(ContainSubstring("- first line"))
				Expect(msg).To(ContainSubstring("  second line"))
				Expect(msg).To(ContainSubstring("  third line"))
			})

			It("Skips nil errors", func() {
				err := &common.ContextualError{
					Context: "niltest",
					Errors:  []error{nil, fmt.Errorf("real error")},
				}
				Expect(err.Error()).To(ContainSubstring("- real error"))
				Expect(err.Error()).NotTo(ContainSubstring("<nil>"))
			})

			It("Unwrap returns underlying error slice", func() {
				errs := []error{fmt.Errorf("1"), fmt.Errorf("2")}
				err := &common.ContextualError{Context: "wrap", Errors: errs}
				Expect(err.Unwrap()).To(Equal(errs))
			})
		})
	})

})

func TestUnitCommons(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Commons")
}
