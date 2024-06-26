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

// ComponentsTableColumnFilter specifies what columns should be emitted from
// the table output format. If not provided to applicable functions (e.g. a
// nil value), a default set of columns is used.
type ComponentsTableColumnFilter struct {
	GroupName     bool
	GroupID       bool
	ComponentName bool
	ComponentID   bool
	Evaluated     bool
	Status        bool
}

// componentsTable represents the tabular output generated for a components
// report. A constructor should be used in order to properly initialize
// embedded fields.
type componentsTable struct {

	// report is used directly and by the embedded *tabwriter.Writer to
	// accumulate formatted component details for display as a final report.
	report strings.Builder

	// tabWriter uses the embedded strings.Builder to collect formatted
	// component details.
	tabWriter *tabwriter.Writer

	// filter controls which columns are emitted by the final report.
	filter ComponentsTableColumnFilter

	// header collects the fields used as a table header.
	header componentsTableRow
}

// componentsTableRow represents a table row in the components table output.
// Fields may be omitted if a value is not intended for use.
type componentsTableRow struct {
	GroupName     string
	GroupID       string
	ComponentName string
	ComponentID   string
	Evaluated     string
	Status        string
}

// FieldsEnabled indicates how many column fields are enabled for display.
func (ctf ComponentsTableColumnFilter) FieldsEnabled() int {
	var num int

	if ctf.GroupName {
		num++
	}

	if ctf.GroupID {
		num++
	}

	if ctf.ComponentName {
		num++
	}

	if ctf.ComponentID {
		num++
	}

	if ctf.Evaluated {
		num++
	}

	if ctf.Status {
		num++
	}

	return num

}

// newComponentsTable handles constructing a new components table for use in
// building a report. This constructor should be used instead of attempting to
// directly instantiate the componentsTable type. The provided filter is used
// to control which columns are emitted in the report.
func newComponentsTable(columnsList ComponentsTableColumnFilter) *componentsTable {
	var table componentsTable

	table.filter = columnsList

	w := tabwriter.NewWriter(&table.report, 4, 4, 3, ' ', 0)
	// w := tabwriter.NewWriter(&table.report, 4, 4, 4, ' ', tabwriter.Debug|tabwriter.DiscardEmptyColumns)

	// See GH-44 regarding issues with lack of spacing between columns in
	// email notifications (when using tabs).
	//
	// w := tabwriter.NewWriter(&table.report, 0, 8, 2, '\t', 0)
	table.tabWriter = w

	return &table
}

// addHeaderRow is used to add a header to the report. Any provided fields not
// enabled for display are ignored when generating the report.
func (ctr *componentsTable) addHeaderRow(headerRow componentsTableRow) {
	// Save as-is for later use.
	ctr.header = headerRow

	// Add row normally.
	ctr.addRow(headerRow)
}

// headerRow returns the components table header as a formatted string with
// tab terminators after each column header.
func (ctr *componentsTable) headerRow() string {
	var output strings.Builder

	if ctr.filter.GroupName {
		_, _ = fmt.Fprint(&output, ctr.header.GroupName, "\t")
	}

	if ctr.filter.GroupID {
		_, _ = fmt.Fprint(&output, ctr.header.GroupID, "\t")
	}

	if ctr.filter.ComponentName {
		_, _ = fmt.Fprint(&output, ctr.header.ComponentName, "\t")
	}

	if ctr.filter.ComponentID {
		_, _ = fmt.Fprint(&output, ctr.header.ComponentID, "\t")
	}

	if ctr.filter.Evaluated {
		_, _ = fmt.Fprint(&output, ctr.header.Evaluated, "\t")
	}

	if ctr.filter.Status {
		_, _ = fmt.Fprint(&output, ctr.header.Status, "\t")
	}

	_, _ = fmt.Fprint(&output, nagios.CheckOutputEOL)

	return output.String()

}

