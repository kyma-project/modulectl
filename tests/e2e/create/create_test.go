//go:build e2e

package create_test

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"strings"

	"k8s.io/apimachinery/pkg/util/yaml"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	v2 "ocm.software/ocm/api/ocm/compdesc/versions/v2"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/github"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/localblob"
	"ocm.software/ocm/api/ocm/extensions/accessmethods/ociartifact"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/lifecycle-manager/api/v1beta2"

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

		_, err := exec.Command("k3d", "registry", "delete", "--all").CombinedOutput()
		Expect(err).ToNot(HaveOccurred())
		_, err = exec.Command("k3d", "registry", "create", "oci.localhost", "--port",
			"5001").CombinedOutput()
		Expect(err).ToNot(HaveOccurred())
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked without config-file arg", func() {
			cmd = createCmd{
				registry: ociRegistry,
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to read file module-config.yaml: open module-config.yaml: no such file or directory"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked without registry arg", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
			}
		})

		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("opts.RegistryURL must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing name", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: missingNameConfig,
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
				registry:         ociRegistry,
				moduleConfigFile: missingVersionConfig,
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
				registry:         ociRegistry,
				moduleConfigFile: missingManifestConfig,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate manifest: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing repository", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: missingRepositoryConfig,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing documentation", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: missingDocumentationConfig,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https repository", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: nonHttpsRepository,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: 'http://github.com/kyma-project/template-operator' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https documentation", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: nonHttpsDocumentation,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: 'http://github.com/kyma-project/template-operator/blob/main/README.md' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with missing icons", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: missingIconsConfig,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: must contain at least one icon: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with duplicate entry in icons", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: duplicateIcons,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Icons: map contains duplicate entries"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid icon - link missing", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: iconsWithoutLink,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: link must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid icon - name missing", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: iconsWithoutName,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: name must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with duplicate entry in resources", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: duplicateResources,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Resources: map contains duplicate entries"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with non https resource", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: nonHttpsResource,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: failed to validate link: 'http://some.other/location/template-operator.yaml' is not using https scheme: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid resource - link missing", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: resourceWithoutLink,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: link must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid resource - name missing", func() {
			cmd = createCmd{
				registry:         ociRegistry,
				moduleConfigFile: resourceWithoutName,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: name must not be empty: invalid Option"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with existing '--registry' and missing '--insecure' flag", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("could not push"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config and dry-run flag", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
				dryRun:           true,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)

				validateMinimalModuleTemplate(template, descriptor)

				By("And descriptor.component.repositoryContexts should be empty")
				Expect(descriptor.RepositoryContexts).To(HaveLen(0))

				By("And descriptor.component.resources should be empty")
				Expect(descriptor.Resources).To(HaveLen(0))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)

				validateMinimalModuleTemplate(template, descriptor)

				By("And descriptor.component.repositoryContexts should be correct")
				Expect(descriptor.RepositoryContexts).To(HaveLen(1))
				repo := descriptor.GetEffectiveRepositoryContext()
				Expect(repo.Object["baseUrl"]).To(Equal(ociRegistry))
				Expect(repo.Object["componentNameMapping"]).To(Equal(string(ocireg.OCIRegistryURLPathMapping)))
				Expect(repo.Object["type"]).To(Equal(ocireg.Type))

				By("And descriptor.component.resources should be correct")
				Expect(descriptor.Resources).To(HaveLen(2))
				resource := descriptor.Resources[0]
				Expect(resource.Name).To(Equal("metadata"))
				Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
				Expect(resource.Type).To(Equal("plainText"))
				Expect(resource.Version).To(Equal(moduleVersion))
				resource = descriptor.Resources[1]
				Expect(resource.Name).To(Equal("raw-manifest"))
				Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
				Expect(resource.Type).To(Equal("directoryTree"))
				Expect(resource.Version).To(Equal(moduleVersion))

				By("And descriptor.component.resources[0].access should be correct")
				resourceAccessSpec0, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[0].Access)
				Expect(err).ToNot(HaveOccurred())
				localBlobAccessSpec0, ok := resourceAccessSpec0.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(localBlobAccessSpec0.GetType()).To(Equal(localblob.Type))
				Expect(localBlobAccessSpec0.LocalReference).To(ContainSubstring("sha256:"))
				Expect(localBlobAccessSpec0.MediaType).To(Equal("application/x-yaml"))

				By("And descriptor.component.resources[1].access should be correct")
				resourceAccessSpec1, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[1].Access)
				Expect(err).ToNot(HaveOccurred())
				localBlobAccessSpec1, ok := resourceAccessSpec1.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(localBlobAccessSpec1.GetType()).To(Equal(localblob.Type))
				Expect(localBlobAccessSpec1.LocalReference).To(ContainSubstring("sha256:"))
				Expect(localBlobAccessSpec1.MediaType).To(Equal("application/x-tar"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
			Expect(cmd.execute()).To(Succeed())
		})
		By("Then invoked with same version that already exists in the registry", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
			}
		})
		By("Then the command should fail with same version exists message", func() {
			err := cmd.execute()
			Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("could not push component version: cannot push component version %s: component version already exists, cannot push the new version",
				moduleVersion)))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
			Expect(cmd.execute()).To(Succeed())
		})
		By("When invoked with same version that already exists in the registry and dry-run flag", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
				dryRun:           true,
			}
		})
		By("Then the command should fail with same version exists message", func() {
			err := cmd.execute()
			Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("component kyma-project.io/module/template-operator in version %s already exists: component version already exists",
				moduleVersion)))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with same version that already exists in the registry, and dry-run flag, and overwrite flag",
			func() {
				cmd = createCmd{
					moduleConfigFile: minimalConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
					overwrite:        true,
					dryRun:           true,
				}
			})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)

				validateMinimalModuleTemplate(template, descriptor)

				By("And descriptor.component.repositoryContexts should be empty")
				Expect(descriptor.RepositoryContexts).To(HaveLen(0))

				By("And descriptor.component.resources should be empty")
				Expect(descriptor.Resources).To(HaveLen(0))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with same version that already exists in the registry and overwrite flag", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				overwrite:        true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			err := cmd.execute()
			Expect(err).Should(Succeed())
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing annotations and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withAnnotationsConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And new annotation should be correctly added")
				annotations := template.Annotations
				Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))
				Expect(annotations["operator.kyma-project.io/doc-url"]).To(Equal("https://kyma-project.io"))

				By("And descriptor.component.resources should be correct")
				resource := descriptor.Resources[0]
				Expect(resource.Version).To(Equal(moduleVersion))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing default-cr and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withDefaultCrConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And descriptor.component.resources should be correct")
				Expect(descriptor.Resources).To(HaveLen(3))
				resource := descriptor.Resources[2]
				Expect(resource.Name).To(Equal("default-cr"))
				Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
				Expect(resource.Type).To(Equal("directoryTree"))
				Expect(resource.Version).To(Equal(moduleVersion))

				By("And descriptor.component.resources[2].access should be correct")
				defaultCRResourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[2].Access)
				Expect(err).ToNot(HaveOccurred())
				defaultCRAccessSpec, ok := defaultCRResourceAccessSpec.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(defaultCRAccessSpec.GetType()).To(Equal(localblob.Type))
				Expect(defaultCRAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
				Expect(defaultCRAccessSpec.MediaType).To(Equal("application/x-tar"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing security-scanner-config and different version",
			func() {
				cmd = createCmd{
					moduleConfigFile: withSecurityConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
				}
			})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And descriptor.component.resources should be correct")
				Expect(descriptor.Resources).To(HaveLen(4))
				resource := descriptor.Resources[0]
				Expect(resource.Name).To(Equal("template-operator"))
				Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))
				Expect(resource.Type).To(Equal("ociArtifact"))
				Expect(resource.Version).To(Equal(moduleVersion))

				resource = descriptor.Resources[1]
				Expect(resource.Name).To(Equal("template-operator"))
				Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))
				Expect(resource.Type).To(Equal("ociArtifact"))
				Expect(resource.Version).To(Equal("2.0.0"))

				resource = descriptor.Resources[2]
				Expect(resource.Name).To(Equal("metadata"))
				Expect(resource.Version).To(Equal(moduleVersion))

				resource = descriptor.Resources[3]
				Expect(resource.Name).To(Equal("raw-manifest"))
				Expect(resource.Version).To(Equal(moduleVersion))

				By("And descriptor.component.resources[0].access should be correct")
				resourceAccessSpec0, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[0].Access)
				Expect(err).ToNot(HaveOccurred())
				ociArtifactAccessSpec, ok := resourceAccessSpec0.(*ociartifact.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(ociArtifactAccessSpec.GetType()).To(Equal(ociartifact.Type))
				Expect(ociArtifactAccessSpec.ImageReference).To(Equal(fmt.Sprintf("europe-docker.pkg.dev/kyma-project/prod/template-operator:%s",
					moduleVersion)))

				By("And descriptor.component.resources[1].access should be correct")
				resourceAccessSpec1, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[1].Access)
				Expect(err).ToNot(HaveOccurred())
				ociArtifactAccessSpec, ok = resourceAccessSpec1.(*ociartifact.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(ociArtifactAccessSpec.GetType()).To(Equal(ociartifact.Type))
				Expect(ociArtifactAccessSpec.ImageReference).To(Equal("europe-docker.pkg.dev/kyma-project/prod/template-operator:2.0.0"))

				By("And descriptor.component.resources[2].access should be correct")
				resourceAccessSpec2, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[2].Access)
				Expect(err).ToNot(HaveOccurred())
				localBlobAccessSpec, ok := resourceAccessSpec2.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(localBlobAccessSpec.GetType()).To(Equal(localblob.Type))
				Expect(localBlobAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
				Expect(localBlobAccessSpec.MediaType).To(Equal("application/x-yaml"))

				By("And descriptor.component.resources[3].access should be correct")
				resourceAccessSpec3, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[3].Access)
				Expect(err).ToNot(HaveOccurred())
				localBlobAccessSpec2, ok := resourceAccessSpec3.(*localblob.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(localBlobAccessSpec2.GetType()).To(Equal(localblob.Type))
				Expect(localBlobAccessSpec2.LocalReference).To(ContainSubstring("sha256:"))
				Expect(localBlobAccessSpec2.MediaType).To(Equal("application/x-tar"))

				By("And descriptor.component.sources should be correct")
				Expect(len(descriptor.Sources)).To(Equal(1))
				source := descriptor.Sources[0]
				sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
				Expect(err).ToNot(HaveOccurred())
				githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
				Expect(ok).To(BeTrue())
				Expect(github.Type).To(Equal(githubAccessSpec.Type))
				Expect(githubAccessSpec.RepoURL).To(Equal("https://github.com/kyma-project/template-operator"))

				By("And module template should not marked as mandatory")
				Expect(template.Spec.Mandatory).To(BeFalse())
				val, ok := template.Labels[shared.IsMandatoryModule]
				Expect(val).To(BeEmpty())
				Expect(ok).To(BeFalse())

				By("And security scan labels should be correct")
				secScanLabels := flatten(descriptor.Sources[0].Labels)
				Expect(secScanLabels).To(HaveKeyWithValue("git.kyma-project.io/ref", "HEAD"))
				Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/rc-tag", moduleVersion))
				Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/language", "golang-mod"))
				Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/dev-branch", "main"))
				Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/subprojects", "false"))
				Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/exclude",
					"**/test/**,**/*_test.go"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with invalid module-config containing not existing security-scanner-config",
			func() {
				cmd = createCmd{
					moduleConfigFile: invalidSecurityConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
				}
			})
		By("Then the command should succeed", func() {
			err := cmd.execute()

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to configure security scanners: failed to parse security config data: security config file does not exist"))
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing mandatory true and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withMandatoryConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should be marked as mandatory")
				Expect(template.Spec.Mandatory).To(BeTrue())
				Expect(template.Labels[shared.IsMandatoryModule]).To(Equal("true"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing manager field and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withManagerConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And spec.manager should be correct")
				manager := template.Spec.Manager
				Expect(manager).ToNot(BeNil())
				Expect(manager.Name).To(Equal("template-operator-controller-manager"))
				Expect(manager.Namespace).To(Equal("template-operator-system"))
				Expect(manager.Version).To(Equal("v1"))
				Expect(manager.Group).To(Equal("apps"))
				Expect(manager.Kind).To(Equal("Deployment"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing manager field without namespace and different version",
			func() {
				cmd = createCmd{
					moduleConfigFile: withNoNamespaceManagerConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
				}
			})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And spec.manager should be correct")
				manager := template.Spec.Manager
				Expect(manager).ToNot(BeNil())
				Expect(manager.Name).To(Equal("template-operator-controller-manager"))
				Expect(manager.Namespace).To(BeEmpty())
				Expect(manager.Version).To(Equal("v1"))
				Expect(manager.Group).To(Equal("apps"))
				Expect(manager.Kind).To(Equal("Deployment"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing associatedResources list", func() {
			cmd = createCmd{
				moduleConfigFile: withAssociatedResourcesConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And spec.associatedResources should be correct")
				resources := template.Spec.AssociatedResources
				Expect(resources).ToNot(BeEmpty())
				Expect(len(resources)).To(Equal(1))
				Expect(resources[0].Group).To(Equal("networking.istio.io"))
				Expect(resources[0].Version).To(Equal("v1alpha3"))
				Expect(resources[0].Kind).To(Equal("Gateway"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config containing resources", func() {
			cmd = createCmd{
				moduleConfigFile: withResources,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain merged .spec.resources", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(template.Spec.Resources).To(HaveLen(2))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(template.Spec.Resources[0].Link).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
					moduleVersion)))
				Expect(template.Spec.Resources[1].Name).To(Equal("someResource"))
				Expect(template.Spec.Resources[1].Link).To(Equal("https://some.other/location/template-operator.yaml"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with minimal valid module-config containing rawManfiest in resources", func() {
			cmd = createCmd{
				moduleConfigFile: withResourcesOverwrite,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain rawManifest value from module-config", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())

				Expect(template.Spec.Resources).To(HaveLen(1))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(template.Spec.Resources[0].Link).To(Equal("https://some.other/location/template-operator.yaml"))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with manifest being a local file reference", func() {
			cmd = createCmd{
				moduleConfigFile: manifestFileref,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				validateTemplateWithFileReference(template, descriptor, moduleVersion)

				By("And template's spec.resources should NOT contain rawManifest")
				Expect(template.Spec.Resources).To(HaveLen(0))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with default CR being a fileref", func() {
			cmd = createCmd{
				moduleConfigFile: defaultCRFileref,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				validateTemplateWithFileReference(template, descriptor, moduleVersion)

				By("And template's spec.resources should contain rawManifest")
				Expect(template.Spec.Resources).To(HaveLen(1))
				Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
				Expect(template.Spec.Resources[0].Link).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
					moduleVersion)))
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing requiresDowntime true and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withRequiresDowntimeConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should have spec.requiresDowntime set to true")
				Expect(template.Spec.RequiresDowntime).To(BeTrue())
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing internal true and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withInternalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should have operator.kyma-project.io/internal label set to true")
				val, ok := template.Labels[shared.InternalLabel]
				Expect(val).To(Equal("true"))
				Expect(ok).To(BeTrue())
			})
		})
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing beta true and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withBetaConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("Then module template should contain the expected content", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				By("And module template should have operator.kyma-project.io/beta label set to true")
				val, ok := template.Labels[shared.BetaLabel]
				Expect(val).To(Equal("true"))
				Expect(ok).To(BeTrue())
			})
		})
	})

	// Test for successful image extraction and merging with security config
	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images from both manifest and security config", func() {
			cmd = createCmd{
				moduleConfigFile: withManifestMixedWithSecurity,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))

			By("And the module template should contain merged and deduplicated images", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResources(descriptor)
				Expect(len(imageResources)).To(Equal(6))

				// Verify specific images are present
				imageNames := extractImageNamesFromResources(imageResources)
				Expect(imageNames).To(ContainElement("template-operator"))
				Expect(imageNames).To(ContainElement("webhook"))
				Expect(imageNames).To(ContainElement("nginx"))
				Expect(imageNames).To(ContainElement("postgres"))
				Expect(imageNames).To(ContainElement("static"))

				// Verify deduplication - should not have duplicates
				imageURLs := extractImageURLsFromResources(imageResources)
				count := 0
				for _, url := range imageURLs {
					if url == "europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.3" {
						count++
					}
				}
				Expect(count).To(Equal(1), "template-operator:1.0.3 should appear only once")
			})
		})
	})

	// Test for env variable image extraction
	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images from env variables", func() {
			cmd = createCmd{
				moduleConfigFile: withManifestEnvVariables,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed and extract images from env variables", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images from env variables", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResources(descriptor)
				imageNames := extractImageNamesFromResources(imageResources)

				// Images from env variables should be included
				Expect(imageNames).To(ContainElement("webhook")) // from WEBHOOK_IMAGE
				Expect(imageNames).To(ContainElement("alpine"))  // from HELPER_IMAGE
				Expect(imageNames).To(ContainElement("migrate")) // from MIGRATION_IMAGE

				// Non-image env values should be excluded
				for _, resource := range imageResources {
					Expect(resource.Name).ToNot(Equal("some-config-value"))
				}
			})
		})
	})

	// Test for containers image extraction
	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images in containers", func() {
			cmd = createCmd{
				moduleConfigFile: withManifestContainers,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed and extract images from containers", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images from containers", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResources(descriptor)
				imageNames := extractImageNamesFromResources(imageResources)

				Expect(imageNames).To(ContainElement("template-operator"))
				Expect(imageNames).To(ContainElement("webhook"))
				Expect(imageNames).To(ContainElement("nginx"))
			})
		})
	})

	// Test for initContainers image extraction
	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images in initContainers", func() {
			cmd = createCmd{
				moduleConfigFile: withManifestInitContainers,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed and extract images from initContainers", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images from initContainers", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResources(descriptor)
				imageNames := extractImageNamesFromResources(imageResources)

				Expect(imageNames).To(ContainElement("busybox"))
				Expect(imageNames).To(ContainElement("migrate"))
				Expect(imageNames).To(ContainElement("alpine"))
			})
		})
	})

	// Test for SHA digest image handling
	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config containing images with SHA digest", func() {
			cmd = createCmd{
				moduleConfigFile: withManifestShaDigest,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should succeed and extract images with SHA digest", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And the module template should contain images with SHA digest", func() {
				template, err := readModuleTemplate(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
				descriptor := getDescriptor(template)
				Expect(descriptor).ToNot(BeNil())

				imageResources := getImageResources(descriptor)

				// Find the image with SHA digest
				foundShaImage := false
				for _, resource := range imageResources {
					if resource.Name == "template-operator" && strings.HasPrefix(resource.Version, "sha256:") {
						foundShaImage = true
						break
					}
				}
				Expect(foundShaImage).To(BeTrue(), "Should contain image with SHA digest")
			})
		})
	})

	// Test for latest/main tag validation failure
	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with manifest containing latest/main tags", func() {
			cmd = createCmd{
				moduleConfigFile: withManifestLatestMainTags,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		By("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("image tag is disallowed"))
		})
	})
})

