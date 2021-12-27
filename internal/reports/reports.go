// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package reports

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/atc0005/check-statuspage/internal/statuspage/components"
	"github.com/atc0005/check-statuspage/internal/textutils"
	"github.com/atc0005/go-nagios"
)

// printStartDate is a helper function to display a component's creation or
// start date (if set) in the desired format for inclusion in summary output.
// func printStartDate(csd components.ComponentStartDate) string {
// 	if !csd.IsSet() {
// 		return ""
// 	}
//
// 	return fmt.Sprintf(
// 		" (Created: %s)",
// 		csd.Format(components.ComponentStartDateLayout),
// 	)
//
// }

// printStatus is a helper function to display a component's status in the
// desired format for inclusion in summary output.
func printStatus(status string) string {
	return strings.ToUpper(strings.ReplaceAll(status, "_", " "))
}

// printVerboseComponent is a helper function to display a component (group or
// subcomponet) in a consistent way throughout the verbose components report.
func printVerboseComponent(component *components.Component, num int) string {
	groupID := "N/A"
	if !component.Group {
		groupID = string(component.GroupID)
	}
	return fmt.Sprintf(
		"* %03d): %s [ID: %s, GroupID: %s, Status: %s]%s",
		num,
		component.Name,
		component.ID,
		groupID,
		printStatus(component.Status),
		nagios.CheckOutputEOL,
	)
}

// printTableHeaderSeparatorRowTmpl is a helper function to generate a
// separator row template for use between the header and data rows. Each
// "column" in the generated separator row template is of the same length as
// the header row column above it.
func printTableHeaderSeparatorRowTmpl(headerRowTmpl string) string {

	var headerSepRowTmpl strings.Builder

	headerTmplItems := strings.Split(headerRowTmpl, "\t")

	// Drop the last trailing tab character from the slice.
	if len(headerTmplItems) > 0 {
		headerTmplItems = headerTmplItems[:len(headerTmplItems)-1]
	}
	for _, item := range headerTmplItems {
		headerSepRowTmpl.WriteString(strings.Repeat("-", len(item)))
		headerSepRowTmpl.WriteString("\t")
	}

	// Ensure that the row template has a trailing slash and a placeholder for
	// a nagios.CheckOutputEOL string.
	headerSepRowTmpl.WriteString("\t%s")

	return headerSepRowTmpl.String()
}

// printTableDataRowTmpl is a helper function to generate a separator row
// template for use between collections of components. The number of "columns"
// in the generated separator row template  is of the same length as the
// header row (or any other in the table).
func printTableDataRowTmpl(headerRowTmpl string) string {

	var dataRowTmpl strings.Builder

	// Don't count the trailing tab as a column when calculating the header
	// row length.
	headerTmplLen := len(strings.Split(headerRowTmpl, "\t")) - 1
	if headerTmplLen < 0 {
		headerTmplLen = 0
	}

	dataRowTmpl.WriteString(strings.Repeat("%s\t", headerTmplLen))
	dataRowTmpl.WriteString("%s")

	return dataRowTmpl.String()
}

// printTableCollectionSeparatorTmpl is a helper function to generate a
// separator row template between collections of components. The number of
// "columns" in the generated separator row template is of the same length as
// the header row (or any other in the table).
func printTableCollectionSeparatorTmpl(headerRowTmpl string) string {

	var separatorTmpl strings.Builder

	// Don't count the trailing tab as a column when calculating the header
	// row length.
	headerTmplLen := len(strings.Split(headerRowTmpl, "\t")) - 1
	if headerTmplLen < 0 {
		headerTmplLen = 0
	}

	separatorTmpl.WriteString(strings.Repeat("\t", headerTmplLen))
	separatorTmpl.WriteString("%s")

	return separatorTmpl.String()
}

