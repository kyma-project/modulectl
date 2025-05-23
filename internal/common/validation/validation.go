package validation

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"

	commonerrors "github.com/kyma-project/modulectl/internal/common/errors"
)

const (
	// // taken from "https://github.com/open-component-model/ocm/blob/4473dacca406e4c84c0ac5e6e14393c659384afc/resources/component-descriptor-v2-schema.yaml#L40"
	moduleNamePattern   = "^[a-z][-a-z0-9]*([.][a-z][-a-z0-9]*)*[.][a-z]{2,}(/[a-z][-a-z0-9_]*([.][a-z][-a-z0-9_]*)*)+$"
	moduleNameMaxLength = 255
	namespaceMaxLength  = 253
	namespacePattern    = "^[a-z0-9]+(?:-[a-z0-9]+)*$"
)

func ValidateModuleName(name string) error {
	if name == "" {
		return fmt.Errorf("opts.ModuleName must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if len(name) > moduleNameMaxLength {
		return fmt.Errorf("opts.ModuleName length must not exceed %q characters: %w", moduleNameMaxLength, commonerrors.ErrInvalidOption)
	}

	if name != strings.ToLower(name) {
		return fmt.Errorf("opts.ModuleName must not contain uppercase letters: %w", commonerrors.ErrInvalidOption)
	}

	if matched, err := regexp.MatchString(moduleNamePattern, name); err != nil {
		return fmt.Errorf("failed to evaluate regex pattern for opts.ModuleName: %w", commonerrors.ErrInvalidOption)
	} else if !matched {
		return fmt.Errorf("opts.ModuleName must match the required pattern, e.g: 'github.com/path-to/your-repo': %w",
			commonerrors.ErrInvalidOption)
	}

	return nil
}

func ValidateModuleVersion(version string) error {
	if version == "" {
		return fmt.Errorf("opts.ModuleVersion must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if err := validateSemanticVersion(version); err != nil {
		return err
	}

	return nil
}

func ValidateModuleNamespace(namespace string) error {
	if namespace == "" {
		return fmt.Errorf("opts.ModuleNamespace must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if err := ValidateNamespace(namespace); err != nil {
		return err
	}

	return nil
}

func ValidateNamespace(namespace string) error {
	if len(namespace) > namespaceMaxLength {
		return fmt.Errorf("opts.ModuleNamespace length must not exceed %q characters: %w",
			namespaceMaxLength,
			commonerrors.ErrInvalidOption)
	}

	if matched, err := regexp.MatchString(namespacePattern, namespace); err != nil {
		return fmt.Errorf("failed to evaluate regex pattern for module namespace: %w", err)
	} else if !matched {
		return fmt.Errorf("namespace must match the required pattern, only small alphanumeric characters and hyphens: %w",
			commonerrors.ErrInvalidOption)
	}

	return nil
}

func ValidateMapEntries(nameLinkMap map[string]string) error {
	for name, link := range nameLinkMap {
		if name == "" {
			return fmt.Errorf("name must not be empty: %w", commonerrors.ErrInvalidOption)
		}

		if link == "" {
			return fmt.Errorf("link must not be empty: %w", commonerrors.ErrInvalidOption)
		}

		if err := ValidateIsValidHTTPSURL(link); err != nil {
			return fmt.Errorf("failed to validate link: %w", err)
		}
	}

	return nil
}

func ValidateIsValidHTTPSURL(input string) error {
	if input == "" {
		return fmt.Errorf("must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	_url, err := url.Parse(input)
	if err != nil {
		return fmt.Errorf("'%s' is not a valid URL: %w", input, commonerrors.ErrInvalidOption)
	}

	if _url.Scheme != "https" {
		return fmt.Errorf("'%s' is not using https scheme: %w", input, commonerrors.ErrInvalidOption)
	}

	return nil
}

func validateSemanticVersion(version string) error {
	_, err := semver.StrictNewVersion(strings.TrimSpace(version))
	if err != nil {
		return fmt.Errorf("opts.ModuleVersion failed to be parsed as semantic version: %w",
			commonerrors.ErrInvalidOption)
	}

	return nil
}

func ValidateGvk(group, version, kind string) error {
	if kind == "" {
		return fmt.Errorf("kind must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if group == "" {
		return fmt.Errorf("group must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	if version == "" {
		return fmt.Errorf("version must not be empty: %w", commonerrors.ErrInvalidOption)
	}

	return nil
}
