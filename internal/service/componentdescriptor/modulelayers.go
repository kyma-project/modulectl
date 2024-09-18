package componentdescriptor

import (
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"github.com/pkg/errors"
	"ocm.software/ocm/api/ocm/compdesc"
	ocmv1 "ocm.software/ocm/api/ocm/compdesc/meta/v1"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/utils/accessobj"
	"os"
)

// Layer encapsulates all necessary data to create an OCI layer
type Layer struct {
	Name          string
	ResourceType  string
	Path          string
	ExcludedFiles []string
}

type Definition struct {
	contentprovider.ModuleConfig
	Layers    []Layer
	Repo      string
	DefaultCR []byte
}

// ResourceDescriptor contains all information to describe a resource
type ResourceDescriptor struct {
	compdesc.Resource
	*Input `json:"input,omitempty"`
}

// ResourceDescriptorList contains a list of all information to describe a resource.
type ResourceDescriptorList struct {
	Resources []ResourceDescriptor `json:"resources"`
}

// AddResources adds the resources in the given resource definitions into the archive and its FS.
// A resource definition is a string with format: NAME:TYPE@PATH, where NAME and TYPE can be omitted and will default to the last path element name and "helm-chart" respectively
func AddResources(
	archive *comparch.ComponentArchive,
	modDef *Definition,
	fs vfs.FileSystem,
	registryCredSelector string,
) error {

	resources, err := generateResources(modDef.Version, registryCredSelector, modDef.Layers...)
	if err != nil {
		return err
	}

	for i, resource := range resources {
		if resource.Input != nil {
			if err := addBlob(fs, archive, &resources[i]); err != nil {
				return err
			}
		}
	}
	return nil
}

// generateResources generates resources by parsing the given definitions.
// Definitions have the following format: NAME:TYPE@PATH
// If a definition does not have a name or type, the name of the last path element is used,
// and it is assumed to be a helm-chart type.
func generateResources(version, registryCredSelector string, defs ...Layer) ([]ResourceDescriptor, error) {
	var res []ResourceDescriptor
	credMatchLabels, err := CreateCredMatchLabels(registryCredSelector)
	if err != nil {
		return nil, err
	}
	for _, d := range defs {
		r := ResourceDescriptor{Input: &Input{}}
		r.Name = d.Name
		r.Input.Path = d.Path
		r.Type = InputType(d.ResourceType)
		r.Version = version
		r.Relation = "local"
		fileInfo, err := os.Stat(r.Input.Path)
		if err != nil {
			return nil, errors.Wrap(err, "Could not determine if resource is a directory")
		}
		dir := fileInfo.IsDir()
		if dir {
			r.Input.Type = "dir"
			compress := true
			r.Input.CompressWithGzip = &compress
			r.Input.ExcludeFiles = d.ExcludedFiles
		} else {
			r.Input.Type = "file"
		}

		if len(credMatchLabels) > 0 {
			r.SetLabels([]ocmv1.Label{{
				Name:  OCIRegistryCredLabel,
				Value: credMatchLabels,
			}})
		}

		res = append(res, r)
	}
	return res, nil
}

func addBlob(fs vfs.FileSystem, archive *comparch.ComponentArchive, resource *ResourceDescriptor) error {
	access, err := AccessForFileOrFolder(fs, resource.Input)
	if err != nil {
		return err
	}

	blobAccess, err := archive.AddBlob(
		accessobj.CachedBlobAccessForDataAccess(archive.GetContext(), access.MimeType(), access), string(resource.Input.Type),
		resource.Resource.Name, nil,
	)
	if err != nil {
		return err
	}

	return archive.SetResource(&resource.ResourceMeta, blobAccess)
}
