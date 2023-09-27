package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

// requireFlags ensures that each of the provided `flags` is supplied to viper,
// either via. cobra flags or env vars
func requireFlags(flags ...string) error {
	var errs []error

	for _, flag := range flags {
		if !viper.IsSet(flag) {
			errs = append(errs, fmt.Errorf("Required flag not provided: %v", flag))
		}
	}

	return errors.Join(errs...)
}
