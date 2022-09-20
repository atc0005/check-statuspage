// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package config

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rs/zerolog"
)

// Updated via Makefile builds. Setting placeholder value here so that
// something resembling a version string will be provided for non-Makefile
// builds.
var version = "x.y.z"

// ErrVersionRequested indicates that the user requested application version
// information.
var ErrVersionRequested = errors.New("version information requested")

// ErrHelpRequested indicates that the user requested application help/usage
// information.
var ErrHelpRequested = errors.New("help/usage information requested")

// ErrConfigNotInitialized indicates that the configuration is not in a usable
// state and application execution can not successfully proceed.
var ErrConfigNotInitialized = errors.New("configuration not initialized")

// AppType represents the type of application that is being
// configured/initialized. Not all application types will use the same
// features and as a result will not accept the same flags. Unless noted
// otherwise, each of the application types are incompatible with each other,
// though some flags are common to all types.
type AppType struct {

	// PluginComponents represents an application used as a monitoring plugin
	// for evaluating Statuspage components.
	PluginComponents bool

	// InspectorComponents represents an application used for one-off or
	// isolated checks against Statuspage components. Unlike a Nagios plugin
	// which is focused on specific attributes resulting in a severity-based
	// outcome, an inspecter application is intended for examining targets for
	// informational/troubleshooting purposes.
	InspectorComponents bool
}

// AppInfo identifies common details about the plugins provided by this
// project.
type AppInfo struct {

	// Name specifies the public name shared by all plugins in this project.
	Name string

	// Version specifies the public version shared by all plugins in this
	// project.
	Version string

	// URL specifies the public repo URL shared by all plugins in this
	// project.
	URL string

	// Plugin indicates which plugin provided by this project is currently
	// executing.
	Plugin string
}

// multiValueStringFlag is a custom type that satisfies the flag.Value
// interface in order to accept multiple string values for some of our flags.
type multiValueStringFlag []string

// String returns a comma separated string consisting of all slice elements.
func (mvs *multiValueStringFlag) String() string {

	// The String() method is called by the flag.isZeroValue function in order
	// to determine whether the output string represents the zero value for a
	// flag. This occurs even if the flag is not specified by the user.

	// From the `flag` package docs:
	// "The flag package may call the String method with a zero-valued
	// receiver, such as a nil pointer."
	if mvs == nil {
		return ""
	}

	return strings.Join(*mvs, ", ")
}

// Set is called once by the flag package, in command line order, for each
// flag present.
func (mvs *multiValueStringFlag) Set(value string) error {

	// split comma-separated string into multiple values, toss leading and
	// trailing whitespace
	items := strings.Split(value, ",")
	for index, item := range items {
		items[index] = strings.TrimSpace(item)
		items[index] = strings.ReplaceAll(items[index], "'", "")
		items[index] = strings.ReplaceAll(items[index], "\"", "")
	}

	// add them to the collection
	*mvs = append(*mvs, items...)

	return nil
}

// ComponentFilter is a custom type that reflects a user specified filter for
// component group and components. If specified together, components must be
// subcomponents of the specified group.
//
// This type satisfies the flag.Value interface in order to accept comma
// separated component group / subcomponent pairings.
//
// NOTE: While the original design required a flag associated with this type
// in order to specify pairings of component group and subcomponents, the
// current design allows for specifying this pairing by way of separate
// component and group flags.
type ComponentFilter struct {

	// Group is an optional component group (aka, "parent" or "container")
	// associated with individual subcomponents.
	Group string

	// Components is a collection of components. If specified with a component
	// group, then all specified components must be subcomponents of the
	// specified component group.
	Components []string
}

