// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/atc0005/go-nagios"

	"github.com/atc0005/check-statuspage/internal/config"
	"github.com/atc0005/check-statuspage/internal/reports"
	"github.com/atc0005/check-statuspage/internal/statuspage"
	"github.com/atc0005/check-statuspage/internal/statuspage/components"

	zlog "github.com/rs/zerolog/log"
)

func main() {

	// Start the timer. We'll use this to emit the plugin runtime as a
	// performance data metric.
	pluginStart := time.Now()

	// Set initial "state" as valid, adjust as we go.
	var nagiosExitState = nagios.ExitState{
		LastError:      nil,
		ExitStatusCode: nagios.StateOKExitCode,
	}

	// defer this from the start so it is the last deferred function to run
	defer nagiosExitState.ReturnCheckResults()

	// Collect last minute details just before ending plugin execution.
	defer func(exitState *nagios.ExitState, start time.Time) {

		// Record plugin runtime, emit this metric regardless of exit
		// point/cause.
		runtimeMetric := nagios.PerformanceData{
			Label: "time",
			Value: fmt.Sprintf("%dms", time.Since(start).Milliseconds()),
		}
		if err := exitState.AddPerfData(false, runtimeMetric); err != nil {
			zlog.Error().
				Err(err).
				Msg("failed to add time (runtime) performance data metric")
		}

	}(&nagiosExitState, pluginStart)

	// Setup configuration by parsing user-provided flags. Note plugin type so
	// that only applicable CLI flags are exposed and any plugin-specific
	// settings are applied.
	cfg, cfgErr := config.New(config.AppType{PluginComponents: true})
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
		nagiosExitState.ServiceOutput = fmt.Sprintf(
			"%s: Error initializing application",
			nagios.StateCRITICALLabel,
		)
		nagiosExitState.LastError = cfgErr
		nagiosExitState.ExitStatusCode = nagios.StateCRITICALExitCode

		return
	}

	// Enable library-level logging if debug logging level is enabled
	// app-wide. Otherwise, explicitly disable library logging output.
	switch {
	case cfg.LoggingLevel == config.LogLevelDebug:
		statuspage.EnableLogging()
		components.EnableLogging()
		reports.EnableLogging()

	default:
		statuspage.DisableLogging()
		components.DisableLogging()
		reports.DisableLogging()
	}

	// Set context deadline equal to user-specified timeout value for plugin
	// runtime/execution.
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout())
	defer cancel()

	// Record thresholds for use as Nagios "Long Service Output" content. This
	// content is shown in the detailed web UI and in notifications generated
	// by Nagios.
	criticalComponentStatuses := components.ServiceStateToComponentStatuses(
		nagios.ServiceState{ExitCode: nagios.StateCRITICALExitCode},
	)
	nagiosExitState.CriticalThreshold = strings.Join(criticalComponentStatuses, ", ")

	warningComponentStatuses := components.ServiceStateToComponentStatuses(
		nagios.ServiceState{ExitCode: nagios.StateWARNINGExitCode},
	)

	nagiosExitState.WarningThreshold = strings.Join(warningComponentStatuses, ", ")

	if cfg.EmitBranding {
		// If enabled, show application details at end of notification
		nagiosExitState.BrandingCallback = config.Branding("Notification generated by ")
	}

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

			nagiosExitState.LastError = err
			nagiosExitState.ServiceOutput = fmt.Sprintf(
				"%s: Failed to decode JSON feed from %q",
				nagios.StateCRITICALLabel,
				cfg.Filename,
			)
			nagiosExitState.ExitStatusCode = nagios.StateCRITICALExitCode

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

			nagiosExitState.LastError = err
			nagiosExitState.ServiceOutput = fmt.Sprintf(
				"%s: Failed to decode JSON feed from %q",
				nagios.StateCRITICALLabel,
				cfg.Filename,
			)
			nagiosExitState.ExitStatusCode = nagios.StateCRITICALExitCode

			return
		}
		log.Debug().Msg("Successfully decoded JSON input")

	}

	if err := componentsSet.Validate(); err != nil {

		log.Error().Msg("Failed to validate JSON feed")

		nagiosExitState.LastError = err
		nagiosExitState.ServiceOutput = fmt.Sprintf(
			"%s: Error validating JSON feed from %q",
			nagios.StateCRITICALLabel,
			feedSource,
		)
		nagiosExitState.ExitStatusCode = nagios.StateCRITICALExitCode

		return

	}

	csFilter := components.Filter(cfg.ComponentFilter())

	switch {
	case cfg.EvalAllComponents:

		log.Debug().Msg("Option to evaluate all components chosen")
		componentsSet.EvalAllComponents = true

	default:

		log.Debug().Msg("Option to evaluate all components not chosen")
		log.Debug().
			Str("group", csFilter.Group).
			Str("components", strings.Join(csFilter.Components, ", ")).
			Msg("Applying user specified components filter to components set")

		if err := componentsSet.Filter(csFilter); err != nil {
			log.Error().
				Err(err).
				Msg("Error applying search terms as filter to components set")

			nagiosExitState.LastError = err
			nagiosExitState.ServiceOutput = fmt.Sprintf(
				"%s: Error filtering components set using specified search terms",
				nagios.StateUNKNOWNLabel,
			)

			nagiosExitState.LongServiceOutput = filterErrAdvice(
				err,
				componentsSet,
				csFilter,
				feedSource,
			)

			nagiosExitState.ExitStatusCode = nagios.StateUNKNOWNExitCode

			return
		}

	}

	// Global stats
	numTotalComponents := componentsSet.NumComponents()
	numTotalComponentGroups := componentsSet.NumGroups()
	numComponentsCritical := componentsSet.NumCriticalState(true)
	numComponentsWarning := componentsSet.NumWarningState(true)
	numComponentsUnknown := componentsSet.NumUnknownState(true)
	numComponentsOK := componentsSet.NumOKState(true)

	// Stats specific to the components being monitored
	numComponentsRemainingCritical := componentsSet.NumCriticalState(false)
	numComponentsRemainingWarning := componentsSet.NumWarningState(false)
	numComponentsRemainingUnknown := componentsSet.NumUnknownState(false)
	numComponentsRemainingOK := componentsSet.NumOKState(false)

	numProblemComponents := componentsSet.NumProblemComponents(true)
	numExcludedComponents := componentsSet.NumExcluded()
	numRemainingProblemComponents := componentsSet.NumProblemComponents(false)
	numExcludedProblemComponents := numProblemComponents - numRemainingProblemComponents

	pd := []nagios.PerformanceData{
		// The `time` (runtime) metric is appended at plugin exit, so do not
		// duplicate it here.
		{
			Label: "all_components",
			Value: fmt.Sprintf("%d", numTotalComponents),
		},
		{
			Label: "all_component_groups",
			Value: fmt.Sprintf("%d", numTotalComponentGroups),
		},
		{
			Label: "all_problem_components",
			Value: fmt.Sprintf("%d", numProblemComponents),
		},
		{
			Label: "excluded_problem_components",
			Value: fmt.Sprintf("%d", numExcludedProblemComponents),
		},
		{
			Label: "remaining_problem_components",
			Value: fmt.Sprintf("%d", numRemainingProblemComponents),
		},
		{
			Label: "all_components_critical",
			Value: fmt.Sprintf("%d", numComponentsCritical),
		},
		{
			Label: "all_components_warning",
			Value: fmt.Sprintf("%d", numComponentsWarning),
		},
		{
			Label: "all_components_unknown",
			Value: fmt.Sprintf("%d", numComponentsUnknown),
		},
		{
			Label: "all_components_ok",
			Value: fmt.Sprintf("%d", numComponentsOK),
		},
		{
			Label: "remaining_components_critical",
			Value: fmt.Sprintf("%d", numComponentsRemainingCritical),
		},
		{
			Label: "remaining_components_warning",
			Value: fmt.Sprintf("%d", numComponentsRemainingWarning),
		},
		{
			Label: "remaining_components_unknown",
			Value: fmt.Sprintf("%d", numComponentsRemainingUnknown),
		},
		{
			Label: "remaining_components_ok",
			Value: fmt.Sprintf("%d", numComponentsRemainingOK),
		},
	}

	// Update logger with new performance data related fields
	log = log.With().
		Int("total_problem_components", numProblemComponents).
		Int("excluded_components", numExcludedComponents).
		Int("excluded_problem_components", numExcludedProblemComponents).
		Int("remaining_problem_components", numRemainingProblemComponents).
		Logger()

	switch {
	case !componentsSet.IsOKState(false):

		log.Error().
			Msg("Non-excluded, non-operational status of statuspage components detected")

		// Set state label and exit code based on most severe
		// status found in the (filtered) collection.
		var stateLabel string
		switch {
		case componentsSet.HasCriticalState(false):
			stateLabel = nagios.StateCRITICALLabel
			nagiosExitState.ExitStatusCode = nagios.StateCRITICALExitCode
			nagiosExitState.LastError = components.ErrComponentWithProblemStatusNotExcluded

		case componentsSet.HasWarningState(false):
			stateLabel = nagios.StateWARNINGLabel
			nagiosExitState.ExitStatusCode = nagios.StateWARNINGExitCode
			nagiosExitState.LastError = components.ErrComponentWithProblemStatusNotExcluded

		case componentsSet.HasUnknownState(false):
			stateLabel = nagios.StateUNKNOWNLabel
			nagiosExitState.ExitStatusCode = nagios.StateUNKNOWNExitCode
			nagiosExitState.LastError = components.ErrComponentWithProblemStatusNotExcluded
		}

		nagiosExitState.ServiceOutput = reports.ComponentsOneLineCheckSummary(
			stateLabel,
			componentsSet,
			false,
		)

		nagiosExitState.LongServiceOutput = reports.ComponentsReport(
			stateLabel,
			csFilter,
			componentsSet,
			cfg.OmitOKComponents,
		)

		if err := nagiosExitState.AddPerfData(false, pd...); err != nil {
			log.Error().
				Err(err).
				Msg("failed to add performance data")
		}

		return

	default:

		// success path

		log.Debug().Msg("Evaluated components are in an operational state")

		nagiosExitState.LastError = nil
		nagiosExitState.ExitStatusCode = nagios.StateOKExitCode

		nagiosExitState.ServiceOutput = reports.ComponentsOneLineCheckSummary(
			nagios.StateOKLabel,
			componentsSet,
			false,
		)

		nagiosExitState.LongServiceOutput = reports.ComponentsReport(
			nagios.StateOKLabel,
			csFilter,
			componentsSet,
			cfg.OmitOKComponents,
		)

		if err := nagiosExitState.AddPerfData(false, pd...); err != nil {
			log.Error().
				Err(err).
				Msg("failed to add performance data")
		}

		return

	}

}
