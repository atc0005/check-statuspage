// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

//nolint:dupl
package main

import "github.com/atc0005/go-nagios"

// As of this writing, GitHub doesn't use component groups, just "flat" or
// top-level components to describe their infrastructure.
var githubEvalComponentsTestEntries = []evalComponentsTestdataFile{
	{
		name:                       "(OK) GitHub, by one valid component ID, complete expected excluded list",
		filenameFlagValue:          "testdata/components/github-components.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "br0l2tvcx85d",
		evalAllComponentsFlagValue: defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			"8l4ygp009s5s", // Git Operations
			"brv1bkgrwx7q", // API Requests
			"4230lsnqdsld", // Webhooks
			"0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			"kr09ddfgbfsf", // Issues
			"hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			"st3j38cctv9l", // GitHub Packages
			"vg70hn9s2tyj", // GitHub Pages
			"h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: false,
	},
	{
		name:                       "(OK) GitHub, by one valid component name, one valid component ID, complete expected excluded list",
		filenameFlagValue:          "testdata/components/github-components.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "GitHub Actions, vg70hn9s2tyj",
		evalAllComponentsFlagValue: defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			"8l4ygp009s5s", // Git Operations
			"brv1bkgrwx7q", // API Requests
			"4230lsnqdsld", // Webhooks
			"0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			"kr09ddfgbfsf", // Issues
			"hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			"st3j38cctv9l", // GitHub Packages
			// "vg70hn9s2tyj", // GitHub Pages
			"h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: false,
	},
	{
		name:                       "(OK) GitHub, by two valid component names, complete expected excluded list",
		filenameFlagValue:          "testdata/components/github-components.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "GitHub Actions, GitHub Pages",
		evalAllComponentsFlagValue: defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			"8l4ygp009s5s", // Git Operations
			"brv1bkgrwx7q", // API Requests
			"4230lsnqdsld", // Webhooks
			"0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			"kr09ddfgbfsf", // Issues
			"hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			"st3j38cctv9l", // GitHub Packages
			// "vg70hn9s2tyj", // GitHub Pages
			"h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: false,
	},
	{
		name:                         "(FAIL) GitHub, by one valid component ID, invalid empty expected excluded list",
		filenameFlagValue:            "testdata/components/github-components.json",
		groupFlag:                    "",
		groupFlagValue:               "",
		componentFlag:                defaultComponentFlag,
		componentFlagValue:           "br0l2tvcx85d",
		evalAllComponentsFlagValue:   defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			// "8l4ygp009s5s", // Git Operations
			// "brv1bkgrwx7q", // API Requests
			// "4230lsnqdsld", // Webhooks
			// "0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			// "kr09ddfgbfsf", // Issues
			// "hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			// "st3j38cctv9l", // GitHub Packages
			// "vg70hn9s2tyj", // GitHub Pages
			// "h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: true,
	},
	{
		name:                         "(FAIL) GitHub, by two valid component names, empty expected excluded list",
		filenameFlagValue:            "testdata/components/github-components.json",
		groupFlag:                    "",
		groupFlagValue:               "",
		componentFlag:                defaultComponentFlag,
		componentFlagValue:           "GitHub Actions, GitHub Pages",
		evalAllComponentsFlagValue:   defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			// "8l4ygp009s5s", // Git Operations
			// "brv1bkgrwx7q", // API Requests
			// "4230lsnqdsld", // Webhooks
			// "0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			// "kr09ddfgbfsf", // Issues
			// "hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			// "st3j38cctv9l", // GitHub Packages
			// "vg70hn9s2tyj", // GitHub Pages
			// "h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: true,
	},
	{
		name:                         "(FAIL) GitHub, by one invalid group name, empty expected excluded list",
		filenameFlagValue:            "testdata/components/github-components.json",
		groupFlag:                    defaultGroupFlag,
		groupFlagValue:               "Fish Tacos",
		componentFlag:                "",
		componentFlagValue:           "",
		evalAllComponentsFlagValue:   defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			// "8l4ygp009s5s", // Git Operations
			// "brv1bkgrwx7q", // API Requests
			// "4230lsnqdsld", // Webhooks
			// "0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			// "kr09ddfgbfsf", // Issues
			// "hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			// "st3j38cctv9l", // GitHub Packages
			// "vg70hn9s2tyj", // GitHub Pages
			// "h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected: true,

		// empty list of expected exclusions matches empty list of actual
		// exclusions due to failed filter which we are permitting
		filterResultsMismatchExpected: false,
	},
	{
		name:                       "(FAIL) GitHub, by three valid component names, invalid complete expected excluded list",
		filenameFlagValue:          "testdata/components/github-components.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "GitHub Actions, GitHub Packages, GitHub Pages",
		evalAllComponentsFlagValue: defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			"8l4ygp009s5s", // Git Operations
			"brv1bkgrwx7q", // API Requests
			"4230lsnqdsld", // Webhooks
			"0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			"kr09ddfgbfsf", // Issues
			"hhtssxt0f5v2", // Pull Requests
			"br0l2tvcx85d", // GitHub Actions
			"st3j38cctv9l", // GitHub Packages
			"vg70hn9s2tyj", // GitHub Pages
			"h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: true,
	},
	{
		name:                       "(OK) GitHub, by three valid component names, valid expected excluded list",
		filenameFlagValue:          "testdata/components/github-components.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "GitHub Actions, GitHub Packages, GitHub Pages",
		evalAllComponentsFlagValue: defaultEvalAllFlagValue,
		expectedExcludedComponentIDs: []string{
			"8l4ygp009s5s", // Git Operations
			"brv1bkgrwx7q", // API Requests
			"4230lsnqdsld", // Webhooks
			"0l2p9nhqnxpd", // 'Visit www.githubstatus.com for more information'
			"kr09ddfgbfsf", // Issues
			"hhtssxt0f5v2", // Pull Requests
			// "br0l2tvcx85d", // GitHub Actions
			// "st3j38cctv9l", // GitHub Packages
			// "vg70hn9s2tyj", // GitHub Pages
			"h2ftsgbw7kmk", // Codespaces
		},
		filterErrorExpected:           false,
		filterResultsMismatchExpected: false,
	},
	{
		name:                          "(OK) GitHub, eval all, empty expected excluded list",
		filenameFlagValue:             "testdata/components/github-components.json",
		groupFlag:                     "",
		groupFlagValue:                "",
		componentFlag:                 "",
		componentFlagValue:            "",
		evalAllComponentsFlagValue:    "true",
		expectedExcludedComponentIDs:  []string{}, // everything should be matched, so no components will be excluded
		filterErrorExpected:           false,
		filterResultsMismatchExpected: false,
	},
}

