package templategenerator

import (
	"fmt"
	"text/template"

	"bytes"
	"github.com/kyma-project/lifecycle-manager/api/shared"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"gopkg.in/yaml.v3"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/compdesc"
	"strings"
)

type FileSystem interface {
	WriteFile(path, content string) error
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) *Service {
	return &Service{
		fileSystem: fileSystem,
	}
}

const (
	modTemplate = `apiVersion: operator.kyma-project.io/v1beta2
kind: ModuleTemplate
metadata:
  name: {{.ResourceName}}
  namespace: {{.Namespace}}
{{- with .Labels}}
  labels:
    {{- range $key, $value := . }}
    {{ printf "%q" $key }}: {{ printf "%q" $value }}
    {{- end}}
{{- end}} 
{{- with .Annotations}}
  annotations:
    {{- range $key, $value := . }}
    {{ printf "%q" $key }}: {{ printf "%q" $value }}
    {{- end}}
{{- end}} 
spec:
  channel: {{.Channel}}
  mandatory: {{.Mandatory}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
`
)

type moduleTemplateData struct {
	ResourceName string
	Namespace    string
	Descriptor   compdesc.ComponentDescriptorVersion
	Channel      string
	Labels       map[string]string
	Annotations  map[string]string
	Mandatory    bool
}

func (s *Service) GenerateModuleTemplate(componentVersionAccess ocm.ComponentVersionAccess,
	moduleConfig *contentprovider.ModuleConfig, templateOutput string, isCrdClusterScoped bool) error {
	descriptor := componentVersionAccess.GetDescriptor()
	ref, err := oci.ParseRef(descriptor.Name)
	if err != nil {
		return fmt.Errorf("failed to parse ref: %w", err)
	}

	cva, err := compdesc.Convert(descriptor)
	if err != nil {
		return fmt.Errorf("failed to convert descriptor: %w", err)
	}

	labels := generateLabels(moduleConfig)
	annotations := generateAnnotations(moduleConfig, isCrdClusterScoped)

	shortName := shortName(ref)
	labels[shared.ModuleName] = shortName
	if moduleConfig.ResourceName == "" {
		moduleConfig.ResourceName = shortName + "-" + moduleConfig.Channel
	}

	moduleTemplate, err := template.New("moduleTemplate").Funcs(template.FuncMap{
		"yaml":   yaml.Marshal,
		"indent": indent,
	}).Parse(modTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse module template: %w", err)
	}

	mtData := moduleTemplateData{
		ResourceName: moduleConfig.ResourceName,
		Namespace:    moduleConfig.Namespace,
		Descriptor:   cva,
		Channel:      moduleConfig.Channel,
		Labels:       labels,
		Annotations:  annotations,
		Mandatory:    moduleConfig.Mandatory,
	}

	w := &bytes.Buffer{}
	if err := moduleTemplate.Execute(w, mtData); err != nil {
		return fmt.Errorf("failed to execute template, %w", err)
	}

	if err := s.fileSystem.WriteFile(w.String(), templateOutput); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func generateLabels(config *contentprovider.ModuleConfig) map[string]string {
	labels := config.Labels

	if labels == nil {
		labels = make(map[string]string)
	}
	if config.Beta {
		labels[shared.BetaLabel] = shared.EnableLabelValue
	}

	if config.Internal {
		labels[shared.InternalLabel] = shared.EnableLabelValue
	}

	return labels
}

func generateAnnotations(config *contentprovider.ModuleConfig, isCrdClusterScoped bool) map[string]string {
	annotations := config.Annotations
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations[shared.ModuleVersionAnnotation] = config.Version
	if isCrdClusterScoped {
		annotations[shared.IsClusterScopedAnnotation] = shared.EnableLabelValue
	} else {
		annotations[shared.IsClusterScopedAnnotation] = shared.DisableLabelValue
	}
	return config.Annotations
}

func indent(n int, in string) string {
	out := strings.Builder{}

	lines := strings.Split(in, "\n")

	// remove empty line at the end of the file if any
	if len(strings.TrimSpace(lines[len(lines)-1])) == 0 {
		lines = lines[:len(lines)-1]
	}

	for i, line := range lines {
		out.WriteString(strings.Repeat(" ", n))
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func shortName(ref oci.RefSpec) string {
	t := strings.Split(ref.Repository, "/")
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}
