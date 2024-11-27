package templategenerator

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/kyma-project/lifecycle-manager/api/shared"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"ocm.software/ocm/api/oci"
	"ocm.software/ocm/api/ocm/compdesc"
	"sigs.k8s.io/yaml"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

var (
	ErrEmptyModuleConfig = errors.New("can not generate module template from empty module config")
	ErrEmptyDescriptor   = errors.New("can not generate module template from empty descriptor")
)

type FileSystem interface {
	WriteFile(path, content string) error
}

type Service struct {
	fileSystem FileSystem
}

func NewService(fileSystem FileSystem) (*Service, error) {
	if fileSystem == nil {
		return nil, fmt.Errorf("%w: fileSystem must not be nil", commonerrors.ErrInvalidArg)
	}

	return &Service{
		fileSystem: fileSystem,
	}, nil
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
  moduleName: {{.ModuleName}}
  version: {{.ModuleVersion}}
  mandatory: {{.Mandatory}}
  info:
    repository: {{.Repository}}
    documentation: {{.Documentation}}
    {{- with .Icons}}
    icons:
      {{- range $key, $value := . }}
    - name: {{ $key }}
      link: {{ $value }}
      {{- end}}
    {{- end}}
{{- with .AssociatedResources}}
  associatedResources:
  {{- range .}}
  - group: {{.Group}}
    version: {{.Version}}
    kind: {{.Kind}}
  {{- end}}
{{- end}}
{{- with .Data}}
  data:
{{. | indent 4}}
{{- end}}
{{- with .Manager}}
  manager:
    name: {{.Name}}
    {{- if .Namespace}}      
    namespace: {{.Namespace}}
    {{- end}}
    group: {{.GroupVersionKind.Group}}
    version: {{.GroupVersionKind.Version}}
    kind: {{.GroupVersionKind.Kind}}
{{- end}}
  descriptor:
{{yaml .Descriptor | printf "%s" | indent 4}}
{{- with .Resources}}
  resources:
    {{- range $key, $value := . }}
  - name: {{ $key }}
    link: {{ $value }}
    {{- end}}
{{- end}}
`
)

type moduleTemplateData struct {
	ModuleName          string
	ResourceName        string
	Namespace           string
	ModuleVersion       string
	Descriptor          compdesc.ComponentDescriptorVersion
	Repository          string
	Documentation       string
	Icons               contentprovider.Icons
	Labels              map[string]string
	Annotations         map[string]string
	Mandatory           bool
	Data                string
	AssociatedResources []*metav1.GroupVersionKind
	Resources           contentprovider.Resources
	Manager             *contentprovider.Manager
}

func (s *Service) GenerateModuleTemplate(
	moduleConfig *contentprovider.ModuleConfig,
	descriptor *compdesc.ComponentDescriptor,
	data []byte,
	isCrdClusterScoped bool,
	templateOutput string,
) error {
	if moduleConfig == nil {
		return ErrEmptyModuleConfig
	}
	if descriptor == nil {
		return ErrEmptyDescriptor
	}

	labels := generateLabels(moduleConfig)
	annotations := generateAnnotations(moduleConfig, isCrdClusterScoped)

	ref, err := oci.ParseRef(descriptor.Name)
	if err != nil {
		return fmt.Errorf("failed to parse ref: %w", err)
	}
	shortName := trimShortNameFromRef(ref)
	labels[shared.ModuleName] = shortName
	moduleTemplateName := shortName + "-" + moduleConfig.Version

	moduleTemplate, err := template.New("moduleTemplate").Funcs(template.FuncMap{
		"yaml":   yaml.Marshal,
		"indent": indent,
	}).Parse(modTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse module template: %w", err)
	}

	cva, err := compdesc.Convert(descriptor)
	if err != nil {
		return fmt.Errorf("failed to convert descriptor: %w", err)
	}

	mtData := moduleTemplateData{
		ModuleName:          shortName,
		ResourceName:        moduleTemplateName,
		Namespace:           moduleConfig.Namespace,
		ModuleVersion:       moduleConfig.Version,
		Descriptor:          cva,
		Repository:          moduleConfig.Repository,
		Documentation:       moduleConfig.Documentation,
		Icons:               moduleConfig.Icons,
		Labels:              labels,
		Annotations:         annotations,
		Mandatory:           moduleConfig.Mandatory,
		AssociatedResources: moduleConfig.AssociatedResources,
		Resources: contentprovider.Resources{
			"rawManifest": moduleConfig.Manifest, // defaults rawManifest to Manifest; may be overwritten by explicitly provided entries
		},
		Manager: moduleConfig.Manager,
	}

	if len(data) > 0 {
		mtData.Data = string(data)
	}

	for name, link := range moduleConfig.Resources {
		mtData.Resources[name] = link
	}

	w := &bytes.Buffer{}
	if err = moduleTemplate.Execute(w, mtData); err != nil {
		return fmt.Errorf("failed to execute template, %w", err)
	}

	if err = s.fileSystem.WriteFile(templateOutput, w.String()); err != nil {
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
	if isCrdClusterScoped {
		annotations[shared.IsClusterScopedAnnotation] = shared.EnableLabelValue
	} else {
		annotations[shared.IsClusterScopedAnnotation] = shared.DisableLabelValue
	}
	return annotations
}

func indent(spaces int, input string) string {
	out := strings.Builder{}

	lines := strings.Split(input, "\n")

	// remove empty line at the end of the file if any
	if len(strings.TrimSpace(lines[len(lines)-1])) == 0 {
		lines = lines[:len(lines)-1]
	}

	for i, line := range lines {
		out.WriteString(strings.Repeat(" ", spaces))
		out.WriteString(line)
		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}
	return out.String()
}

func trimShortNameFromRef(ref oci.RefSpec) string {
	t := strings.Split(ref.Repository, "/")
	if len(t) == 0 {
		return ""
	}
	return t[len(t)-1]
}
