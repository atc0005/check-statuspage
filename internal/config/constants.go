// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package config

const myAppName string = "check-statuspage"
const myAppURL string = "https://github.com/atc0005/" + myAppName

// Flag names. Exported so that they can be referenced from tests.
const (
	HelpFlagLong                    string = "help"
	HelpFlagShort                   string = "h"
	VersionFlagLong                 string = "version"
	VersionFlagShort                string = "v"
	BrandingFlag                    string = "branding"
	ComponentsListFlagLong          string = "component"
	ComponentsListFlagShort         string = "c"
	ComponentGroupFlagLong          string = "group"
	ComponentGroupFlagShort         string = "g"
	EvalAllComponentsFlagLong       string = "eval-all"
	EvalAllComponentsFlagShort      string = "ea"
	InspectorOutputFormatFlagLong   string = "output-format"
	InspectorOutputFormatFlagShort  string = "fmt"
	OmitOKComponentsFlagLong        string = "omit-ok"
	OmitOKComponentsFlagShort       string = "ook"
	OmitSummaryResultsFlagLong      string = "omit-summary"
	OmitSummaryResultsFlagShort     string = "os"
	URLFlagLong                     string = "url"
	URLFlagShort                    string = "u"
	FilenameFlagLong                string = "filename"
	FilenameFlagShort               string = "f"
	AllowUnknownJSONFieldsFlagLong  string = "allow-unknown-fields"
	AllowUnknownJSONFieldsFlagShort string = "auf"
	ReadLimitFlagLong               string = "read-limit"
	ReadLimitFlagShort              string = "rl"
	TimeoutFlagLong                 string = "timeout"
	TimeoutFlagShort                string = "t"
	LogLevelFlagLong                string = "log-level"
	LogLevelFlagShort               string = "ll"
)

// Common or shared flag help text
const (
	helpFlagHelp                   string = "Emit this help text"
	versionFlagHelp                string = "Whether to display application version and then immediately exit application."
	logLevelFlagHelp               string = "Sets log level to one of disabled, panic, fatal, error, warn, info, debug or trace."
	urlFlagHelp                    string = "The fully-qualified URL of a Statuspage API/JSON feed (e.g., https://www.githubstatus.com/api/v2/components.json)."
	filenameFlagHelp               string = "The fully-qualified filename of a previously downloaded Statuspage API/JSON feed (e.g., /tmp/statuspage/github/components.json)."
	timeoutRuntimeFlagHelp         string = "Timeout value in seconds allowed before an execution attempt is abandoned and an error returned."
	readLimitFlagHelp              string = "Limit in bytes used to help prevent abuse when reading input that could be larger than expected. The default value is nearly 4x the largest observed (formatted) feed size."
	allowUnknownJSONFieldsFlagHelp string = "Whether unknown JSON fields encountered while decoding JSON data should be ignored."
	omitOKComponentsFlagHelp       string = "Whether listed components in results output should be limited to just those in a non-operational state. Does not apply to all output formats."
	omitSummaryResultsFlagHelp     string = "Whether summary in results output should be omitted."
)

// Inspector type application flag help text
const (
	inspectorOutputFormatFlagHelp string = "Sets output format to one of overview, table, verbose, debug, list or json."
)

// Plugin type application flag help text
const (
	brandingFlagHelp          string = "Toggles emission of branding details with plugin status details. This output is disabled by default."
	componentsListFlagHelp    string = "One or more comma-separated component (name or ID) values. Can be used by itself or with the flag to specify a component group. If used with the component group flag, all specified components are required to be subcomponents of the group."
	componentGroupFlagHelp    string = "A single name or ID value for a component group. Can be used by itself or with the flag to specify a list of components. If used with the components flag all specified components are required to be subcomponents of the group."
	evalAllComponentsFlagHelp string = "Whether all components should be evaluated. Incompatible with flag to specify list of components, component group or component group set."
)

// Default flag settings if not overridden by user input
const (
	defaultURL                    string = ""
	defaultFilename               string = ""
	defaultLogLevel               string = "info"
	defaultComponentGroup         string = ""
	defaultHelp                   bool   = false
	defaultBranding               bool   = false
	defaultOmitOKComponents       bool   = false
	defaultOmitSummaryResults     bool   = false
	defaultEvalAllComponents      bool   = false
	defaultDisplayVersionAndExit  bool   = false
	defaultAllowUnknownJSONFields bool   = false
	defaultRuntimeTimeout         int    = 10

	// Set a read limit to help prevent abuse from unexpected/overly large
	// input. The limit set here is OVERLY generous and is unlikely to be met
	// unless something is broken.
	//
	// For reference formatted components JSON feeds weigh in around:
	//
	// - DUO (209 KB)
	// - Qualys (152 KB)
	// - Linode (38 KB)
	// - Box (30 KB)
	// - GitHub (6 KB)
	defaultReadLimit int64 = 1 * MB

	defaultInspectorOutputFormat string = InspectorOutputFormatTable
)

// Application and plugin types provided by this project. These values are
// used as labels in logging and report output. See also the PluginType struct
// type used to indicate what plugin is executing.
const (
	PluginComponentsAppType    string = "plugin-components"
	PluginComponentsAppName    string = "check_components"
	InspectorComponentsAppType string = "inspector-components"
	InspectorComponentsAppName string = "lscs"
)

// ThresholdNotUsed indicates that a plugin is not using a specific threshold.
// This is visible in locations where Long Service Output text is displayed.
const ThresholdNotUsed string = "Not used."

const (

	// LogLevelDisabled maps to zerolog.Disabled logging level
	LogLevelDisabled string = "disabled"

	// LogLevelPanic maps to zerolog.PanicLevel logging level
	LogLevelPanic string = "panic"

	// LogLevelFatal maps to zerolog.FatalLevel logging level
	LogLevelFatal string = "fatal"

	// LogLevelError maps to zerolog.ErrorLevel logging level
	LogLevelError string = "error"

	// LogLevelWarn maps to zerolog.WarnLevel logging level
	LogLevelWarn string = "warn"

	// LogLevelInfo maps to zerolog.InfoLevel logging level
	LogLevelInfo string = "info"

	// LogLevelDebug maps to zerolog.DebugLevel logging level
	LogLevelDebug string = "debug"

	// LogLevelTrace maps to zerolog.TraceLevel logging level
	LogLevelTrace string = "trace"
)

// MB represents 1 Megabyte
const MB int64 = 1048576

// Supported Inspector type application output formats
const (
	InspectorOutputFormatOverview string = "overview"
	InspectorOutputFormatTable    string = "table"
	InspectorOutputFormatVerbose  string = "verbose"
	InspectorOutputFormatDebug    string = "debug"
	InspectorOutputFormatIDsList  string = "list"
	InspectorOutputFormatJSON     string = "json"
)
