// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package config

import (
	"flag"
	"os"
)

// handleFlagsConfig handles toggling the exposure of specific configuration
// flags to the user. This behavior is controlled via the specified
// application type as set by each cmd. Based on the application's specified
// type, a smaller subset of flags specific to each type are exposed along
// with a set common to all application types.
func (c *Config) handleFlagsConfig(flagSet *flag.FlagSet, appType AppType) error {

	// Flags specific to one plugin type or the other
	switch {
	case appType.PluginComponents:

		flagSet.BoolVar(&c.EmitBranding, BrandingFlag, defaultBranding, brandingFlagHelp)

		flagSet.Var(&c.componentsList, ComponentsListFlagShort, componentsListFlagHelp+" (shorthand)")
		flagSet.Var(&c.componentsList, ComponentsListFlagLong, componentsListFlagHelp)

		flagSet.StringVar(&c.componentGroup, ComponentGroupFlagShort, defaultComponentGroup, componentGroupFlagHelp+" (shorthand)")
		flagSet.StringVar(&c.componentGroup, ComponentGroupFlagLong, defaultComponentGroup, componentGroupFlagHelp)

		flagSet.BoolVar(&c.EvalAllComponents, EvalAllComponentsFlagShort, defaultEvalAllComponents, evalAllComponentsFlagHelp+" (shorthand)")
		flagSet.BoolVar(&c.EvalAllComponents, EvalAllComponentsFlagLong, defaultEvalAllComponents, evalAllComponentsFlagHelp)

	case appType.InspectorComponents:

		flagSet.StringVar(&c.InspectorOutputFormat, InspectorOutputFormatFlagShort, defaultInspectorOutputFormat, inspectorOutputFormatFlagHelp+" (shorthand)")
		flagSet.StringVar(&c.InspectorOutputFormat, InspectorOutputFormatFlagLong, defaultInspectorOutputFormat, inspectorOutputFormatFlagHelp)

	}

	// Shared flags for all application types

	flagSet.BoolVar(&c.OmitOKComponents, OmitOKComponentsFlagShort, defaultOmitOKComponents, omitOKComponentsFlagHelp+" (shorthand)")
	flagSet.BoolVar(&c.OmitOKComponents, OmitOKComponentsFlagLong, defaultOmitOKComponents, omitOKComponentsFlagHelp)

	flagSet.StringVar(&c.URL, URLFlagShort, defaultURL, urlFlagHelp+" (shorthand)")
	flagSet.StringVar(&c.URL, URLFlagLong, defaultURL, urlFlagHelp)

	flagSet.StringVar(&c.Filename, FilenameFlagShort, defaultFilename, filenameFlagHelp+" (shorthand)")
	flagSet.StringVar(&c.Filename, FilenameFlagLong, defaultFilename, filenameFlagHelp)

	flagSet.BoolVar(&c.AllowUnknownJSONFields, AllowUnknownJSONFieldsFlagShort, defaultAllowUnknownJSONFields, allowUnknownJSONFieldsFlagHelp+" (shorthand)")
	flagSet.BoolVar(&c.AllowUnknownJSONFields, AllowUnknownJSONFieldsFlagLong, defaultAllowUnknownJSONFields, allowUnknownJSONFieldsFlagHelp)

	flagSet.Int64Var(&c.ReadLimit, ReadLimitFlagShort, defaultReadLimit, readLimitFlagHelp+" (shorthand)")
	flagSet.Int64Var(&c.ReadLimit, ReadLimitFlagLong, defaultReadLimit, readLimitFlagHelp)

	flagSet.IntVar(&c.timeout, TimeoutFlagShort, defaultRuntimeTimeout, timeoutRuntimeFlagHelp+" (shorthand)")
	flagSet.IntVar(&c.timeout, TimeoutFlagLong, defaultRuntimeTimeout, timeoutRuntimeFlagHelp)

	flagSet.StringVar(&c.LoggingLevel, LogLevelFlagShort, defaultLogLevel, logLevelFlagHelp+" (shorthand)")
	flagSet.StringVar(&c.LoggingLevel, LogLevelFlagLong, defaultLogLevel, logLevelFlagHelp)

	flagSet.BoolVar(&c.ShowVersion, VersionFlagShort, defaultDisplayVersionAndExit, versionFlagHelp+" (shorthand)")
	flagSet.BoolVar(&c.ShowVersion, VersionFlagLong, defaultDisplayVersionAndExit, versionFlagHelp)

	// Allow our function to override the default Help output
	flagSet.Usage = Usage(flagSet)

	// parse flag definitions from the argument list
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		return err
	}

	return nil

}