// addHeaderSeparator generates a separator row between the header and data
// rows. Each "column" in the generated separator row template is of the same
// length as the header row column above it. Any columns not enabled for
// display are also omitted from the separator row.
func (ctr *componentsTable) addHeaderSeparator() {

	var headerSepRowTmpl strings.Builder

	headerTmplItems := strings.Split(ctr.headerRow(), "\t")

	// Drop the last trailing tab character from the slice.
	if len(headerTmplItems) > 0 {
		headerTmplItems = headerTmplItems[:len(headerTmplItems)-1]
	}

	for _, item := range headerTmplItems {
		headerSepRowTmpl.WriteString(strings.Repeat("-", len(item)))
		headerSepRowTmpl.WriteString("\t")
	}

	headerSepRowTmpl.WriteString(nagios.CheckOutputEOL)

	_, _ = fmt.Fprint(ctr.tabWriter, headerSepRowTmpl.String())

}

// addCollectionSeparator generates a separator row between collections of
// components. The number of "columns" in the generated separator row is of
// the same length as the header row (or any other in the table). Any columns
// not enabled for display are also omitted from the separator row.
func (ctr *componentsTable) addCollectionSeparator() {
	numFields := ctr.filter.FieldsEnabled()

	_, _ = fmt.Fprint(
		ctr.tabWriter,
		strings.Repeat("\t", numFields),
		nagios.CheckOutputEOL,
	)
}

// addRow adds a new row to the components table output. Any columns not
// enabled for display are omitted from the output.
func (ctr *componentsTable) addRow(row componentsTableRow) {

	if ctr.filter.GroupName {
		_, _ = fmt.Fprint(ctr.tabWriter, row.GroupName, "\t")
	}

	if ctr.filter.GroupID {
		_, _ = fmt.Fprint(ctr.tabWriter, row.GroupID, "\t")
	}

	if ctr.filter.ComponentName {
		_, _ = fmt.Fprint(ctr.tabWriter, row.ComponentName, "\t")
	}

	if ctr.filter.ComponentID {
		_, _ = fmt.Fprint(ctr.tabWriter, row.ComponentID, "\t")
	}

	if ctr.filter.Evaluated {
		_, _ = fmt.Fprint(ctr.tabWriter, row.Evaluated, "\t")
	}

	if ctr.filter.Status {
		_, _ = fmt.Fprint(ctr.tabWriter, row.Status, "\t")
	}

	_, _ = fmt.Fprint(ctr.tabWriter, nagios.CheckOutputEOL)

}

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

