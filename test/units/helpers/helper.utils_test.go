package helpers_test

import (
	"fmt"
	"testing"

	"github.com/gopher-fleece/gleece/v2/definitions"
	"github.com/gopher-fleece/gleece/v2/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Helpers", func() {
	Context("Definitions Helpers", func() {
		Context("ConvertToHttpStatus", func() {
			It("Returns an error if given code string cannot be parsed", func() {
				code, err := definitions.ConvertToHttpStatus("cannot be parsed")
				Expect(err).To(MatchError(ContainSubstring("invalid syntax")))
				Expect(code).To(BeEquivalentTo(0))
			})

			It("Returns an error if parsed code is not a valid HTTP Status code", func() {
				code, err := definitions.ConvertToHttpStatus("999")
				Expect(err).To(MatchError(ContainSubstring("is not a valid HTTP status code")))
				Expect(code).To(BeEquivalentTo(0))
			})

			It("Returns a correctly parsed HttpStatusCode if given code string is valid", func() {
				for _, codeUint := range definitions.GetValidHttpStatusCodes() {
					code, err := definitions.ConvertToHttpStatus(fmt.Sprintf("%d", codeUint))
					Expect(err).To(BeNil())
					Expect(code).To(BeEquivalentTo(codeUint))
				}
			})
		})

		Context("PermissionStringToFileMod", func() {
			It("should return no error for valid permissions", func() {
				validPermissions := []string{
					"0000", // No permissions
					"0755", // Standard rwx for owner, read-execute for others
					"0777", // Full read-write-execute for all
					"0700", // Owner-only rwx
					"0644", // rw-r--r--
				}

				for _, perm := range validPermissions {
					_, err := definitions.PermissionStringToFileMod(perm)
					Expect(err).To(BeNil(), fmt.Sprintf("Expected no error for permission %s", perm))
				}
			})

			It("should return an error for invalid permissions", func() {
				invalidPermissions := []string{
					"07778", // Invalid, has a bit outside 07777
					"10000", // Out of range for UNIX permission
					"abc",   // Non-octal string
				}

				for _, perm := range invalidPermissions {
					_, err := definitions.PermissionStringToFileMod(perm)
					Expect(err).To(HaveOccurred(), fmt.Sprintf("Expected error for permission %s", perm))
				}
			})

			It("should return an error for empty permission string", func() {
				_, err := definitions.PermissionStringToFileMod("")
				Expect(err).To(HaveOccurred(), "Expected error for empty permission string")
			})

			It("should return an error for permissions with bits outside valid range", func() {
				_, err := definitions.PermissionStringToFileMod("10000") // 010000 is outside valid permission range
				Expect(err).To(HaveOccurred(), "Expected error for permission 10000")
			})

			It("should not return error for permissions with valid sticky/setuid/setgid bits (07777)", func() {
				_, err := definitions.PermissionStringToFileMod("7777") // Valid, sticky + rwx
				Expect(err).To(BeNil(), "Expected no error for permission 7777")
			})

			It("should return error for permission with non-octal digits", func() {
				_, err := definitions.PermissionStringToFileMod("abcd")
				Expect(err).To(HaveOccurred(), "Expected error for non-octal permission string")
			})

			It("should return no error for boundary valid permissions", func() {
				_, err := definitions.PermissionStringToFileMod("0000") // No permissions
				Expect(err).To(BeNil(), "Expected no error for permission 0000")

				_, err = definitions.PermissionStringToFileMod("0777") // Full permissions
				Expect(err).To(BeNil(), "Expected no error for permission 0777")
			})
		})
	})
})

func TestUnits(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Helpers")
}
