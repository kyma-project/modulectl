package create_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/create"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

func Test_NewService_ReturnsError_WhenModuleConfigServiceIsNil(t *testing.T) {
	_, err := create.NewService(nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "moduleConfigService")
}

func Test_CreateModule_ReturnsError_WhenModuleConfigFileIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withModuleConfigFile("").build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.ModuleConfigFile")
}

func Test_CreateModule_ReturnsError_WhenOutIsNil(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withOut(nil).build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.Out")
}

func Test_CreateModule_ReturnsError_WhenGitRemoteIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withGitRemote("").build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.GitRemote")
}

func Test_CreateModule_ReturnsError_WhenCredentialsIsInInvalidFormat(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withCredentials("user").build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.Credentials")
}

func Test_CreateModule_ReturnsError_WhenTemplateOutputIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withTemplateOutput("").build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.TemplateOutput")
}

func Test_CreateModule_ReturnsError_WhenRegistryURLIsInInvalidFormat(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withRegistryURL("test").build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.RegistryURL")
}

func Test_CreateModule_ReturnsError_WhenParseModuleConfigReturnsError(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceParseErrorStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().build()

	err = svc.CreateModule(opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to parse module config file")
}

func Test_CreateModule_ReturnsError_WhenGetDefaultCRPathReturnsError(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceDefaultCRErrorStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().build()

	err = svc.CreateModule(opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download default CR file")
}

func Test_CreateModule_ReturnsError_WhenGetManifestPathReturnsError(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceManifestErrorStub{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().build()

	err = svc.CreateModule(opts)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to download Manifest file")
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
		withGitRemote("origin").
		withTemplateOutput("test").
		withCredentials("user:password")
}

func (b *createOptionsBuilder) build() create.Options {
	return b.options
}

func (b *createOptionsBuilder) withOut(out iotools.Out) *createOptionsBuilder {
	b.options.Out = out
	return b
}

func (b *createOptionsBuilder) withModuleConfigFile(moduleConfigFile string) *createOptionsBuilder {
	b.options.ModuleConfigFile = moduleConfigFile
	return b
}

func (b *createOptionsBuilder) withRegistryURL(registryURL string) *createOptionsBuilder {
	b.options.RegistryURL = registryURL
	return b
}

func (b *createOptionsBuilder) withGitRemote(gitRemote string) *createOptionsBuilder {
	b.options.GitRemote = gitRemote
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

// Test Stubs
type moduleConfigServiceStub struct{}

func (*moduleConfigServiceStub) ParseModuleConfig(_ string) (*contentprovider.ModuleConfig, error) {
	return &contentprovider.ModuleConfig{}, nil
}

func (*moduleConfigServiceStub) ValidateModuleConfig(_ *contentprovider.ModuleConfig) error {
	return nil
}

func (*moduleConfigServiceStub) GetDefaultCRPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceStub) GetManifestPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceStub) CleanupTempFiles() []error {
	return nil
}

type moduleConfigServiceParseErrorStub struct{}

func (*moduleConfigServiceParseErrorStub) ParseModuleConfig(_ string) (*contentprovider.ModuleConfig, error) {
	return nil, errors.New("failed to read module config file")
}

func (*moduleConfigServiceParseErrorStub) ValidateModuleConfig(_ *contentprovider.ModuleConfig) error {
	return nil
}

func (*moduleConfigServiceParseErrorStub) GetDefaultCRPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceParseErrorStub) GetManifestPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceParseErrorStub) CleanupTempFiles() []error {
	return nil
}

type moduleConfigServiceDefaultCRErrorStub struct{}

func (*moduleConfigServiceDefaultCRErrorStub) ParseModuleConfig(_ string) (*contentprovider.ModuleConfig, error) {
	return &contentprovider.ModuleConfig{
		DefaultCRPath: "test",
	}, nil
}

func (*moduleConfigServiceDefaultCRErrorStub) ValidateModuleConfig(_ *contentprovider.ModuleConfig) error {
	return nil
}

func (*moduleConfigServiceDefaultCRErrorStub) GetDefaultCRPath(_ string) (string, error) {
	return "", errors.New("failed to download default CR file")
}

func (*moduleConfigServiceDefaultCRErrorStub) GetManifestPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceDefaultCRErrorStub) CleanupTempFiles() []error {
	return nil
}

type moduleConfigServiceManifestErrorStub struct{}

func (*moduleConfigServiceManifestErrorStub) ParseModuleConfig(_ string) (*contentprovider.ModuleConfig, error) {
	return &contentprovider.ModuleConfig{
		ManifestPath: "test",
	}, nil
}

func (*moduleConfigServiceManifestErrorStub) ValidateModuleConfig(_ *contentprovider.ModuleConfig) error {
	return nil
}

func (*moduleConfigServiceManifestErrorStub) GetDefaultCRPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceManifestErrorStub) GetManifestPath(_ string) (string, error) {
	return "", errors.New("failed to download Manifest file")
}

func (*moduleConfigServiceManifestErrorStub) CleanupTempFiles() []error {
	return nil
}
