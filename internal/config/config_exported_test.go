// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// Package config_test is used to limit interaction with the config package to
// just exported items.
package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/atc0005/check-statuspage/internal/config"
)

// Shared flags and values across various tests
const (
	defaultFilenameFlag                   = "--" + config.FilenameFlagLong
	defaultFilenameFlagValue              = "testdata/components/instructure-components.json"
	defaultURLFlag                        = "--" + config.URLFlagLong
	defaultURLFlagValue                   = "https://status.instructure.com/api/v2/components.json"
	defaultLogLevelFlag                   = "--" + config.LogLevelFlagLong
	defaultLogLevelFlagValue              = config.LogLevelInfo
	defaultComponentFlag                  = "--" + config.ComponentsListFlagLong
	defaultComponentFlagValue             = "9dlvqx1drp3d" // valid 'Canvas' component ID value
	defaultGroupFlag                      = "--" + config.ComponentGroupFlagLong
	defaultGroupFlagValue                 = "41wg86q5vc14" // valid 'Canvas' group ID value
	defaultTimeoutFlag                    = "--" + config.TimeoutFlagLong
	defaultTimeoutFlagValue               = "10"
	defaultReadLimitFlag                  = "--" + config.ReadLimitFlagLong
	defaultReadLimitFlagValue             = "1048576" // 1 MB
	defaultInspectorOutputFormatFlag      = "--" + config.InspectorOutputFormatFlagLong
	defaultInspectorOutputFormatFlagValue = config.InspectorOutputFormatTable // see config.supportedInspectorOutputFormats()
)

type configFlags struct {
	name               string
	filenameFlag       string
	filenameFlagValue  string
	urlFlag            string
	urlFlagValue       string
	logLevelFlag       string
	logLevelFlagValue  string
	componentFlag      string
	componentFlagValue string
	groupFlag          string
	groupFlagValue     string
	readLimitFlag      string
	readLimitFlagValue string
	timeoutFlag        string
	timeoutFlagValue   string
	errorExpected      bool
}