// Test helper functions

func readModuleTemplate(filepath string) (*v1beta2.ModuleTemplate, error) {
	moduleTemplate := &v1beta2.ModuleTemplate{}
	moduleFile, err := os.ReadFile(filepath)
	if err != nil && len(moduleFile) > 0 {
		return nil, err
	}
	err = yaml.Unmarshal(moduleFile, moduleTemplate)
	if err != nil {
		return nil, err
	}
	return moduleTemplate, err
}

func getDescriptor(template *v1beta2.ModuleTemplate) *compdesc.ComponentDescriptor {
	ocmDesc, err := compdesc.Decode(
		template.Spec.Descriptor.Raw,
		[]compdesc.DecodeOption{compdesc.DisableValidation(true)}...)
	if err != nil {
		return nil
	}

	return ocmDesc
}

func flatten(labels ocmv1.Labels) map[string]string {
	labelsMap := make(map[string]string)
	for _, l := range labels {
		var value string
		_ = yaml.Unmarshal(l.Value, &value)
		labelsMap[l.Name] = value
	}
	return labelsMap
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

// Helper functions for image resource testing
func getImageResources(descriptor *compdesc.ComponentDescriptor) []compdesc.Resource {
	var imageResources []compdesc.Resource
	for _, resource := range descriptor.Resources {
		if resource.Type == "ociArtifact" {
			imageResources = append(imageResources, resource)
		}
	}
	return imageResources
}

func extractImageNamesFromResources(resources []compdesc.Resource) []string {
	var names []string
	for _, resource := range resources {
		if ociSpec, ok := resource.Access.(*ociartifact.AccessSpec); ok {
			ref := ociSpec.ImageReference
			parts := strings.Split(ref, "/")
			last := parts[len(parts)-1]
			name := strings.SplitN(last, ":", 2)[0]
			name = strings.SplitN(name, "@", 2)[0]
			names = append(names, name)
		} else {
			names = append(names, resource.Name)
		}
	}
	return names
}

func extractImageURLsFromResources(resources []compdesc.Resource) []string {
	var urls []string
	for _, resource := range resources {
		if ociSpec, ok := resource.Access.(*ociartifact.AccessSpec); ok {
			urls = append(urls, ociSpec.ImageReference)
		}
	}
	return urls
}

func validateMinimalModuleTemplate(template *v1beta2.ModuleTemplate, descriptor *compdesc.ComponentDescriptor) {
	Expect(descriptor).ToNot(BeNil())
	Expect(descriptor.SchemaVersion()).To(Equal(v2.SchemaVersion))
	Expect(template.Name).To(Equal(fmt.Sprintf("template-operator-%s", moduleVersion)))

	By("And spec.info should be correct")
	Expect(template.Spec.ModuleName).To(Equal("template-operator"))
	Expect(template.Spec.Version).To(Equal(moduleVersion))
	Expect(template.Spec.Info.Repository).To(Equal("https://github.com/kyma-project/template-operator"))
	Expect(template.Spec.Info.Documentation).To(Equal("https://github.com/kyma-project/template-operator/blob/main/README.md"))
	Expect(template.Spec.Info.Icons).To(HaveLen(1))
	Expect(template.Spec.Info.Icons[0].Name).To(Equal("module-icon"))
	Expect(template.Spec.Info.Icons[0].Link).To(Equal("https://github.com/kyma-project/template-operator/blob/main/docs/assets/logo.png"))

	By("And annotations should be correct")
	annotations := template.Annotations
	Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))

	By("And descriptor.component.sources should contain repository entry")
	Expect(len(descriptor.Sources)).To(Equal(1))
	source := descriptor.Sources[0]
	sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
	Expect(err).ToNot(HaveOccurred())
	githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
	Expect(ok).To(BeTrue())
	Expect(github.Type).To(Equal(githubAccessSpec.Type))
	Expect(githubAccessSpec.RepoURL).To(Equal("https://github.com/kyma-project/template-operator"))

	By("And module template should not marked as mandatory")
	Expect(template.Spec.Mandatory).To(BeFalse())
	val, ok := template.Labels[shared.IsMandatoryModule]
	Expect(val).To(BeEmpty())
	Expect(ok).To(BeFalse())

	By("And spec.associatedResources should be empty")
	Expect(template.Spec.AssociatedResources).To(BeEmpty())

	By("And spec.manager should be nil")
	Expect(template.Spec.Manager).To(BeNil())

	By("And spec.resources should contain rawManifest")
	Expect(template.Spec.Resources).To(HaveLen(1))
	Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
	Expect(template.Spec.Resources[0].Link).To(Equal(fmt.Sprintf("https://github.com/kyma-project/template-operator/releases/download/%s/template-operator.yaml",
		moduleVersion)))

	By("And spec.requiresDowntime should be set to false")
	Expect(template.Spec.RequiresDowntime).To(BeFalse())

	By("And module template should not have operator.kyma-project.io/internal label")
	val, ok = template.Labels[shared.InternalLabel]
	Expect(val).To(BeEmpty())
	Expect(ok).To(BeFalse())

	By("And module template should not have operator.kyma-project.io/beta label")
	val, ok = template.Labels[shared.BetaLabel]
	Expect(val).To(BeEmpty())
	Expect(ok).To(BeFalse())
}

