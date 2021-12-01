// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package config

import (
	"fmt"
	"strings"
)

// validate verifies all Config struct fields have been provided acceptable
// values.
func (c Config) validate(appType AppType) error {

	// Flags specific to one plugin type or the other
	switch {
	case appType.PluginComponents:

		componentOrGroupSpecified := func() bool {
			switch {
			case c.componentGroup != "":
				return true
			case len(c.componentsList) > 0:
				return true
			}

			return false
		}

		switch {
		case c.EvalAllComponents && componentOrGroupSpecified():
			return fmt.Errorf(
				"invalid combination of flags; "+
					"%s flag is incompatible with %q or %q flag",
				EvalAllComponentsFlagLong,
				ComponentsListFlagLong,
				ComponentGroupFlagLong,
			)

		case !c.EvalAllComponents && !componentOrGroupSpecified():
			return fmt.Errorf(
				"missing component values; must specify one of"+
					" %s, %s or %s flags",
				EvalAllComponentsFlagLong,
				ComponentsListFlagLong,
				ComponentGroupFlagLong,
			)
		}

		// Assert that group, component flags were not provided only
		// whitespace characters.
		switch {
		case c.componentGroup != "":
			if strings.TrimSpace(c.componentGroup) == "" {
				return fmt.Errorf(
					"whitespace only group value provided to %s flag",
					ComponentGroupFlagLong,
				)
			}

		case len(c.componentsList) > 0:
			for _, component := range c.componentsList {
				if strings.TrimSpace(component) == "" {
					return fmt.Errorf(
						"whitespace only component value provided to %s flag",
						ComponentsListFlagLong,
					)
				}
			}
		}

	case appType.InspectorComponents:

		supportedFormats := supportedInspectorOutputFormats()
		isSupportedOutputFormat := func(specified string, supported []string) bool {
			for _, supportedFormat := range supported {
				if strings.EqualFold(specified, supportedFormat) {
					return true
				}
			}

			return false
		}

		if !isSupportedOutputFormat(c.InspectorOutputFormat, supportedFormats) {
			return fmt.Errorf(
				"invalid output format specified; got %v, expected one of %v",
				c.InspectorOutputFormat,
				supportedFormats,
			)
		}

	}

	// shared validation checks

	if c.URL == "" && c.Filename == "" {
		return fmt.Errorf("components feed URL or filename not provided")
	}

	if c.URL != "" && c.Filename != "" {
		return fmt.Errorf(
			"invalid combination of flags; only one of %s or %s flags are permitted",
			URLFlagLong,
			FilenameFlagLong,
		)
	}

	if c.Timeout() < 1 {
		return fmt.Errorf("invalid timeout value %d provided", c.Timeout())
	}

	requestedLoggingLevel := strings.ToLower(c.LoggingLevel)
	if _, ok := loggingLevels[requestedLoggingLevel]; !ok {
		return fmt.Errorf("invalid logging level %q", c.LoggingLevel)
	}

	// Optimist
	return nil

}
