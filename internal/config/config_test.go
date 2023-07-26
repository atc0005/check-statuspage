// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

// These tests are defined in the same package as the code we are testing so
// that we can directly interact with both exported and unexported package
// items.
package config

import (
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/atc0005/check-statuspage/internal/textutils"
)

var expectedPluginComponentsFlags = []string{
	BrandingFlag,
	VerboseFlag,
	ComponentsListFlagShort,
	ComponentsListFlagLong,
	ComponentGroupFlagShort,
	ComponentGroupFlagLong,
	EvalAllComponentsFlagShort,
	EvalAllComponentsFlagLong,
}

var expectedInspectorComponentsFlags = []string{
	InspectorOutputFormatFlagShort,
	InspectorOutputFormatFlagLong,
}

var expectedSharedFlags = []string{
	HelpFlagLong,
	HelpFlagShort,
	OmitOKComponentsFlagShort,
	OmitOKComponentsFlagLong,
	OmitSummaryResultsFlagShort,
	OmitSummaryResultsFlagLong,
	URLFlagShort,
	URLFlagLong,
	FilenameFlagShort,
	FilenameFlagLong,
	AllowUnknownJSONFieldsFlagShort,
	AllowUnknownJSONFieldsFlagLong,
	ReadLimitFlagShort,
	ReadLimitFlagLong,
	TimeoutFlagShort,
	TimeoutFlagLong,
	LogLevelFlagShort,
	LogLevelFlagLong,
	VersionFlagShort,
	VersionFlagLong,
}

// TestExpectedPluginComponentsFlags tests defined config flags for the
// components plugin against a list of expected flags for the plugin. This is
// done to help prevent documentation from getting out of date with config
// flag changes.
func TestExpectedPluginComponentsFlags(t *testing.T) {

	// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang

	// Save old command-line arguments so that we can restore them later
	oldArgs := os.Args

	// Defer restoring original command-line arguments
	defer func() { os.Args = oldArgs }()

	// Note to self: Don't add/escape double-quotes here. The shell strips
	// them away and the application never sees them.
	os.Args = []string{
		PluginComponentsAppName,
		"--" + FilenameFlagLong, "placeholder",
		"--" + LogLevelFlagLong, "placeholder",
		"--" + ComponentGroupFlagLong, "placeholder",
		"--" + ComponentsListFlagLong, "placeholder",
	}

	var config Config
	appType := AppType{PluginComponents: true}
	config.App = AppInfo{
		Name:    myAppName,
		Version: version,
		URL:     myAppURL,
		Plugin:  appTypeLabel(appType),
	}

	config.flagSet = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	if err := config.handleFlagsConfig(appType); err != nil {
		t.Fatalf(
			"ERROR: Failed to set flags configuration: %v",
			err,
		)
	}

	totalExpectedFlagsCount := len(expectedSharedFlags) + len(expectedPluginComponentsFlags)

	definedFlags := make([]string, 0, totalExpectedFlagsCount)
	config.flagSet.VisitAll(func(f *flag.Flag) {
		definedFlags = append(definedFlags, f.Name)
	})
	definedFlagsCount := len(definedFlags)

	if totalExpectedFlagsCount != len(definedFlags) {
		t.Errorf(
			"ERROR: Expected %d defined flags for %s; got %d defined flags",
			totalExpectedFlagsCount,
			PluginComponentsAppName,
			definedFlagsCount,
		)
	} else {
		t.Logf(
			"OK: Num Flags expected (%d) matches num flags defined (%d)",
			totalExpectedFlagsCount,
			definedFlagsCount,
		)
	}

	// combine the shared and dedicated flag lists
	expectedFlags := make([]string, 0, totalExpectedFlagsCount)
	expectedFlags = append(expectedFlags, expectedSharedFlags...)
	expectedFlags = append(expectedFlags, expectedPluginComponentsFlags...)

	// Assert that each defined flag is represented exactly by an entry in the
	// list of expected flags. Since we have already compared the length of
	// each collection (defined vs expected), we don't have to compare in the
	// opposite direction to assert that each collection is equal.
	for _, definedFlag := range definedFlags {
		if !textutils.InList(definedFlag, expectedFlags, false) {
			t.Errorf(
				"ERROR: defined flag %q is not in the list of expected flags",
				definedFlag,
			)
		} else {
			t.Logf(
				"OK: defined flag %q is in the list of expected flags",
				definedFlag,
			)
		}
	}
	t.Log("OK: Defined flags match expected flags")

}

