package scaffold_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path"

	"gopkg.in/yaml.v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	markerFileData = "test-marker"
)

var _ = Describe("Create Scaffold Command", Ordered, func() {
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

		var cmd createScaffoldCmd
		It("When `modulectl create scaffold` command is invoked without any args", func() {
			cmd = createScaffoldCmd{}
		})

		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the create config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-create-config.yaml"))

			By("And the create config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-create-config.yaml")
			expectedModConf := (&moduleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given a directory with an existing create configuration file", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("scaffold-create-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `modulectl create scaffold` command is invoked without any args", func() {
			cmd = createScaffoldCmd{}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("create config file already exists"))

			By("And no files should be generated")
			Expect(filesIn(workDir)).Should(HaveLen(1))
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-create-config.yaml"))
			Expect(getMarkerFileData("scaffold-create-config.yaml")).Should(Equal(markerFileData))
		})
	})

	Context("Given a directory with an existing create configuration file", func() {
		BeforeAll(func() {
			setup()
			Expect(createMarkerFile("scaffold-create-config.yaml")).To(Succeed())
		})
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `modulectl create scaffold` command is invoked with --overwrite flag", func() {
			cmd = createScaffoldCmd{
				overwrite: true,
			}
		})

		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And two files are generated")
			Expect(filesIn(workDir)).Should(HaveLen(2))

			By("And the manifest file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("manifest.yaml"))

			By("And the create config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-create-config.yaml"))

			By("And the create config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-create-config.yaml")
			expectedModConf := (&moduleConfigBuilder{}).defaults().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})

	Context("Given an empty directory", func() {
		BeforeAll(func() { setup() })
		AfterAll(func() { teardown() })

		var cmd createScaffoldCmd
		It("When `modulectl create scaffold` command args override defaults", func() {
			cmd = createScaffoldCmd{
				moduleName:                    "github.com/custom/create",
				moduleVersion:                 "3.2.1",
				moduleChannel:                 "custom",
				moduleConfigFileFlag:          "custom-create-config.yaml",
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

			By("And the create config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("custom-create-config.yaml"))

			By("And the create config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "custom-create-config.yaml")
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

		var cmd createScaffoldCmd
		It("When `modulectl create scaffold` command is invoked with arguments that match existing files names", func() {
			cmd = createScaffoldCmd{
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

			By("And the create config file is generated")
			Expect(filesIn(workDir)).Should(ContainElement("scaffold-create-config.yaml"))

			By("And create config contains expected entries")
			actualModConf := moduleConfigFromFile(workDir, "scaffold-create-config.yaml")
			expectedModConf := cmd.toConfigBuilder().get()
			Expect(actualModConf).To(BeEquivalentTo(expectedModConf))
		})
	})
})

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

	scaffoldDir, err := os.MkdirTemp("", "create_scaffold_test")
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

type createScaffoldCmd struct {
	moduleName                    string
	moduleVersion                 string
	moduleChannel                 string
	moduleConfigFileFlag          string
	genDefaultCRFlag              string
	genSecurityScannersConfigFlag string
	genManifestFlag               string
	overwrite                     bool
}

func (cmd *createScaffoldCmd) execute() error {
	var command *exec.Cmd

	args := []string{"create", "scaffold"}

	if cmd.moduleName != "" {
		args = append(args, "--create-name="+cmd.moduleName)
	}

	if cmd.moduleVersion != "" {
		args = append(args, "--create-version="+cmd.moduleVersion)
	}

	if cmd.moduleChannel != "" {
		args = append(args, "--create-channel="+cmd.moduleChannel)
	}

	if cmd.moduleConfigFileFlag != "" {
		args = append(args, "--create-config="+cmd.moduleConfigFileFlag)
	}

	if cmd.genDefaultCRFlag != "" {
		args = append(args, "--gen-default-cr="+cmd.genDefaultCRFlag)
	}

	if cmd.genSecurityScannersConfigFlag != "" {
		args = append(args, "--gen-security-config="+cmd.genSecurityScannersConfigFlag)
	}

	if cmd.genManifestFlag != "" {
		args = append(args, "--gen-manifest="+cmd.genManifestFlag)
	}

	if cmd.overwrite {
		args = append(args, "--overwrite=true")
	}

	command = exec.Command("modulectl", args...)
	cmdOut, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create scaffold command failed with output: %s and error: %w", cmdOut, err)
	}
	return nil
}

func (cmd *createScaffoldCmd) toConfigBuilder() *moduleConfigBuilder {
	res := &moduleConfigBuilder{}
	res.defaults()
	if cmd.moduleName != "" {
		res.withName(cmd.moduleName)
	}
	if cmd.moduleVersion != "" {
		res.withVersion(cmd.moduleVersion)
	}
	if cmd.moduleChannel != "" {
		res.withChannel(cmd.moduleChannel)
	}
	if cmd.genDefaultCRFlag != "" {
		res.withDefaultCRPath(cmd.genDefaultCRFlag)
	}
	if cmd.genSecurityScannersConfigFlag != "" {
		res.withSecurityScannersPath(cmd.genSecurityScannersConfigFlag)
	}
	if cmd.genManifestFlag != "" {
		res.withManifestPath(cmd.genManifestFlag)
	}
	return res
}

// moduleConfigBuilder is used to simplify create.Config creation for testing purposes
type moduleConfigBuilder struct {
	moduleConfig
}

func (mcb *moduleConfigBuilder) get() *moduleConfig {
	res := mcb.moduleConfig
	return &res
}

func (mcb *moduleConfigBuilder) withName(val string) *moduleConfigBuilder {
	mcb.Name = val
	return mcb
}

func (mcb *moduleConfigBuilder) withVersion(val string) *moduleConfigBuilder {
	mcb.Version = val
	return mcb
}

func (mcb *moduleConfigBuilder) withChannel(val string) *moduleConfigBuilder {
	mcb.Channel = val
	return mcb
}

func (mcb *moduleConfigBuilder) withManifestPath(val string) *moduleConfigBuilder {
	mcb.ManifestPath = val
	return mcb
}

func (mcb *moduleConfigBuilder) withDefaultCRPath(val string) *moduleConfigBuilder {
	mcb.DefaultCRPath = val
	return mcb
}

func (mcb *moduleConfigBuilder) withSecurityScannersPath(val string) *moduleConfigBuilder {
	mcb.Security = val
	return mcb
}

func (mcb *moduleConfigBuilder) defaults() *moduleConfigBuilder {
	return mcb.
		withName("kyma-project.io/create/mymodule").
		withVersion("0.0.1").
		withChannel("regular").
		withManifestPath("manifest.yaml")
}

// This is a copy of the moduleConfig struct from internal/scaffold/contentprovider/moduleconfig.go
// to not make the moduleConfig public just for the sake of testing.
// It is expected that the moduleConfig struct will be made public in the future when introducing more commands.
// Once it is public, this struct should be removed.
type moduleConfig struct {
	Name          string            `yaml:"name" comment:"required, the name of the Module"`
	Version       string            `yaml:"version" comment:"required, the version of the Module"`
	Channel       string            `yaml:"channel" comment:"required, channel that should be used in the ModuleTemplate"`
	ManifestPath  string            `yaml:"manifest" comment:"required, relative path or remote URL to the manifests"`
	Mandatory     bool              `yaml:"mandatory" comment:"optional, default=false, indicates whether the create is mandatory to be installed on all clusters"`
	DefaultCRPath string            `yaml:"defaultCR" comment:"optional, relative path or remote URL to a YAML file containing the default CR for the create"`
	ResourceName  string            `yaml:"resourceName" comment:"optional, default={name}-{channel}, when channel is 'none', the default is {name}-{version}, the name for the ModuleTemplate that will be created"`
	Namespace     string            `yaml:"namespace" comment:"optional, default=kcp-system, the namespace where the ModuleTemplate will be deployed"`
	Security      string            `yaml:"security" comment:"optional, name of the security scanners config file"`
	Internal      bool              `yaml:"internal" comment:"optional, default=false, determines whether the ModuleTemplate should have the internal flag or not"`
	Beta          bool              `yaml:"beta" comment:"optional, default=false, determines whether the ModuleTemplate should have the beta flag or not"`
	Labels        map[string]string `yaml:"labels" comment:"optional, additional labels for the ModuleTemplate"`
	Annotations   map[string]string `yaml:"annotations" comment:"optional, additional annotations for the ModuleTemplate"`
}
