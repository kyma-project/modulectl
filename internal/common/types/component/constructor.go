package component

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/kyma-project/modulectl/internal/common"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/git"
	"github.com/kyma-project/modulectl/internal/service/image"
)

const (
	GithubSourceType = "Github"
	GithubAccessType = "gitHub"

	OCIArtifactResourceType     = "ociArtifact"
	OCIArtifactResourceRelation = "external"
	OCIArtifactAccessType       = "ociArtifact"

	DirectoryTreeResourceType = "directoryTree"
	DirectoryInputType        = "dir"

	PlainTextResourceType = "PlainText"
	BinaryResourceInput   = "binary"
)

type Provider struct {
	Name   string  `yaml:"name"`
	Labels []Label `yaml:"labels,omitempty"`
}

type Input struct {
	Type         string   `yaml:"type"`
	Path         string   `yaml:"path,omitempty"`
	Data         string   `yaml:"data,omitempty"`
	Compress     bool     `yaml:"compress,omitempty"`
	IncludeFiles []string `yaml:"includeFiles,omitempty"`
}

type Access struct {
	Type           string `yaml:"type"`
	ImageReference string `yaml:"imageReference,omitempty"`
	RepoUrl        string `yaml:"repoUrl,omitempty"`
	Commit         string `yaml:"commit,omitempty"`
}

type Label struct {
	Name    string `yaml:"name"`
	Value   string `yaml:"value"`
	Version string `yaml:"version,omitempty"`
}

type Resource struct {
	Name     string  `yaml:"name"`
	Type     string  `yaml:"type"`
	Version  string  `yaml:"version,omitempty"`
	Relation string  `yaml:"relation,omitempty"`
	Labels   []Label `yaml:"labels,omitempty"`
	Input    *Input  `yaml:"input,omitempty"`
	Access   *Access `yaml:"access,omitempty"`
}

type Source struct {
	Name    string  `yaml:"name"`
	Type    string  `yaml:"type"`
	Version string  `yaml:"version,omitempty"`
	Labels  []Label `yaml:"labels,omitempty"`
	Input   *Input  `yaml:"input,omitempty"`
	Access  *Access `yaml:"access,omitempty"`
}

type Component struct {
	Name      string     `yaml:"name"`
	Version   string     `yaml:"version"`
	Provider  Provider   `yaml:"provider"`
	Labels    []Label    `yaml:"labels,omitempty"`
	Resources []Resource `yaml:"resources"`
	Sources   []Source   `yaml:"sources,omitempty"`
}

type Constructor struct {
	Components []Component `yaml:"components"`
}

func NewConstructor(componentName, componentVersion string) *Constructor {
	return &Constructor{
		Components: []Component{
			{
				Name:    componentName,
				Version: componentVersion,
				Provider: Provider{
					Name: common.ProviderName,
					Labels: []Label{
						{common.BuiltByLabelKey, common.BuiltByLabelValue, common.VersionV1},
					},
				},
				Resources: make([]Resource, 0),
				Sources:   make([]Source, 0),
			},
		},
	}
}

func (c *Constructor) AddGitSource(gitRepoURL, commitHash string) {
	source := Source{
		Name:    common.OCMIdentityName,
		Type:    GithubSourceType,
		Version: c.Components[0].Version,
		Labels: []Label{
			{
				Name:    common.RefLabel,
				Value:   git.HeadRef,
				Version: common.OCMVersion,
			},
		},
		Access: &Access{
			Type:    GithubAccessType,
			RepoUrl: gitRepoURL,
			Commit:  commitHash,
		},
	}

	c.Components[0].Sources = append(c.Components[0].Sources, source)
}

func (c *Constructor) AddLabel(key, value, version string) {
	labels := c.Components[0].Labels
	labelValue := Label{
		Name:    key,
		Value:   value,
		Version: version,
	}
	labels = append(labels, labelValue)
	c.Components[0].Labels = labels
}

func (c *Constructor) AddLabelToSources(key, value, version string) {
	for index, source := range c.Components[0].Sources {
		labels := source.Labels
		labelValue := Label{
			Name:    key,
			Value:   value,
			Version: version,
		}
		labels = append(labels, labelValue)
		c.Components[0].Sources[index].Labels = labels
	}
}

func (c *Constructor) AddImageAsResource(imageInfos []*image.ImageInfo) {
	for _, imageInfo := range imageInfos {
		version, resourceName := resources.GenerateOCMVersionAndName(imageInfo)
		resource := Resource{
			Name:     resourceName,
			Type:     OCIArtifactResourceType,
			Relation: OCIArtifactResourceRelation,
			Version:  version,
			Labels: []Label{
				{
					Name:    fmt.Sprintf("%s/%s", common.SecScanBaseLabelKey, common.TypeLabelKey),
					Value:   common.ThirdPartyImageLabelValue,
					Version: common.OCMVersion,
				},
			},
			Access: &Access{
				Type:           OCIArtifactAccessType,
				ImageReference: imageInfo.FullURL,
			},
		}
		c.Components[0].Resources = append(c.Components[0].Resources, resource)
	}
}

func (c *Constructor) AddRawManifestResource(manifestPath string) {
	c.Components[0].Resources = append(c.Components[0].Resources, Resource{
		Name:    common.RawManifestResourceName,
		Type:    DirectoryTreeResourceType,
		Version: c.Components[0].Version,
		Input: &Input{
			Type:         DirectoryInputType,
			Path:         filepath.Dir(manifestPath),
			Compress:     true,
			IncludeFiles: []string{filepath.Base(manifestPath)},
		},
	})
}

func (c *Constructor) AddDefaultCRResource(defaultCRPath string) {
	c.Components[0].Resources = append(c.Components[0].Resources, Resource{
		Name:    common.DefaultCRResourceName,
		Type:    DirectoryTreeResourceType,
		Version: c.Components[0].Version,
		Input: &Input{
			Type:         DirectoryInputType,
			Path:         filepath.Dir(defaultCRPath),
			Compress:     true,
			IncludeFiles: []string{filepath.Base(defaultCRPath)},
		},
	})
}

func (c *Constructor) AddMetadataResource(moduleConfig *contentprovider.ModuleConfig) {
	yamlData, err := resources.GenerateMetadataYaml(moduleConfig)
	if err != nil {
		return
	}
	c.Components[0].Resources = append(c.Components[0].Resources, Resource{
		Name:    common.MetadataResourceName,
		Type:    PlainTextResourceType,
		Version: c.Components[0].Version,
		Input: &Input{
			Type: BinaryResourceInput,
			Data: base64.StdEncoding.EncodeToString(yamlData),
		},
	})
}

func (c *Constructor) CreateComponentConstructorFile(filePath string) error {
	marshal, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("unable to marshal component constructor: %w", err)
	}

	filePermission := 0o600
	if err = os.WriteFile(filePath, marshal, os.FileMode(filePermission)); err != nil {
		return fmt.Errorf("unable to write component constructor file: %w", err)
	}
	return nil
}