// sharedComponentsConfigFlagTestCases is a collection of test cases
// applicable to testing flag validation for both the components CLI and
// plugin. The components CLI test ignores fields in these test cases which
// are not applicable to the CLI app, though some CLI-specific test cases may
// mix in unsupported flags in order to confirm expected validation failure
// logic.
var sharedComponentsConfigFlagTestCases = []configFlags{
	{
		name:               "Valid long filename flag set",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid short filename flag set",
		filenameFlag:       "-f",
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid long url flag set",
		filenameFlag:       "",
		filenameFlagValue:  "",
		urlFlag:            defaultURLFlag,
		urlFlagValue:       defaultURLFlagValue,
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid short url flag set",
		filenameFlag:       "",
		filenameFlagValue:  "",
		urlFlag:            "-u",
		urlFlagValue:       defaultURLFlagValue,
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Invalid use of both url flag and filename flag set",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "-u",
		urlFlagValue:       defaultURLFlagValue,
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      true,
	},
	{
		name:               "Valid timeout value",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   "30",
		errorExpected:      false,
	},
	{
		name:               "Invalid timeout value",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  defaultLogLevelFlagValue,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   "-500",
		errorExpected:      true,
	},
	{
		name:               "Valid logging level disabled",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelDisabled,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level panic",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelPanic,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level fatal",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelFatal,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level error",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelError,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level warn",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelWarn,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level info",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelInfo,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level debug",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelDebug,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
	{
		name:               "Valid logging level trace",
		filenameFlag:       defaultFilenameFlag,
		filenameFlagValue:  defaultFilenameFlagValue,
		urlFlag:            "",
		urlFlagValue:       "",
		logLevelFlag:       defaultLogLevelFlag,
		logLevelFlagValue:  config.LogLevelTrace,
		componentFlag:      defaultComponentFlag,
		componentFlagValue: defaultComponentFlagValue,
		groupFlag:          defaultGroupFlag,
		groupFlagValue:     defaultGroupFlagValue,
		readLimitFlag:      defaultReadLimitFlag,
		readLimitFlagValue: defaultReadLimitFlagValue,
		timeoutFlag:        defaultTimeoutFlag,
		timeoutFlagValue:   defaultTimeoutFlagValue,
		errorExpected:      false,
	},
}

// testComponentsConfigFlagTestCases is a helper function which allows for
// deduplication of test logic for the Test*ConfigFlags functions.
func testComponentsConfigFlagTestCases(
	t *testing.T,
	name string,
	errorExpected bool,
	appType config.AppType,
) {
	t.Log("os.Args:", os.Args)
	t.Logf("appType: %#v", appType)

	_, got := config.New(appType)
	errorOccurred := got != nil
	switch {
	case !errorOccurred && errorExpected:
		t.Fatalf(
			"ERROR | test: %q; got: %v, expected error: %t",
			name,
			got,
			errorExpected,
		)
	case errorOccurred && !errorExpected:
		t.Fatalf(
			"ERROR | test: %q; got: %v, expected error: %t",
			name,
			got,
			errorExpected,
		)
	default:
		t.Logf(
			"OK | test: %q; got: %v, expected error: %t",
			name,
			got,
			errorExpected,
		)
	}
}

// TestInspectorComponentsConfigFlags exercises the config package to check
// for changes in configuration validation logic for the components CLI app.
// This is another attempt to help prevent documentation from getting out of
// date with changes to the config package.
func TestInspectorComponentsConfigFlags(t *testing.T) {

	t.Log("Processing sharedComponentsConfigFlagTestCases")
	for _, test := range sharedComponentsConfigFlagTestCases {

		t.Run("Shared_"+test.name, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			// Note to self: Don't add/escape double-quotes here. The shell strips
			// them away and the application never sees them.
			flagsAndValuesInOrder := []string{
				config.InspectorComponentsAppName,
				test.filenameFlag, test.filenameFlagValue,
				test.urlFlag, test.urlFlagValue,
				test.logLevelFlag, test.logLevelFlagValue,
				test.readLimitFlag, test.readLimitFlagValue,
				test.timeoutFlag, test.timeoutFlagValue,
			}

			for i, item := range flagsAndValuesInOrder {

				if strings.TrimSpace(item) != "" {
					os.Args = append(os.Args, item)
				} else {
					t.Logf("Skipping item %d due to empty value", i)
				}
			}

			testComponentsConfigFlagTestCases(t, test.name, test.errorExpected, config.AppType{InspectorComponents: true})
		})
	}

	// A mix of standard/expected success cases which mirror the shared test
	// cases and some specific test cases using flags unsupported by the
	// components CLI app.
	ourTestCases := []struct {
		name                  string
		flagsAndValuesInOrder []string
		errorExpected         bool
	}{
		{
			name: "Invalid mix of URL and Filename flags",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultURLFlag, defaultURLFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
			},
			errorExpected: true,
		},
		{
			name: "Valid URL flag",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultURLFlag, defaultURLFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
			},
			errorExpected: false,
		},
		{
			name: "Valid Filename flag",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
			},
			errorExpected: false,
		},
		{
			name: "Unsupported group flag",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultGroupFlag, defaultGroupFlagValue,
			},
			errorExpected: true,
		},
		{
			name: "Unsupported component flag",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultComponentFlag, defaultComponentFlagValue,
			},
			errorExpected: true,
		},
		{
			name: "Valid Output format 'overview'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, config.InspectorOutputFormatOverview,
			},
			errorExpected: false,
		},
		{
			name: "Valid Output format 'table'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, config.InspectorOutputFormatTable,
			},
			errorExpected: false,
		},
		{
			name: "Valid Output format 'verbose'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, config.InspectorOutputFormatVerbose,
			},
			errorExpected: false,
		},
		{
			name: "Valid Output format 'debug'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, config.InspectorOutputFormatDebug,
			},
			errorExpected: false,
		},
		{
			name: "Valid Output format 'list'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, config.InspectorOutputFormatIDsList,
			},
			errorExpected: false,
		},
		{
			name: "Valid Output format 'json'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, config.InspectorOutputFormatJSON,
			},
			errorExpected: false,
		},
		{
			name: "Invalid Output format 'tacos'",
			flagsAndValuesInOrder: []string{
				config.InspectorComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultInspectorOutputFormatFlag, "tacos",
			},
			errorExpected: true,
		},
	}

	t.Log("Processing ourTestCases")
	for _, test := range ourTestCases {

		t.Run(test.name, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			for i, item := range test.flagsAndValuesInOrder {

				if strings.TrimSpace(item) != "" {
					os.Args = append(os.Args, item)
				} else {
					t.Logf("Skipping item %d due to empty value", i)
				}
			}

			testComponentsConfigFlagTestCases(t, test.name, test.errorExpected, config.AppType{InspectorComponents: true})
		})
	}

}

