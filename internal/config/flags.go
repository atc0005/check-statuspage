// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package config

import (
	"fmt"
	"os"
)

// handleFlagsConfig handles toggling the exposure of specific configuration
// flags to the user. This behavior is controlled via the specified
// application type as set by each cmd. Based on the application's specified
// type, a smaller subset of flags specific to each type are exposed along
// with a set common to all application types.
func (c *Config) handleFlagsConfig(appType AppType) error {

	if c == nil {
		return fmt.Errorf(
			"nil configuration, cannot process flags: %w",
			ErrConfigNotInitialized,
		)
	}

	// Flags specific to one plugin type or the other
	switch {
	case appType.PluginComponents:

		c.flagSet.BoolVar(&c.EmitBranding, BrandingFlag, defaultBranding, brandingFlagHelp)

		c.flagSet.BoolVar(&c.ShowVerbose, VerboseFlag, defaultVerbose, verboseFlagHelp)

		c.flagSet.Var(&c.componentsList, ComponentsListFlagShort, componentsListFlagHelp+shorthandFlagSuffix)
		c.flagSet.Var(&c.componentsList, ComponentsListFlagLong, componentsListFlagHelp)

		c.flagSet.StringVar(&c.componentGroup, ComponentGroupFlagShort, defaultComponentGroup, componentGroupFlagHelp+shorthandFlagSuffix)
		c.flagSet.StringVar(&c.componentGroup, ComponentGroupFlagLong, defaultComponentGroup, componentGroupFlagHelp)

		c.flagSet.BoolVar(&c.EvalAllComponents, EvalAllComponentsFlagShort, defaultEvalAllComponents, evalAllComponentsFlagHelp+shorthandFlagSuffix)
		c.flagSet.BoolVar(&c.EvalAllComponents, EvalAllComponentsFlagLong, defaultEvalAllComponents, evalAllComponentsFlagHelp)

	case appType.InspectorComponents:

		c.flagSet.StringVar(&c.InspectorOutputFormat, InspectorOutputFormatFlagShort, defaultInspectorOutputFormat, inspectorOutputFormatFlagHelp+shorthandFlagSuffix)
		c.flagSet.StringVar(&c.InspectorOutputFormat, InspectorOutputFormatFlagLong, defaultInspectorOutputFormat, inspectorOutputFormatFlagHelp)

	}

	// Shared flags for all application types

	c.flagSet.BoolVar(&c.ShowHelp, HelpFlagShort, defaultHelp, helpFlagHelp+shorthandFlagSuffix)
	c.flagSet.BoolVar(&c.ShowHelp, HelpFlagLong, defaultHelp, helpFlagHelp)

	c.flagSet.BoolVar(&c.OmitOKComponents, OmitOKComponentsFlagShort, defaultOmitOKComponents, omitOKComponentsFlagHelp+shorthandFlagSuffix)
	c.flagSet.BoolVar(&c.OmitOKComponents, OmitOKComponentsFlagLong, defaultOmitOKComponents, omitOKComponentsFlagHelp)

	c.flagSet.StringVar(&c.URL, URLFlagShort, defaultURL, urlFlagHelp+shorthandFlagSuffix)
	c.flagSet.StringVar(&c.URL, URLFlagLong, defaultURL, urlFlagHelp)

	c.flagSet.StringVar(&c.Filename, FilenameFlagShort, defaultFilename, filenameFlagHelp+shorthandFlagSuffix)
	c.flagSet.StringVar(&c.Filename, FilenameFlagLong, defaultFilename, filenameFlagHelp)

	c.flagSet.BoolVar(&c.AllowUnknownJSONFields, AllowUnknownJSONFieldsFlagShort, defaultAllowUnknownJSONFields, allowUnknownJSONFieldsFlagHelp+shorthandFlagSuffix)
	c.flagSet.BoolVar(&c.AllowUnknownJSONFields, AllowUnknownJSONFieldsFlagLong, defaultAllowUnknownJSONFields, allowUnknownJSONFieldsFlagHelp)

	c.flagSet.Int64Var(&c.ReadLimit, ReadLimitFlagShort, defaultReadLimit, readLimitFlagHelp+shorthandFlagSuffix)
	c.flagSet.Int64Var(&c.ReadLimit, ReadLimitFlagLong, defaultReadLimit, readLimitFlagHelp)

	c.flagSet.IntVar(&c.timeout, TimeoutFlagShort, defaultRuntimeTimeout, timeoutRuntimeFlagHelp+shorthandFlagSuffix)
	c.flagSet.IntVar(&c.timeout, TimeoutFlagLong, defaultRuntimeTimeout, timeoutRuntimeFlagHelp)

	c.flagSet.StringVar(&c.LoggingLevel, LogLevelFlagShort, defaultLogLevel, logLevelFlagHelp+shorthandFlagSuffix)
	c.flagSet.StringVar(&c.LoggingLevel, LogLevelFlagLong, defaultLogLevel, logLevelFlagHelp)

	c.flagSet.BoolVar(&c.ShowVersion, VersionFlagShort, defaultDisplayVersionAndExit, versionFlagHelp+shorthandFlagSuffix)
	c.flagSet.BoolVar(&c.ShowVersion, VersionFlagLong, defaultDisplayVersionAndExit, versionFlagHelp)

	c.flagSet.BoolVar(&c.OmitSummaryResults, OmitSummaryResultsFlagShort, defaultOmitSummaryResults, omitSummaryResultsFlagHelp+shorthandFlagSuffix)
	c.flagSet.BoolVar(&c.OmitSummaryResults, OmitSummaryResultsFlagLong, defaultOmitSummaryResults, omitSummaryResultsFlagHelp)

	// Allow our function to override the default Help output.
	//
	// Override default of stderr as destination for help output. This allows
	// Nagios XI and similar monitoring systems to call plugins with the
	// `--help` flag and have it display within the Admin web UI.
	c.flagSet.Usage = Usage(c.flagSet, os.Stdout)

	// parse flag definitions from the argument list
	return c.flagSet.Parse(os.Args[1:])
}
