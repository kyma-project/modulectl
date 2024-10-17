package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
	"github.com/kyma-project/modulectl/internal/service/contentprovider"
)

const (
	// // taken from "https://github.com/open-component-model/ocm/blob/4473dacca406e4c84c0ac5e6e14393c659384afc/resources/component-descriptor-v2-schema.yaml#L40"
	moduleNamePattern   = "^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$"
	moduleNameMaxLength = 255
	channelMinLength    = 3
	channelMaxLength    = 32
	channelPattern      = "^[a-z]+$"
	namespaceMaxLength  = 253
	namespacePattern    = "^[a-z0-9]+(?:-[a-z0-9]+)*$"
)

func ValidateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("%w: opts.ModuleName must not be empty", commonerrors.ErrInvalidOption)
	}

	if len(name) > moduleNameMaxLength {
		return fmt.Errorf("%w: opts.ModuleName length must not exceed %q characters", commonerrors.ErrInvalidOption,
			moduleNameMaxLength)
	}

	if matched, err := regexp.MatchString(moduleNamePattern, name); err != nil {
		return fmt.Errorf("%w: failed to evaluate regex pattern for opts.ModuleName", commonerrors.ErrInvalidOption)
	} else if !matched {
		return fmt.Errorf("%w: opts.ModuleName must match the required pattern, e.g: 'github.com/path-to/your-repo'",
			commonerrors.ErrInvalidOption)
	}

	return nil
}

func ValidateModuleVersion(version string) error {
	if version == "" {
		return fmt.Errorf("%w: opts.ModuleVersion must not be empty", commonerrors.ErrInvalidOption)
	}

	if err := validateSemanticVersion(version); err != nil {
		return err
	}

	return nil
}

func ValidateModuleChannel(channel string) error {
	if channel == "" {
		return fmt.Errorf("%w: opts.ModuleChannel must not be empty", commonerrors.ErrInvalidOption)
	}

	if len(channel) > channelMaxLength {
		return fmt.Errorf("%w: opts.ModuleChannel length must not exceed %q characters", commonerrors.ErrInvalidOption,
			channelMaxLength)
	}

	if len(channel) < channelMinLength {
		return fmt.Errorf("%w: opts.ModuleChannel length must be at least %q characters", commonerrors.ErrInvalidOption,
			channelMinLength)
	}

	if matched, err := regexp.MatchString(channelPattern, channel); err != nil {
		return fmt.Errorf("%w: failed to evaluate regex pattern for opts.ModuleChannel", commonerrors.ErrInvalidOption)
	} else if !matched {
		return fmt.Errorf("%w: opts.ModuleChannel must match the required pattern, only characters from a-z are allowed",
			commonerrors.ErrInvalidOption)
	}

	return nil
}

func ValidateModuleNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("%w: opts.ModuleNamespace must not be empty", commonerrors.ErrInvalidOption)
	}

	if len(namespace) > namespaceMaxLength {
		return fmt.Errorf("%w: opts.ModuleNamespace length must not exceed %q characters",
			commonerrors.ErrInvalidOption,
			namespaceMaxLength)
	}

	if matched, err := regexp.MatchString(namespacePattern, namespace); err != nil {
		return fmt.Errorf("failed to evaluate regex pattern for module namespace: %w", err)
	} else if !matched {
		return fmt.Errorf("%w: namespace must match the required pattern, only small alphanumeric characters and hyphens",
			commonerrors.ErrInvalidOption)
	}

	return nil
}

func ValidateResources(resources contentprovider.ResourcesMap) error {
	for name, link := range resources {
		if name == "" {
			return fmt.Errorf("%w: name must not be empty", commonerrors.ErrInvalidOption)
		}

		if link == "" {
			return fmt.Errorf("%w: link must not be empty", commonerrors.ErrInvalidOption)
		}

		if err := ValidateIsValidHttpsUrl(link); err != nil {
			return err
		}
	}

	return nil
}

func ValidateIsValidHttpsUrl(input string) error {
	_url, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("%w: link %s is not a valid URL", commonerrors.ErrInvalidOption, input)
	}

	if _url.Scheme != "https" {
		return fmt.Errorf("%w: link %s is not using https scheme", commonerrors.ErrInvalidOption, input)
	}

	return nil
}

func validateSemanticVersion(version string) error {
	_, err := semver.StrictNewVersion(strings.TrimSpace(version))
	if err != nil {
		return fmt.Errorf("%w: opts.ModuleVersion failed to parse as semantic version: %w",
			commonerrors.ErrInvalidOption, err)
	}

	return nil
}
