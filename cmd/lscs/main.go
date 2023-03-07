// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

//go:generate go-winres make --product-version=git-tag --file-version=git-tag

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	// "github.com/kr/pretty"
	// "github.com/hexops/valast"
	"github.com/sanity-io/litter"

	"github.com/atc0005/check-statuspage/internal/config"
	"github.com/atc0005/check-statuspage/internal/reports"
	"github.com/atc0005/check-statuspage/internal/statuspage"
	"github.com/atc0005/check-statuspage/internal/statuspage/components"

	zlog "github.com/rs/zerolog/log"
)

func main() {

	// Setup configuration by parsing user-provided flags. Note plugin type so
	// that only applicable CLI flags are exposed and any plugin-specific
	// settings are applied.
	cfg, cfgErr := config.New(config.AppType{InspectorComponents: true})
	switch {
	case errors.Is(cfgErr, config.ErrVersionRequested):
		fmt.Println(config.Version())

		return

	case errors.Is(cfgErr, config.ErrHelpRequested):
		fmt.Println(cfg.Help())

		return

	case cfgErr != nil:
		// We're using the standalone Err function from rs/zerolog/log as we
		// do not have a working configuration.
		zlog.Err(cfgErr).Msg("Error initializing application")

		return
	}

	// Enable library-level logging if debug logging level is enabled
	// app-wide. Otherwise, explicitly disable library logging output.
	switch {
	case cfg.LoggingLevel == config.LogLevelDebug:
		statuspage.EnableLogging()

	default:
		statuspage.DisableLogging()
	}

	// Set context deadline equal to user-specified timeout value for
	// runtime/execution.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
	defer cancel()

	log := cfg.Log.With().
		Str("filename", cfg.Filename).
		Str("url", cfg.URL).
		Int64("read_limit", cfg.ReadLimit).
		Bool("allow_unknown_fields", cfg.AllowUnknownJSONFields).
		Logger()

	// Process one of local file or remote URL. Rely on config package
	// validation to prevent the user from specifying both.
	var componentsSet *components.Set
	var feedSource string
	switch {

	case cfg.Filename != "":

		feedSource = cfg.Filename

		log.Debug().Msg("Decoding JSON input")

		var err error
		componentsSet, err = components.NewFromFile(cfg.Filename, cfg.ReadLimit, cfg.AllowUnknownJSONFields)
		if err != nil {
			log.Error().Err(err).Msg("Error decoding JSON feed")

			return
		}
		log.Debug().Msg("Successfully decoded JSON input")

	case cfg.URL != "":

		feedSource = cfg.URL

		log.Debug().Msg("Decoding JSON input")

		var err error
		componentsSet, err = components.NewFromURL(
			ctx,
			cfg.URL,
			cfg.ReadLimit,
			cfg.AllowUnknownJSONFields,
			cfg.UserAgent(),
		)
		if err != nil {
			log.Error().Err(err).Msg("Error decoding JSON feed")

			return
		}
		log.Debug().Msg("Successfully decoded JSON input")

	}

	if err := componentsSet.Validate(); err != nil {
		log.Error().
			Err(err).
			Str("feed_source", feedSource).
			Msg("Failed to validate JSON feed")

		return
	}

	switch cfg.InspectorOutputFormat {

	case config.InspectorOutputFormatOverview:
		fmt.Print(reports.ComponentsOverview(componentsSet, cfg.OmitOKComponents))

	case config.InspectorOutputFormatTable:
		// Enable all columns in filter
		columnFilter := reports.ComponentsTableColumnFilter{
			GroupName:     true,
			GroupID:       true,
			ComponentName: true,
			ComponentID:   true,
			Evaluated:     true,
			Status:        true,
		}

		// Generate table, providing our "use everything" filter.
		fmt.Print(reports.ComponentsTable(componentsSet, cfg.OmitOKComponents, cfg.OmitSummaryResults, &columnFilter))

	case config.InspectorOutputFormatVerbose:
		fmt.Print(reports.ComponentsVerbose(componentsSet, cfg.OmitOKComponents))

	case config.InspectorOutputFormatDebug:
		// fmt.Printf("%# v", pretty.Formatter(componentsSet))
		// fmt.Println(valast.String(componentsSet))
		litter.Dump(componentsSet)

	case config.InspectorOutputFormatJSON:
		s, err := json.MarshalIndent(componentsSet, "", "\t")
		if err != nil {
			log.Error().Err(err).
				Str("feed_source", feedSource).
				Msg("Failed to marshal parsed components as JSON output")

			return
		}
		fmt.Print(string(s))

	case config.InspectorOutputFormatIDsList:
		fmt.Print(reports.ComponentsIDList(componentsSet))

	default:
		fmt.Printf(
			"unknown output format chosen: %q\n",
			cfg.InspectorOutputFormat,
		)

		return

	}

}
