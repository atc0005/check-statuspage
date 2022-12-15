// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/atc0005/check-statuspage/internal/config"
	"github.com/atc0005/check-statuspage/internal/statuspage/components"
	"github.com/atc0005/check-statuspage/internal/textutils"
	"github.com/atc0005/go-nagios"

	"github.com/google/go-cmp/cmp"
)

// Shared flags and values across various tests
const (
	defaultFilenameFlag = "--" + config.FilenameFlagLong
	// defaultFilenameFlagValue    = "testdata/components/instructure-components.json"
	// defaultURLFlag = "--" + config.URLFlagLong
	// defaultURLFlagValue         = "https://status.instructure.com/api/v2/components.json"
	defaultLogLevelFlag         = "--" + config.LogLevelFlagLong
	defaultLogLevelFlagValue    = config.LogLevelInfo
	defaultComponentFlag        = "--" + config.ComponentsListFlagLong
	defaultComponentFlagValue   = "9dlvqx1drp3d" // valid 'Canvas' component ID value
	defaultGroupFlag            = "--" + config.ComponentGroupFlagLong
	defaultGroupFlagValue       = "41wg86q5vc14" // valid 'Canvas' group ID value
	defaultEvalAllFlag          = "--" + config.EvalAllComponentsFlagLong
	defaultEvalAllFlagValue     = "false"
	defaultTimeoutFlag          = "--" + config.TimeoutFlagLong
	defaultTimeoutFlagValue     = "10"
	defaultReadLimitFlag        = "--" + config.ReadLimitFlagLong
	defaultReadLimitFlagValue   = "1048576" // 1 MB
	defaultAllowUnknownJSONFlag = "--" + config.AllowUnknownJSONFieldsFlagLong
)

const (
	prefixIGNORE = "IGNORE"
	prefixOK     = "OK"
)

// evalComponentsTestdataFile represents a table test entry for the
// TestEvalComponentsFromTestdataFiles function.
type evalComponentsTestdataFile struct {
	name                          string
	filenameFlagValue             string
	groupFlag                     string
	groupFlagValue                string
	componentFlag                 string
	componentFlagValue            string
	evalAllComponentsFlagValue    string
	expectedExcludedComponentIDs  []string
	filterErrorExpected           bool
	filterResultsMismatchExpected bool
}

// evalPluginStatusTestdataFile represents a table test entry for the
// TestPluginStatusFromTestdataFiles function.
type evalPluginStatusTestdataFile struct {
	name                         string
	filenameFlagValue            string
	groupFlag                    string
	groupFlagValue               string
	componentFlag                string
	componentFlagValue           string
	evalAllComponentsFlagValue   string
	expectedPluginStatus         nagios.ServiceState
	pluginStatusMismatchExpected bool
}

// itemsFromList1NotInList2 is a helper function to quickly identify items
// from one list that are not in a second list.
func itemsFromList1NotInList2(t *testing.T, list1 []string, list2 []string) []string {
	t.Helper()

	var diff []string

	for _, item := range list1 {
		if !textutils.InList(item, list2, true) {
			diff = append(diff, item)
		}
	}

	return diff
}

func shouldIgnoreError(t *testing.T, err error, errorExpected bool) bool {
	t.Helper()

	errorOccurred := err != nil

	switch {
	case !errorOccurred && errorExpected:
		return false
	case errorOccurred && !errorExpected:
		return false
	default:
		return true
	}
}

func shouldIgnoreCmpResult(t *testing.T, isEqual bool, mismatchExpected bool) bool {
	t.Helper()

	mismatch := !isEqual

	// t.Log("DEBUG: isEqual:", isEqual)
	// t.Log("DEBUG: mismatch:", mismatch)
	// t.Log("DEBUG: mismatchExpected:", mismatchExpected)

	switch {
	case !mismatch && mismatchExpected:
		// t.Log("DEBUG: !mismatch && mismatchExpected triggered")
		return false
	case mismatch && !mismatchExpected:
		// t.Log("DEBUG: mismatch && !mismatchExpected triggered")
		return false
	default:
		// t.Log("DEBUG: default triggered")
		return true
	}
}

