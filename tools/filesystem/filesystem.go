package filesystem

import (
	"errors"
	"fmt"
	"os"
)

type Helper struct{}

func (u *Helper) FileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil
	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, fmt.Errorf("failed check if file exists %s: %w", path, err)
	}
}

const perm = 0o600

func (u *Helper) WriteFile(path, content string) error {
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		return fmt.Errorf("failed to write file %s: %w", path, err)
	}

	return nil
}

func (u *Helper) ReadFile(path string) ([]byte, error) {
	fileContent, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return fileContent, nil
}