// ComponentsVerbose generates a verbose report for use as LongServiceOutput
// content, intended for display by the monitoring system either via web UI or
// one or more notifications.
//
// This report is intended to display verbose details for a feed to help
// troubleshoot final results for a statuspage feed.
func ComponentsVerbose(componentsSet *components.Set, omitOKComponents bool, verbose bool) string {

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

	if verbose {
		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

		componentsReportHeader(&report, componentsSet)

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	if omitOKComponents {
		_, _ = fmt.Fprint(
			&report,
			"NOTE: Omitting OK/operational components as requested.",
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)
	}

	if componentsSet.NumGroups() > 0 {

		_, _ = fmt.Fprintf(
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
			_, _ = fmt.Fprint(&report, printVerboseComponent(component, i+1))
			componentsEmitted = true
		}

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	if componentsSet.NumTopLevel() > 0 {

		_, _ = fmt.Fprintf(
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
			_, _ = fmt.Fprint(&report, printVerboseComponent(component, i+1))
			componentsEmitted = true
		}

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	if componentsSet.NumSubcomponents() > 0 {

		_, _ = fmt.Fprintf(
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
			_, _ = fmt.Fprint(&report, printVerboseComponent(component, i+1))
			componentsEmitted = true
		}

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)
	}

	// Evaluate all components, regardless of "ignore" or exclusion
	// status.
	if !componentsSet.IsOKState(true) {

		_, _ = fmt.Fprintf(
			&report,
			"%sComponents (%d) in a non-OK state:%s%s",
			nagios.CheckOutputEOL,
			componentsSet.NumProblemComponents(true),
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)

		for i, component := range componentsSet.ProblemComponents(true) {
			if !component.IsOKState() {
				_, _ = fmt.Fprint(&report, printVerboseComponent(component, i+1))
			}
		}

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

	}

	switch {
	case !componentsEmitted && omitOKComponents:
		_, _ = fmt.Fprint(
			&report,
			"* All components are operational. No problems to report.",
			nagios.CheckOutputEOL,
		)
	case !componentsEmitted && !omitOKComponents:
		_, _ = fmt.Fprint(
			&report,
			"* Skipping OK components was not requested, but no components were emitted. Bug?",
			nagios.CheckOutputEOL,
		)
	}

	componentsStatusSummary(&report, componentsSet, omitOKComponents)

	_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

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
func ComponentsOverview(componentsSet *components.Set, omitOKComponents bool, verbose bool) string {

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

	if verbose {
		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

		componentsReportHeader(&report, componentsSet)

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL, nagios.CheckOutputEOL)
	}

	if omitOKComponents {
		_, _ = fmt.Fprint(
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

			_, _ = fmt.Fprintf(
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

				_, _ = fmt.Fprintf(
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

						_, _ = fmt.Fprintf(
							&report,
							"\t%s%s",
							subcomponent.Name,
							nagios.CheckOutputEOL,
						)
					default:
						componentsEmitted = true

						_, _ = fmt.Fprintf(
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

	_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

	switch {
	case !componentsEmitted && omitOKComponents:
		_, _ = fmt.Fprint(
			&report,
			"* All components are operational. No problems to report.",
			nagios.CheckOutputEOL,
		)
	case !componentsEmitted && !omitOKComponents:
		_, _ = fmt.Fprint(
			&report,
			"* Skipping OK components was not requested, but no components were emitted. Bug?",
			nagios.CheckOutputEOL,
		)
	}

	if len(errsEncountered) > 0 {
		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

		_, _ = fmt.Fprintf(
			&report,
			"Errors encountered while generating this report:%s",
			nagios.CheckOutputEOL,
		)
		for i, err := range errsEncountered {
			_, _ = fmt.Fprintf(
				&report,
				"* %02d): %s%s",
				i+1,
				err,
				nagios.CheckOutputEOL,
			)
		}
	}

	_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsStatusSummary(&report, componentsSet, omitOKComponents)

	_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

	return report.String()

}

// ComponentsTable generates a report in a tabular format for use as
// LongServiceOutput content, intended for display by the monitoring system
// either via web UI or one or more notifications. If specified, only
// non-operational status components will be listed, otherwise all components
// defined for the given Statuspage will be displayed.
//
// If provided, the given columns list filter will be used to determine which
// details are emitted for applicable components. If not specified (e.g., a
// nil value is given), a default set of details are emitted for each
// applicable component.
func ComponentsTable(
	componentsSet *components.Set,
	omitOKComponents bool,
	omitResultsSummary bool,
	columnsList *ComponentsTableColumnFilter,
	verbose bool,
) string {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsTable func.\n",
			time.Since(funcTimeStart),
		)
	}()

	// If not specified, assume that all columns should be displayed.
	var chosenColumns ComponentsTableColumnFilter
	switch {
	case columnsList != nil:
		chosenColumns = *columnsList
	default:
		chosenColumns = ComponentsTableColumnFilter{
			GroupName:     true,
			GroupID:       true,
			ComponentName: true,
			ComponentID:   true,
			Evaluated:     true,
			Status:        true,
		}
	}

	componentsTable := newComponentsTable(chosenColumns)

	// w := tabwriter.NewWriter(&report, 4, 4, 4, ' ', 0)

	// A collection of errors (if any) encountered while generating this
	// report.
	var errsEncountered []error

	if verbose {
		_, _ = fmt.Fprint(&componentsTable.report, nagios.CheckOutputEOL)

		componentsReportHeader(&componentsTable.report, componentsSet)

		_, _ = fmt.Fprint(&componentsTable.report, nagios.CheckOutputEOL)
	}

	if omitOKComponents {
		_, _ = fmt.Fprint(
			&componentsTable.report,
			"NOTE: Omitting OK/operational components as requested.",
			nagios.CheckOutputEOL,
			nagios.CheckOutputEOL,
		)
	}

	switch {
	case componentsSet.NumGroups() > 0:

		headerRow := componentsTableRow{
			GroupName:     "GROUP NAME",
			GroupID:       "GROUP ID",
			ComponentName: "COMPONENT NAME",
			ComponentID:   "COMPONENT ID",
			Evaluated:     "EVALUATED",
			Status:        "STATUS",
		}

		componentsTable.addHeaderRow(headerRow)

	default:

		headerRow := componentsTableRow{
			ComponentName: "COMPONENT NAME",
			ComponentID:   "COMPONENT ID",
			Evaluated:     "EVALUATED",
			Status:        "STATUS",
		}

		componentsTable.addHeaderRow(headerRow)
	}

	componentsTable.addHeaderSeparator()

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

				componentsTable.addRow(componentsTableRow{
					// Empty group name and id column values specified here
					// since this is a top-level component and there are
					// groups defined.
					GroupName:     "",
					GroupID:       "",
					ComponentName: component.Name,
					ComponentID:   component.ID,
					Evaluated:     evaluated,
					Status:        printStatus(component.Status),
				})
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

				componentsTable.addRow(componentsTableRow{
					ComponentName: component.Name,
					ComponentID:   component.ID,
					Evaluated:     evaluated,
					Status:        printStatus(component.Status),
				})
			}
		}

		if separatorRowNeeded {
			componentsEmitted = true
			componentsTable.addCollectionSeparator()
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

					componentsTable.addRow(componentsTableRow{
						GroupName:     group.Parent.Name,
						GroupID:       group.Parent.ID,
						ComponentName: subcomponent.Name,
						ComponentID:   subcomponent.ID,
						Evaluated:     evaluated,
						Status:        printStatus(subcomponent.Status),
					})

				}

				if separatorRowNeeded {
					componentsTable.addCollectionSeparator()
				}

			}

			if separatorRowNeeded {
				componentsEmitted = true
			}
		}
	}

	if !componentsEmitted {
		componentsTable.addRow(componentsTableRow{
			GroupName:     "N/A",
			GroupID:       "N/A",
			ComponentName: "N/A",
			ComponentID:   "N/A",
			Evaluated:     "N/A",
			Status:        "N/A",
		})
	}

	_, _ = fmt.Fprint(&componentsTable.report, nagios.CheckOutputEOL)

	if err := componentsTable.tabWriter.Flush(); err != nil {
		errsEncountered = append(errsEncountered, err)
	}

	if len(errsEncountered) > 0 {
		_, _ = fmt.Fprint(&componentsTable.report, nagios.CheckOutputEOL)

		_, _ = fmt.Fprintf(
			&componentsTable.report,
			"Errors encountered while generating this report:%s",
			nagios.CheckOutputEOL,
		)
		for i, err := range errsEncountered {
			_, _ = fmt.Fprintf(
				&componentsTable.report,
				"* %02d): %s%s",
				i+1,
				err,
				nagios.CheckOutputEOL,
			)
		}
	}

	_, _ = fmt.Fprint(&componentsTable.report, nagios.CheckOutputEOL)

	if !omitResultsSummary {
		componentsStatusSummary(&componentsTable.report, componentsSet, omitOKComponents)
	}

	_, _ = fmt.Fprint(&componentsTable.report, nagios.CheckOutputEOL)

	return componentsTable.report.String()

}

