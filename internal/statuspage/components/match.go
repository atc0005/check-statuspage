// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package components

import (
	"fmt"
	"strings"
)

// componentIsValidSubcomponent is a helper function intended to evaluate one
// or more components to determine if they are a subcomponent of a collection
// of groups in an index of component ID to Component.
func componentIsValidSubcomponent(
	groupIndex map[string]*Component,
	components ...*Component,
) bool {
	for _, component := range components {
		_, ok := groupIndex[string(component.GroupID)]
		if ok {
			return true
		}
	}

	return false
}

// matchGroupComponents evaluates components in the set using a given Filter.
// If a component group match is made, it is recorded in the given index of
// component ID to Component.
func (cs *Set) matchGroupComponents(filter Filter, matches map[string]*Component) error {

	// Attempt to retrieve by name first.
	components, err := cs.GetComponentsByName(filter.Group)
	switch {
	case err == nil:

		// This can return one or many matches which can include top-level
		// components (ignored), subcomponents (also ignored) or component
		// groups (what we're looking for). As long as we find the group we're
		// after from the retrieval or "search" results, we're ok.
		componentsListHasComponentGroup := func(components []*Component) bool {
			for _, component := range components {
				if component.Group {
					return true
				}
			}
			return false
		}

		// Indicate that at least one component group was not found.
		if !componentsListHasComponentGroup(components) {
			return fmt.Errorf(
				"failed to apply filter '%s' to components set; "+
					"no matching component groups found: %w",
				filter,
				ErrComponentIsNotComponentGroup,
			)
		}

		for _, component := range components {

			logger.Printf("Evaluating component as group: %s", component)
			switch {

			case component.Group:
				logger.Printf("Component is component group: %s", component)
				matches[component.ID] = component
				continue

			default:
				logger.Printf("Component is NOT component group: %s", component)
			}

		}

		return nil

	default:
		// Fallback to trying to retrieve by ID.
		component, err := cs.GetComponentByID(filter.Group)
		if err != nil {
			return fmt.Errorf(
				"failed to apply filter '%s' to components set: %w",
				filter,
				ErrComponentGroupNotFound,
			)
		}

		logger.Printf("Evaluating component as group: %s", component)
		if !component.Group {
			return fmt.Errorf(
				"failed to apply filter '%s' to components set: %w",
				filter,
				ErrComponentIsNotComponentGroup,
			)
		}
		matches[component.ID] = component
	}

	return nil
}

// matchComponents evaluates and then matches components in the set using a
// given Filter.
func (cs *Set) matchComponents(
	filter Filter,
	groupIndex map[string]*Component,
	componentMatches map[string]*Component,
) error {

	logger.Printf("Evaluating %d filter.Components", len(filter.Components))

	for _, compSearchVal := range filter.Components {

		logger.Printf("Retrieving component(s) for component search value: %q", compSearchVal)
		components, err := cs.retrieveComponentByNameOrID(filter, compSearchVal)
		if err != nil {
			return err
		}
		logger.Printf(
			"Successfully retrieved %d component(s) for search value %q",
			len(components),
			compSearchVal,
		)

		for _, component := range components {
			logger.Printf(
				"Assert that search value %q is not a component group ID",
				compSearchVal,
			)
			if component.Group && strings.EqualFold(compSearchVal, component.ID) {
				return fmt.Errorf(
					"given component search value %q is an ID value for component group: %s",
					compSearchVal,
					component,
				)
			}
		}
		logger.Printf(
			"Successfully asserted that search value %q is not a component group ID",
			compSearchVal,
		)

		isAllGroupComponents := func(components []*Component) bool {
			for _, component := range components {
				if !component.Group {
					return false
				}
			}
			return true
		}

		logger.Printf(
			"Assert that at least one retrieved component is not a group for search value: %q",
			compSearchVal,
		)
		if isAllGroupComponents(components) {
			return fmt.Errorf(
				"all retrieved components for given search value %q are component groups",
				compSearchVal,
			)
		}
		logger.Printf(
			"Successfully asserted that at least one retrieved component is not a group for search value: %q",
			compSearchVal,
		)

		logger.Printf("Retrieving subcomponents for %d groups", len(groupIndex))
		// NOTE: You can have multiple groups if retrieving by name.
		err = matchComponents(filter, groupIndex, componentMatches, components...)
		if err != nil {
			return err
		}
		logger.Printf(
			"Successfully retrieved %d subcomponents for %d groups",
			len(componentMatches),
			len(groupIndex),
		)
	}

	return nil
}

