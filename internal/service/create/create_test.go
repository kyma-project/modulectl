package create_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"ocm.software/ocm/api/ocm/compdesc"
	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/common/types/component"
	"github.com/kyma-project/modulectl/internal/service/componentarchive"
	"github.com/kyma-project/modulectl/internal/service/componentdescriptor/resources"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/create"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

func Test_NewService_ReturnsError_WhenModuleConfigServiceIsNil(t *testing.T) {
	_, err := create.NewService(nil, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{}, &componentArchiveServiceStub{}, &registryServiceStub{},
		&ModuleTemplateServiceStub{}, &CRDParserServiceStub{}, &ModuleResourceServiceStub{},
		&imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{}, &fileResolverStub{},
		&fileExistsStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	require.Contains(t, err.Error(), "moduleConfigService")
}

func Test_CreateModule_ReturnsError_WhenModuleConfigFileIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withModuleConfigFile("").build()

	err = svc.Run(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	require.Contains(t, err.Error(), "opts.ConfigFile")
}

func Test_CreateModule_ReturnsError_WhenOutIsNil(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withOut(nil).build()

	err = svc.Run(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	require.Contains(t, err.Error(), "opts.Out")
}

func Test_CreateModule_ReturnsError_WhenCredentialsIsInInvalidFormat(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withCredentials("user").build()

	err = svc.Run(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	require.Contains(t, err.Error(), "opts.Credentials")
}

func Test_CreateModule_ReturnsError_WhenTemplateOutputIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withTemplateOutput("").build()

	err = svc.Run(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	require.Contains(t, err.Error(), "opts.TemplateOutput")
}

func Test_CreateModule_ReturnsError_WhenParseAndValidateModuleConfigReturnsError(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceParseErrorStub{}, &gitSourcesServiceStub{},
		&securityConfigServiceStub{}, &componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().build()

	err = svc.Run(opts)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to read module config file")
}

func Test_CreateModule_ReturnsError_WhenResolvingManifestFilePathReturnsError(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverErrorStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().build()

	err = svc.Run(opts)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to resolve file")
}

func Test_CreateModule_ReturnsError_WhenResolvingDefaultCRFilePathReturnsError(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverErrorStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().build()

	err = svc.Run(opts)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to resolve file")
}

func Test_CreateModule_ReturnsError_WhenModuleSourcesGitDirectoryIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withModuleSourcesGitDirectory("").build()

	err = svc.Run(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	require.Contains(t, err.Error(), "opts.ModuleSourcesGitDirectory must not be empty")
}

func Test_CreateModule_ReturnsError_WhenModuleSourcesIsNotGitDirectory(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{}, &gitSourcesServiceStub{}, &securityConfigServiceStub{},
		&componentConstructorServiceStub{},
		&componentArchiveServiceStub{}, &registryServiceStub{}, &ModuleTemplateServiceStub{}, &CRDParserServiceStub{},
		&ModuleResourceServiceStub{}, &imageVersionVerifierStub{}, &manifestServiceStub{}, &fileResolverStub{},
		&fileResolverStub{},
		&fileExistsStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withModuleSourcesGitDirectory(".").build()

	err = svc.Run(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	require.Contains(t, err.Error(),
		"currently configured module-sources-git-directory \".\" must point to a valid git repository")
}

type createOptionsBuilder struct {
	options create.Options
}

func newCreateOptionsBuilder() *createOptionsBuilder {
	builder := &createOptionsBuilder{
		options: create.Options{},
	}

	return builder.
		withOut(iotools.NewDefaultOut(io.Discard)).
		withModuleConfigFile("create-module-config.yaml").
		withRegistryURL("https://registry.kyma.cx").
		withTemplateOutput("test").
		withCredentials("user:password").
		withModuleSourcesGitDirectory("../../../")
}

func (b *createOptionsBuilder) build() create.Options {
	return b.options
}

func (b *createOptionsBuilder) withOut(out iotools.Out) *createOptionsBuilder {
	b.options.Out = out
	return b
}

func (b *createOptionsBuilder) withModuleConfigFile(moduleConfigFile string) *createOptionsBuilder {
	b.options.ConfigFile = moduleConfigFile
	return b
}

func (b *createOptionsBuilder) withRegistryURL(registryURL string) *createOptionsBuilder {
	b.options.RegistryURL = registryURL
	return b
}

func (b *createOptionsBuilder) withTemplateOutput(templateOutput string) *createOptionsBuilder {
	b.options.TemplateOutput = templateOutput
	return b
}

func (b *createOptionsBuilder) withCredentials(credentials string) *createOptionsBuilder {
	b.options.Credentials = credentials
	return b
}

func (b *createOptionsBuilder) withModuleSourcesGitDirectory(moduleSourcesGitDirectory string) *createOptionsBuilder {
	b.options.ModuleSourcesGitDirectory = moduleSourcesGitDirectory
	return b
}

type fileExistsStub struct{}

func (*fileExistsStub) FileExists(_ string) (bool, error) {
	return true, nil
}

func (*fileExistsStub) ReadFile(_ string) ([]byte, error) {
	return nil, nil
}

type fileResolverStub struct{}

func (*fileResolverStub) Resolve(_ contentprovider.UrlOrLocalFile, _ string) (string, error) {
	return "/tmp/some-file.yaml", nil
}

func (*fileResolverStub) CleanupTempFiles() []error {
	return nil
}

type fileResolverErrorStub struct{}

func (*fileResolverErrorStub) Resolve(_ contentprovider.UrlOrLocalFile, _ string) (string, error) {
	return "", errors.New("failed to resolve file")
}

func (*fileResolverErrorStub) CleanupTempFiles() []error {
	return []error{errors.New("failed to cleanup temp files")}
}

type moduleConfigServiceStub struct{}

func (*moduleConfigServiceStub) ParseAndValidateModuleConfig(_ string) (*contentprovider.ModuleConfig, error) {
	var fileRef contentprovider.UrlOrLocalFile
	if err := fileRef.FromString("default-cr.yaml"); err != nil {
		return nil, err
	}
	return &contentprovider.ModuleConfig{
		Name:      "kyma-project.io/module/telemetry",
		DefaultCR: fileRef,
		Version:   "1.43.1",
	}, nil
}

type moduleConfigServiceParseErrorStub struct{}

func (*moduleConfigServiceParseErrorStub) ParseAndValidateModuleConfig(_ string) (
	*contentprovider.ModuleConfig, error,
) {
	return nil, errors.New("failed to read module config file")
}

type gitSourcesServiceStub struct{}

func (s *gitSourcesServiceStub) AddGitSourcesToConstructor(_ *component.Constructor,
	_, _ string,
) error {
	return nil
}

func (*gitSourcesServiceStub) AddGitSources(_ *compdesc.ComponentDescriptor,
	_, _, _ string,
) error {
	return nil
}

type securityConfigServiceStub struct{}

func (s *securityConfigServiceStub) AppendSecurityScanConfigToConstructor(_ *component.Constructor,
	_ contentprovider.SecurityScanConfig,
) {
}

func (*securityConfigServiceStub) ParseSecurityConfigData(_ string) (*contentprovider.SecurityScanConfig, error) {
	return &contentprovider.SecurityScanConfig{}, nil
}

func (*securityConfigServiceStub) AppendSecurityScanConfig(_ *compdesc.ComponentDescriptor,
	_ contentprovider.SecurityScanConfig,
) error {
	return nil
}

type componentConstructorServiceStub struct{}

func (c *componentConstructorServiceStub) AddImagesToConstructor(_ *component.Constructor,
	_ []string,
) error {
	return nil
}

func (c *componentConstructorServiceStub) AddResources(_ *component.Constructor,
	_ *types.ResourcePaths,
) error {
	return nil
}

func (c *componentConstructorServiceStub) CreateConstructorFile(_ *component.Constructor,
	_ string,
) error {
	return nil
}

type componentArchiveServiceStub struct{}

func (*componentArchiveServiceStub) CreateComponentArchive(_ *compdesc.ComponentDescriptor) (
	*comparch.ComponentArchive,
	error,
) {
	return &comparch.ComponentArchive{}, nil
}

func (*componentArchiveServiceStub) AddModuleResourcesToArchive(_ componentarchive.ComponentArchive,
	_ []resources.Resource,
) error {
	return nil
}

type registryServiceStub struct{}

func (*registryServiceStub) PushComponentVersion(_ *comparch.ComponentArchive, _, _ bool,
	_, _ string,
) error {
	return nil
}

func (*registryServiceStub) GetComponentVersion(_ *comparch.ComponentArchive, _ bool,
	_, _ string,
) (cpi.ComponentVersionAccess, error) {
	var componentVersion cpi.ComponentVersionAccess
	return componentVersion, nil
}

func (*registryServiceStub) ExistsComponentVersion(_ *comparch.ComponentArchive, _ bool,
	_, _ string,
) (bool, error) {
	return false, nil
}

type ModuleTemplateServiceStub struct{}

func (*ModuleTemplateServiceStub) GenerateModuleTemplate(_ *contentprovider.ModuleConfig,
	_ *compdesc.ComponentDescriptor,
	_ []byte, _ bool, _ string,
) error {
	return nil
}

type CRDParserServiceStub struct{}

func (*CRDParserServiceStub) IsCRDClusterScoped(_ *types.ResourcePaths) (bool, error) {
	return false, nil
}

type ModuleResourceServiceStub struct{}

func (*ModuleResourceServiceStub) GenerateModuleResources(
	_ *types.ResourcePaths,
	_ string,
) []resources.Resource {
	return []resources.Resource{}
}

type imageVersionVerifierStub struct{}

func (*imageVersionVerifierStub) VerifyModuleResources(_ *contentprovider.ModuleConfig,
	_ string,
) error {
	return nil
}

type manifestServiceStub struct{}

func (*manifestServiceStub) ExtractImagesFromManifest(_ string) ([]string, error) {
	return []string{"image1:latest", "image2:v1.0"}, nil
}
