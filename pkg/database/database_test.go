package database_test

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/SchSeba/slack-ai-assistant/pkg/database"
)

var _ = Describe("Database", func() {
	var (
		db     *database.Database
		tmpDir string
		dbPath string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "slack-ai-test-*")
		Expect(err).NotTo(HaveOccurred())

		dbPath = filepath.Join(tmpDir, "test.db")
		db, err = database.NewDatabase(dbPath)
		Expect(err).NotTo(HaveOccurred())
		Expect(db).NotTo(BeNil())

		err = db.AutoMigrate()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if db != nil {
			Expect(db.Close()).To(Succeed())
		}
		if tmpDir != "" {
			if err := os.RemoveAll(tmpDir); err != nil {
				GinkgoWriter.Printf("Warning: failed to remove temp dir %s: %v\n", tmpDir, err)
			}
		}
	})

	Describe("NewDatabase", func() {
		Context("when creating a new database", func() {
			It("should create a database instance successfully", func() {
				tempDir, err := os.MkdirTemp("", "test-*")
				Expect(err).NotTo(HaveOccurred())
				defer func() {
					if err := os.RemoveAll(tempDir); err != nil {
						GinkgoWriter.Printf("Warning: failed to remove temp dir %s: %v\n", tempDir, err)
					}
				}()

				testPath := filepath.Join(tempDir, "test.db")
				testDB, err := database.NewDatabase(testPath)
				Expect(err).NotTo(HaveOccurred())
				Expect(testDB).NotTo(BeNil())

				Expect(testDB.Close()).To(Succeed())
			})

			It("should handle invalid database path gracefully", func() {
				invalidPath := "/invalid/path/that/does/not/exist/test.db"
				testDB, err := database.NewDatabase(invalidPath)
				Expect(err).To(HaveOccurred())
				Expect(testDB).To(BeNil())
			})
		})
	})

	Describe("AutoMigrate", func() {
		It("should migrate the schema successfully", func() {
			tempDir, err := os.MkdirTemp("", "test-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					GinkgoWriter.Printf("Warning: failed to remove temp dir %s: %v\n", tempDir, err)
				}
			}()

			testPath := filepath.Join(tempDir, "migrate_test.db")
			testDB, err := database.NewDatabase(testPath)
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				if err := testDB.Close(); err != nil {
					GinkgoWriter.Printf("Warning: failed to close test DB: %v\n", err)
				}
			}()

			err = testDB.AutoMigrate()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("CreateSlackThreadWithSlug", func() {
		Context("when creating a new slack thread record", func() {
			It("should create the record successfully", func() {
				err := db.CreateSlackThreadWithSlug("thread123", "slug456")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should allow creating multiple different records", func() {
				err := db.CreateSlackThreadWithSlug("thread1", "slug1")
				Expect(err).NotTo(HaveOccurred())

				err = db.CreateSlackThreadWithSlug("thread2", "slug2")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should fail when creating duplicate slack thread", func() {
				err := db.CreateSlackThreadWithSlug("duplicate_thread", "slug1")
				Expect(err).NotTo(HaveOccurred())

				err = db.CreateSlackThreadWithSlug("duplicate_thread", "slug2")
				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("GetSlugForThread", func() {
		Context("when retrieving an existing thread", func() {
			BeforeEach(func() {
				err := db.CreateSlackThreadWithSlug("existing_thread", "existing_slug")
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the correct slug and found=true", func() {
				slug, found, err := db.GetSlugForThread("existing_thread")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeTrue())
				Expect(slug).To(Equal("existing_slug"))
			})
		})

		Context("when retrieving a non-existing thread", func() {
			It("should return empty slug and found=false", func() {
				slug, found, err := db.GetSlugForThread("non_existing_thread")
				Expect(err).NotTo(HaveOccurred())
				Expect(found).To(BeFalse())
				Expect(slug).To(BeEmpty())
			})
		})
	})

	Describe("Close", func() {
		It("should close the database connection successfully", func() {
			tempDir, err := os.MkdirTemp("", "test-*")
			Expect(err).NotTo(HaveOccurred())
			defer func() {
				if err := os.RemoveAll(tempDir); err != nil {
					GinkgoWriter.Printf("Warning: failed to remove temp dir %s: %v\n", tempDir, err)
				}
			}()

			testPath := filepath.Join(tempDir, "close_test.db")
			testDB, err := database.NewDatabase(testPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(testDB).NotTo(BeNil())

			err = testDB.Close()
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
