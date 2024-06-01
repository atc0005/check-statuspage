// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atc0005/check-statuspage/internal/config"
	"github.com/atc0005/check-statuspage/internal/statuspage/components"
	"github.com/atc0005/go-nagios"
)

// filterErrAdvice is a small helper function used to evaluate the specific
// filter error that occurred and offer the user some feedback or advice for
// resolving it.
func filterErrAdvice(err error, cs *components.Set, filter components.Filter, feedSrc string) string {

	var tryAgainMsg strings.Builder

	_, _ = fmt.Fprintf(
		&tryAgainMsg,
		"Specified filter: %s%s",
		filter,
		nagios.CheckOutputEOL,
	)

	switch {
	case errors.Is(err, components.ErrComponentSetFilterWhitespaceGroupField):
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"Double-check provided component group name or ID value (whitespace only value received).%s",
			nagios.CheckOutputEOL,
		)

	case errors.Is(err, components.ErrComponentSetFilterWhitespaceComponentsField):
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"Double-check provided component name or ID values (whitespace only value received).%s",
			nagios.CheckOutputEOL,
		)

	case errors.Is(err, components.ErrComponentGroupNotFound):
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"Double-check provided component group name or ID values (provided value not found).%s",
			nagios.CheckOutputEOL,
		)

	case errors.Is(err, components.ErrComponentIsNotValidSubcomponent):
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"Double-check provided component group and subcomponent name or ID values "+
				"(mismatch between group/subcomponent values).%s",
			nagios.CheckOutputEOL,
		)

	case errors.Is(err, components.ErrComponentNotFound):
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"Double-check provided component name or ID values (provided value not found).%s",
			nagios.CheckOutputEOL,
		)

	// NOTE: While this plugin supports evaluating all components (and
	// therefore results in an empty filter), this error is only
	// returned if filtering is enabled, but an empty filter provided
	// for the filtering stage. While unlikely to occur, we can offer
	// some useful feedback to the user to assist with that scenario.
	case errors.Is(err, components.ErrComponentSetFilterEmpty):
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"While both component group and components list are optional, "+
				"one is required unless evaluating all components.%s",
			nagios.CheckOutputEOL,
		)

		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"If you wish to evaluate all components, use the %s flag and omit filtering options.%s",
			config.EvalAllComponentsFlagLong,
			nagios.CheckOutputEOL,
		)

	default:
		_, _ = fmt.Fprintf(
			&tryAgainMsg,
			"%sPlease recheck provided filter values.%s",
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)

	}

	_, _ = fmt.Fprintf(
		&tryAgainMsg,
		"%sIf in doubt, please use the %s tool to view all provided components of the %q feed (%s).%s",
		nagios.CheckOutputEOL,
		config.InspectorComponentsAppName,
		cs.Page.Name,
		feedSrc,
		nagios.CheckOutputEOL,
	)

	return tryAgainMsg.String()

}
