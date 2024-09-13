package scaffold_test

import (
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/common/types"
	"github.com/kyma-project/modulectl/internal/service/scaffold"
	iotools "github.com/kyma-project/modulectl/tools/io"
)

func Test_NewService_ReturnsError_WhenModuleConfigServiceIsNil(t *testing.T) {
	_, err := scaffold.NewService(
		nil,
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "moduleConfigService")
}

func Test_NewService_ReturnsError_WhenManifestServiceIsNil(t *testing.T) {
	_, err := scaffold.NewService(
		&moduleConfigForceExplicitOverwriteErrorStub{},
		nil,
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "manifestService")
}

func Test_NewService_ReturnsError_WhenDefaultCRServiceIsNil(t *testing.T) {
	_, err := scaffold.NewService(
		&moduleConfigForceExplicitOverwriteErrorStub{},
		&fileGeneratorErrorStub{},
		nil,
		&fileGeneratorErrorStub{})

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "defaultCRService")
}

func Test_NewService_ReturnsError_WhenSecurityConfigServiceIsNil(t *testing.T) {
	_, err := scaffold.NewService(
		&moduleConfigForceExplicitOverwriteErrorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{},
		nil)

	require.ErrorIs(t, err, commonerrors.ErrInvalidArg)
	assert.Contains(t, err.Error(), "securityConfigService")
}

func Test_CreateScaffold_ReturnsError_WhenModuleConfigServiceForceExplicitOverwriteReturnsError(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigForceExplicitOverwriteErrorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.ErrorIs(t, result, errOverwriteError)
}

func Test_CreateScaffold_ReturnsError_WhenGeneratingManifestFileFails(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigGenerateFileErrorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.ErrorIs(t, result, scaffold.ErrGeneratingFile)
	require.ErrorIs(t, result, errSomeFileGeneratorError)
	assert.Contains(t, result.Error(), "manifest.yaml")
}

func Test_CreateScaffold_Succeeds_WhenGeneratingManifestFile(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.NoError(t, result)
}

func Test_CreateScaffold_Succeeds_WhenDefaultCRFileIsNotConfigured(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigStub{},
		&fileGeneratorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().withDefaultCRFileName("").build())

	require.NoError(t, result)
}

func Test_CreateScaffold_ReturnsError_WhenGeneratingDefaultCRFileFails(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigGenerateFileErrorStub{},
		&fileGeneratorStub{},
		&fileGeneratorErrorStub{},
		&fileGeneratorErrorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.ErrorIs(t, result, scaffold.ErrGeneratingFile)
	require.ErrorIs(t, result, errSomeFileGeneratorError)
	assert.Contains(t, result.Error(), "default-cr.yaml")
}

func Test_CreateScaffold_Succeeds_WhenGeneratingDefaultCRFile(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.NoError(t, result)
}

func Test_CreateScaffold_ReturnsError_WhenGeneratingSecurityConfigFileFails(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigGenerateFileErrorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{},
		&fileGeneratorErrorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.ErrorIs(t, result, scaffold.ErrGeneratingFile)
	require.ErrorIs(t, result, errSomeFileGeneratorError)
	assert.Contains(t, result.Error(), "security-config.yaml")
}

func Test_CreateScaffold_Succeeds_WhenGeneratingSecurityConfigFile(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.NoError(t, result)
}

func Test_CreateScaffold_ReturnsError_WhenGeneratingModuleConfigReturnsError(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigGenerateFileErrorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.ErrorIs(t, result, scaffold.ErrGeneratingFile)
	require.ErrorIs(t, result, errSomeFileGeneratorError)
	assert.Contains(t, result.Error(), "module-config.yaml")
}

func Test_CreateScaffold_Succeeds(t *testing.T) {
	svc, _ := scaffold.NewService(
		&moduleConfigStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{},
		&fileGeneratorStub{})

	result := svc.Run(newScaffoldOptionsBuilder().build())

	require.NoError(t, result)
}

// Test Stubs

var (
	errSomeFileGeneratorError = errors.New("some file generator error")
	errSomeUnexpectedError    = errors.New("if you see this error, something went wrong in the test setup")
)