// TestLoadButNotDecodeTestdataFiles attempts to load each listed JSON
// testdata file. Each test table entry notes whether an error is expected and
// asserts that result. No decoding or validation is performed of JSON
// testdata content.
func TestLoadButNotDecodeTestdataFiles(t *testing.T) {

	tests := []struct {
		name          string
		filename      string
		errorExpected bool
	}{
		{filename: "testdata/components/box-components.json", errorExpected: false},
		{filename: "testdata/components/cisco-urgentnotices-components.json", errorExpected: false},
		{filename: "testdata/components/ciscoamp-components.json", errorExpected: false},
		{filename: "testdata/components/ciscointersight-components.json", errorExpected: false},
		{filename: "testdata/components/ciscokinetic-components.json", errorExpected: false},
		{filename: "testdata/components/coinbase-components.json", errorExpected: false},
		{filename: "testdata/components/digitalocean-components.json", errorExpected: false},
		{filename: "testdata/components/dropbox-components.json", errorExpected: false},
		{filename: "testdata/components/duo-components.json", errorExpected: false},
		{filename: "testdata/components/github-components-with-problem.json", errorExpected: false},
		{filename: "testdata/components/github-components.json", errorExpected: false},
		{filename: "testdata/components/instructure-components.json", errorExpected: false},
		{filename: "testdata/components/intercom-components.json", errorExpected: false},
		{filename: "testdata/components/lastpass-components.json", errorExpected: false},
		{filename: "testdata/components/linode-components-problems.json", errorExpected: false},
		{filename: "testdata/components/linode-components.json", errorExpected: false},
		{filename: "testdata/components/newrelic-components.json", errorExpected: false},
		{filename: "testdata/components/qualys-components.json", errorExpected: false},
		{filename: "testdata/components/reddit-components.json", errorExpected: false},
		{filename: "testdata/components/squarespace-components.json", errorExpected: false},
		{filename: "testdata/components/twilio-components.json", errorExpected: false},
		{filename: "testdata/components/this-file-does-not-exist-components.json", errorExpected: true},
	}

	for _, test := range tests {

		testName := filepath.Base(test.filename)

		t.Run(testName, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			// The testdata directory is two levels up
			// normalizedFullyQualifiedFilename := filepath.Join("../../", test.filename)
			// normalizedFullyQualifiedFilename, err := filepath.Abs(normalizedFullyQualifiedFilename)
			// if err != nil {
			// 	t.Fatalf("Failed to normalize filename %q: %v", test.filename, err)
			// }

			// fmt.Println(normalizedFullyQualifiedFilename)

			// files, _ := ioutil.ReadDir(".")
			// for _, f := range files {
			// 	fmt.Println(f.Name())
			// }

			// Note to self: Don't add/escape double-quotes here. The shell strips
			// them away and the application never sees them.
			flagsAndValuesInOrder := []string{
				config.PluginComponentsAppName,
				defaultFilenameFlag, filepath.Join("../../", test.filename),
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				defaultComponentFlag, defaultComponentFlagValue,
				defaultAllowUnknownJSONFlag,
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

			// Expected to pass
			cfg, err := config.New(config.AppType{PluginComponents: true})
			if err != nil {
				t.Fatalf("Failed to instantiate configuration: %v", err)
			}

			_, err = components.NewFromFile(cfg.Filename, cfg.ReadLimit, cfg.AllowUnknownJSONFields)
			switch shouldIgnoreError(t, err, test.errorExpected) {
			case true:
				t.Logf(
					"OK | test: %q; got: %v, expected error: %t",
					testName,
					err,
					test.errorExpected,
				)
			case false:
				t.Fatalf(
					"ERROR | test: %q; got: %v, expected error: %t",
					testName,
					err,
					test.errorExpected,
				)
			}

		})

	}

}

