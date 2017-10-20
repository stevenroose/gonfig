package gonfig

import (
	"fmt"
	"os"
	"reflect"

	"github.com/spf13/pflag"
)

const (
	defaultHelpDescription = "print this help menu"
)

// createFlagSet builds the flagset for the options in the setup.
func createFlagSet(s *setup) *pflag.FlagSet {
	flagSet := pflag.NewFlagSet(os.Args[0], pflag.ContinueOnError)
	flagSet.SortFlags = false

	for _, opt := range s.allOpts {
		if opt.isParent {
			// Parents are skipped, we should only add the children.
			continue
		}

		switch opt.value.Type().Kind() {
		case reflect.Bool:
			var def bool
			if opt.defaul == "true" {
				def = true
			}
			flagSet.BoolP(opt.fullId(), opt.short, def, opt.desc)

		case reflect.Slice:
			if opt.value.Type().Elem().Kind() == reflect.Uint8 {
				// Special case for byte slices.
				flagSet.StringP(opt.fullId(), opt.short, opt.defaul, opt.desc)
				break
			}
			defSlice, err := readAsCSV(opt.defaul)
			if err != nil {
				panic(fmt.Sprintf(
					"error parsing default value '%s' for slice variable %s: %s",
					opt.defaul, opt.fullId(), err))
			}
			flagSet.StringSliceP(opt.fullId(), opt.short, defSlice, opt.desc)

		default:
			// We use strings for everything else since there is no visual
			// difference in the help output and we need logic for parsing
			// values from into the target type anyhow.
			flagSet.StringP(opt.fullId(), opt.short, opt.defaul, opt.desc)
		}
	}

	if s.conf.HelpEnable {
		desc := s.conf.HelpDescription
		if desc == "" {
			desc = defaultHelpDescription
		}

		//TODO test help message
		flagSet.BoolP("help", "h", false, desc)
	}

	return flagSet
}

// initFlags makes sure that the flagset should only be initialized once.
// This method initializes the flagset and stores it; when called a second
// time, it just returns nil.
func initFlags(s *setup) error {
	// Check if already initialized.
	if s.flagSet != nil {
		return nil
	}

	s.flagSet = createFlagSet(s)

	if err := s.flagSet.Parse(os.Args[1:]); err != nil {
		return err
	}

	return nil
}

// parseFlags parses the command line flags for all config options
// and writes the values that have been found in place.
func parseFlags(s *setup) error {
	if err := initFlags(s); err != nil {
		return err
	}

	for _, opt := range s.allOpts {
		if opt.isParent {
			// Parents are skipped, we should only add the children.
			continue
		}

		// Prevent storing empty (unset) values.
		if !s.flagSet.Changed(opt.fullId()) {
			continue
		}

		flag := s.flagSet.Lookup(opt.fullId())
		stringValue := flag.Value.String()

		if opt.isSlice {
			// Trim the square brackets of the string.
			stringValue = stringValue[1 : len(stringValue)-1]
		}

		if err := opt.setValueByString(stringValue); err != nil {
			return fmt.Errorf("error parsing flag %s: %s", opt.fullId(), err)
		}
	}

	return nil
}

// lookupConfigFileFlag looks for the config file in the command line flags.
func lookupConfigFileFlag(s *setup, configOpt *option) (string, error) {
	if err := initFlags(s); err != nil {
		return "", err
	}

	return s.flagSet.Lookup(configOpt.id).Value.String(), nil
}
