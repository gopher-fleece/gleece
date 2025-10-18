package gast_test

import (
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/gopher-fleece/gleece/gast"
	"github.com/gopher-fleece/gleece/infrastructure/logger"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Unit Tests - Versioning", func() {
	var tmpFilePath string
	var fileContent = []byte("initial content")

	Describe("FileVersion", func() {

		BeforeEach(func() {
			tmpDir := GinkgoT().TempDir()
			tmpFilePath = filepath.Join(tmpDir, "test.txt")
			err := os.WriteFile(tmpFilePath, fileContent, 0644)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("NewFileVersion", func() {
			It("Creates FileVersion from existing file", func() {
				fv, err := gast.NewFileVersion(tmpFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(fv.Path).To(Equal(tmpFilePath))
				Expect(fv.Hash).ToNot(BeEmpty())
			})

			It("Returns error for nonexistent file", func() {
				_, err := gast.NewFileVersion("nonexistent.file")
				Expect(err).To(HaveOccurred())
			})

			It("Returns error if hash computation fails", func() {
				Expect(os.Chmod(tmpFilePath, 0000)).To(Succeed())
				defer os.Chmod(tmpFilePath, 0644)

				_, err := gast.NewFileVersion(tmpFilePath)
				if runtime.GOOS != "windows" {
					Expect(err).To(HaveOccurred())
				}
			})
		})

		Describe("NewFileVersionFromAstFile", func() {
			It("Fails for bad AST/FileSet combo", func() {
				file := &ast.File{}
				fSet := token.NewFileSet()
				_, err := gast.NewFileVersionFromAstFile(file, fSet)
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("Equals", func() {
			It("Returns true for identical metadata", func() {
				fv1, _ := gast.NewFileVersion(tmpFilePath)
				time.Sleep(time.Millisecond * 10) // Ensure OS timestamps don't match coincidentally
				fv2 := fv1
				Expect(fv1.Equals(&fv2)).To(BeTrue())
			})

			It("Returns false for different hash", func() {
				fv1, _ := gast.NewFileVersion(tmpFilePath)

				// Modify file
				Expect(os.WriteFile(tmpFilePath, []byte("changed content"), 0644)).To(Succeed())

				fv2, _ := gast.NewFileVersion(tmpFilePath)
				Expect(fv1.Equals(&fv2)).To(BeFalse())
			})
		})

		Describe("String", func() {
			It("Returns correct string format", func() {
				fv, _ := gast.NewFileVersion(tmpFilePath)
				s := fv.String()
				Expect(s).To(ContainSubstring(fv.Path))
				Expect(s).To(ContainSubstring(fv.Hash))
			})
		})

		Describe("HasChanged", func() {
			It("Returns false if file unchanged", func() {
				fv, _ := gast.NewFileVersion(tmpFilePath)
				changed, err := fv.HasChanged(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(changed).To(BeFalse())
			})

			It("Detects file modification", func() {
				fv, _ := gast.NewFileVersion(tmpFilePath)

				// Delay to ensure mtime change
				time.Sleep(time.Millisecond * 10)

				// Modify content
				Expect(os.WriteFile(tmpFilePath, []byte("updated content"), 0644)).To(Succeed())

				changed, err := fv.HasChanged(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(changed).To(BeTrue())
			})

			It("Updates metadata if selfUpdate = true", func() {
				fv, _ := gast.NewFileVersion(tmpFilePath)

				// Delay and update file
				time.Sleep(time.Millisecond * 10)
				Expect(os.WriteFile(tmpFilePath, []byte("again updated"), 0644)).To(Succeed())

				changed, err := fv.HasChanged(true)
				Expect(err).ToNot(HaveOccurred())
				Expect(changed).To(BeTrue())

				// Should now be unchanged
				changedAgain, err := fv.HasChanged(false)
				Expect(err).ToNot(HaveOccurred())
				Expect(changedAgain).To(BeFalse())
			})

			It("Returns error if file stat fails", func() {
				fv, _ := gast.NewFileVersion(tmpFilePath)

				// Remove the file so os.Stat will fail
				Expect(os.Remove(tmpFilePath)).To(Succeed())

				changed, err := fv.HasChanged(false)
				Expect(err).To(HaveOccurred())
				Expect(changed).To(BeTrue())
			})

			It("Returns error if hash fails", func() {
				// Make file unreadable
				Expect(os.Chmod(tmpFilePath, 0000)).To(Succeed())
				defer os.Chmod(tmpFilePath, 0644) // Reset so cleanup doesn't panic

				if runtime.GOOS != "windows" {
					fv, _ := gast.NewFileVersion(tmpFilePath)
					changed, err := fv.HasChanged(false)
					Expect(err).To(HaveOccurred())
					Expect(changed).To(BeTrue()) // We default to true in case someone forgets to check error
				}
			})

			It("Returns true and error if hash fails after stat succeeds", func() {

				// Ensure modtime is newer to avoid short-circuit
				Expect(os.Chtimes(tmpFilePath, time.Now(), time.Now().Add(time.Second))).To(Succeed())

				// Break file perms to cause Sha256File failure
				Expect(os.Chmod(tmpFilePath, 0000)).To(Succeed())
				defer os.Chmod(tmpFilePath, 0644)

				if runtime.GOOS != "windows" {
					fv, _ := gast.NewFileVersion(tmpFilePath)
					changed, err := fv.HasChanged(false)
					Expect(err).To(HaveOccurred())
					Expect(changed).To(BeTrue()) // We default to true in case someone forgets to check error
				}
			})

		})
	})
})

func TestGastFileVersion(t *testing.T) {
	logger.SetLogLevel(logger.LogLevelNone)
	RegisterFailHandler(Fail)
	RunSpecs(t, "Unit Tests - Versioning")
}