// String implements the Stringer interface, providing a human-readable
// version of user specified component group set values.
func (cgs *ComponentFilter) String() string {

	// The String() method is called by the flag.isZeroValue function in order
	// to determine whether the output string represents the zero value for a
	// flag. This occurs even if the flag is not specified by the user.

	// From the `flag` package docs:
	// "The flag package may call the String method with a zero-valued
	// receiver, such as a nil pointer."
	if cgs == nil {
		return ""
	}

	return fmt.Sprintf(
		`{Group: "%s", Components: "%s"}`,
		cgs.Group,
		strings.Join(cgs.Components, ", "),
	)
}

// Set is called once by the flag package, in command line order, for each
// flag present.
//
// We require a minimum of values. The first is the component group, the
// second is a subcomponent of that group. There can be many subcomponents in
// a single component group. Alternatively, a special keyword "ALL" is
// recognized in place of a valid subcomponent name. Validation of group and
// subcomponent names is handled by a later step.
func (cgs *ComponentFilter) Set(value string) error {

	const minNumValues int = 2

	// Split comma-separated string into multiple values, toss whitespace.
	items := strings.Split(value, ",")

	if len(items) < minNumValues {
		return fmt.Errorf(
			"error processing flag; string %q provides %d values, expected minimum %d values",
			value,
			len(items),
			minNumValues,
		)
	}

	for i := range items {
		items[i] = strings.TrimSpace(items[i])
		items[i] = strings.ReplaceAll(items[i], "'", "")
		items[i] = strings.ReplaceAll(items[i], "\"", "")
	}

	// Set group using first value
	cgs.Group = items[0]

	// Expand components list using remaining values
	cgs.Components = append(cgs.Components, items[1:]...)

	return nil

}

// Config represents the application configuration as specified via
// command-line flags.
type Config struct {

	// URL is the fully-qualified URL of a Statuspage API/JSON feed.
	URL string

	// Filename is the fully-qualified filename of a previously downloaded
	// Statuspage API/JSON feed.
	Filename string

	// LoggingLevel is the supported logging level for this application.
	LoggingLevel string

	// InspectorOutputFormat is the output format used for Inspector type
	// applications.
	InspectorOutputFormat string

	// App represents common details about the plugins provided by this
	// project.
	App AppInfo

	// flagSet provides a useful hook to allow evaluating defined flags
	// against a list of expected flags. This field is exported so that the
	// flagset is accessible to tests from within this package and from
	// outside of the config package.
	flagSet *flag.FlagSet

	// componentGroup is a component group specified by the user. This field
	// is set when the user opts to not specify sets, but rather a single
	// component group for evaluation.
	componentGroup string

	// componentGroupSet is the set of component group and subcomponents for
	// that group specified by the user. This field is set when the user opts
	// to not specify individual components.
	// componentGroupSet ComponentFilter

	// componentsList is a collection of individual components specified by
	// the user. This field is set when the user opts to not specify sets.
	componentsList multiValueStringFlag

	// Log is an embedded zerolog Logger initialized via config.New().
	Log zerolog.Logger

	// ReadLimit is a limit in bytes set to help prevent abuse when reading
	// input that could be larger than expected. The default value is overly
	// generous and is unlikely to be met unless something is broken.
	ReadLimit int64

	// timeout is the value in seconds allowed before an execution attempt is
	// abandoned and an error returned.
	timeout int

	// EmitBranding controls whether "generated by" text is included at the
	// bottom of application output. This output is included in the Nagios
	// dashboard and notifications. This output may not mix well with branding
	// output from other tools such as atc0005/send2teams which also insert
	// their own branding output.
	EmitBranding bool

	// AllowUnknownJSONFields controls whether the JSON decoder aborts when
	// encountering fields not defined in a destination struct type.
	AllowUnknownJSONFields bool

	// ShowVersion indicates whether the user opted to display only the
	// version string and then immediately exit the application.
	ShowVersion bool

	// ShowHelp indicates whether the user opted to display usage information
	// and exit the application.
	ShowHelp bool

	// OmitOKComponents indicates whether the user opted to omit components in
	// an OK or operational status from results output. This setting does not
	// apply to all output formats.
	OmitOKComponents bool

	// OmitSummaryResults indicates whether the user opted to omit Summary
	// results at the end of plugin execution.
	OmitSummaryResults bool

	// EvalAllComponents indicates whether the user opted to evaluate all
	// components instead of specific components.
	EvalAllComponents bool
}