type moduleConfigForceExplicitOverwriteErrorStub struct{}

var errOverwriteError = errors.New("overwrite error")

func (*moduleConfigForceExplicitOverwriteErrorStub) ForceExplicitOverwrite(_, _ string, _ bool) error {
	return errOverwriteError
}

func (*moduleConfigForceExplicitOverwriteErrorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return errSomeUnexpectedError
}

type moduleConfigGenerateFileErrorStub struct{}

func (*moduleConfigGenerateFileErrorStub) ForceExplicitOverwrite(_, _ string, _ bool) error {
	return nil
}

func (*moduleConfigGenerateFileErrorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return errSomeFileGeneratorError
}

type moduleConfigStub struct{}

func (*moduleConfigStub) ForceExplicitOverwrite(_, _ string, _ bool) error {
	return nil
}

func (*moduleConfigStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return nil
}

type fileGeneratorErrorStub struct{}

func (*fileGeneratorErrorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return errSomeFileGeneratorError
}

type fileGeneratorStub struct{}

func (*fileGeneratorStub) GenerateFile(_ iotools.Out, _ string, _ types.KeyValueArgs) error {
	return nil
}

// Test Options Builder

type scaffoldOptionsBuilder struct {
	options scaffold.Options
}

func newScaffoldOptionsBuilder() *scaffoldOptionsBuilder {
	builder := &scaffoldOptionsBuilder{
		options: scaffold.Options{},
	}

	return builder.
		withOut(iotools.NewDefaultOut(io.Discard)).
		withDirectory("./").
		withModuleConfigFileName("scaffold-module-config.yaml").
		withManifestFileName("manifest.yaml").
		withDefaultCRFileName("default-cr.yaml").
		withModuleConfigFileOverwrite(false).
		withSecurityConfigFileName("security-config.yaml").
		withModuleName("github.com/kyma-project/test").
		withModuleVersion("0.0.1").
		withModuleChannel("experimental")
}

func (b *scaffoldOptionsBuilder) build() scaffold.Options {
	return b.options
}

func (b *scaffoldOptionsBuilder) withOut(out iotools.Out) *scaffoldOptionsBuilder {
	b.options.Out = out
	return b
}

func (b *scaffoldOptionsBuilder) withDirectory(directory string) *scaffoldOptionsBuilder {
	b.options.Directory = directory
	return b
}

func (b *scaffoldOptionsBuilder) withModuleConfigFileName(moduleConfigFileName string) *scaffoldOptionsBuilder {
	b.options.ModuleConfigFileName = moduleConfigFileName
	return b
}

func (b *scaffoldOptionsBuilder) withModuleConfigFileOverwrite(moduleConfigFileOverwrite bool) *scaffoldOptionsBuilder {
	b.options.ModuleConfigFileOverwrite = moduleConfigFileOverwrite
	return b
}

func (b *scaffoldOptionsBuilder) withManifestFileName(manifestFileName string) *scaffoldOptionsBuilder {
	b.options.ManifestFileName = manifestFileName
	return b
}

func (b *scaffoldOptionsBuilder) withDefaultCRFileName(defaultCRFileName string) *scaffoldOptionsBuilder {
	b.options.DefaultCRFileName = defaultCRFileName
	return b
}

func (b *scaffoldOptionsBuilder) withSecurityConfigFileName(securityConfigFileName string) *scaffoldOptionsBuilder {
	b.options.SecurityConfigFileName = securityConfigFileName
	return b
}

func (b *scaffoldOptionsBuilder) withModuleName(moduleName string) *scaffoldOptionsBuilder {
	b.options.ModuleName = moduleName
	return b
}

func (b *scaffoldOptionsBuilder) withModuleVersion(moduleVersion string) *scaffoldOptionsBuilder {
	b.options.ModuleVersion = moduleVersion
	return b
}

func (b *scaffoldOptionsBuilder) withModuleChannel(moduleChannel string) *scaffoldOptionsBuilder {
	b.options.ModuleChannel = moduleChannel
	return b
}