var githubEvalPluginStatusTestEntries = []evalPluginStatusTestdataFile{
	{
		name:                       "(FAIL) GitHub, eval all, invalid plugin status",
		filenameFlagValue:          "testdata/components/github-components-with-problem.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              "",
		componentFlagValue:         "",
		evalAllComponentsFlagValue: "true",
		expectedPluginStatus: nagios.ServiceState{
			Label:    nagios.StateOKLabel,
			ExitCode: nagios.StateOKExitCode,
		},
		pluginStatusMismatchExpected: true, // actual status is WARNING, we note OK
	},
	{
		name:                       "(OK) GitHub, eval all, valid plugin status",
		filenameFlagValue:          "testdata/components/github-components-with-problem.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              "",
		componentFlagValue:         "",
		evalAllComponentsFlagValue: "true",
		expectedPluginStatus: nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		},
		pluginStatusMismatchExpected: false,
	},
	{
		name:                       "(OK) Github, by component name, valid plugin status",
		filenameFlagValue:          "testdata/components/github-components-with-problem.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "GitHub Actions",
		evalAllComponentsFlagValue: "false",
		expectedPluginStatus: nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		},
		pluginStatusMismatchExpected: false,
	},
	{
		name:                       "(OK) Github, by component id, valid plugin status",
		filenameFlagValue:          "testdata/components/github-components-with-problem.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "br0l2tvcx85d",
		evalAllComponentsFlagValue: "false",
		expectedPluginStatus: nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		},
		pluginStatusMismatchExpected: false,
	},
	{
		name:                       "(FAIL) GitHub, by component name, invalid plugin status",
		filenameFlagValue:          "testdata/components/github-components-with-problem.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "GitHub Pages",
		evalAllComponentsFlagValue: "false",
		expectedPluginStatus: nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		},
		pluginStatusMismatchExpected: true, // actual status is OK
	},
	{
		name:                       "(FAIL) GitHub, by component id, invalid plugin status",
		filenameFlagValue:          "testdata/components/github-components-with-problem.json",
		groupFlag:                  "",
		groupFlagValue:             "",
		componentFlag:              defaultComponentFlag,
		componentFlagValue:         "vg70hn9s2tyj",
		evalAllComponentsFlagValue: "false",
		expectedPluginStatus: nagios.ServiceState{
			Label:    nagios.StateWARNINGLabel,
			ExitCode: nagios.StateWARNINGExitCode,
		},
		pluginStatusMismatchExpected: true, // actual status is OK
	},
}
