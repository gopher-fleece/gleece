package definitions_test

import (
	"testing"

	"github.com/gopher-fleece/gleece/common"
	"github.com/gopher-fleece/gleece/definitions"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Definitions", func() {
	var _ = Describe("TypeMetadata Equals", func() {

		Context("AliasMetadata Equals", func() {
			It("returns true for identical AliasMetadata", func() {
				a := definitions.AliasMetadata{
					Name:      "LengthUnits",
					AliasType: "string",
					Values:    []string{"Meter", "Kilometer"},
				}
				b := definitions.AliasMetadata{
					Name:      "LengthUnits",
					AliasType: "string",
					Values:    []string{"Meter", "Kilometer"},
				}
				Expect(a.Equals(b)).To(BeTrue())
			})

			It("returns false when name differs", func() {
				a := definitions.AliasMetadata{Name: "A"}
				b := definitions.AliasMetadata{Name: "B"}
				Expect(a.Equals(b)).To(BeFalse())
			})

			It("returns false when alias type differs", func() {
				a := definitions.AliasMetadata{AliasType: "int"}
				b := definitions.AliasMetadata{AliasType: "string"}
				Expect(a.Equals(b)).To(BeFalse())
			})

			It("returns false when values differ in length", func() {
				a := definitions.AliasMetadata{Values: []string{"a", "b"}}
				b := definitions.AliasMetadata{Values: []string{"a"}}
				Expect(a.Equals(b)).To(BeFalse())
			})

			It("returns false when values differ in content", func() {
				a := definitions.AliasMetadata{Values: []string{"a", "b"}}
				b := definitions.AliasMetadata{Values: []string{"a", "c"}}
				Expect(a.Equals(b)).To(BeFalse())
			})
		})

		Context("TypeMetadata Equals", func() {
			var base definitions.TypeMetadata

			BeforeEach(func() {
				base = definitions.TypeMetadata{
					Name:                "Foo",
					PkgPath:             "example.com/pkg",
					DefaultPackageAlias: "pkg",
					Description:         "A test type",
					Import:              common.ImportTypeNone,
					IsUniverseType:      true,
					IsByAddress:         false,
					SymbolKind:          common.SymKindStruct,
					AliasMetadata: &definitions.AliasMetadata{
						Name:      "Units",
						AliasType: "string",
						Values:    []string{"One", "Two"},
					},
				}
			})

			It("returns true for identical TypeMetadata", func() {
				copy := base
				Expect(base.Equals(copy)).To(BeTrue())
			})

			It("returns false if Name differs", func() {
				other := base
				other.Name = "Bar"
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if PkgPath differs", func() {
				other := base
				other.PkgPath = "other/pkg"
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if Description differs", func() {
				other := base
				other.Description = "Different"
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if Import type differs", func() {
				other := base
				other.Import = common.ImportTypeAlias
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if IsUniverseType differs", func() {
				other := base
				other.IsUniverseType = false
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if SymbolKind differs", func() {
				other := base
				other.SymbolKind = common.SymKindInterface
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if AliasMetadata is nil on one side", func() {
				other := base
				other.AliasMetadata = nil
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns true if both AliasMetadata are nil", func() {
				base.AliasMetadata = nil
				other := base
				other.AliasMetadata = nil
				Expect(base.Equals(other)).To(BeTrue())
			})

			It("returns false if AliasMetadata content differs", func() {
				other := base
				// A simple assignment copies the underlying slice ptr so we must do a manual copy.
				copyAlias := *base.AliasMetadata
				copyAlias.Values = []string{"Different"} // safely mutate copy
				other.AliasMetadata = &copyAlias

				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if IsByAddress differs", func() {
				other := base
				other.IsByAddress = true
				Expect(base.Equals(other)).To(BeFalse())
			})

			It("returns false if DefaultPackageAlias differs", func() {
				other := base
				other.DefaultPackageAlias = "different"
				Expect(base.Equals(other)).To(BeFalse())
			})
		})
	})
})

func TestUnitCommons(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Definitions")
}