// ComponentsIDList generates a multi-column list of component IDs. This
// multi-column list can be used to populate component ID fields in tests and
// other batch processing tasks.
func ComponentsIDList(componentsSet *components.Set, verbose bool) string {

	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsIDList func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder

	columnsPerRow := 4

	if verbose {
		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

		componentsReportHeader(&report, componentsSet)

		_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL, nagios.CheckOutputEOL)
	}

	notExcludedComponents := componentsSet.NotExcludedComponents()
	notExcluded := make([]string, 0, len(notExcludedComponents))

	for _, component := range notExcludedComponents {
		notExcluded = append(notExcluded, component.ID)
	}

	for i := 0; i < len(notExcluded); {
		for j := 0; j < columnsPerRow && i < len(notExcluded); j++ {
			_, _ = fmt.Fprintf(&report, `"%s", `, notExcluded[i])
			i++
		}
		_, _ = fmt.Fprintln(&report)
	}

	_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

	componentsStatusSummary(&report, componentsSet, true)

	_, _ = fmt.Fprint(&report, nagios.CheckOutputEOL)

	return report.String()

}

// componentsStatusSummary generates a brief summary of high-level component
// details. This summary is written to the provided io.Writer.
func componentsStatusSummary(
	w io.Writer,
	componentsSet *components.Set,
	omitOKComponents bool,
) {
	_, _ = fmt.Fprintf(
		w,
		"%sSummary:%s%s",
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Page: %s (%s)%s",
		componentsSet.Page.Name,
		componentsSet.Page.URL,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Last Updated (%s): %s%s",
		componentsSet.Page.TimeZone,
		componentsSet.Page.UpdatedAt.Format(time.RFC3339),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Last Updated (%s): %s%s",
		"Local",
		componentsSet.Page.UpdatedAt.Local().Format(time.DateTime+" PM"),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Filtering applied to components set: %t%s",
		componentsSet.FilterApplied,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Evaluating all components in the set: %t%s",
		componentsSet.EvalAllComponents,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Omitting OK/operational components (if requested): %t%s",
		omitOKComponents,
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Number of total top-level components: %d%s",
		componentsSet.NumTopLevel(),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Number of total component groups: %d%s",
		componentsSet.NumGroups(),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Number of total subcomponents: %d%s",
		componentsSet.NumSubcomponents(),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Number of total problem components: %d%s",
		componentsSet.NumProblemComponents(true),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Number of ignored problem components: %d%s",
		componentsSet.NumProblemComponents(true)-componentsSet.NumProblemComponents(false),
		nagios.CheckOutputEOL,
	)

	_, _ = fmt.Fprintf(
		w,
		"* Number of remaining problem components: %d%s",
		componentsSet.NumProblemComponents(false),
		nagios.CheckOutputEOL,
	)

}

