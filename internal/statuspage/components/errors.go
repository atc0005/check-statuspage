// Copyright 2021 Adam Chalkley
//
// https://github.com/atc0005/check-statuspage
//
// Licensed under the MIT License. See LICENSE file in the project root for
// full license information.

package components

import "errors"

// ErrComponentSetFilterEmpty indicates that a given components set filter is
// empty.
var ErrComponentSetFilterEmpty = errors.New(
	"given component set filter is empty",
)

// ErrComponentSetFilterWhitespaceGroupField indicates that a given components
// set filter contains a whitespace only group field value.
var ErrComponentSetFilterWhitespaceGroupField = errors.New(
	"given component set filter contains whitespace only group value",
)

// ErrComponentSetFilterWhitespaceComponentsField indicates that a given
// components set filter contains a whitespace only components list field
// value.
var ErrComponentSetFilterWhitespaceComponentsField = errors.New(
	"given component set filter contains whitespace only components list value",
)

// ErrComponentSetFilterInvalid indicates that a given components set filter
// is in an unknown or invalid state. This error condition is unlikely to
// occur.
var ErrComponentSetFilterInvalid = errors.New(
	"given component set filter is invalid",
)

// ErrComponentSetValidationFailed indicates that validating decode JSON data
// has failed.
var ErrComponentSetValidationFailed = errors.New(
	"decoded components endpoint JSON data validation failed",
)

// var ErrURLNotParsed = errors.New(
// 	"failed to parse given path as URL",
// )

// ErrSubcomponentNotFound indicates that a subcomponent was not found when
// attempting to apply a given filter to a components set or when searching
// by name or id values.
var ErrSubcomponentNotFound = errors.New(
	"subcomponent not found",
)

// ErrComponentGroupNotFound indicates that a component group was not found
// when attempting to apply a given filter to a components set or when
// searching by name or id values.
var ErrComponentGroupNotFound = errors.New(
	"component group not found",
)

// ErrComponentIsNotValidSubcomponent indicates that a specified component was
// found to not be a member of a specified component group when attempting to
// apply a given filter to a components set.
var ErrComponentIsNotValidSubcomponent = errors.New(
	"component is not a member of specified group",
)

// ErrComponentNotFound indicates that a component was not found when
// attempting to apply a given filter to a components set or when searching by
// name or id values.
var ErrComponentNotFound = errors.New(
	"component not found",
)

// ErrComponentIsNotComponentGroup indicates that a specified component was
// found to not be a component group when attempting to apply a given filter
// to a components set.
var ErrComponentIsNotComponentGroup = errors.New(
	"component is not a component group",
)

// ErrComponentStatusDegradedPerformance indicates that a component was found
// to have a degraded performance status.
var ErrComponentStatusDegradedPerformance = errors.New(
	"component has degraded performance status",
)

// ErrComponentStatusPartialOutage indicates that a component was found to
// have a partial outage status.
var ErrComponentStatusPartialOutage = errors.New(
	"component has partial outage status",
)

// ErrComponentStatusMajorOutage indicates that a component was found to have
// a major outage status.
var ErrComponentStatusMajorOutage = errors.New(
	"component has major outage status",
)

// ErrComponentStatusUnderMaintenance indicates that a component was found to
// have an under maintenance status.
var ErrComponentStatusUnderMaintenance = errors.New(
	"component has under maintenance status",
)

// ErrComponentWithProblemStatusNotExcluded indicates that a component with a
// non-operational status was not excluded from evaluation. This is a
// user-facing error, intended for display in detailed output.
var ErrComponentWithProblemStatusNotExcluded = errors.New(
	"component with non-operational status not excluded from evaluation",
)
