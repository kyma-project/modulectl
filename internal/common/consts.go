package common

const (
	ProviderName         = "kyma-project.io"
	BuiltByLabelKey      = ProviderName + "/built-by"
	BuiltByLabelValue    = "modulectl"
	VersionV1, VersionV2 = "v1", "v2"

	OCMIdentityName = "module-sources"
	RefLabel        = "git.kyma-project.io/ref"
	OCMVersion      = "v1"

	SecScanBaseLabelKey       = "scan.security.kyma-project.io"
	TypeLabelKey              = "type"
	ThirdPartyImageLabelValue = "third-party-image"

	ModuleImageResourceName    = "module-image"
	RawManifestResourceName    = "raw-manifest"
	DefaultCRResourceName      = "default-cr"
	ModuleTemplateResourceName = "moduletemplate"
)