// TestHelpFlag asserts that specifying help flags is both successful and
// output contains all expected flags for the application type (e.g.,
// components plugin vs components cli app).
func TestHelpFlag(t *testing.T) {

	// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang

	// Save old command-line arguments so that we can restore them later
	oldArgs := os.Args

	// Defer restoring original command-line arguments
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name    string
		appName string
		appType AppType
		flag    string
	}{
		{
			name:    "Components plugin, short help flag",
			appName: PluginComponentsAppName,
			appType: AppType{PluginComponents: true},
			flag:    HelpFlagShort,
		},
		{
			name:    "Components plugin, long help flag",
			appName: PluginComponentsAppName,
			appType: AppType{PluginComponents: true},
			flag:    HelpFlagLong,
		},
		{
			name:    "Components CLI app, short help flag",
			appName: InspectorComponentsAppName,
			appType: AppType{InspectorComponents: true},
			flag:    HelpFlagShort,
		},
		{
			name:    "Components CLI app, long help flag",
			appName: InspectorComponentsAppName,
			appType: AppType{InspectorComponents: true},
			flag:    HelpFlagLong,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

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
				test.appName, test.flag,
			}

			for i, item := range flagsAndValuesInOrder {
				if strings.TrimSpace(item) != "" {
					os.Args = append(os.Args, item)
				} else {
					t.Logf("Skipping item %d due to empty value", i)
				}
			}

			t.Log("INFO: Old os.Args before rewriting:\n", oldArgs)
			t.Log("INFO: New os.Args before init config:\n", os.Args)

			var config Config
			config.App = AppInfo{
				Name:    myAppName,
				Version: version,
				URL:     myAppURL,
				Plugin:  appTypeLabel(test.appType),
			}

			config.flagSet = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

			if err := config.handleFlagsConfig(test.appType); err != nil {
				t.Fatalf(
					"ERROR: Failed to set flags configuration: %v",
					err,
				)
			}

			helpOutput := config.Help()

			// t.Logf("helpOutput: %s", helpOutput)

			// Quick fail: look for known failure text in output.
			if strings.Contains(helpOutput, ErrConfigNotInitialized.Error()) {
				t.Fatalf(
					"ERROR: Help text indicates uninitialized configuration: %s",
					ErrConfigNotInitialized,
				)
			}

			// combine the shared and dedicated flag lists
			expectedFlags := make([]string, 0, len(expectedSharedFlags)+len(expectedPluginComponentsFlags))
			switch {
			case test.appType.InspectorComponents:
				expectedFlags = append(expectedFlags, expectedSharedFlags...)
				expectedFlags = append(expectedFlags, expectedInspectorComponentsFlags...)
			case test.appType.PluginComponents:
				expectedFlags = append(expectedFlags, expectedSharedFlags...)
				expectedFlags = append(expectedFlags, expectedPluginComponentsFlags...)

			}

			// Ensure that we get useful help output by confirming that each
			// defined flag is represented in the generated output.
			for _, expectedFlag := range expectedFlags {
				switch {
				case !strings.Contains(helpOutput, expectedFlag):
					t.Fatalf(
						"ERROR: Help text is missing usage information for flag %q",
						expectedFlag,
					)

				default:
					t.Logf(
						"OK: Help text usage information for flag %q found",
						expectedFlag,
					)
				}
			}
		})
	}

}

// TestExpectedInspectorComponentsFlags tests defined config flags for the
// components inspector app against a list of expected flags for the CLI app.
// This is done to help prevent documentation from getting out of date with
// config flag changes.
func TestExpectedInspectorComponentsFlags(t *testing.T) {

	// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang

	// Save old command-line arguments so that we can restore them later
	oldArgs := os.Args

	// Defer restoring original command-line arguments
	defer func() { os.Args = oldArgs }()

	// Note to self: Don't add/escape double-quotes here. The shell strips
	// them away and the application never sees them.
	os.Args = []string{
		InspectorComponentsAppName,
		"--" + FilenameFlagLong, "placeholder",
		"--" + LogLevelFlagLong, "placeholder",
	}

	var config Config
	appType := AppType{InspectorComponents: true}
	config.App = AppInfo{
		Name:    myAppName,
		Version: version,
		URL:     myAppURL,
		Plugin:  appTypeLabel(appType),
	}

	config.flagSet = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	if err := config.handleFlagsConfig(appType); err != nil {
		t.Fatalf(
			"ERROR: Failed to set flags configuration: %v",
			err,
		)
	}

	totalExpectedFlagsCount := len(expectedSharedFlags) + len(expectedInspectorComponentsFlags)

	definedFlags := make([]string, 0, totalExpectedFlagsCount)
	config.flagSet.VisitAll(func(f *flag.Flag) {
		definedFlags = append(definedFlags, f.Name)
	})
	definedFlagsCount := len(definedFlags)

	if totalExpectedFlagsCount != len(definedFlags) {
		t.Errorf(
			"ERROR: Expected %d defined flags for %s; got %d defined flags",
			totalExpectedFlagsCount,
			InspectorComponentsAppName,
			definedFlagsCount,
		)
	} else {
		t.Logf(
			"OK: Num Flags expected (%d) matches num flags defined (%d)",
			totalExpectedFlagsCount,
			definedFlagsCount,
		)
	}

	// combine the shared and dedicated flag lists
	expectedFlags := make([]string, 0, totalExpectedFlagsCount)
	expectedFlags = append(expectedFlags, expectedSharedFlags...)
	expectedFlags = append(expectedFlags, expectedInspectorComponentsFlags...)

	// Assert that each defined flag is represented exactly by an entry in the
	// list of expected flags. Since we have already compared the length of
	// each collection (defined vs expected), we don't have to compare in the
	// opposite direction to assert that each collection is equal.
	for _, definedFlag := range definedFlags {
		if !textutils.InList(definedFlag, expectedFlags, false) {
			t.Errorf(
				"ERROR: defined flag %q is not in the list of expected flags",
				definedFlag,
			)
		} else {
			t.Logf(
				"OK: defined flag %q is in the list of expected flags",
				definedFlag,
			)
		}
	}
	t.Log("OK: Defined flags match expected flags")

}