// Usage is a custom override for the default Help text provided by the flag
// package. Here we prepend some additional metadata to the existing output.
func Usage(flagSet *flag.FlagSet, w io.Writer) func() {

	// Make one attempt to override output so that calling Config.Help() later
	// will have a chance to also override the output destination.
	flag.CommandLine.SetOutput(w)

	switch {

	// Unintialized flagset, provide stub usage information.
	case flagSet == nil:
		return func() {
			fmt.Fprintln(w, "Failed to initialize configuration; nil FlagSet")
		}

	// Non-nil flagSet, proceed
	default:

		// Make one attempt to override output so that calling Config.Help()
		// later will have a chance to also override the output destination.
		flagSet.SetOutput(w)

		return func() {
			fmt.Fprintln(flag.CommandLine.Output(), "\n"+Version()+"\n")
			fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
			flagSet.PrintDefaults()
		}
	}
}

// Help emits application usage information to the previously configured
// destination for usage and error messages.
func (c *Config) Help() string {

	var helpTxt strings.Builder

	// Override previously specified output destination, redirect to Builder.
	flag.CommandLine.SetOutput(&helpTxt)

	switch {

	// Handle nil configuration initialization.
	case c == nil || c.flagSet == nil:
		// Fallback message noting the issue.
		fmt.Fprintln(&helpTxt, ErrConfigNotInitialized)

	default:
		// Emit expected help output to builder.
		c.flagSet.SetOutput(&helpTxt)
		c.flagSet.Usage()

	}

	return helpTxt.String()
}

// Version emits application name, version and repo location.
func Version() string {
	return fmt.Sprintf("%s %s (%s)", myAppName, version, myAppURL)
}

// Branding accepts a message and returns a function that concatenates that
// message with version information. This function is intended to be called as
// a final step before application exit after any other output has already
// been emitted.
func Branding(msg string) func() string {
	return func() string {
		return strings.Join([]string{msg, Version()}, "")
	}
}

// appTypeLabel is used as a lookup to return the a application label
// associated with the active/specified AppType.
func appTypeLabel(appType AppType) string {

	var label string

	switch {
	case appType.PluginComponents:
		label = PluginComponentsAppType

	case appType.InspectorComponents:
		label = InspectorComponentsAppType

	default:
		label = "ERROR: Please report this; AppType collection is missing an entry"

	}

	return label

}

// New is a factory function that produces a new Config object based on user
// provided flag and config file values. It is responsible for validating
// user-provided values and initializing the logging settings used by this
// application.
func New(appType AppType) (*Config, error) {
	var config Config

	// NOTE: Need to make sure we allow execution to continue on encountered
	// errors. This is so that we can check for those errors as return values
	// both within the main apps and tests for this package.
	config.flagSet = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	if err := config.handleFlagsConfig(appType); err != nil {
		return nil, fmt.Errorf(
			"failed to set flags configuration: %w",
			err,
		)
	}

	config.App = AppInfo{
		Name:    myAppName,
		Version: version,
		URL:     myAppURL,
		Plugin:  appTypeLabel(appType),
	}

	switch {

	// The configuration was successfully initialized, so we're good with
	// returning it for use by the caller.
	case config.ShowVersion:
		return &config, ErrVersionRequested

	// The configuration was successfully initialized, so we're good with
	// returning it for use by the caller.
	case config.ShowHelp:
		return &config, ErrHelpRequested
	}

	if err := config.validate(appType); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// initialize logging just as soon as validation is complete
	if err := config.setupLogging(appType); err != nil {
		return nil, fmt.Errorf(
			"failed to set logging configuration: %w",
			err,
		)
	}

	return &config, nil

}
