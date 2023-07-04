// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-vmware
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package main

import (
	"github.com/rs/zerolog"

	"github.com/atc0005/check-statuspage/internal/reports"
	"github.com/atc0005/check-statuspage/internal/statuspage"
	"github.com/atc0005/check-statuspage/internal/statuspage/components"
)

func handleLibraryLogging() {
	switch {
	case zerolog.GlobalLevel() == zerolog.DebugLevel ||
		zerolog.GlobalLevel() == zerolog.TraceLevel:

		statuspage.EnableLogging()
		components.EnableLogging()
		reports.EnableLogging()

	default:

		statuspage.DisableLogging()
		components.DisableLogging()
		reports.DisableLogging()
	}
}