func componentsReportHeader(w io.Writer, componentsSet *components.Set) {
	_, _ = fmt.Fprintf(
		w,
		"%s (%s)%s",
		componentsSet.Page.Name,
		componentsSet.Page.URL,
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
	_ string,
	filter components.Filter,
	componentsSet *components.Set,
	omitOKComponents bool,
	omitSummaryResults bool,
	verbose bool,
) string {
	funcTimeStart := time.Now()

	defer func() {
		logger.Printf(
			"It took %v to execute ComponentsReport func.\n",
			time.Since(funcTimeStart),
		)
	}()

	var report strings.Builder

	// TODO: Move this to the bottom or put behind a verbose flag?
	switch {
	case !componentsSet.EvalAllComponents:
		if verbose {
			_, _ = fmt.Fprintf(
				&report,
				"Specified filter: %s%s",
				filter,
				nagios.CheckOutputEOL,
			)
		}
	default:
		if verbose {
			_, _ = fmt.Fprintf(
				&report,
				"NOTE: Evaluating all components as requested.%s",
				nagios.CheckOutputEOL,
			)
		}
	}

	// Skip emitting ID values in report in order to generate less "noisy"
	// output for quick review.
	columnFilter := ComponentsTableColumnFilter{
		GroupName:     true,
		GroupID:       false,
		ComponentName: true,
		ComponentID:   false,
		Evaluated:     false,
		Status:        true,
	}

	// Disable Group fields if there are no component groups to display.
	if componentsSet.NumGroups() == 0 {
		columnFilter.GroupID = false
		columnFilter.GroupName = false
	}

	fullTableOutputComponentsLimit := 50
	switch {

	case omitOKComponents:
		_, _ = fmt.Fprint(&report, ComponentsTable(componentsSet, true, omitSummaryResults, &columnFilter, verbose))

	case componentsSet.NumComponents() > fullTableOutputComponentsLimit:

		_, _ = fmt.Fprintf(
			&report,
			"NOTE: Component count (%d) is higher than display limit (%d);"+
				" overriding default display of OK components.%s",
			componentsSet.NumComponents(),
			fullTableOutputComponentsLimit,
			nagios.CheckOutputEOL,
		)

		_, _ = fmt.Fprint(&report, ComponentsTable(componentsSet, true, omitSummaryResults, &columnFilter, verbose))

	default:
		_, _ = fmt.Fprint(&report, ComponentsTable(componentsSet, false, omitSummaryResults, &columnFilter, verbose))
	}

	return report.String()
}
