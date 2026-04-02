//go:build e2e

package create_test

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"
	"sigs.k8s.io/yaml"

	"github.com/kyma-project/modulectl/internal/common"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	moduleVersion = "1.0.3"
)

var _ = Describe("Test 'create' command", Ordered, func() {
	BeforeEach(func() {
		for _, file := range filesIn("/tmp/") {
			if file == "template.yaml" {
				err := os.Remove(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
			}
		}
	})

	// Validation error tests

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked without config-file arg", func() {
			cmd = createCmd{
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to read file module-config.yaml: open module-config.yaml: no such file or directory"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing name", func() {
			cmd = createCmd{
				moduleConfigFile:          missingNameConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("opts.ModuleName must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing version", func() {
			cmd = createCmd{
				moduleConfigFile:          missingVersionConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("opts.ModuleVersion must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing manifest", func() {
			cmd = createCmd{
				moduleConfigFile:          missingManifestConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate manifest: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing repository", func() {
			cmd = createCmd{
				moduleConfigFile:          missingRepositoryConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing documentation", func() {
			cmd = createCmd{
				moduleConfigFile:          missingDocumentationConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing team", func() {
			cmd = createCmd{
				moduleConfigFile:          missingTeamConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate team: must not be empty when security scan is enabled: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https repository", func() {
			cmd = createCmd{
				moduleConfigFile:          nonHttpsRepository,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: 'http://github.com/kyma-project/template-operator' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https documentation", func() {
			cmd = createCmd{
				moduleConfigFile:          nonHttpsDocumentation,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: 'http://github.com/kyma-project/template-operator/blob/main/README.md' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing icons", func() {
			cmd = createCmd{
				moduleConfigFile:          missingIconsConfig,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: must contain at least one icon: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with duplicate entry in icons", func() {
			cmd = createCmd{
				moduleConfigFile:          duplicateIcons,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Icons: map contains duplicate entries"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid icon - link missing", func() {
			cmd = createCmd{
				moduleConfigFile:          iconsWithoutLink,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: link must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid icon - name missing", func() {
			cmd = createCmd{
				moduleConfigFile:          iconsWithoutName,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: name must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with duplicate entry in resources", func() {
			cmd = createCmd{
				moduleConfigFile:          duplicateResources,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Resources: map contains duplicate entries"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https resource", func() {
			cmd = createCmd{
				moduleConfigFile:          nonHttpsResource,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: failed to validate link: 'http://some.other/location/template-operator.yaml' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid resource - link missing", func() {
			cmd = createCmd{
				moduleConfigFile:          resourceWithoutLink,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: link must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid resource - name missing", func() {
			cmd = createCmd{
				moduleConfigFile:          resourceWithoutName,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: name must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with a non git directory for module-sources-git-directory arg", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				moduleSourcesGitDirectory: "/tmp/not-a-git-dir",
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(
				err.Error(),
			).Should(ContainSubstring("currently configured module-sources-git-directory \"/tmp/not-a-git-dir\" must point to a valid git repository: invalid Option"))
		})
	})

	// Happy path tests using the component constructor path

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile:          minimalConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And component constructor file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("component-constructor.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				By("And template should have no descriptor")
				Expect(string(template.Spec.Descriptor.Raw)).To(MatchJSON(`{}`))

				By("And template should have basic info")
				Expect(template.Spec.ModuleName).To(Equal("template-operator"))
				Expect(template.Spec.Version).To(Equal(moduleVersion))
				Expect(template.Spec.Info.Repository).To(Equal("https://github.com/kyma-project/template-operator"))
				Expect(
					template.Spec.Info.Documentation,
				).To(Equal("https://github.com/kyma-project/template-operator/blob/main/README.md"))

				By("And template should have rawManifest resource")
				Expect(template.Spec.Resources).To(HaveLen(1))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(
					template.Spec.Resources[0].Link,
				).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
					moduleVersion)))

				By("And annotations should include cluster scope annotation")
				Expect(template.Annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))

				By("And spec.requiresDowntime should be set to false")
				Expect(template.Spec.RequiresDowntime).To(BeFalse())
			})

			By("And the component constructor file should contain expected content", func() {
				constructorData, err := os.ReadFile(constructorFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(constructorData)).To(ContainSubstring("components:"))
				Expect(string(constructorData)).To(ContainSubstring("name: kyma-project.io/module/template-operator"))
				Expect(string(constructorData)).To(ContainSubstring("version: " + moduleVersion))
				Expect(string(constructorData)).To(ContainSubstring("resources:"))
				Expect(string(constructorData)).To(ContainSubstring(common.RawManifestResourceName))
				Expect(string(constructorData)).To(ContainSubstring(common.ModuleTemplateResourceName))
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with valid module-config containing annotations", func() {
			cmd = createCmd{
				moduleConfigFile:          withAnnotationsConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected annotations", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				annotations := template.Annotations
				Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))
				Expect(annotations["operator.kyma-project.io/doc-url"]).To(Equal("https://kyma-project.io"))
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with valid module-config containing default-cr", func() {
			cmd = createCmd{
				moduleConfigFile:          withDefaultCrConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain default CR data", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.Spec.Data).ToNot(BeNil())
			})

			By("And component constructor should contain default-cr resource", func() {
				constructorData, err := os.ReadFile(constructorFilePath)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(constructorData)).To(ContainSubstring(common.DefaultCRResourceName))
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with valid module-config containing associated resources", func() {
			cmd = createCmd{
				moduleConfigFile:          withAssociatedResourcesConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain associated resources", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.Spec.AssociatedResources).ToNot(BeEmpty())
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with valid module-config with requiresDowntime set to true", func() {
			cmd = createCmd{
				moduleConfigFile:          withRequiresDowntimeConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should have requiresDowntime set to true", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.Spec.RequiresDowntime).To(BeTrue())
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with valid module-config with internal flag", func() {
			cmd = createCmd{
				moduleConfigFile:          withInternalConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should have internal label", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.Labels[shared.InternalLabel]).To(Equal("true"))
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		constructorFilePath := "/tmp/component-constructor.yaml"
		By("When invoked with valid module-config with beta flag", func() {
			cmd = createCmd{
				moduleConfigFile:          withBetaConfig,
				output:                    templateOutputPath,
				outputConstructorFile:     constructorFilePath,
				moduleSourcesGitDirectory: templateOperatorPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should have beta label", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				Expect(template.Labels[shared.BetaLabel]).To(Equal("true"))
			})

			By("And cleanup temporary files", func() {
				if _, err := os.Stat(constructorFilePath); err == nil {
					os.Remove(constructorFilePath)
				}
			})
		})
	})
})

func readModuleTemplate(filepath string) (*v1beta2.ModuleTemplate, error) {
	moduleTemplate := &v1beta2.ModuleTemplate{}
	moduleFile, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	err = yaml.Unmarshal(moduleFile, moduleTemplate)
	if err != nil {
		return nil, err
	}
	return moduleTemplate, nil
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