// matchComponents is a helper function used to perform the bulk of the
// evaluation and matching logic for the Set.matchComponents() method.
func matchComponents(
	filter Filter,
	groupIndex map[string]*Component,
	componentsIndex map[string]*Component,
	components ...*Component,
) error {

	switch {

	// If a group was specified, assert that each component is a valid
	// subcomponent.
	case filter.Group != "":

		logger.Print("Component group was specified")

		logger.Print("Asserting that retrieved components contain at least one valid subcomponent")
		for _, component := range components {
			logger.Printf("Component: %s", component)
		}
		if !componentIsValidSubcomponent(groupIndex, components...) {
			return fmt.Errorf(
				"failed to apply filter to components set: %w",
				ErrComponentIsNotValidSubcomponent,
			)
		}

		logger.Print("Confirmed that retrieved components contain at least one valid subcomponent")
		for _, component := range components {
			if componentIsValidSubcomponent(groupIndex, component) {
				logger.Printf("Recording component as subcomponent match: %s", component)
				componentsIndex[component.ID] = component
			}
			logger.Printf("Ignoring component; not a valid subcomponent: %s", component)
		}

		return nil

	// If a group was not specified, record each matched component
	// as-is.
	default:

		logger.Print("Component group was not specified, recording each matched component as-is")

		for _, component := range components {
			if component.Group {
				logger.Printf("Ignoring group component: %s", component)
				continue
			}
			logger.Printf("Recording component as component match: %s", component)
			componentsIndex[component.ID] = component
		}

		return nil

	}
}

func (cs *Set) excludeUnmatchedComponents(matchedComponents map[string]*Component) {

	for k := range matchedComponents {
		logger.Printf("Matched Component will not be excluded: %s", matchedComponents[k])
	}

	// Evaluate all components, mark any not in the matched components index
	// as excluded.
	for i := range cs.Components {
		if _, ok := matchedComponents[cs.Components[i].ID]; !ok {
			cs.Components[i].Exclude = true
		}
	}

	// If we reach this point, indicate that we have successfully filtered the
	// Component Set.
	cs.FilterApplied = true
}

// recordSubcomponents retrieves all subcomponents for each group in the group
// index and records them in the given components index. An error is returned
// if one is encountered when retrieving subcomponents.
func (cs *Set) recordSubcomponents(
	filter Filter,
	groupIndex map[string]*Component,
	componentsIndex map[string]*Component,
) error {
	logger.Print("No components explicitly listed, evaluating all components for a group")

	// gather ALL components for each given group (id)
	for groupID := range groupIndex {

		logger.Printf("Retrieving ComponentGroup for group id %q", groupID)

		componentGroup, err := cs.GetGroupByID(groupID)
		if err != nil {
			return fmt.Errorf(
				"failed to apply filter '%s' to components set: %w",
				filter,
				ErrComponentGroupNotFound,
			)
		}

		logger.Printf(
			"Successfully retrieved ComponentGroup %s for group id %q:",
			componentGroup,
			groupID,
		)

		for _, component := range componentGroup.Subcomponents {
			logger.Printf("Recording component as component match: %s", component)
			componentsIndex[component.ID] = component
		}
	}

	return nil
}

func (cs *Set) retrieveComponentByNameOrID(filter Filter, searchKey string) ([]*Component, error) {

	logger.Printf("Attempt to get components by specified name first: %q", searchKey)
	components, err := cs.GetComponentsByName(searchKey)
	if err != nil {
		logger.Printf("Error occurred searching for component by name: %q", searchKey)
		logger.Printf("Fall back to retrieving component by ID: %q", searchKey)
		component, err := cs.GetComponentByID(searchKey)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to retrieve components using filter '%s': %w",
				filter,
				ErrComponentNotFound,
			)
		}

		logger.Printf("Retrieved component using ID value: %q", searchKey)

		return []*Component{component}, nil

	}

	return components, nil

}