// ComponentsVerbose generates a verbose report for use as LongServiceOutput
// content, intended for display by the monitoring system either via web UI or
// one or more notifications.
//
// This report is intended to display verbose details for a feed to help
// troubleshoot final results for a statuspage feed.
func ComponentsVerbose(componentsSet *components.Set, omitOKComponents bool) string {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsVerbose func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder

	// Track whether we've emitted anything by the time we finish evaluating
	// all standalone/top-level components, component groups and
	// subcomponents. If we haven't, we'll emit a "N/A" in place of normal
	// column details.
	componentsEmitted := false

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsReportHeader(&report, componentsSet)

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	if omitOKComponents {
		fmt.Fprint(
			&report,
			"NOTE: Omitting OK/operational components as requested.",
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)
	}

	if componentsSet.NumGroups() > 0 {

		fmt.Fprintf(
			&report,
			"%sComponent Groups (%d):%s%s",
			nagios.CheckOutputEOL,
			componentsSet.NumGroups(),
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)

		for i, component := range componentsSet.Groups() {
			if component.IsOKState() && omitOKComponents {
				continue
			}
			fmt.Fprint(&report, printVerboseComponent(component, i+1))
			componentsEmitted = true
		}

		fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	if componentsSet.NumTopLevel() > 0 {

		fmt.Fprintf(
			&report,
			"%sTop-level components (%d):%s%s",
			nagios.CheckOutputEOL,
			componentsSet.NumTopLevel(),
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)

		for i, component := range componentsSet.TopLevel() {
			if component.IsOKState() && omitOKComponents {
				continue
			}
			fmt.Fprint(&report, printVerboseComponent(component, i+1))
			componentsEmitted = true
		}

		fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	if componentsSet.NumSubcomponents() > 0 {

		fmt.Fprintf(
			&report,
			"%sSubcomponents (%d):%s%s",
			nagios.CheckOutputEOL,
			componentsSet.NumSubcomponents(),
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)

		for i, component := range componentsSet.Subcomponents() {
			if component.IsOKState() && omitOKComponents {
				continue
			}
			fmt.Fprint(&report, printVerboseComponent(component, i+1))
			componentsEmitted = true
		}

		fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	// Evaluate all components, regardless of "ignore" or exclusion
	// status.
	if !componentsSet.IsOKState(true) {

		fmt.Fprintf(
			&report,
			"%sComponents (%d) in a non-OK state:%s%s",
			nagios.CheckOutputEOL,
			componentsSet.NumProblemComponents(true),
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)

		for i, component := range componentsSet.ProblemComponents(true) {
			if !component.IsOKState() {
				fmt.Fprint(&report, printVerboseComponent(component, i+1))
			}
		}

		fmt.Fprint(&report, nagios.CheckOutputEOL)

	}

	switch {
	case !componentsEmitted && omitOKComponents:
		fmt.Fprint(
			&report,
			"* All components are operational. No problems to report.",
			nagios.CheckOutputEOL,
		)
	case !componentsEmitted && !omitOKComponents:
		fmt.Fprint(
			&report,
			"* Skipping OK components was not requested, but no components were emitted. Bug?",
			nagios.CheckOutputEOL,
		)
	}

	componentsStatusSummary(&report, componentsSet, omitOKComponents)

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	// Emit this based on http client used to retrieve JSON feed"
	//
	// fmt.Fprintf(
	// 	&report,
	// 	"Plugin User Agent: %s%s",
	// 	"TODO: GH-4"
	// 	nagios.CheckOutputEOL,
	// )

	return report.String()

}

// ComponentsOverview generates a report for use as LongServiceOutput content,
// intended for display by the monitoring system either via web UI or one or
// more notifications.
//
// This report is intended to provide a very rough equivalent to viewing the
// statuspage for a service.
func ComponentsOverview(componentsSet *components.Set, omitOKComponents bool) string {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsOverview func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder

	// A collection of errors (if any) encountered while generating this
	// report.
	var errsEncountered []error

	// Track whether we've emitted anything by the time we finish evaluating
	// all standalone/top-level components, component groups and
	// subcomponents. If we haven't, we'll emit a "N/A" in place of normal
	// column details.
	componentsEmitted := false

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsReportHeader(&report, componentsSet)

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	if omitOKComponents {
		fmt.Fprint(
			&report,
			"NOTE: Omitting OK/operational components as requested.",
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)
	}

	if componentsSet.NumTopLevel() > 0 {

		for _, component := range componentsSet.TopLevel() {

			if component.IsOKState() && omitOKComponents {
				continue
			}

			componentsEmitted = true

			fmt.Fprintf(
				&report,
				"%s [%s]%s",
				component.Name,
				printStatus(component.Status),
				nagios.CheckOutputEOL,
			)
		}

	}

	// Generate a listing of component groups with their subcomponents, if
	// available.
	if componentsSet.NumGroups() > 0 {
		allComponentGroups, err := componentsSet.GetAllGroups()
		switch {
		case err != nil:
			errsEncountered = append(errsEncountered, err)

		default:

			for _, group := range allComponentGroups {

				if group.Parent.IsOKState() && omitOKComponents {
					continue
				}

				componentsEmitted = true

				fmt.Fprintf(
					&report,
					"%s%s [%s]%s",
					nagios.CheckOutputEOL,
					group.Parent.Name,
					printStatus(group.Parent.Status),
					nagios.CheckOutputEOL,
				)

				for _, subcomponent := range group.Subcomponents {
					switch {
					case subcomponent.IsOKState():

						if omitOKComponents {
							continue
						}

						componentsEmitted = true

						fmt.Fprintf(
							&report,
							"\t%s%s",
							subcomponent.Name,
							nagios.CheckOutputEOL,
						)
					default:
						componentsEmitted = true

						fmt.Fprintf(
							&report,
							"\t%s [%s]%s",
							subcomponent.Name,
							printStatus(subcomponent.Status),
							nagios.CheckOutputEOL,
						)
					}
				}
			}
		}
	}

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	switch {
	case !componentsEmitted && omitOKComponents:
		fmt.Fprint(
			&report,
			"* All components are operational. No problems to report.",
			nagios.CheckOutputEOL,
		)
	case !componentsEmitted && !omitOKComponents:
		fmt.Fprint(
			&report,
			"* Skipping OK components was not requested, but no components were emitted. Bug?",
			nagios.CheckOutputEOL,
		)
	}

	if len(errsEncountered) > 0 {
		fmt.Fprint(&report, nagios.CheckOutputEOL)

		fmt.Fprintf(
			&report,
			"Errors encountered while generating this report:%s",
			nagios.CheckOutputEOL,
		)
		for i, err := range errsEncountered {
			fmt.Fprintf(
				&report,
				"* %02d): %s%s",
				i+1,
				err,
				nagios.CheckOutputEOL,
			)
		}
	}

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsStatusSummary(&report, componentsSet, omitOKComponents)

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	return report.String()

}

// ComponentsTable generates a report in a tabular format for use as
// LongServiceOutput content, intended for display by the monitoring system
// either via web UI or one or more notifications. If specified, only
// non-operational status components will be listed, otherwise all components
// defined for the given Statuspage will be displayed.
//
// This report provides component details in a quick reference format with the
// intention of making it easier to craft service checks using name or ID
// values.
func ComponentsTable(componentsSet *components.Set, omitOKComponents bool) string {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsTable func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder
	// w := tabwriter.NewWriter(&report, 4, 4, 4, ' ', 0)
	w := tabwriter.NewWriter(&report, 4, 4, 4, ' ', 0)

	// A collection of errors (if any) encountered while generating this
	// report.
	var errsEncountered []error

	// Add some lead-in spacing to better separate any (potential) earlier log
	// messages from summary output.
	fmt.Fprint(w, nagios.CheckOutputEOL)

	componentsReportHeader(&report, componentsSet)

	fmt.Fprint(&report, nagios.CheckOutputEOL, nagios.CheckOutputEOL)

	if omitOKComponents {
		fmt.Fprint(
			&report,
			"NOTE: Omitting OK/operational components as requested.",
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)
	}

	var headerRowTmpl string
	switch {
	case componentsSet.NumGroups() > 0:
		headerRowTmpl = "GROUP NAME\tGROUP ID\tCOMPONENT NAME\tCOMPONENT ID\tEVALUATED\tSTATUS\t%s"
	default:
		headerRowTmpl = "COMPONENT NAME\tID\tEVALUATED\tSTATUS\t%s"
	}

	separatorHeaderTmpl := printTableHeaderSeparatorRowTmpl(headerRowTmpl)
	separatorRowTmpl := printTableCollectionSeparatorTmpl(headerRowTmpl)
	rowTmpl := printTableDataRowTmpl(headerRowTmpl)

	fmt.Fprintf(w, headerRowTmpl, nagios.CheckOutputEOL)
	fmt.Fprintf(w, separatorHeaderTmpl, nagios.CheckOutputEOL)

	// Used to indicate whether a component has been evaluated or not excluded
	// from eligibility of determining the overall plugin state by its componentsSet.
	// This determination is made by the components Set Filter() method.
	var evaluated string
	if !componentsSet.FilterApplied && !componentsSet.EvalAllComponents {
		evaluated = "N/A"
	}

	// Track whether we've emitted anything by the time we finish evaluating
	// all standalone/top-level components, component groups and
	// subcomponents. If we haven't, we'll emit a "N/A" in place of normal
	// column details.
	componentsEmitted := false

	if componentsSet.NumTopLevel() > 0 {

		separatorRowNeeded := false
		switch {
		case componentsSet.NumGroups() > 0:
			for _, component := range componentsSet.TopLevel() {

				if component.IsOKState() && omitOKComponents {
					continue
				}

				separatorRowNeeded = true

				if componentsSet.FilterApplied || componentsSet.EvalAllComponents {
					evaluated = strconv.FormatBool(!component.Exclude)
				}

				fmt.Fprintf(
					w,
					rowTmpl,
					"",
					"",
					component.Name,
					component.ID,
					evaluated,
					printStatus(component.Status),
					nagios.CheckOutputEOL,
				)
			}
		default:
			for _, component := range componentsSet.TopLevel() {

				if component.IsOKState() && omitOKComponents {
					continue
				}

				separatorRowNeeded = true

				if componentsSet.FilterApplied || componentsSet.EvalAllComponents {
					evaluated = strconv.FormatBool(!component.Exclude)
				}

				fmt.Fprintf(
					w,
					rowTmpl,
					component.Name,
					component.ID,
					evaluated,
					printStatus(component.Status),
					nagios.CheckOutputEOL,
				)
			}
		}

		if separatorRowNeeded {
			componentsEmitted = true
			fmt.Fprintf(w, separatorRowTmpl, nagios.CheckOutputEOL)
		}

	}

	// Generate a listing of component groups with their subcomponents, if
	// available.
	if componentsSet.NumGroups() > 0 {

		separatorRowNeeded := false

		allComponentGroups, err := componentsSet.GetAllGroups()
		switch {
		case err != nil:
			errsEncountered = append(errsEncountered, err)

		default:

			for _, group := range allComponentGroups {

				if group.Parent.IsOKState() && omitOKComponents {
					continue
				}

				for _, subcomponent := range group.Subcomponents {

					if subcomponent.IsOKState() && omitOKComponents {
						continue
					}

					separatorRowNeeded = true
					if componentsSet.FilterApplied || componentsSet.EvalAllComponents {
						evaluated = strconv.FormatBool(!subcomponent.Exclude)
					}

					fmt.Fprintf(
						w,
						rowTmpl,
						group.Parent.Name,
						group.Parent.ID,
						subcomponent.Name,
						subcomponent.ID,
						evaluated,
						printStatus(subcomponent.Status),
						nagios.CheckOutputEOL,
					)
				}

				if separatorRowNeeded {
					fmt.Fprintf(w, separatorRowTmpl, nagios.CheckOutputEOL)
				}

			}

			if separatorRowNeeded {
				componentsEmitted = true
			}
		}
	}

	if !componentsEmitted {
		fmt.Fprintf(
			w,
			rowTmpl,
			"N/A",
			"N/A",
			"N/A",
			"N/A",
			"N/A",
			"N/A",
			nagios.CheckOutputEOL,
		)
	}

	fmt.Fprint(w, nagios.CheckOutputEOL)

	if err := w.Flush(); err != nil {
		errsEncountered = append(errsEncountered, err)
		// logger.Printf("Error flushing tabwriter: %v", err)
	}

	if len(errsEncountered) > 0 {
		fmt.Fprint(w, nagios.CheckOutputEOL)

		fmt.Fprintf(
			w,
			"Errors encountered while generating this report:%s",
			nagios.CheckOutputEOL,
		)
		for i, err := range errsEncountered {
			fmt.Fprintf(
				w,
				"* %02d): %s%s",
				i+1,
				err,
				nagios.CheckOutputEOL,
			)
		}
	}

	fmt.Fprint(w, nagios.CheckOutputEOL)

	componentsStatusSummary(w, componentsSet, omitOKComponents)

	fmt.Fprint(w, nagios.CheckOutputEOL)

	return report.String()

}

// ComponentsIDList generates a multi-column list of component IDs. This
// multi-column list can be used to populate component ID fields in tests and
// other batch processing tasks.
func ComponentsIDList(componentsSet *components.Set) string {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsIDList func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder

	// A collection of errors (if any) encountered while generating this
	// report.
	var errsEncountered []error

	columnsPerRow := 4

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsReportHeader(&report, componentsSet)

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	notExcludedComponents := componentsSet.NotExcludedComponents()
	notExcluded := make([]string, 0, len(notExcludedComponents))

	for _, component := range notExcludedComponents {
		notExcluded = append(notExcluded, component.ID)
	}

	for i := 0; i < len(notExcluded); {
		for j := 0; j < columnsPerRow && i < len(notExcluded); j++ {
			fmt.Fprintf(&report, `"%s", `, notExcluded[i])
			i++
		}
		fmt.Fprintln(&report)
	}

	if len(errsEncountered) > 0 {
		fmt.Fprint(&report, nagios.CheckOutputEOL)

		fmt.Fprintf(
			&report,
			"Errors encountered while generating this report:%s",
			nagios.CheckOutputEOL,
		)
		for i, err := range errsEncountered {
			fmt.Fprintf(
				&report,
				"* %02d): %s%s",
				i+1,
				err,
				nagios.CheckOutputEOL,
			)
		}
	}

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsStatusSummary(&report, componentsSet, true)

	fmt.Fprint(&report, nagios.CheckOutputEOL)

	return report.String()

}

// componentsStatusSummary generates a brief summary of high-level component
// details. This summary is written to the provided io.Writer.
func componentsStatusSummary(
	w io.Writer,
	componentsSet *components.Set,
	omitOKComponents bool,
) {
	fmt.Fprintf(
		w,
		"%sSummary:%s%s",
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Filtering applied to components set: %t%s",
		componentsSet.FilterApplied,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Evaluating all components in the set: %t%s",
		componentsSet.EvalAllComponents,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Omitting OK/operational components (if requested): %t%s",
		omitOKComponents,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Number of total top-level components: %d%s",
		componentsSet.NumTopLevel(),
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Number of total component groups: %d%s",
		componentsSet.NumGroups(),
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Number of total subcomponents: %d%s",
		componentsSet.NumSubcomponents(),
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Number of total problem components: %d%s",
		componentsSet.NumProblemComponents(true),
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Number of ignored problem components: %d%s",
		componentsSet.NumProblemComponents(true)-componentsSet.NumProblemComponents(false),
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"* Number of remaining problem components: %d%s",
		componentsSet.NumProblemComponents(false),
		nagios.CheckOutputEOL,
	)

}

func componentsReportHeader(w io.Writer, componentsSet *components.Set) {

	fmt.Fprintf(
		w,
		"Page: %s (%s)%s",
		componentsSet.Page.Name,
		componentsSet.Page.URL,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"Time Zone: %s%s",
		componentsSet.Page.TimeZone,
		nagios.CheckOutputEOL,
	)

	fmt.Fprintf(
		w,
		"Last Updated: %s%s",
		componentsSet.Page.UpdatedAt.Format(time.RFC3339),
		nagios.CheckOutputEOL,
	)
}

// ComponentsOneLineCheckSummary is used to generate a one-line Nagios service
// check results summary. This is the line most prominent in notifications.
func ComponentsOneLineCheckSummary(
	stateLabel string,
	componentsSet *components.Set,
	evalExcluded bool,
) string {
	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsOneLineCheckSummary func.\n",
			time.Since(funcTimeStart),
		)
	}()

	evaluatedComponents := componentsSet.NumComponents() - componentsSet.NumExcluded()
	problemComponents := componentsSet.ProblemComponents((evalExcluded))
	numProblemComponents := len(problemComponents)
	problemStatusIdx := make(map[string]int)

	for _, component := range problemComponents {
		problemStatusIdx[component.Status]++
	}

	serviceState := componentsSet.ServiceState(evalExcluded)
	potentialStatuses := components.ServiceStateToComponentStatuses(serviceState)

	componentStatuses := make([]string, 0, len(problemStatusIdx))
	for status, count := range problemStatusIdx {
		if textutils.InList(status, potentialStatuses, false) {
			statusTally := fmt.Sprintf("%s (%d)", status, count)
			componentStatuses = append(componentStatuses, statusTally)
		}
	}

	var statusTallies string
	generalStatus := "component has a non-operational status"
	if numProblemComponents > 1 || numProblemComponents == 0 {
		generalStatus = "components have a non-operational status"
	}
	if numProblemComponents > 0 {
		statusTallies = "[" + strings.Join(componentStatuses, ", ") + "]"
	}

	// WARNING: 2 "Qualys, Inc." evaluted components have a non-operational status (11 evaluated, 258 total) [under_maintenance (2)]
	// WARNING: 0 "Qualys, Inc." evaluted components have a non-operational status (Y evaluated, Z total)
	summaryTmpl := "%s: %d evaluated %q %s (%d evaluated, %d total) %s"
	return fmt.Sprintf(
		summaryTmpl,
		stateLabel,
		numProblemComponents,
		componentsSet.Page.Name,
		generalStatus,
		evaluatedComponents,
		componentsSet.NumComponents(),
		statusTallies,
	)

}

// ComponentsReport generates a summary of evaluated component status
// information along with specific verbose details intended to aid in
// troubleshooting check results at a glance.
//
// This information is provided for use with the Long Service Output field
// commonly displayed on the detailed service check results display in the web
// UI or in the body of many notifications.
func ComponentsReport(
	stateLabel string,
	filter components.Filter,
	componentsSet *components.Set,
	omitOKComponents bool,
) string {
	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsReport func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder

	switch {
	case !componentsSet.EvalAllComponents:
		fmt.Fprintf(
			&report,
			"Specified filter: %s%s",
			filter,
			nagios.CheckOutputEOL,
		)
	default:
		fmt.Fprintf(
			&report,
			"NOTE: Evaluating all components as requested.%s",
			nagios.CheckOutputEOL,
		)
	}

	fullTableOutputComponentsLimit := 50
	switch {
	case componentsSet.NumComponents() <= fullTableOutputComponentsLimit || !omitOKComponents:
		fmt.Fprint(&report, ComponentsTable(componentsSet, false))
	default:
		fmt.Fprint(&report, ComponentsTable(componentsSet, true))
	}

	return report.String()
}
