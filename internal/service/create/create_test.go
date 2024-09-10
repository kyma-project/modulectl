package create_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
	"github.com/kyma-project/modulectl/internal/service/create"
	iotools "github.com/kyma-project/modulectl/tools/io"
	"io"
)

func Test_NewService_ReturnsError_WhenModuleConfigServiceIsNil(t *testing.T) {
	_, err := create.NewService(nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "moduleConfigService")
}

type moduleConfigServiceMock struct{}

func (*moduleConfigServiceMock) ParseModuleConfig(_ string) (*contentprovider.ModuleConfig, error) {
	return nil, nil
}

func (*moduleConfigServiceMock) ValidateModuleConfig(_ *contentprovider.ModuleConfig) error {
	return nil
}

func (*moduleConfigServiceMock) GetDefaultCRPath(_ string) (string, error) {
	return "", nil
}

func (*moduleConfigServiceMock) GetManifestPath(_ string) (string, error) {
	return "", nil
}

func Test_CreateModule_ReturnsError_WhenModuleConfigFileIsEmpty(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceMock{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withModuleConfigFile("").build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.ModuleConfigFile")
}

func Test_CreateModule_ReturnsError_WhenOutIsNil(t *testing.T) {
	svc, err := create.NewService(&moduleConfigServiceMock{})
	require.NoError(t, err)

	opts := newCreateOptionsBuilder().withOut(nil).build()

	err = svc.CreateModule(opts)

	require.ErrorIs(t, err, commonerrors.ErrInvalidOption)
	assert.Contains(t, err.Error(), "opts.Out")
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
		withTemplateOutput("test")
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