func validateTemplateWithFileReference(template *v1beta2.ModuleTemplate, descriptor *compdesc.ComponentDescriptor,
	version string,
) {
	Expect(descriptor).ToNot(BeNil())
	Expect(descriptor.SchemaVersion()).To(Equal(v2.SchemaVersion))

	Expect(template).ToNot(BeNil())
	Expect(template.Name).To(Equal("template-operator-" + version))
	Expect(template.Spec.ModuleName).To(Equal("template-operator"))
	Expect(template.Spec.Version).To(Equal(version))

	By("And descriptor.component.resources should be correct")
	Expect(descriptor.Resources).To(HaveLen(3))

	By("And descriptor.component.resources for manifest should be correct")
	resource := descriptor.Resources[1]
	Expect(resource.Name).To(Equal("raw-manifest"))
	Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
	Expect(resource.Type).To(Equal("directoryTree"))
	Expect(resource.Version).To(Equal(version))

	By("And descriptor.component.resources.access for raw-manifest should be correct")
	manifestResourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
	Expect(err).ToNot(HaveOccurred())
	manifestAccessSpec, ok := manifestResourceAccessSpec.(*localblob.AccessSpec)
	Expect(ok).To(BeTrue())
	Expect(manifestAccessSpec.GetType()).To(Equal(localblob.Type))
	Expect(manifestAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
	Expect(manifestAccessSpec.MediaType).To(Equal("application/x-tar"))
	Expect(manifestAccessSpec.ReferenceName).To(Equal("raw-manifest"))

	By("And descriptor.component.resources for default CR should be correct")
	resource = descriptor.Resources[2]
	Expect(resource.Name).To(Equal("default-cr"))
	Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
	Expect(resource.Type).To(Equal("directoryTree"))
	Expect(resource.Version).To(Equal(version))

	By("And descriptor.component.resources.access for default-cr should be correct")
	defaultCRResourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(resource.Access)
	Expect(err).ToNot(HaveOccurred())
	defaultCRAccessSpec, ok := defaultCRResourceAccessSpec.(*localblob.AccessSpec)
	Expect(ok).To(BeTrue())
	Expect(defaultCRAccessSpec.GetType()).To(Equal(localblob.Type))
	Expect(defaultCRAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
	Expect(defaultCRAccessSpec.MediaType).To(Equal("application/x-tar"))
	Expect(defaultCRAccessSpec.ReferenceName).To(Equal("default-cr"))
}
