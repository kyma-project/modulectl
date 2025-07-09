//go:build e2e

package create_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"ocm.software/ocm/api/ocm/extensions/repositories/ocireg"
)

var _ = Describe("Test 'create' command with private registry", Ordered, func() {
	BeforeEach(func() {
		for _, file := range filesIn("/tmp/") {
			if file == "template.yaml" {
				err := os.Remove(templateOutputPath)
				Expect(err).ToNot(HaveOccurred())
			}
		}
	})

	It("Given 'modulectl create' command", func() {
		var cmd createCmd
		By("When invoked with valid module-config with insecure private OCI registry and registry credentials",
			func() {
				cmd = createCmd{
					moduleConfigFile: minimalConfig,
					registry:         privateOciRegistry,
					insecure:         true,
					output:           templateOutputPath,
					registryCreds:    ociRegistryCreds,
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
				repo := descriptor.GetEffectiveRepositoryContext()
				Expect(repo.Object["baseUrl"]).To(Equal(privateOciRegistry))
				Expect(repo.Object["componentNameMapping"]).To(Equal(string(ocireg.OCIRegistryURLPathMapping)))
				Expect(repo.Object["type"]).To(Equal(ocireg.Type))
			})
		})
	})
})
