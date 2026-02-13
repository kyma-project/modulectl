//go:build e2e

package scaffold_test

import (
	"io/fs"
	"os"
	"path"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v3"
)

const (
	markerFileData = "test-marker"
)

var _ = Describe("Test 'scaffold' command", Ordered, func() {
	var initialDir string
	var workDir string

	setup := func() {
		var err error
		initialDir, err = os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		workDir = resolveWorkingDirectory()
		err = os.Chdir(workDir)
		Expect(err).ToNot(HaveOccurred())
	}

	teardown := func() {
		err := os.Chdir(initialDir)
		Expect(err).ToNot(HaveOccurred())
		cleanupWorkingDirectory(workDir)
		workDir = ""
		initialDir = ""
	}

	Context("Given an empty directory", func() {
		BeforeAll(func() { setup() })
		AfterAll(func() { teardown() })

		var cmd scaffoldCmd
		It("When `modulectl scaffold` command is invoked without any args", func() {
			cmd = scaffoldCmd{moduleTeam: "jellyfish"}
		})

		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And the module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := (&moduleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given a directory with an existing module configuration file", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("scaffold-module-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd scaffoldCmd
		It("When `modulectl scaffold` command is invoked without any args", func() {
			cmd = scaffoldCmd{
				moduleTeam: "jellyfish",
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("module config file already exists"))

			By("And no files should be generated")
			Expect(filesIn(workDir)).Should(HaveLen(1))
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))
			Expect(getMarkerFileData("scaffold-module-config.yaml")).Should(Equal(markerFileData))
		})
	})

	Context("Given a directory with an existing module configuration file", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("scaffold-module-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd scaffoldCmd
		It("When `modulectl scaffold` command is invoked with --overwrite flag", func() {
			cmd = scaffoldCmd{
				moduleTeam: "jellyfish",
				overwrite:  true,
			}
		})

		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And the module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := (&moduleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given an empty directory", func() {
		BeforeAll(func() { setup() })
		AfterAll(func() { teardown() })

		var cmd scaffoldCmd
		It("When `modulectl scaffold` command args override defaults", func() {
			cmd = scaffoldCmd{
				moduleName:                    "github.com/custom/module",
				moduleVersion:                 "3.2.1",
				moduleTeam:                    "custom-team",
				moduleConfigFileFlag:          "custom-module-config.yaml",
				genManifestFlag:               "custom-manifest.yaml",
				genDefaultCRFlag:              "custom-default-cr.yaml",
				genSecurityScannersConfigFlag: "custom-security-scanners-config.yaml",
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And four files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(4))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-manifest.yaml"))

			By("And the defaultCR file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-default-cr.yaml"))

			By("And the security-scanners-config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-security-scanners-config.yaml"))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-module-config.yaml"))

			By("And the module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "custom-module-config.yaml")
			expectedModConf := cmd.toConfigBuilder().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given a directory with existing files", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("custom-manifest.yaml")).To(Succeed())
			Expect(createMarkerFile("custom-default-cr.yaml")).To(Succeed())
			Expect(createMarkerFile("custom-security-scanners-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd scaffoldCmd
		It("When `modulectl scaffold` command is invoked with arguments that match existing files names", func() {
			cmd = scaffoldCmd{
				moduleTeam:                    "jellyfish",
				genManifestFlag:               "custom-manifest.yaml",
				genDefaultCRFlag:              "custom-default-cr.yaml",
				genSecurityScannersConfigFlag: "custom-security-scanners-config.yaml",
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And there should be four files in the directory")
			Expect(filesIn(workDir)).Should(HaveLen(4))

			By("And the manifest file is reused (not generated)")
			Expect(getMarkerFileData("custom-manifest.yaml")).Should(Equal(markerFileData))

			By("And the defaultCR file is reused (not generated)")
			Expect(getMarkerFileData("custom-default-cr.yaml")).Should(Equal(markerFileData))

			By("And the security-scanners-config file is reused (not generated)")
			Expect(getMarkerFileData("custom-security-scanners-config.yaml")).Should(Equal(markerFileData))

			By("And the module config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-module-config.yaml"))

			By("And module config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-module-config.yaml")
			expectedModConf := cmd.toConfigBuilder().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})
})

// Test helper functions

func getMarkerFileData(name string) string {
	data, err := os.ReadFile(name)
	Expect(err).ToNot(HaveOccurred())
	return string(data)
}

func createMarkerFile(name string) error {
	err := os.WriteFile(name, []byte(markerFileData), 0o600)
	return err
}

func moduleConfigFromFile(dir, fileName string) *moduleConfig {
	filePath := path.Join(dir, fileName)
	data, err := os.ReadFile(filePath)
	Expect(err).ToNot(HaveOccurred())
	res := moduleConfig{}
	err = yaml.Unmarshal(data, &res)
	Expect(err).ToNot(HaveOccurred())
	return &res
}

func filesIn(dir string) []string {
	fi, err := os.Stat(dir)
	Expect(err).ToNot(HaveOccurred())
	Expect(fi.IsDir()).To(BeTrueBecause("The provided path should be a directory: %s", dir))

	dirFs := os.DirFS(dir)
	entries, err := fs.ReadDir(dirFs, ".")
	Expect(err).ToNot(HaveOccurred())

	var res []string
	for _, ent := range entries {
		if ent.Type().IsRegular() {
			res = append(res, ent.Name())
		}
	}

	return res
}

func resolveWorkingDirectory() string {
	scaffoldDir := os.Getenv("SCAFFOLD_DIR")
	if len(scaffoldDir) > 0 {
		return scaffoldDir
	}

	scaffoldDir, err := os.MkdirTemp("", "scaffold_test")
	if err != nil {
		Fail(err.Error())
	}
	return scaffoldDir
}

func cleanupWorkingDirectory(path string) {
	if len(os.Getenv("SCAFFOLD_DIR")) == 0 {
		_ = os.RemoveAll(path)
	}
}