// TestPluginComponentsConfigFlags exercises the config package to check for
// changes in configuration validation logic for the components plugin. This
// is another attempt to help prevent documentation from getting out of date
// with changes to the config package.
func TestPluginComponentsConfigFlags(t *testing.T) {

	t.Log("Processing sharedComponentsConfigFlagTestCases")
	for _, test := range sharedComponentsConfigFlagTestCases {

		t.Run("Shared_"+test.name, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			// Note to self: Don't add/escape double-quotes here. The shell strips
			// them away and the application never sees them.
			flagsAndValuesInOrder := []string{
				config.PluginComponentsAppName,
				test.filenameFlag, test.filenameFlagValue,
				test.urlFlag, test.urlFlagValue,
				test.logLevelFlag, test.logLevelFlagValue,
				test.groupFlag, test.groupFlagValue,
				test.componentFlag, test.componentFlagValue,
				test.readLimitFlag, test.readLimitFlagValue,
				test.timeoutFlag, test.timeoutFlagValue,
			}

			for i, item := range flagsAndValuesInOrder {

				if strings.TrimSpace(item) != "" {
					os.Args = append(os.Args, item)
				} else {
					t.Logf("Skipping item %d due to empty value", i)
				}
			}

			testComponentsConfigFlagTestCases(t, test.name, test.errorExpected, config.AppType{PluginComponents: true})
		})
	}

	// A mix of standard/expected success cases which mirror the shared test
	// cases and some specific test cases using flags unsupported by the
	// components plugin.
	ourTestCases := []struct {
		name                  string
		flagsAndValuesInOrder []string
		errorExpected         bool
	}{
		{
			name: "Invalid mix of URL and Filename flags",
			flagsAndValuesInOrder: []string{
				config.PluginComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultURLFlag, defaultURLFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultComponentFlag, defaultComponentFlagValue,
			},
			errorExpected: true,
		},
		{
			name: "Valid URL flag, specify component",
			flagsAndValuesInOrder: []string{
				config.PluginComponentsAppName,
				defaultURLFlag, defaultURLFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultComponentFlag, defaultComponentFlagValue,
			},
			errorExpected: false,
		},
		{
			name: "Valid Filename flag, specify component",
			flagsAndValuesInOrder: []string{
				config.PluginComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultComponentFlag, defaultComponentFlagValue,
			},
			errorExpected: false,
		},
		{
			name: "Unsupported output format flag, specify component",
			flagsAndValuesInOrder: []string{
				config.PluginComponentsAppName,
				defaultFilenameFlag, defaultFilenameFlagValue,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultComponentFlag, defaultComponentFlagValue,
				defaultInspectorOutputFormatFlag, defaultInspectorOutputFormatFlagValue,
			},
			errorExpected: true,
		},
	}

	t.Log("Processing ourTestCases")
	for _, test := range ourTestCases {

		t.Run(test.name, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			for i, item := range test.flagsAndValuesInOrder {

				if strings.TrimSpace(item) != "" {
					os.Args = append(os.Args, item)
				} else {
					t.Logf("Skipping item %d due to empty value", i)
				}
			}

			testComponentsConfigFlagTestCases(t, test.name, test.errorExpected, config.AppType{PluginComponents: true})
		})
	}

}
