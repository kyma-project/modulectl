package ocirepo

import (
	"errors"
	"fmt"

	"ocm.software/ocm/api/ocm/cpi"
	"ocm.software/ocm/api/ocm/extensions/repositories/comparch"
	"ocm.software/ocm/api/ocm/tools/transfer"
	"ocm.software/ocm/api/ocm/tools/transfer/transferhandler/standard"
	"ocm.software/ocm/api/utils/misc"
)

type OCIRepo struct{}

var errComponentVersionAlreadyExists = errors.New("component version already exists, cannot push the new version")

func (o *OCIRepo) GetComponentVersion(archive *comparch.ComponentArchive,
	repo cpi.Repository,
) (cpi.ComponentVersionAccess, error) {
	version, err := repo.LookupComponentVersion(archive.GetName(), archive.GetVersion())
	if err != nil {
		return nil, fmt.Errorf("failed to get component version: %w", err)
	}

	return version, nil
}

func (o *OCIRepo) PushComponentVersion(archive *comparch.ComponentArchive, repo cpi.Repository,
	overwrite bool,
) error {
	exists, _ := repo.ExistsComponentVersion(archive.GetName(), archive.GetVersion())
	if exists && !overwrite {
		return fmt.Errorf("cannot push component version %s: %w",
			archive.GetVersion(), errComponentVersionAlreadyExists)
	}

	transferHandler, err := standard.New(standard.Overwrite(overwrite))
	if err != nil {
		return fmt.Errorf("failed to setup archive transfer: %w", err)
	}

	if err := transfer.TransferVersion(
		misc.NewLoggingPrinter(archive.GetContext().Logger()), nil, archive, repo,
		transferHandler); err != nil {
		return fmt.Errorf("failed to finish component transfer: %w", err)
	}

	return nil
}
