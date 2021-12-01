// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package config

import (
	"fmt"
	"time"
)

// Timeout converts the user-specified plugin runtime/execution timeout value
// in seconds to an appropriate time duration value for use with setting
// context deadline value.
func (c Config) Timeout() time.Duration {
	return time.Duration(c.timeout) * time.Second
}

// UserAgent returns a string usable as-is as a custom user agent for plugins
// provided by this project.
func (c Config) UserAgent() string {

	// Default User Agent: (Go-http-client/1.1)
	// https://datatracker.ietf.org/doc/html/draft-ietf-httpbis-p2-semantics-22#section-5.5.3
	return fmt.Sprintf(
		"%s/%s",
		c.App.Name,
		c.App.Version,
	)
}

// ComponentFilter returns the user-specified component filter values.
// Combinations of component group and subcomponents, just a group or the list
// of individual components that should be monitored are returned, wrapped in
// a ComponentFilter type. Configuration validation prevents use of
// incompatible flag types.
func (c Config) ComponentFilter() ComponentFilter {
	return ComponentFilter{
		Group:      c.componentGroup,
		Components: c.componentsList,
	}
}

// supportedInspectorOutputFormats returns a list of valid output formats used
// by Inspector type applications in this project. This list is intended to be
// used for validating the user-specified output format.
func supportedInspectorOutputFormats() []string {
	return []string{
		InspectorOutputFormatOverview,
		InspectorOutputFormatTable,
		InspectorOutputFormatVerbose,
		InspectorOutputFormatDebug,
		InspectorOutputFormatIDsList,
		InspectorOutputFormatJSON,
	}
}
