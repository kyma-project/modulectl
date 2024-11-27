//go:build e2e

package create_test

import (
	"io/fs"
	"os"

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

var _ = Describe("Test 'create' command", Ordered, func() {
	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked without any args", func() {
			cmd = createCmd{}
		})

		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to read file module-config.yaml: open module-config.yaml: no such file or directory"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with missing name", func() {
			cmd = createCmd{
				moduleConfigFile: missingNameConfig,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("invalid Option: opts.ModuleName must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with missing version", func() {
			cmd = createCmd{
				moduleConfigFile: missingVersionConfig,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("invalid Option: opts.ModuleVersion must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with missing manifest", func() {
			cmd = createCmd{
				moduleConfigFile: missingManifestConfig,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate manifest: invalid Option: must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with missing repository", func() {
			cmd = createCmd{
				moduleConfigFile: missingRepositoryConfig,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: invalid Option: must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with missing documentation", func() {
			cmd = createCmd{
				moduleConfigFile: missingDocumentationConfig,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: invalid Option: must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with non https repository", func() {
			cmd = createCmd{
				moduleConfigFile: nonHttpsRepository,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate repository: invalid Option: 'http://github.com/kyma-project/template-operator' is not using https scheme"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with non https documentation", func() {
			cmd = createCmd{
				moduleConfigFile: nonHttpsDocumentation,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate documentation: invalid Option: 'http://github.com/kyma-project/template-operator/blob/main/README.md' is not using https scheme"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with missing icons", func() {
			cmd = createCmd{
				moduleConfigFile: missingIconsConfig,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: invalid Option: must contain at least one icon"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with duplicate entry in icons", func() {
			cmd = createCmd{
				moduleConfigFile: duplicateIcons,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Icons: map contains duplicate entries"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with invalid icon - link missing", func() {
			cmd = createCmd{
				moduleConfigFile: iconsWithoutLink,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: invalid Option: link must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with invalid icon - name missing", func() {
			cmd = createCmd{
				moduleConfigFile: iconsWithoutName,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate module icons: invalid Option: name must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with duplicate entry in resources", func() {
			cmd = createCmd{
				moduleConfigFile: duplicateResources,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config file: failed to unmarshal Resources: map contains duplicate entries"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with non https resource", func() {
			cmd = createCmd{
				moduleConfigFile: nonHttpsResource,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: failed to validate link: invalid Option: 'http://some.other/location/template-operator.yaml' is not using https scheme"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with invalid resource - link missing", func() {
			cmd = createCmd{
				moduleConfigFile: resourceWithoutLink,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: invalid Option: link must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with invalid resource - name missing", func() {
			cmd = createCmd{
				moduleConfigFile: resourceWithoutName,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate resources: invalid Option: name must not be empty"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with '--config-file' using valid file", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And no module template file should be generated")
			currentDir, err := os.Getwd()
			Expect(err).ToNot(HaveOccurred())
			Expect(filesIn(currentDir)).Should(Not(ContainElement("template.yaml")))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with existing '--registry' and missing '--insecure' flag", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("Error: could not push"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with minimal valid module-config", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(descriptor.SchemaVersion()).To(Equal(v2.SchemaVersion))
			Expect(template.Name).To(Equal("template-operator-1.0.0"))

			By("And spec.info should be correct")
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.0"))
			Expect(template.Spec.Info.Repository).To(Equal("https://github.com/kyma-project/template-operator"))
			Expect(template.Spec.Info.Documentation).To(Equal("https://github.com/kyma-project/template-operator/blob/main/README.md"))
			Expect(template.Spec.Info.Icons).To(HaveLen(1))
			Expect(template.Spec.Info.Icons[0].Name).To(Equal("module-icon"))
			Expect(template.Spec.Info.Icons[0].Link).To(Equal("https://github.com/kyma-project/template-operator/blob/main/docs/assets/logo.png"))

			By("And annotations should be correct")
			annotations := template.Annotations
			Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))

			By("And descriptor.component.repositoryContexts should be correct")
			Expect(descriptor.RepositoryContexts).To(HaveLen(1))
			repo := descriptor.GetEffectiveRepositoryContext()
			Expect(repo.Object["baseUrl"]).To(Equal(ociRegistry))
			Expect(repo.Object["componentNameMapping"]).To(Equal(string(ocireg.OCIRegistryURLPathMapping)))
			Expect(repo.Object["type"]).To(Equal(ocireg.Type))

			By("And descriptor.component.resources should be correct")
			Expect(descriptor.Resources).To(HaveLen(1))
			resource := descriptor.Resources[0]
			Expect(resource.Name).To(Equal("raw-manifest"))
			Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
			Expect(resource.Type).To(Equal("directory"))
			Expect(resource.Version).To(Equal("1.0.0"))

			By("And descriptor.component.resources[0].access should be correct")
			resourceAccessSpec1, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[0].Access)
			Expect(err).ToNot(HaveOccurred())
			localBlobAccessSpec, ok := resourceAccessSpec1.(*localblob.AccessSpec)
			Expect(ok).To(BeTrue())
			Expect(localBlobAccessSpec.GetType()).To(Equal(localblob.Type))
			Expect(localBlobAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
			Expect(localBlobAccessSpec.MediaType).To(Equal("application/x-tar"))

			By("And descriptor.component.sources should contain repository entry")
			Expect(len(descriptor.Sources)).To(Equal(1))
			source := descriptor.Sources[0]
			sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
			Expect(err).ToNot(HaveOccurred())
			githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
			Expect(ok).To(BeTrue())
			Expect(github.Type).To(Equal(githubAccessSpec.Type))
			Expect(githubAccessSpec.RepoURL).To(Equal("https://github.com/kyma-project/template-operator"))

			By("And spec.mandatory should be false")
			Expect(template.Spec.Mandatory).To(BeFalse())

			By("And spec.associatedResources should be empty")
			Expect(template.Spec.AssociatedResources).To(BeEmpty())

			By("And spec.manager should be nil")
			Expect(template.Spec.Manager).To(BeNil())

			By("And spec.resources should contain rawManifest")
			Expect(template.Spec.Resources).To(HaveLen(1))
			Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
			Expect(template.Spec.Resources[0].Link).To(Equal("https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with same version that already exists in the registry", func() {
			cmd = createCmd{
				moduleConfigFile: minimalConfig,
				registry:         ociRegistry,
				insecure:         true,
			}
		})
		It("Then the command should fail with same version exists message", func() {
			err := cmd.execute()
			Expect(err.Error()).Should(ContainSubstring("could not push component version: component version already exists"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing annotations and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withAnnotationsConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(template.Name).To(Equal("template-operator-1.0.1"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.1"))

			By("And new annotation should be correctly added")
			annotations := template.Annotations
			Expect(annotations[shared.IsClusterScopedAnnotation]).To(Equal("false"))
			Expect(annotations["operator.kyma-project.io/doc-url"]).To(Equal("https://kyma-project.io"))

			By("And descriptor.component.resources should be correct")
			resource := descriptor.Resources[0]
			Expect(resource.Version).To(Equal("1.0.1"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing default-cr and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withDefaultCrConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(template.Name).To(Equal("template-operator-1.0.2"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.2"))

			By("And descriptor.component.resources should be correct")
			Expect(descriptor.Resources).To(HaveLen(2))
			resource := descriptor.Resources[1]
			Expect(resource.Name).To(Equal("default-cr"))
			Expect(resource.Relation).To(Equal(ocmv1.LocalRelation))
			Expect(resource.Type).To(Equal("directory"))
			Expect(resource.Version).To(Equal("1.0.2"))

			By("And descriptor.component.resources[1].access should be correct")
			defaultCRResourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[1].Access)
			Expect(err).ToNot(HaveOccurred())
			defaultCRAccessSpec, ok := defaultCRResourceAccessSpec.(*localblob.AccessSpec)
			Expect(ok).To(BeTrue())
			Expect(defaultCRAccessSpec.GetType()).To(Equal(localblob.Type))
			Expect(defaultCRAccessSpec.LocalReference).To(ContainSubstring("sha256:"))
			Expect(defaultCRAccessSpec.MediaType).To(Equal("application/x-tar"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing security-scanner-config and different version",
			func() {
				cmd = createCmd{
					moduleConfigFile: withSecurityConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
				}
			})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(template.Name).To(Equal("template-operator-1.0.3"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.3"))

			By("And descriptor.component.resources should be correct")
			Expect(descriptor.Resources).To(HaveLen(3))
			resource := descriptor.Resources[0]
			Expect(resource.Name).To(Equal("template-operator"))
			Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))
			Expect(resource.Type).To(Equal("ociArtifact"))
			Expect(resource.Version).To(Equal("1.0.1"))

			resource = descriptor.Resources[1]
			Expect(resource.Name).To(Equal("template-operator"))
			Expect(resource.Relation).To(Equal(ocmv1.ExternalRelation))
			Expect(resource.Type).To(Equal("ociArtifact"))
			Expect(resource.Version).To(Equal("2.0.0"))

			resource = descriptor.Resources[2]
			Expect(resource.Name).To(Equal("raw-manifest"))
			Expect(resource.Version).To(Equal("1.0.3"))

			By("And descriptor.component.resources[0].access should be correct")
			resourceAccessSpec0, err := ocm.DefaultContext().AccessSpecForSpec(descriptor.Resources[0].Access)
			Expect(err).ToNot(HaveOccurred())
			ociArtifactAccessSpec, ok := resourceAccessSpec0.(*ociartifact.AccessSpec)
			Expect(ok).To(BeTrue())
			Expect(ociArtifactAccessSpec.GetType()).To(Equal(ociartifact.Type))
			Expect(ociArtifactAccessSpec.ImageReference).To(Equal("europe-docker.pkg.dev/kyma-project/prod/template-operator:1.0.1"))

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
			Expect(localBlobAccessSpec.MediaType).To(Equal("application/x-tar"))

			By("And descriptor.component.sources should be correct")
			Expect(len(descriptor.Sources)).To(Equal(1))
			source := descriptor.Sources[0]
			sourceAccessSpec, err := ocm.DefaultContext().AccessSpecForSpec(source.Access)
			Expect(err).ToNot(HaveOccurred())
			githubAccessSpec, ok := sourceAccessSpec.(*github.AccessSpec)
			Expect(ok).To(BeTrue())
			Expect(github.Type).To(Equal(githubAccessSpec.Type))
			Expect(githubAccessSpec.RepoURL).To(Equal("https://github.com/kyma-project/template-operator"))

			By("And spec.mandatory should be false")
			Expect(template.Spec.Mandatory).To(BeFalse())

			By("And security scan labels should be correct")
			secScanLabels := flatten(descriptor.Sources[0].Labels)
			Expect(secScanLabels).To(HaveKeyWithValue("git.kyma-project.io/ref", "HEAD"))
			Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/rc-tag", "1.0.1"))
			Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/language", "golang-mod"))
			Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/dev-branch", "main"))
			Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/subprojects", "false"))
			Expect(secScanLabels).To(HaveKeyWithValue("scan.security.kyma-project.io/exclude",
				"**/test/**,**/*_test.go"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with invalid module-config containing not existing security-scanner-config",
			func() {
				cmd = createCmd{
					moduleConfigFile: invalidSecurityConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
				}
			})
		It("Then the command should succeed", func() {
			err := cmd.execute()

			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to configure security scanners: failed to parse security config data: security config file does not exist"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing mandatory true and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withMandatoryConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(template.Name).To(Equal("template-operator-1.0.4"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.4"))

			By("And spec.mandatory should be true")
			Expect(template.Spec.Mandatory).To(BeTrue())
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing manager field and different version", func() {
			cmd = createCmd{
				moduleConfigFile: withManagerConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(template.Name).To(Equal("template-operator-1.0.5"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.5"))

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

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing manager field without namespace and different version",
			func() {
				cmd = createCmd{
					moduleConfigFile: withNoNamespaceManagerConfig,
					registry:         ociRegistry,
					insecure:         true,
					output:           templateOutputPath,
				}
			})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())
			Expect(template.Name).To(Equal("template-operator-1.0.6"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.6"))

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

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with valid module-config containing associatedResources list", func() {
			cmd = createCmd{
				moduleConfigFile: withAssociatedResourcesConfig,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain the expected content", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())
			descriptor := getDescriptor(template)
			Expect(descriptor).ToNot(BeNil())

			Expect(template.Name).To(Equal("template-operator-1.0.7"))
			Expect(template.Spec.ModuleName).To(Equal("template-operator"))
			Expect(template.Spec.Version).To(Equal("1.0.7"))

			By("And spec.associatedResources should be correct")
			resources := template.Spec.AssociatedResources
			Expect(resources).ToNot(BeEmpty())
			Expect(len(resources)).To(Equal(1))
			Expect(resources[0].Group).To(Equal("networking.istio.io"))
			Expect(resources[0].Version).To(Equal("v1alpha3"))
			Expect(resources[0].Kind).To(Equal("Gateway"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with minimal valid module-config containing resources", func() {
			cmd = createCmd{
				moduleConfigFile: withResources,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain merged .spec.resources", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(template.Spec.Resources).To(HaveLen(2))
			Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
			Expect(template.Spec.Resources[0].Link).To(Equal("https://github.com/kyma-project/template-operator/releases/download/1.0.1/template-operator.yaml"))
			Expect(template.Spec.Resources[1].Name).To(Equal("someResource"))
			Expect(template.Spec.Resources[1].Link).To(Equal("https://some.other/location/template-operator.yaml"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with minimal valid module-config containing rawManfiest in resources", func() {
			cmd = createCmd{
				moduleConfigFile: withResourcesOverwrite,
				registry:         ociRegistry,
				insecure:         true,
				output:           templateOutputPath,
			}
		})
		It("Then the command should succeed", func() {
			Expect(cmd.execute()).To(Succeed())

			By("And module template file should be generated")
			Expect(filesIn("/tmp/")).Should(ContainElement("template.yaml"))
		})
		It("Then module template should contain rawManifest value from module-config", func() {
			template, err := readModuleTemplate(templateOutputPath)
			Expect(err).ToNot(HaveOccurred())

			Expect(template.Spec.Resources).To(HaveLen(1))
			Expect(template.Spec.Resources[0].Name).To(Equal("rawManifest"))
			Expect(template.Spec.Resources[0].Link).To(Equal("https://some.other/location/template-operator.yaml"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with manifest being a fileref", func() {
			cmd = createCmd{
				moduleConfigFile: manifestFileref,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate manifest: invalid Option: './template-operator.yaml' is not using https scheme"))
		})
	})

	Context("Given 'modulectl create' command", func() {
		var cmd createCmd
		It("When invoked with default CR being a fileref", func() {
			cmd = createCmd{
				moduleConfigFile: defaultCRFileref,
			}
		})
		It("Then the command should fail", func() {
			err := cmd.execute()
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("failed to parse module config: failed to validate module config: failed to validate default CR: invalid Option: '/tmp/default-sample-cr.yaml' is not using https scheme"))
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

func getDescriptor(template *v1beta2.ModuleTemplate) *v1beta2.Descriptor {
	if template.Spec.Descriptor.Object != nil {
		desc, ok := template.Spec.Descriptor.Object.(*v1beta2.Descriptor)
		if !ok || desc == nil {
			return nil
		}
		return desc
	}
	ocmDesc, err := compdesc.Decode(
		template.Spec.Descriptor.Raw,
		[]compdesc.DecodeOption{compdesc.DisableValidation(true)}...)
	if err != nil {
		return nil
	}
	template.Spec.Descriptor.Object = &v1beta2.Descriptor{ComponentDescriptor: ocmDesc}
	desc, ok := template.Spec.Descriptor.Object.(*v1beta2.Descriptor)
	if !ok {
		return nil
	}

	return desc
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