// TestLoadAndDecodeAndValidateTestdataFiles attempts to load each listed JSON
// testdata file and then attempts to decode it. Each test table entry notes
// whether an error is expected and asserts that result. No validation is
// performed of JSON testdata content.
func TestLoadAndDecodeAndValidateTestdataFiles(t *testing.T) {

	// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang

	// Save old command-line arguments so that we can restore them later
	oldArgs := os.Args

	// Defer restoring original command-line arguments
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name                      string
		filenameFlag              string
		filenameFlagValue         string
		urlFlag                   string
		urlFlagValue              string
		logLevelFlag              string
		logLevelFlagValue         string
		componentFlag             string
		componentFlagValue        string
		groupFlag                 string
		groupFlagValue            string
		readLimitFlag             string
		readLimitFlagValue        string
		timeoutFlag               string
		timeoutFlagValue          string
		allowUnknownJSONFlag      string
		allowUnknownJSONFlagValue string
		errorLoadingFileExpected  bool
		errorDecodingFileExpected bool
	}{
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/box-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/cisco-urgentnotices-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/ciscoamp-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/ciscointersight-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/ciscokinetic-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/coinbase-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/digitalocean-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/dropbox-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/duo-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/github-components-with-problem.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/github-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/instructure-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/intercom-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/lastpass-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/linode-components-problems.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/linode-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/newrelic-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/qualys-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/reddit-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/squarespace-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			// name:                      "Placeholder for this test case",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/twilio-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  false,
			errorDecodingFileExpected: false,
		},
		{
			name:                      "Invalid filename for components JSON feed",
			filenameFlag:              defaultFilenameFlag,
			filenameFlagValue:         "testdata/components/notreal-components.json",
			urlFlag:                   "",
			urlFlagValue:              "",
			logLevelFlag:              defaultLogLevelFlag,
			logLevelFlagValue:         defaultLogLevelFlagValue,
			componentFlag:             defaultComponentFlag,
			componentFlagValue:        defaultComponentFlagValue,
			groupFlag:                 defaultGroupFlag,
			groupFlagValue:            defaultGroupFlagValue,
			readLimitFlag:             defaultReadLimitFlag,
			readLimitFlagValue:        defaultReadLimitFlagValue,
			timeoutFlag:               defaultTimeoutFlag,
			timeoutFlagValue:          defaultTimeoutFlagValue,
			allowUnknownJSONFlag:      defaultAllowUnknownJSONFlag,
			allowUnknownJSONFlagValue: "false",
			errorLoadingFileExpected:  true,
			errorDecodingFileExpected: true, // the file does not exist, so decoding is not possible
		},
	}

	for _, test := range tests {

		var testName string
		if test.name != "" {
			testName = test.name
		} else {
			testName = filepath.Base(test.filenameFlagValue)
		}

		t.Run(testName, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			normalizedFullyQualifiedFilename := filepath.Join("../../", test.filenameFlagValue)

			flagsAndValuesInOrder := []string{
				config.PluginComponentsAppName,
				test.filenameFlag, normalizedFullyQualifiedFilename,
				test.logLevelFlag, test.logLevelFlagValue,
				test.readLimitFlag, test.readLimitFlagValue,
				test.timeoutFlag, test.timeoutFlagValue,
				test.componentFlag, test.componentFlagValue,
				test.allowUnknownJSONFlag + "=" + test.allowUnknownJSONFlagValue,
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

			// Expected to pass
			cfg, err := config.New(config.AppType{PluginComponents: true})
			if err != nil {
				t.Fatalf("Failed to instantiate configuration: %v", err)
			}

			componentsSet, err := components.NewFromFile(cfg.Filename, cfg.ReadLimit, cfg.AllowUnknownJSONFields)
			switch shouldIgnoreError(t, err, test.errorLoadingFileExpected) {
			case true:
				t.Logf(
					"OK | test: %q; got: %v, expected error: %t",
					testName,
					err,
					test.errorLoadingFileExpected,
				)
			case false:
				t.Fatalf(
					"ERROR | test: %q; got: %v, expected error: %t",
					testName,
					err,
					test.errorLoadingFileExpected,
				)
			}

			err = componentsSet.Validate()
			switch shouldIgnoreError(t, err, test.errorDecodingFileExpected) {
			case true:
				t.Logf(
					"OK | test: %q; got: %v, expected error: %t",
					testName,
					err,
					test.errorLoadingFileExpected,
				)
			case false:
				t.Fatalf(
					"ERROR | test: %q; got: %v, expected error: %t",
					testName,
					err,
					test.errorLoadingFileExpected,
				)
			}

		})
	}

}

