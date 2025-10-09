// package shared contains constants used across the codebase
package shared

import (
	"errors"
	"fmt"
)

var (
	ErrConfig error = fmt.Errorf("configuration error")
)

func ConfigError(m string, err error) error {
	return errors.Join(ErrConfig, fmt.Errorf("%s: %w", m, err))
}

func IsConfigError(err error) bool {
	return errors.Is(err, ErrConfig)
}
