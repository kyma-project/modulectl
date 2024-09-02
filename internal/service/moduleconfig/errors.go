package moduleconfig

import "errors"

var ErrFileExists = errors.New("module config file already exists. Use the overwrite option to overwrite it")