// TestEvalComponentsFromTestdataFiles performs the same loading and decoding
// of testdata files as other functions in this test suite, but success for
// both steps is assumed. Instead, this function focuses on evaluating given
// component and component group combinations to assert that filtering
// behavior works as intended.
func TestEvalComponentsFromTestdataFiles(t *testing.T) {

	// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang

	// Save old command-line arguments so that we can restore them later
	oldArgs := os.Args

	// Defer restoring original command-line arguments
	defer func() { os.Args = oldArgs }()

	// Load a battery of test cases specific to multiple Statuspage powered
	// sites. Further test cases can be crafted based on other sites found in
	// the 'testdata/components/' path.
	var tests []evalComponentsTestdataFile
	tests = append(tests, boxEvalComponentsTestEntries...)
	tests = append(tests, githubEvalComponentsTestEntries...)
	tests = append(tests, qualysEvalComponentsTestEntries...)
	tests = append(tests, instructureEvalComponentsTestEntries...)

	for _, test := range tests {

		var testName string
		if test.name != "" {
			testName = test.name
		} else {
			testName = filepath.Base(test.filenameFlagValue)
		}

		t.Run(testName, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			normalizedFullyQualifiedFilename := filepath.Join("../../", test.filenameFlagValue)

			flagsAndValuesInOrder := []string{
				config.PluginComponentsAppName,
				defaultFilenameFlag, normalizedFullyQualifiedFilename,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				test.componentFlag, test.componentFlagValue,
				test.groupFlag, test.groupFlagValue,
				defaultAllowUnknownJSONFlag + "=" + "false",
				defaultEvalAllFlag + "=" + test.evalAllComponentsFlagValue,
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

			// Expected to succeed (no potential failure allowance)
			cfg, err := config.New(config.AppType{PluginComponents: true})
			if err != nil {
				t.Fatalf("Failed to instantiate configuration: %v", err)
			}

			// Expected to succeed (no potential failure allowance)
			componentsSet, err := components.NewFromFile(cfg.Filename, cfg.ReadLimit, cfg.AllowUnknownJSONFields)
			if err != nil {
				t.Fatalf("Failed to initialize components set: %v", err)
			}

			// Expected to succeed (no potential failure allowance)
			err = componentsSet.Validate()
			if err != nil {
				t.Fatalf("Failed to validate components set: %v", err)
			}

			switch {
			case cfg.EvalAllComponents:
				t.Log("Option to evaluate all components chosen")
				componentsSet.EvalAllComponents = true

			default:
				t.Log("Option to evaluate all components not chosen")

				// Apply filter (success depdendent on specific test case)
				csFilter := components.Filter(cfg.ComponentFilter())
				err := componentsSet.Filter(csFilter)
				switch shouldIgnoreError(t, err, test.filterErrorExpected) {
				case true:
					// Don't mark a filter error as "OK", but rather note that
					// we're ignoring it if requested by the test case.
					prefix := prefixIGNORE
					if err == nil {
						prefix = prefixOK
					}
					t.Logf(
						"%s | test: %q; got: %v, expected error: %t",
						prefix,
						testName,
						err,
						test.filterErrorExpected,
					)
				case false:
					t.Errorf(
						"ERROR | test: %q; got: %v, expected error: %t",
						testName,
						err,
						test.filterErrorExpected,
					)

					// Add extra emphasis to make this scenario prominent.
					if err == nil && test.filterErrorExpected {
						t.Errorf("ERROR: Filter successfully applied, but expected to fail.")
					}

					t.FailNow()
				}
			}

			// Evaluate filtering results against test case expected results.
			excludedIDs := func() []string {
				excludedComponents := componentsSet.ExcludedComponents()
				idVals := make([]string, 0, len(excludedComponents))

				for _, component := range excludedComponents {
					idVals = append(idVals, component.ID)
				}

				return idVals
			}()

			// Gather components not excluded from evaluation so we can list
			// them as part of troubleshooting output.
			notExcludedList := func() []string {
				notExcludedComponents := componentsSet.NotExcludedComponents()
				notExcluded := make([]string, 0, len(notExcludedComponents))

				for i, component := range notExcludedComponents {
					item := fmt.Sprintf("(%02d) %s", i+1, component)
					notExcluded = append(notExcluded, item)
				}

				return notExcluded
			}()

			formatList := func(list []string) string {
				var report strings.Builder

				columnsPerRow := 6

				for i := 0; i < len(list); {
					for j := 0; j < columnsPerRow && i < len(list); j++ {
						fmt.Fprintf(&report, `"%s", `, list[i])
						i++
					}
					fmt.Fprintln(&report)
				}

				return report.String()
			}

			t.Logf(
				"Expected Excluded IDs (%d): \n%s",
				len(test.expectedExcludedComponentIDs),
				formatList(test.expectedExcludedComponentIDs),
			)

			t.Logf(
				"Actual Excluded IDs (%d): \n%s",
				len(excludedIDs),
				formatList(excludedIDs),
			)

			t.Logf(
				"Actual Not Excluded Components: \n%s",
				strings.Join(notExcludedList, "\n"),
			)

			sort.Strings(excludedIDs)
			sort.Strings(test.expectedExcludedComponentIDs)

			exclusionsMatch := cmp.Equal(excludedIDs, test.expectedExcludedComponentIDs)

			switch shouldIgnoreCmpResult(t, exclusionsMatch, test.filterResultsMismatchExpected) {
			case true:
				// Don't mark an exclusions mismatch as "OK", but rather note
				// that we're ignoring it if requested by the test case.
				prefix := prefixIGNORE
				if exclusionsMatch {
					prefix = prefixOK
				}
				t.Logf(
					"%s: got %d included, %d excluded components, expected %d excluded",
					prefix,
					len(notExcludedList),
					len(excludedIDs),
					len(test.expectedExcludedComponentIDs),
				)

				if d := cmp.Diff(excludedIDs, test.expectedExcludedComponentIDs); d != "" {
					t.Logf("(-got, +want)\n:%s", d)
				}

				if !exclusionsMatch {
					t.Log("IGNORE: Ignoring mismatch (as requested) between actual excluded components and test case expected IDs list")
				}

			case false:
				t.Errorf(
					"ERROR: got %d included, %d excluded components, expected %d excluded",
					len(notExcludedList),
					len(excludedIDs),
					len(test.expectedExcludedComponentIDs),
				)

				if d := cmp.Diff(excludedIDs, test.expectedExcludedComponentIDs); d != "" {
					t.Errorf("(-got, +want)\n:%s", d)
				}

				// Triggered when more items are matched by a filter (and not
				// excluded) than the test case indicates is expected.
				notExcludedAsExpected := itemsFromList1NotInList2(t, test.expectedExcludedComponentIDs, excludedIDs)
				for _, componentID := range notExcludedAsExpected {

					// Expected to succeed (no potential failure allowance)
					component, err := componentsSet.GetComponentByID(componentID)
					if err != nil {
						t.Fatalf("ERROR: Failed to retrieve component by ID %s", componentID)
					}

					// We rely on earlier switch statement to flag the test as
					// failed or not based on whether the test indicates that
					// a mismatch is expected. Here, we just log the mismatch
					// for potential troubleshooting purposes.
					t.Errorf(
						"Component %s not excluded as expected (tt exclude list)",
						component,
					)
				}

				t.FailNow()

			}

		})

	}
}

// TestPluginStatusFromTestdataFiles performs the same loading, decoding and
// filtering of components using testdata files as other functions in this
// test suite, but success for these steps is assumed. Instead, this function
// focuses on evaluating overall plugin status to assert that service state
// results for evaluated components is as expected..
func TestPluginStatusFromTestdataFiles(t *testing.T) {

	// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang

	// Save old command-line arguments so that we can restore them later
	oldArgs := os.Args

	// Defer restoring original command-line arguments
	defer func() { os.Args = oldArgs }()

	// Load a battery of test cases specific to multiple Statuspage powered
	// sites. Further test cases can be crafted based on other sites found in
	// the 'testdata/components/' path.
	var tests []evalPluginStatusTestdataFile
	tests = append(tests, boxEvalPluginStatusTestEntries...)
	tests = append(tests, githubEvalPluginStatusTestEntries...)
	tests = append(tests, qualysEvalPluginStatusTestEntries...)
	tests = append(tests, instructureEvalPluginStatusTestEntries...)

	for _, test := range tests {

		var testName string
		if test.name != "" {
			testName = test.name
		} else {
			testName = filepath.Base(test.filenameFlagValue)
		}

		t.Run(testName, func(t *testing.T) {

			// Save old command-line arguments so that we can restore them later
			// https://stackoverflow.com/questions/33723300/how-to-test-the-passing-of-arguments-in-golang
			oldArgs := os.Args

			// Defer restoring original command-line arguments
			defer func() { os.Args = oldArgs }()

			// Clear out any entries added by `go test` or leftovers from
			// previous test cases.
			os.Args = nil

			normalizedFullyQualifiedFilename := filepath.Join("../../", test.filenameFlagValue)

			flagsAndValuesInOrder := []string{
				config.PluginComponentsAppName,
				defaultFilenameFlag, normalizedFullyQualifiedFilename,
				defaultLogLevelFlag, defaultLogLevelFlagValue,
				defaultReadLimitFlag, defaultReadLimitFlagValue,
				defaultTimeoutFlag, defaultTimeoutFlagValue,
				test.componentFlag, test.componentFlagValue,
				test.groupFlag, test.groupFlagValue,
				defaultAllowUnknownJSONFlag + "=" + "false",
				defaultEvalAllFlag + "=" + test.evalAllComponentsFlagValue,
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

			// Expected to succeed (no potential failure allowance)
			cfg, err := config.New(config.AppType{PluginComponents: true})
			if err != nil {
				t.Fatalf("Failed to instantiate configuration: %v", err)
			}

			// Expected to succeed (no potential failure allowance)
			componentsSet, err := components.NewFromFile(cfg.Filename, cfg.ReadLimit, cfg.AllowUnknownJSONFields)
			if err != nil {
				t.Fatalf("Failed to initialize components set: %v", err)
			}

			// Expected to succeed (no potential failure allowance)
			err = componentsSet.Validate()
			if err != nil {
				t.Fatalf("Failed to validate components set: %v", err)
			}

			switch {
			case cfg.EvalAllComponents:
				t.Log("Option to evaluate all components chosen")
				componentsSet.EvalAllComponents = true

			default:
				t.Log("Option to evaluate all components not chosen")

				// Expected to succeed (no potential failure allowance)
				csFilter := components.Filter(cfg.ComponentFilter())
				err := componentsSet.Filter(csFilter)
				if err != nil {
					t.Fatalf("Failed to apply filter to components set: %v", err)
				}
			}

			// Test plugin status. Don't evaluate excluded components per
			// earlier filtering.
			serviceState := componentsSet.ServiceState(false)
			// t.Logf("serviceState: %#v", serviceState)
			// t.Logf("test.expectedPluginStatus: %#v", test.expectedPluginStatus)
			serviceStateEqual := cmp.Equal(serviceState, test.expectedPluginStatus)
			// t.Log("serviceStateEqual:", serviceStateEqual)
			switch shouldIgnoreCmpResult(t, serviceStateEqual, test.pluginStatusMismatchExpected) {
			case true:
				// Don't mark a service state mismatch as "OK", but rather
				// note that we're ignoring it if requested by the test case.
				prefix := prefixIGNORE
				if serviceStateEqual {
					prefix = prefixOK
				}
				t.Logf(
					"%s: got %#v, expected %#v",
					prefix,
					serviceState,
					test.expectedPluginStatus,
				)

				if d := cmp.Diff(serviceState, test.expectedPluginStatus); d != "" {
					t.Logf("(-got, +want)\n:%s", d)
				}

				if !serviceStateEqual {
					t.Log("IGNORE: Ignoring mismatch (as requested) between actual plugin state and test case expected state")
				}

			case false:
				switch {
				case !serviceStateEqual:
					t.Errorf(
						"ERROR: got %#v, expected %#v",
						serviceState,
						test.expectedPluginStatus,
					)

					if d := cmp.Diff(serviceState, test.expectedPluginStatus); d != "" {
						t.Errorf("(-got, +want)\n:%s", d)
					}

				case serviceStateEqual && test.pluginStatusMismatchExpected:
					t.Logf(
						"NOTE: got %#v, expected %#v",
						serviceState,
						test.expectedPluginStatus,
					)

					t.Errorf("Service State matched, but mismatch expected per test case")
				}

				t.FailNow()

			}

		})

	}
}

// TestEmptyClientPerfDataAndConstructedPluginProducesDefaultTimeMetric
// asserts that omitted performance data from client code produces a default
// time metric when using the Plugin constructor.
func TestEmptyClientPerfDataAndConstructedPluginProducesDefaultTimeMetric(t *testing.T) {
	t.Parallel()

	// Setup Plugin type the same way that client code using the
	// constructor would.
	plugin := nagios.NewPlugin()

	// Performance Data metrics are not emitted if we do not supply a
	// ServiceOutput value.
	plugin.ServiceOutput = "TacoTuesday"

	var outputBuffer strings.Builder

	plugin.SetOutputTarget(&outputBuffer)

	// os.Exit calls break tests
	plugin.SkipOSExit()

	// Process exit state, emit output to our output buffer.
	plugin.ReturnCheckResults()

	want := fmt.Sprintf(
		"%s | %s",
		plugin.ServiceOutput,
		"'time'=",
	)

	got := outputBuffer.String()

	if !strings.Contains(got, want) {
		t.Errorf("ERROR: Plugin output does not contain the expected time metric")
		t.Errorf("\nwant %q\ngot %q", want, got)
	} else {
		t.Logf("OK: Emitted performance data contains the expected time metric.")
	}
}
