<!-- omit in toc -->

# check-statuspage

Go-based tooling to interact with status page APIs hosted by Atlassian
Statuspage; NOT affiliated with or endorsed by Atlassian.

[![Latest Release](https://img.shields.io/github/release/atc0005/check-statuspage.svg?style=flat-square)](https://github.com/atc0005/check-statuspage/releases/latest)
[![Go Reference](https://pkg.go.dev/badge/github.com/atc0005/check-statuspage.svg)](https://pkg.go.dev/github.com/atc0005/check-statuspage)
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/atc0005/check-statuspage)](https://github.com/atc0005/check-statuspage)
[![Lint and Build](https://github.com/atc0005/check-statuspage/actions/workflows/lint-and-build.yml/badge.svg)](https://github.com/atc0005/check-statuspage/actions/workflows/lint-and-build.yml)
[![Project Analysis](https://github.com/atc0005/check-statuspage/actions/workflows/project-analysis.yml/badge.svg)](https://github.com/atc0005/check-statuspage/actions/workflows/project-analysis.yml)
[![Push Validation](https://github.com/atc0005/check-statuspage/actions/workflows/push-validation.yml/badge.svg)](https://github.com/atc0005/check-statuspage/actions/workflows/push-validation.yml)

<!-- omit in toc -->
## Table of Contents

- [check-statuspage](#check-statuspage)
  - [Project home](#project-home)
  - [Overview](#overview)
    - [Output](#output)
    - [Performance Data](#performance-data)
      - [`check_statuspage_components` Plugin](#check_statuspage_components-plugin)
        - [NOTES](#notes)
    - [`check_statuspage_components`](#check_statuspage_components)
    - [`lscs`](#lscs)
  - [Features](#features)
  - [Changelog](#changelog)
  - [Requirements](#requirements)
    - [Building source code](#building-source-code)
    - [Running](#running)
  - [Installation](#installation)
    - [From source](#from-source)
    - [Using provided binaries](#using-provided-binaries)
      - [Linux](#linux)
      - [Other operating systems](#other-operating-systems)
  - [Configuration options](#configuration-options)
    - [Threshold calculations](#threshold-calculations)
      - [`check_statuspage_components`](#check_statuspage_components-1)
    - [Command-line arguments](#command-line-arguments)
      - [`check_statuspage_components`](#check_statuspage_components-2)
      - [`lscs`](#lscs-1)
    - [Configuration file](#configuration-file)
  - [Examples](#examples)
    - [`check_statuspage_components` Nagios plugin](#check_statuspage_components-nagios-plugin)
      - [CLI invocations](#cli-invocations)
        - [Evaluate all subcomponents in a group](#evaluate-all-subcomponents-in-a-group)
        - [Evaluate a specific top-level component](#evaluate-a-specific-top-level-component)
      - [Command definition](#command-definition)
    - [`lscs` CLI app](#lscs-cli-app)
      - [CLI invocation](#cli-invocation)
        - [The `table` format (default)](#the-table-format-default)
        - [The `overview` format](#the-overview-format)
        - [The `verbose` format](#the-verbose-format)
        - [The `list` format](#the-list-format)
        - [Other supported formats](#other-supported-formats)
  - [License](#license)
  - [References](#references)

## Project home

See [our GitHub repo][repo-url] for the latest code, to file an issue or
submit improvements for review and potential inclusion into the project.

Just to be 100% clear: this project is not affiliated with or endorsed by
Atlassian.

## Overview

This repo contains various tools and plugins used to monitor status pages
powered by Atlassian Statuspage.

| Plugin or Tool Name           | Description                                                  |
| ----------------------------- | ------------------------------------------------------------ |
| `lscs`                        | a CLI app to list `components` in multiple output formats.   |
| `check_statuspage_components` | Nagios plugin used to monitor one, many or all `components`. |

### Output

The output for plugins in this project is designed to provide the one-line
summary needed by Nagios (and other monitoring systems) for quick
identification of a problem while providing longer, more detailed information
for display within the web UI, use in email and Teams notifications
([atc0005/send2teams](https://github.com/atc0005/send2teams)).

By default, output intended for processing by Nagios is sent to `stdout` and
output intended for troubleshooting by the sysadmin is sent to `stderr`.
Output emitted to `stderr` is  configurable via the `--log-level` flag.

For some monitoring systems or addons (e.g., Icinga Web 2, Nagios XI), the
`stderr` output is mixed in with the `stdout` output in the web UI for the
service check. This may add visual noise when viewing the service check
output. For those cases, you may wish to explicitly disable the output to
`stderr` by using the `--log-level "disabled"` CLI flag & value.

### Performance Data

Initial support has been added for emitting Performance Data / Metrics, but
refinement suggestions are welcome.

Consult the tables below for the metrics implemented thus far.

Please add to an existing
[Discussion](https://github.com/atc0005/check-statuspage/discussions) thread
(if applicable) or [open a new
one](https://github.com/atc0005/check-statuspage/discussions/new) with any
feedback that you may have. Thanks in advance!

#### `check_statuspage_components` Plugin

| Emitted Performance Data / Metric | Meaning                                                                           |
| --------------------------------- | --------------------------------------------------------------------------------- |
| `time`                            | Runtime for plugin                                                                |
| `all_component_groups`            | Number of non-excluded component Groups                                           |
| `all_components`                  | Number of non-excluded components `*`                                             |
| `all_components_critical`         | Number of non-excluded components in a `CRITICAL` state                           |
| `all_components_ok`               | Number of non-excluded components in an `OK` state                                |
| `all_components_unknown`          | Number of non-excluded components in an `UNKNOWN` state                           |
| `all_components_warning`          | Number of non-excluded components in a `WARNING` state                            |
| `all_problem_components`          | Number of non-excluded components in a "problem" (non-`OK`) state                 |
| `excluded_problem_components`     | Number of excluded components (all states)                                        |
| `remaining_components_critical`   | Number of components in a `CRITICAL` state remaining *after* exclusions           |
| `remaining_components_ok`         | Number of components in an `OK` state remaining *after* exclusions                |
| `remaining_components_unknown`    | Number of components in an `UNKNOWN` state remaining *after* exclusions           |
| `remaining_components_warning`    | Number of components in a `WARNING` state remaining *after* exclusions            |
| `remaining_problem_components`    | Number of components in a "problem" (non-`OK`) state remaining *after* exclusions |

##### NOTES

- the `all_components` metric (`*`)
  - includes top-level components (not part of a component Group)
  - includes subcomponents (part of a component Group)
  - *includes* component Groups
- the `*_components_STATE` metrics (where `STATE` is `critical`, `ok`,
  `unknown`, `warning`)
  - includes top-level components (not part of a component Group)
  - includes subcomponents (part of a component Group)
  - *excludes* component Groups
- subcomponents are not currently reported as independent values
- top-level / standalone components (those outside of a component Group) are
  not currently reported as independent values

### `check_statuspage_components`

Nagios plugin used to monitor the status of one, many or all `components`
(aka, "services") of a Statuspage powered site. See the [configuration
options](#configuration-options) section for details regarding supported flags
and values.

### `lscs`

Small CLI app used to generate an overview of `components` (aka, "services")
of a Statuspage powered site. This tool can be useful for quickly inspecting a
new Statuspage powered site to retrieve `component` names or IDs for
monitoring via the `check_statuspage_components` plugin.

## Features

- Plugin for monitoring an Atlassian Statuspage powered site
  - the status of `components` (aka, "services") specified by one or many
    top-level components, component groups (all subcomponents) or component
    group and subcomponents

- CLI app to list `components` from an Atlassian Statuspage powered site
  - multiple output formats
  - `overview`
  - `table`
  - `verbose`
  - `debug`
  - `list` (mostly used to assist with crafting test cases)
  - `json`

- User-specified input sources
  - local file
    - useful for testing
    - useful to help prevent hitting API limits (e.g., refresh via a cron job)
  - remote URL
    - e.g., <https://status.linode.com/api/v2/components.json>
    - most common scenario

- Optional, leveled logging using `rs/zerolog` package
  - JSON-format output (to `stderr`)
  - choice of `disabled`, `panic`, `fatal`, `error`, `warn`, `info` (the
    default), `debug` or `trace`.

- Optional, user-specified timeout value for plugin execution

- Optional, user-specified read limit
  - helps protect against excessive/unexpected input size)

- Optional override to allow unknown JSON fields in input source
  - future proofing
  - could be useful as a hotfix until plugins in this project can be updated
    to handle future JSON input changes

- Optional support for omitting `components` in an `OK` state (aka,
  `operational` status)
  - help focus on just the `components` with a "problem" status

- Optional support for omitting summary in results output

## Changelog

See the [`CHANGELOG.md`](CHANGELOG.md) file for the changes associated with
each release of this application. Changes that have been merged to `master`,
but not yet an official release may also be noted in the file under the
`Unreleased` section. A helpful link to the Git commit history since the last
official release is also provided for further review.

## Requirements

The following is a loose guideline. Other combinations of Go and operating
systems for building and running tools from this repo may work, but have not
been tested.

### Building source code

- Go
  - see this project's `go.mod` file for *preferred* version
  - this project tests against [officially supported Go
    releases][go-supported-releases]
    - the most recent stable release (aka, "stable")
    - the prior, but still supported release (aka, "oldstable")
- GCC
  - if building with custom options (as the provided `Makefile` does)
- `make`
  - if using the provided `Makefile`

### Running

- Windows 10
- Ubuntu Linux 18.04+

## Installation

### From source

1. [Download][go-docs-download] Go
1. [Install][go-docs-install] Go
   - NOTE: Pay special attention to the remarks about `$HOME/.profile`
1. Clone the repo
   1. `cd /tmp`
   1. `git clone https://github.com/atc0005/check-statuspage`
   1. `cd check-statuspage`
1. Install dependencies (optional)
   - for Ubuntu Linux
     - `sudo apt-get install make gcc`
   - for CentOS Linux
     - `sudo yum install make gcc`
   - for Windows
     - Emulated environments (*easier*)
       - Skip all of this and build using the default `go build` command in
         Windows (see below for use of the `-mod=vendor` flag)
       - build using Windows Subsystem for Linux Ubuntu environment and just
         copy out the Windows binaries from that environment
       - If already running a Docker environment, use a container with the Go
         tool-chain already installed
       - If already familiar with LXD, create a container and follow the
         installation steps given previously to install required dependencies
     - Native tooling (*harder*)
       - see the StackOverflow Question `32127524` link in the
         [References](references.md) section for potential options for
         installing `make` on Windows
       - see the mingw-w64 project homepage link in the
         [References](references.md) section for options for installing `gcc`
         and related packages on Windows
1. Build binaries
   - for the current operating system, explicitly using bundled dependencies
         in top-level `vendor` folder
     - `go build -mod=vendor ./cmd/check_statuspage_components/`
     - `go build -mod=vendor ./cmd/lscs/`
   - for all supported platforms (where `make` is installed)
      - `make all`
   - for use on Windows
      - `make windows`
   - for use on Linux
     - `make linux`
1. Copy the newly compiled binary from the applicable `/tmp` subdirectory path
   (based on the clone instructions in this section) below and deploy where
   needed.
   - if using `Makefile`
     - look in `/tmp/check-statuspage/release_assets/check_statuspage_components/`
     - look in `/tmp/check-statuspage/release_assets/lscs/`
   - if using `go build`
     - look in `/tmp/check-statuspage/`
1. Review [configuration options](#configuration-options) and
   [examples](#examples) sections usage details.

### Using provided binaries

#### Linux

1. Download plugins from the [Latest
   release](https://github.com/atc0005/check-statuspage/releases/latest) that you
   are interested in
1. Review [configuration options](#configuration-options) and
   [examples](#examples) sections usage details.

#### Other operating systems

Binaries for Windows are currently available, but may be discontinued if there
isn't sufficient interest.

If you would benefit from precompiled binaries for other platforms, please let
us know by opening a new issue or responding to an existing issue with an
up-vote. See <https://golang.org/doc/install/source> for a list of supported
architectures and operating systems.

## Configuration options

### Threshold calculations

#### `check_statuspage_components`

This table lists equivalent Nagios plugin states and Statuspage status values.
Though these are not technically *thresholds*, the plugin lists them as such
for quick reference.

| Nagios State | Statuspage Status                                             |
| ------------ | ------------------------------------------------------------- |
| `OK`         | `operational`                                                 |
| `WARNING`    | `under_maintenance`, `partial_outage`, `degraded_performance` |
| `CRITICAL`   | `major_outage`                                                |

### Command-line arguments

- Use the `-h` or `--help` flag to display current usage information.
- Flags marked as **`required`** must be set via CLI flag.
- Flags *not* marked as required are for settings where a useful default is
  already defined, but may be overridden if desired.

#### `check_statuspage_components`

| Flag                          | Required  | Default   | Repeat | Possible                                                                | Description                                                                                                                                                                                                                                    |
| ----------------------------- | --------- | --------- | ------ | ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `branding`                    | No        | `false`   | No     | `branding`                                                              | Toggles emission of branding details with plugin status details. This output is disabled by default.                                                                                                                                           |
| `h`, `help`                   | No        | `false`   | No     | `h`, `help`                                                             | Show Help text along with the list of supported flags.                                                                                                                                                                                         |
| `v`, `version`                | No        | `false`   | No     | `v`, `version`                                                          | Whether to display application version and then immediately exit application.                                                                                                                                                                  |
| `ll`, `log-level`             | No        | `info`    | No     | `disabled`, `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace` | Log message priority filter. Log messages with a lower level are ignored. Log messages are sent to `stderr` by default. See [Output](#output) for more information.                                                                            |
| `t`, `timeout`                | No        | `10`      | No     | *positive whole number of seconds*                                      | Timeout value in seconds allowed before a plugin execution attempt is abandoned and an error returned.                                                                                                                                         |
| `f`, `filename`               | **Maybe** |           | No     | *fully-qualified path to a Statuspage components JSON file*             | The fully-qualified filename of a previously downloaded Statuspage API/JSON feed (e.g., /tmp/statuspage/github/components.json). This option is incompatible with the `--url` flag.                                                            |
| `u`, `url`                    | **Maybe** |           | No     | *valid https URL*                                                       | The fully-qualified URL of a Statuspage API/JSON feed (e.g., <https://www.githubstatus.com/api/v2/components.json>)..                                                                                                                          |
| `g`, `group`                  | **Maybe** |           | No     | *valid name or ID value of component group*                             | A single name or ID value for a component group. Can be used by itself or with the flag to specify a list of components. If used with the components flag all specified components are required to be subcomponents of the group.              |
| `c`, `component`              | **Maybe** |           | No     | *valid name or ID value of component*                                   | One or more comma-separated component (name or ID) values. Can be used by itself or with the flag to specify a component group. If used with the component group flag, all specified components are required to be subcomponents of the group. |
| `ea`, `eval-all`              | **Maybe** | `false`   | No     | `true`, `false`                                                         | Whether all components should be evaluated. Incompatible with flag to specify list of components, component group or component group set.                                                                                                      |
| `ook`, `omit-ok`              | No        | `false`   | No     | `true`, `false`                                                         | Whether listed components in results output should be limited to just those in a non-operational state.                                                                                                                                        |
| `os`, `omit-summary`          | No        | `false`   | No     | `true`, `false`                                                         | Whether summary in results output should be omitted.                                                                                                                                                                                           |
| `rl`, `read-limit`            | No        | `1048576` | No     | *valid whole number of bytes*                                           | Limit in bytes used to help prevent abuse when reading input that could be larger than expected. The default value is nearly 4x the largest observed (formatted) feed size.                                                                    |
| `auf`, `allow-unknown-fields` | No        | `false`   | No     | `true`, `false`                                                         | Whether unknown JSON fields encountered while decoding JSON data should be ignored.                                                                                                                                                            |

#### `lscs`

| Flag                          | Required  | Default   | Repeat | Possible                                                                | Description                                                                                                                                                                                                                                    |
| ----------------------------- | --------- | --------- | ------ | ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `h`, `help`                   | No        | `false`   | No     | `h`, `help`                                                             | Show Help text along with the list of supported flags.                                                                                                                                                                                         |
| `v`, `version`                | No        | `false`   | No     | `v`, `version`                                                          | Whether to display application version and then immediately exit application.                                                                                                                                                                  |
| `ll`, `log-level`             | No        | `info`    | No     | `disabled`, `panic`, `fatal`, `error`, `warn`, `info`, `debug`, `trace` | Log message priority filter. Log messages with a lower level are ignored. Log messages are sent to `stderr` by default. See [Output](#output) for more information.                                                                            |
| `t`, `timeout`                | No        | `10`      | No     | *positive whole number of seconds*                                      | Timeout value in seconds allowed before a plugin execution attempt is abandoned and an error returned.                                                                                                                                         |
| `f`, `filename`               | **Maybe** |           | No     | *fully-qualified path to a Statuspage components JSON file*             | The fully-qualified filename of a previously downloaded Statuspage API/JSON feed (e.g., /tmp/statuspage/github/components.json). This option is incompatible with the `--url` flag.                                                            |
| `u`, `url`                    | **Maybe** |           | No     | *valid https URL*                                                       | The fully-qualified URL of a Statuspage API/JSON feed (e.g., <https://www.githubstatus.com/api/v2/components.json>)..                                                                                                                          |
| `g`, `group`                  | **Maybe** |           | No     | *valid name or ID value of component group*                             | A single name or ID value for a component group. Can be used by itself or with the flag to specify a list of components. If used with the components flag all specified components are required to be subcomponents of the group.              |
| `c`, `component`              | **Maybe** |           | No     | *valid name or ID value of component*                                   | One or more comma-separated component (name or ID) values. Can be used by itself or with the flag to specify a component group. If used with the component group flag, all specified components are required to be subcomponents of the group. |
| `ook`, `omit-ok`              | No        | `false`   | No     | `true`, `false`                                                         | Whether listed components in results output should be limited to just those in a non-operational state. Applies to `table`, `overview` and `verbose` formats.                                                                                  |
| `os`, `omit-summary`          | No        | `false`   | No     | `true`, `false`                                                         | Whether summary in results output should be omitted.                                                                                                                                                                                           |
| `rl`, `read-limit`            | No        | `1048576` | No     | *valid whole number of bytes*                                           | Limit in bytes used to help prevent abuse when reading input that could be larger than expected. The default value is nearly 4x the largest observed (formatted) feed size.                                                                    |
| `auf`, `allow-unknown-fields` | No        | `false`   | No     | `true`, `false`                                                         | Whether unknown JSON fields encountered while decoding JSON data should be ignored.                                                                                                                                                            |
| `fmt`, `output-format`        | No        | `table`   | No     | `overview`, `table`, `verbose`, `debug`, `list`, `json`                 | Sets output format. The default format is `table`.                                                                                                                                                                                             |

### Configuration file

Not currently supported. This feature may be added later if there is
sufficient interest.

## Examples

Entries in this section attempt to provide a brief overview of usage. While
the `--filename` flag is predominantly used to provide reproducible results
(based on the `testdata/components` files), the `--url` flag is also used for
contrast.

### `check_statuspage_components` Nagios plugin

#### CLI invocations

##### Evaluate all subcomponents in a group

We evaluate all subcomponents in the entire `Canvas` component group for a
non-OK or non-operational status. The most severe (non-excluded) status is
used as the final plugin state.

Here we use a cached testdata file:

```console
$ /usr/lib/nagios/plugins/check_statuspage_components --filename testdata/components/instructure-components.json --group 'Canvas'
{"level":"error","version":"check-statuspage x.y.z (https://github.com/atc0005/check-statuspage)","logging_level":"info","app_type":"plugin-components","timeout":"10s","filename":"testdata/components/instructure-components.json","url":"","read_limit":1048576,"allow_unknown_fields":false,"total_problem_components":8,"excluded_components":10,"excluded_problem_components":7,"remaining_problem_components":1,"time":"2021-12-25T16:13:49-06:00","caller":"/mnt/t/github/check-statuspage/cmd/check_statuspage_components/main.go:319","message":"Non-excluded, non-operational status of statuspage components detected"}
WARNING: 1 evaluated "Instructure" component has a non-operational status (14 evaluated, 24 total) [degraded_performance (1)]

**ERRORS**

* component with non-operational status not excluded from evaluation

**THRESHOLDS**

* CRITICAL: major_outage
* WARNING: under_maintenance, partial_outage, degraded_performance

**DETAILED INFO**

Specified filter: {Group: "Canvas", Components: ""}

Page: Instructure (https://status.instructure.com)
Time Zone: America/Denver
Last Updated: 2021-12-07T11:07:15-07:00


GROUP NAME        GROUP ID        COMPONENT NAME              COMPONENT ID    EVALUATED    STATUS
----------        --------        --------------              ------------    ---------    ------
Canvas            41wg86q5vc14    Canvas                      9dlvqx1drp3d    true         DEGRADED PERFORMANCE
Canvas            41wg86q5vc14    — Catalog                   jw0fn0dnpcgn    true         OPERATIONAL
Canvas            41wg86q5vc14    — Chat                      57p1tjtk1yq0    true         OPERATIONAL
Canvas            41wg86q5vc14    — Collaboration             zxq967k6np07    true         OPERATIONAL
Canvas            41wg86q5vc14    — Commons                   z5p8qvl1hj1y    true         OPERATIONAL
Canvas            41wg86q5vc14    — Conferences               mtytktcmbk6p    true         OPERATIONAL
Canvas            41wg86q5vc14    — Document previewing       ch8dsykb6hln    true         OPERATIONAL
Canvas            41wg86q5vc14    — Imports and exports       qt6q9hfpbljc    true         OPERATIONAL
Canvas            41wg86q5vc14    — Media tools               knh34j1129ft    true         OPERATIONAL
Canvas            41wg86q5vc14    — Notifications             6pnn3zwfyzxz    true         OPERATIONAL
Canvas            41wg86q5vc14    — Support: Phones/Chat      d7cxm3fbff4h    true         OPERATIONAL
Canvas            41wg86q5vc14    — Support: Webform/email    zlyh32dxbwjj    true         OPERATIONAL
Canvas            41wg86q5vc14    — ePortfolios               tlhdyd68vb55    true         OPERATIONAL
Canvas            41wg86q5vc14    — Gauge                     3q12z77wvfjp    true         OPERATIONAL

Portfolium        9c01dg04bfg5    Website                     j7jp6sq831c2    false        DEGRADED PERFORMANCE
Portfolium        9c01dg04bfg5    Web Application             100xy482gkyf    false        DEGRADED PERFORMANCE
Portfolium        9c01dg04bfg5    EDU Platform                c8zkn4rlhvw6    false        DEGRADED PERFORMANCE

MasteryConnect    qw5j90r2w7k1    Assessments                 v6m5nhwgtshj    false        DEGRADED PERFORMANCE
MasteryConnect    qw5j90r2w7k1    Benchmarks                  jt1kl5fj472f    false        DEGRADED PERFORMANCE
MasteryConnect    qw5j90r2w7k1    Portal                      142661pcf7h1    false        DEGRADED PERFORMANCE
MasteryConnect    qw5j90r2w7k1    Reporting                   xwqppk51m3mm    false        DEGRADED PERFORMANCE




Summary:

* Filtering applied to components set: true
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 0
* Number of total component groups: 3
* Number of total subcomponents: 21
* Number of total problem components: 8
* Number of ignored problem components: 7
* Number of remaining problem components: 1


 | 'all_components'=24;;;; 'all_component_groups'=3;;;; 'all_problem_components'=8;;;; 'excluded_problem_components'=7;;;; 'remaining_problem_components'=1;;;; 'all_components_critical'=0;;;; 'all_components_warning'=11;;;; 'all_components_unknown'=0;;;; 'all_components_ok'=13;;;; 'remaining_components_critical'=0;;;; 'remaining_components_warning'=1;;;; 'remaining_components_unknown'=0;;;; 'remaining_components_ok'=13;;;; 'time'=4ms;;;;
```

and here we use live data:

```console
$ /usr/lib/nagios/plugins/check_statuspage_components --url https://status.instructure.com/api/v2/components.json --group 'Canvas'
OK: 0 evaluated "Instructure" components have a non-operational status (14 evaluated, 24 total)

**ERRORS**

* None

**THRESHOLDS**

* CRITICAL: major_outage
* WARNING: under_maintenance, partial_outage, degraded_performance

**DETAILED INFO**

Specified filter: {Group: "Canvas", Components: ""}

Page: Instructure (https://status.instructure.com)
Time Zone: America/Denver
Last Updated: 2021-12-22T07:25:53-07:00


GROUP NAME        GROUP ID        COMPONENT NAME              COMPONENT ID    EVALUATED    STATUS
----------        --------        --------------              ------------    ---------    ------
Canvas            41wg86q5vc14    Canvas                      9dlvqx1drp3d    true         OPERATIONAL
Canvas            41wg86q5vc14    — Catalog                   jw0fn0dnpcgn    true         OPERATIONAL
Canvas            41wg86q5vc14    — Chat                      57p1tjtk1yq0    true         OPERATIONAL
Canvas            41wg86q5vc14    — Collaboration             zxq967k6np07    true         OPERATIONAL
Canvas            41wg86q5vc14    — Commons                   z5p8qvl1hj1y    true         OPERATIONAL
Canvas            41wg86q5vc14    — Conferences               mtytktcmbk6p    true         OPERATIONAL
Canvas            41wg86q5vc14    — Document previewing       ch8dsykb6hln    true         OPERATIONAL
Canvas            41wg86q5vc14    — Imports and exports       qt6q9hfpbljc    true         OPERATIONAL
Canvas            41wg86q5vc14    — Media tools               knh34j1129ft    true         OPERATIONAL
Canvas            41wg86q5vc14    — Notifications             6pnn3zwfyzxz    true         OPERATIONAL
Canvas            41wg86q5vc14    — Support: Phones/Chat      d7cxm3fbff4h    true         OPERATIONAL
Canvas            41wg86q5vc14    — Support: Webform/email    zlyh32dxbwjj    true         OPERATIONAL
Canvas            41wg86q5vc14    — ePortfolios               tlhdyd68vb55    true         OPERATIONAL
Canvas            41wg86q5vc14    — Gauge                     3q12z77wvfjp    true         OPERATIONAL

Portfolium        9c01dg04bfg5    Website                     j7jp6sq831c2    false        OPERATIONAL
Portfolium        9c01dg04bfg5    Web Application             100xy482gkyf    false        OPERATIONAL
Portfolium        9c01dg04bfg5    EDU Platform                c8zkn4rlhvw6    false        OPERATIONAL

MasteryConnect    qw5j90r2w7k1    Assessments                 v6m5nhwgtshj    false        OPERATIONAL
MasteryConnect    qw5j90r2w7k1    Benchmarks                  jt1kl5fj472f    false        OPERATIONAL
MasteryConnect    qw5j90r2w7k1    Portal                      142661pcf7h1    false        OPERATIONAL
MasteryConnect    qw5j90r2w7k1    Reporting                   xwqppk51m3mm    false        OPERATIONAL




Summary:

* Filtering applied to components set: true
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 0
* Number of total component groups: 3
* Number of total subcomponents: 21
* Number of total problem components: 0
* Number of ignored problem components: 0
* Number of remaining problem components: 0


 | 'all_components'=24;;;; 'all_component_groups'=3;;;; 'all_problem_components'=0;;;; 'excluded_problem_components'=0;;;; 'remaining_problem_components'=0;;;; 'all_components_critical'=0;;;; 'all_components_warning'=0;;;; 'all_components_unknown'=0;;;; 'all_components_ok'=24;;;; 'remaining_components_critical'=0;;;; 'remaining_components_warning'=0;;;; 'remaining_components_unknown'=0;;;; 'remaining_components_ok'=14;;;; 'time'=553ms;;;;
```

See the [configuration options](#configuration-options) section for all
command-line settings supported by this plugin along with descriptions of
each.

##### Evaluate a specific top-level component

We evaluate a specific top-level component named `Hosted DNS Service` by id
value for a non-OK or non-operational status.

Here we use a cached testdata file when Linode was experiencing an outage of
some services, but not the specific service we opted to evaluate:

```console
$ `/usr/lib/nagios/plugins/check_statuspage_components --filename testdata/components/linode-components-problems.json --component mmgkdgcjgnbl
OK: 0 evaluated "Linode" components have a non-operational status (1 evaluated, 68 total)

**ERRORS**

* None

**THRESHOLDS**

* CRITICAL: major_outage
* WARNING: under_maintenance, partial_outage, degraded_performance

**DETAILED INFO**

Specified filter: {Group: "", Components: "mmgkdgcjgnbl"}

Page: Linode (https://status.linode.com)
Time Zone: Etc/UTC
Last Updated: 2021-12-17T22:23:41Z


GROUP NAME                  GROUP ID        COMPONENT NAME                                       COMPONENT ID    EVALUATED    STATUS
----------                  --------        --------------                                       ------------    ---------    ------
                                            Linode.com                                           qx710vq78w73    false        OPERATIONAL
                                            Linode Manager and API                               61mfhwczrxzt    false        OPERATIONAL
                                            Hosted DNS Service                                   mmgkdgcjgnbl    true         OPERATIONAL
                                            Longview                                             v3rbxq0mztjx    false        OPERATIONAL

Regions                     kb0cf2d0rryb    US-East (Newark)                                     v8zpnn61fcdm    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    US-Central (Dallas)                                  s8355f9s5hpx    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    US-West (Fremont)                                    v5mcdxdt9dtv    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    US-Southeast (Atlanta)                               ss2n92gy5xkk    false        OPERATIONAL
Regions                     kb0cf2d0rryb    CA-Central (Toronto)                                 ptln4fx0jhrt    false        OPERATIONAL
Regions                     kb0cf2d0rryb    EU-West (London)                                     jzjmxld11n92    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    EU-Central (Frankfurt)                               bxcfhjdf5c4k    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    AP-South (Singapore)                                 ymhgjx06c93n    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    AP-Northeast-2 (Tokyo 2)                             94q3fkblbvs3    false        PARTIAL OUTAGE
Regions                     kb0cf2d0rryb    AP-West (Mumbai)                                     cc7n9z4yz3gw    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-Southeast (Sydney)                                lq833jghn099    false        OPERATIONAL

Backups                     drpn3k8bk56t    US-East (Newark) Backups                             tyb1r72rvfrx    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-Central (Dallas) Backups                          kprtdgjdqfx5    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-West (Fremont) Backups                            4ny59fgq6cln    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-Southeast (Atlanta) Backups                       qczfz4brkf1b    false        OPERATIONAL
Backups                     drpn3k8bk56t    CA-Central (Toronto) Backups                         l9plxr7kgdtp    false        OPERATIONAL
Backups                     drpn3k8bk56t    EU-West (London) Backups                             1bdfwxqf66rj    false        OPERATIONAL
Backups                     drpn3k8bk56t    EU-Central (Frankfurt) Backups                       24vh5f8lfkld    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-South (Singapore) Backups                         1x47wvj7sk31    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-Northeast-2 (Tokyo 2) Backups                     99qq411dqngl    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-West (Mumbai) Backups                             gpfk34v7m7bx    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-Southeast (Sydney) Backups                        dfdf5d1hcrkt    false        OPERATIONAL

Block Storage               m5wrr6m00mvh    US-East (Newark) Block Storage                       zq4h2zstlwhz    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-Central (Dallas) Block Storage                    j256v14cqldk    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-West (Fremont) Block Storage                      fvq65tvc6vk2    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    CA-Central (Toronto) Block Storage                   wh3l2nzfvrj6    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    EU-West (London) Block Storage                       fmy3yfmxmcx6    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    EU-Central (Frankfurt) Block Storage                 h3p36ttc972l    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-South (Singapore) Block Storage                   6y55232d871t    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-Northeast-2 (Tokyo 2) Block Storage               kdrjlntx44yj    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-West (Mumbai) Block Storage                       q9b5j34kd21x    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-Southeast (Sydney) Block Storage                  vz46kd91my1k    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-Southeast (Atlanta) Block Storage                 9clykhgs02xl    false        OPERATIONAL

NodeBalancers               7q62h0wl818b    US-East (Newark) NodeBalancers                       3382rw9rdyby    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-Central (Dallas) NodeBalancers                    jyl125r91hhl    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-West (Fremont) NodeBalancers                      gqlf5lr6g4xj    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-Southeast (Atlanta) NodeBalancers                 362rg0lr43dz    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    CA-Central (Toronto) NodeBalancers                   l3w6b1ck9f5h    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    EU-West (London) NodeBalancers                       c45ljxpwx5m4    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    EU-Central (Frankfurt)  NodeBalancers                y2slqzr87kdz    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-South (Singapore) NodeBalancers                   pyg6cyddyh92    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-Northeast-2 (Tokyo 2) NodeBalancers               c3m26py2z00j    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-West (Mumbai) NodeBalancers                       fb81v7x1rjd1    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-Southeast (Sydney) NodeBalancers                  n2sbx9jll8cz    false        OPERATIONAL

Object Storage              4bvp5l14grs6    US-East (Newark) Object Storage                      q55c1gfpf7m4    false        OPERATIONAL
Object Storage              4bvp5l14grs6    EU-Central (Frankfurt) Object Storage                zm7300chzddp    false        OPERATIONAL
Object Storage              4bvp5l14grs6    AP-South (Singapore) Object Storage                  l8y4k4tz708b    false        OPERATIONAL
Object Storage              4bvp5l14grs6    US-Southeast (Atlanta) Object Storage                y5gcmn9bstbg    false        OPERATIONAL

Linode Kubernetes Engine    lplvcng25jv0    US-East (Newark) Linode Kubernetes Engine            bv6jc1sr32l5    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    US-Central (Dallas) Linode Kubernetes Engine         gwsl8rcc9jmh    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    US-West (Fremont) Linode Kubernetes Engine           xd4qzq00pq6b    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    CA-Central (Toronto) Linode Kubernetes Engine        sc9zctkdfwsp    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    EU-West (London) Linode Kubernetes Engine            p2jqcf059vgf    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    EU-Central (Frankfurt) Linode Kubernetes Engine      f5g6kbf5tnf2    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-South (Singapore) Linode Kubernetes Engine        wzr2h8y6b14r    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-Northeast-2 (Tokyo 2) Linode Kubernetes Engine    gdvwymjjd06s    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-West (Mumbai) Linode Kubernetes Engine            0r45ydtwlrz3    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-Southeast (Sydney) Linode Kubernetes Engine       mzvhlfd50msv    false        OPERATIONAL




Summary:

* Filtering applied to components set: true
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 4
* Number of total component groups: 6
* Number of total subcomponents: 58
* Number of total problem components: 7
* Number of ignored problem components: 7
* Number of remaining problem components: 0


 | 'all_components'=68;;;; 'all_component_groups'=6;;;; 'all_problem_components'=7;;;; 'excluded_problem_components'=7;;;; 'remaining_problem_components'=0;;;; 'all_components_critical'=0;;;; 'all_components_warning'=8;;;; 'all_components_unknown'=0;;;; 'all_components_ok'=60;;;; 'remaining_components_critical'=0;;;; 'remaining_components_warning'=0;;;; 'remaining_components_unknown'=0;;;; 'remaining_components_ok'=1;;;; 'time'=3ms;;;;
```

and for constrast, here we use another cached testdata file when Linode was
not experiencing any outages:

```console
$ /usr/lib/nagios/plugins/check_statuspage_components --filename testdata/components/linode-components.json --component mmgkdgcjgnbl
OK: 0 evaluated "Linode" components have a non-operational status (1 evaluated, 68 total)

**ERRORS**

* None

**THRESHOLDS**

* CRITICAL: major_outage
* WARNING: under_maintenance, partial_outage, degraded_performance

**DETAILED INFO**

Specified filter: {Group: "", Components: "mmgkdgcjgnbl"}

Page: Linode (https://status.linode.com)
Time Zone: Etc/UTC
Last Updated: 2021-12-01T10:57:29Z


GROUP NAME                  GROUP ID        COMPONENT NAME                                       COMPONENT ID    EVALUATED    STATUS
----------                  --------        --------------                                       ------------    ---------    ------
                                            Linode.com                                           qx710vq78w73    false        OPERATIONAL
                                            Linode Manager and API                               61mfhwczrxzt    false        OPERATIONAL
                                            Hosted DNS Service                                   mmgkdgcjgnbl    true         OPERATIONAL
                                            Longview                                             v3rbxq0mztjx    false        OPERATIONAL

Regions                     kb0cf2d0rryb    US-East (Newark)                                     v8zpnn61fcdm    false        OPERATIONAL
Regions                     kb0cf2d0rryb    US-Central (Dallas)                                  s8355f9s5hpx    false        OPERATIONAL
Regions                     kb0cf2d0rryb    US-West (Fremont)                                    v5mcdxdt9dtv    false        OPERATIONAL
Regions                     kb0cf2d0rryb    US-Southeast (Atlanta)                               ss2n92gy5xkk    false        OPERATIONAL
Regions                     kb0cf2d0rryb    CA-Central (Toronto)                                 ptln4fx0jhrt    false        OPERATIONAL
Regions                     kb0cf2d0rryb    EU-West (London)                                     jzjmxld11n92    false        OPERATIONAL
Regions                     kb0cf2d0rryb    EU-Central (Frankfurt)                               bxcfhjdf5c4k    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-South (Singapore)                                 ymhgjx06c93n    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-Northeast-2 (Tokyo 2)                             94q3fkblbvs3    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-West (Mumbai)                                     cc7n9z4yz3gw    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-Southeast (Sydney)                                lq833jghn099    false        OPERATIONAL

Backups                     drpn3k8bk56t    US-East (Newark) Backups                             tyb1r72rvfrx    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-Central (Dallas) Backups                          kprtdgjdqfx5    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-West (Fremont) Backups                            4ny59fgq6cln    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-Southeast (Atlanta) Backups                       qczfz4brkf1b    false        OPERATIONAL
Backups                     drpn3k8bk56t    CA-Central (Toronto) Backups                         l9plxr7kgdtp    false        OPERATIONAL
Backups                     drpn3k8bk56t    EU-West (London) Backups                             1bdfwxqf66rj    false        OPERATIONAL
Backups                     drpn3k8bk56t    EU-Central (Frankfurt) Backups                       24vh5f8lfkld    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-South (Singapore) Backups                         1x47wvj7sk31    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-Northeast-2 (Tokyo 2) Backups                     99qq411dqngl    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-West (Mumbai) Backups                             gpfk34v7m7bx    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-Southeast (Sydney) Backups                        dfdf5d1hcrkt    false        OPERATIONAL

Block Storage               m5wrr6m00mvh    US-East (Newark) Block Storage                       zq4h2zstlwhz    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-Central (Dallas) Block Storage                    j256v14cqldk    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-West (Fremont) Block Storage                      fvq65tvc6vk2    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    CA-Central (Toronto) Block Storage                   wh3l2nzfvrj6    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    EU-West (London) Block Storage                       fmy3yfmxmcx6    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    EU-Central (Frankfurt) Block Storage                 h3p36ttc972l    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-South (Singapore) Block Storage                   6y55232d871t    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-Northeast-2 (Tokyo 2) Block Storage               kdrjlntx44yj    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-West (Mumbai) Block Storage                       q9b5j34kd21x    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-Southeast (Sydney) Block Storage                  vz46kd91my1k    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-Southeast (Atlanta) Block Storage                 9clykhgs02xl    false        OPERATIONAL

NodeBalancers               7q62h0wl818b    US-East (Newark) NodeBalancers                       3382rw9rdyby    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-Central (Dallas) NodeBalancers                    jyl125r91hhl    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-West (Fremont) NodeBalancers                      gqlf5lr6g4xj    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-Southeast (Atlanta) NodeBalancers                 362rg0lr43dz    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    CA-Central (Toronto) NodeBalancers                   l3w6b1ck9f5h    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    EU-West (London) NodeBalancers                       c45ljxpwx5m4    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    EU-Central (Frankfurt)  NodeBalancers                y2slqzr87kdz    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-South (Singapore) NodeBalancers                   pyg6cyddyh92    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-Northeast-2 (Tokyo 2) NodeBalancers               c3m26py2z00j    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-West (Mumbai) NodeBalancers                       fb81v7x1rjd1    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-Southeast (Sydney) NodeBalancers                  n2sbx9jll8cz    false        OPERATIONAL

Object Storage              4bvp5l14grs6    US-East (Newark) Object Storage                      q55c1gfpf7m4    false        OPERATIONAL
Object Storage              4bvp5l14grs6    EU-Central (Frankfurt) Object Storage                zm7300chzddp    false        OPERATIONAL
Object Storage              4bvp5l14grs6    AP-South (Singapore) Object Storage                  l8y4k4tz708b    false        OPERATIONAL
Object Storage              4bvp5l14grs6    US-Southeast (Atlanta) Object Storage                y5gcmn9bstbg    false        OPERATIONAL

Linode Kubernetes Engine    lplvcng25jv0    US-East (Newark) Linode Kubernetes Engine            bv6jc1sr32l5    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    US-Central (Dallas) Linode Kubernetes Engine         gwsl8rcc9jmh    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    US-West (Fremont) Linode Kubernetes Engine           xd4qzq00pq6b    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    CA-Central (Toronto) Linode Kubernetes Engine        sc9zctkdfwsp    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    EU-West (London) Linode Kubernetes Engine            p2jqcf059vgf    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    EU-Central (Frankfurt) Linode Kubernetes Engine      f5g6kbf5tnf2    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-South (Singapore) Linode Kubernetes Engine        wzr2h8y6b14r    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-Northeast-2 (Tokyo 2) Linode Kubernetes Engine    gdvwymjjd06s    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-West (Mumbai) Linode Kubernetes Engine            0r45ydtwlrz3    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-Southeast (Sydney) Linode Kubernetes Engine       mzvhlfd50msv    false        OPERATIONAL




Summary:

* Filtering applied to components set: true
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 4
* Number of total component groups: 6
* Number of total subcomponents: 58
* Number of total problem components: 0
* Number of ignored problem components: 0
* Number of remaining problem components: 0


 | 'all_components'=68;;;; 'all_component_groups'=6;;;; 'all_problem_components'=0;;;; 'excluded_problem_components'=0;;;; 'remaining_problem_components'=0;;;; 'all_components_critical'=0;;;; 'all_components_warning'=0;;;; 'all_components_unknown'=0;;;; 'all_components_ok'=68;;;; 'remaining_components_critical'=0;;;; 'remaining_components_warning'=0;;;; 'remaining_components_unknown'=0;;;; 'remaining_components_ok'=1;;;; 'time'=4ms;;;;
```

and here we use live data where Linode was not experiencing any outages:

```console
$ `/usr/lib/nagios/plugins/check_statuspage_components --url https://status.linode.com/api/v2/components.json --component mmgkdgcjgnbl
OK: 0 evaluated "Linode" components have a non-operational status (1 evaluated, 68 total)

**ERRORS**

* None

**THRESHOLDS**

* CRITICAL: major_outage
* WARNING: under_maintenance, partial_outage, degraded_performance

**DETAILED INFO**

Specified filter: {Group: "", Components: "mmgkdgcjgnbl"}

Page: Linode (https://status.linode.com)
Time Zone: Etc/UTC
Last Updated: 2021-12-26T12:40:58Z


GROUP NAME                  GROUP ID        COMPONENT NAME                                       COMPONENT ID    EVALUATED    STATUS
----------                  --------        --------------                                       ------------    ---------    ------
                                            Linode.com                                           qx710vq78w73    false        OPERATIONAL
                                            Linode Manager and API                               61mfhwczrxzt    false        OPERATIONAL
                                            Hosted DNS Service                                   mmgkdgcjgnbl    true         OPERATIONAL
                                            Longview                                             v3rbxq0mztjx    false        OPERATIONAL

Regions                     kb0cf2d0rryb    US-East (Newark)                                     v8zpnn61fcdm    false        OPERATIONAL
Regions                     kb0cf2d0rryb    US-Central (Dallas)                                  s8355f9s5hpx    false        OPERATIONAL
Regions                     kb0cf2d0rryb    US-West (Fremont)                                    v5mcdxdt9dtv    false        OPERATIONAL
Regions                     kb0cf2d0rryb    US-Southeast (Atlanta)                               ss2n92gy5xkk    false        OPERATIONAL
Regions                     kb0cf2d0rryb    CA-Central (Toronto)                                 ptln4fx0jhrt    false        OPERATIONAL
Regions                     kb0cf2d0rryb    EU-West (London)                                     jzjmxld11n92    false        OPERATIONAL
Regions                     kb0cf2d0rryb    EU-Central (Frankfurt)                               bxcfhjdf5c4k    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-South (Singapore)                                 ymhgjx06c93n    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-Northeast-2 (Tokyo 2)                             94q3fkblbvs3    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-West (Mumbai)                                     cc7n9z4yz3gw    false        OPERATIONAL
Regions                     kb0cf2d0rryb    AP-Southeast (Sydney)                                lq833jghn099    false        OPERATIONAL

Backups                     drpn3k8bk56t    US-East (Newark) Backups                             tyb1r72rvfrx    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-Central (Dallas) Backups                          kprtdgjdqfx5    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-West (Fremont) Backups                            4ny59fgq6cln    false        OPERATIONAL
Backups                     drpn3k8bk56t    US-Southeast (Atlanta) Backups                       qczfz4brkf1b    false        OPERATIONAL
Backups                     drpn3k8bk56t    CA-Central (Toronto) Backups                         l9plxr7kgdtp    false        OPERATIONAL
Backups                     drpn3k8bk56t    EU-West (London) Backups                             1bdfwxqf66rj    false        OPERATIONAL
Backups                     drpn3k8bk56t    EU-Central (Frankfurt) Backups                       24vh5f8lfkld    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-South (Singapore) Backups                         1x47wvj7sk31    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-Northeast-2 (Tokyo 2) Backups                     99qq411dqngl    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-West (Mumbai) Backups                             gpfk34v7m7bx    false        OPERATIONAL
Backups                     drpn3k8bk56t    AP-Southeast (Sydney) Backups                        dfdf5d1hcrkt    false        OPERATIONAL

Block Storage               m5wrr6m00mvh    US-East (Newark) Block Storage                       zq4h2zstlwhz    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-Central (Dallas) Block Storage                    j256v14cqldk    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-West (Fremont) Block Storage                      fvq65tvc6vk2    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    CA-Central (Toronto) Block Storage                   wh3l2nzfvrj6    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    EU-West (London) Block Storage                       fmy3yfmxmcx6    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    EU-Central (Frankfurt) Block Storage                 h3p36ttc972l    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-South (Singapore) Block Storage                   6y55232d871t    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-Northeast-2 (Tokyo 2) Block Storage               kdrjlntx44yj    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-West (Mumbai) Block Storage                       q9b5j34kd21x    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    AP-Southeast (Sydney) Block Storage                  vz46kd91my1k    false        OPERATIONAL
Block Storage               m5wrr6m00mvh    US-Southeast (Atlanta) Block Storage                 9clykhgs02xl    false        OPERATIONAL

NodeBalancers               7q62h0wl818b    US-East (Newark) NodeBalancers                       3382rw9rdyby    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-Central (Dallas) NodeBalancers                    jyl125r91hhl    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-West (Fremont) NodeBalancers                      gqlf5lr6g4xj    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    US-Southeast (Atlanta) NodeBalancers                 362rg0lr43dz    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    CA-Central (Toronto) NodeBalancers                   l3w6b1ck9f5h    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    EU-West (London) NodeBalancers                       c45ljxpwx5m4    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    EU-Central (Frankfurt)  NodeBalancers                y2slqzr87kdz    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-South (Singapore) NodeBalancers                   pyg6cyddyh92    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-Northeast-2 (Tokyo 2) NodeBalancers               c3m26py2z00j    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-West (Mumbai) NodeBalancers                       fb81v7x1rjd1    false        OPERATIONAL
NodeBalancers               7q62h0wl818b    AP-Southeast (Sydney) NodeBalancers                  n2sbx9jll8cz    false        OPERATIONAL

Object Storage              4bvp5l14grs6    US-East (Newark) Object Storage                      q55c1gfpf7m4    false        OPERATIONAL
Object Storage              4bvp5l14grs6    EU-Central (Frankfurt) Object Storage                zm7300chzddp    false        OPERATIONAL
Object Storage              4bvp5l14grs6    AP-South (Singapore) Object Storage                  l8y4k4tz708b    false        OPERATIONAL
Object Storage              4bvp5l14grs6    US-Southeast (Atlanta) Object Storage                y5gcmn9bstbg    false        OPERATIONAL

Linode Kubernetes Engine    lplvcng25jv0    US-East (Newark) Linode Kubernetes Engine            bv6jc1sr32l5    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    US-Central (Dallas) Linode Kubernetes Engine         gwsl8rcc9jmh    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    US-West (Fremont) Linode Kubernetes Engine           xd4qzq00pq6b    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    CA-Central (Toronto) Linode Kubernetes Engine        sc9zctkdfwsp    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    EU-West (London) Linode Kubernetes Engine            p2jqcf059vgf    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    EU-Central (Frankfurt) Linode Kubernetes Engine      f5g6kbf5tnf2    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-South (Singapore) Linode Kubernetes Engine        wzr2h8y6b14r    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-Northeast-2 (Tokyo 2) Linode Kubernetes Engine    gdvwymjjd06s    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-West (Mumbai) Linode Kubernetes Engine            0r45ydtwlrz3    false        OPERATIONAL
Linode Kubernetes Engine    lplvcng25jv0    AP-Southeast (Sydney) Linode Kubernetes Engine       mzvhlfd50msv    false        OPERATIONAL




Summary:

* Filtering applied to components set: true
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 4
* Number of total component groups: 6
* Number of total subcomponents: 58
* Number of total problem components: 0
* Number of ignored problem components: 0
* Number of remaining problem components: 0


 | 'all_components'=68;;;; 'all_component_groups'=6;;;; 'all_problem_components'=0;;;; 'excluded_problem_components'=0;;;; 'remaining_problem_components'=0;;;; 'all_components_critical'=0;;;; 'all_components_warning'=0;;;; 'all_components_unknown'=0;;;; 'all_components_ok'=68;;;; 'remaining_components_critical'=0;;;; 'remaining_components_warning'=0;;;; 'remaining_components_unknown'=0;;;; 'remaining_components_ok'=1;;;; 'time'=552ms;;;;
```

#### Command definition

The command definition file below defines three commands. Each command
explicitly excludes "OK" components in order to keep the output manageable.
Remove the `omit-ok` flag if you wish to use the built-in components output
limit to control the number of components emitted.

```shell
# /etc/nagios-plugins/config/statuspage-components.cfg

# Evaluate all components for a specified component group.
define command{
    command_name    check_statuspage_components_group
    command_line    $USER1$/check_statuspage_components --url '$ARG1$' --group '$ARG2$' --omit-ok --log-level info
    }

# Evaluate one or more components for a (single) specified component group.
#
# NOTE: Multiple components can be specified as a comma separated string using
# a mix of component names and/or ID values.
define command{
    command_name    check_statuspage_components_group_and_select_subcomponents
    command_line    $USER1$/check_statuspage_components --url '$ARG1$' --group '$ARG2$' --component '$ARG3$' --omit-ok --log-level info
    }

# Evaluate one or more components regardless of component group.
#
# NOTE: Multiple components can be specified as a comma separated string using
# a mix of component names and/or ID values.
define command{
    command_name    check_statuspage_components_list
    command_line    $USER1$/check_statuspage_components --url '$ARG1$' --component '$ARG2$' --omit-ok --log-level info
    }
```

See the [configuration options](#configuration-options) section for all
command-line settings supported by this plugin along with descriptions of
each.

### `lscs` CLI app

#### CLI invocation

We evaluate a local testdata file to illustrate different supported output
formats.

##### The `table` format (default)

```console
$ /usr/local/bin/lscs --filename testdata/components/instructure-components.json --output-format table

Page: Instructure (https://status.instructure.com)
Time Zone: America/Denver
Last Updated: 2021-12-07T11:07:15-07:00


GROUP NAME        GROUP ID        COMPONENT NAME              COMPONENT ID    EVALUATED    STATUS
----------        --------        --------------              ------------    ---------    ------
Canvas            41wg86q5vc14    Canvas                      9dlvqx1drp3d    N/A          DEGRADED PERFORMANCE
Canvas            41wg86q5vc14    — Catalog                   jw0fn0dnpcgn    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Chat                      57p1tjtk1yq0    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Collaboration             zxq967k6np07    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Commons                   z5p8qvl1hj1y    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Conferences               mtytktcmbk6p    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Document previewing       ch8dsykb6hln    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Imports and exports       qt6q9hfpbljc    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Media tools               knh34j1129ft    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Notifications             6pnn3zwfyzxz    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Support: Phones/Chat      d7cxm3fbff4h    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Support: Webform/email    zlyh32dxbwjj    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — ePortfolios               tlhdyd68vb55    N/A          OPERATIONAL
Canvas            41wg86q5vc14    — Gauge                     3q12z77wvfjp    N/A          OPERATIONAL

Portfolium        9c01dg04bfg5    Website                     j7jp6sq831c2    N/A          DEGRADED PERFORMANCE
Portfolium        9c01dg04bfg5    Web Application             100xy482gkyf    N/A          DEGRADED PERFORMANCE
Portfolium        9c01dg04bfg5    EDU Platform                c8zkn4rlhvw6    N/A          DEGRADED PERFORMANCE

MasteryConnect    qw5j90r2w7k1    Assessments                 v6m5nhwgtshj    N/A          DEGRADED PERFORMANCE
MasteryConnect    qw5j90r2w7k1    Benchmarks                  jt1kl5fj472f    N/A          DEGRADED PERFORMANCE
MasteryConnect    qw5j90r2w7k1    Portal                      142661pcf7h1    N/A          DEGRADED PERFORMANCE
MasteryConnect    qw5j90r2w7k1    Reporting                   xwqppk51m3mm    N/A          DEGRADED PERFORMANCE




Summary:

* Filtering applied to components set: false
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 0
* Number of total component groups: 3
* Number of total subcomponents: 21
* Number of total problem components: 8
* Number of ignored problem components: 0
* Number of remaining problem components: 8
```

##### The `overview` format

```console
$ /usr/local/bin/lscs --filename testdata/components/instructure-components.json --output-format overview

Page: Instructure (https://status.instructure.com)
Time Zone: America/Denver
Last Updated: 2021-12-07T11:07:15-07:00


Canvas [DEGRADED PERFORMANCE]
        Canvas [DEGRADED PERFORMANCE]
        — Catalog
        — Chat
        — Collaboration
        — Commons
        — Conferences
        — Document previewing
        — Imports and exports
        — Media tools
        — Notifications
        — Support: Phones/Chat
        — Support: Webform/email
        — ePortfolios
        — Gauge

Portfolium [DEGRADED PERFORMANCE]
        Website [DEGRADED PERFORMANCE]
        Web Application [DEGRADED PERFORMANCE]
        EDU Platform [DEGRADED PERFORMANCE]

MasteryConnect [DEGRADED PERFORMANCE]
        Assessments [DEGRADED PERFORMANCE]
        Benchmarks [DEGRADED PERFORMANCE]
        Portal [DEGRADED PERFORMANCE]
        Reporting [DEGRADED PERFORMANCE]



Summary:

* Filtering applied to components set: false
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 0
* Number of total component groups: 3
* Number of total subcomponents: 21
* Number of total problem components: 8
* Number of ignored problem components: 0
* Number of remaining problem components: 8
```

##### The `verbose` format

```console
$ /usr/local/bin/lscs --filename testdata/components/instructure-components.json --output-format verbose

Page: Instructure (https://status.instructure.com)
Time Zone: America/Denver
Last Updated: 2021-12-07T11:07:15-07:00


Component Groups (3):

* 001): Canvas [ID: 41wg86q5vc14, GroupID: N/A, Status: DEGRADED PERFORMANCE]
* 002): Portfolium [ID: 9c01dg04bfg5, GroupID: N/A, Status: DEGRADED PERFORMANCE]
* 003): MasteryConnect [ID: qw5j90r2w7k1, GroupID: N/A, Status: DEGRADED PERFORMANCE]


Subcomponents (21):

* 001): Website [ID: j7jp6sq831c2, GroupID: 9c01dg04bfg5, Status: DEGRADED PERFORMANCE]
* 002): Canvas [ID: 9dlvqx1drp3d, GroupID: 41wg86q5vc14, Status: DEGRADED PERFORMANCE]
* 003): Assessments [ID: v6m5nhwgtshj, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 004): Benchmarks [ID: jt1kl5fj472f, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 005): Web Application [ID: 100xy482gkyf, GroupID: 9c01dg04bfg5, Status: DEGRADED PERFORMANCE]
* 006): — Catalog [ID: jw0fn0dnpcgn, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 007): — Chat [ID: 57p1tjtk1yq0, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 008): EDU Platform [ID: c8zkn4rlhvw6, GroupID: 9c01dg04bfg5, Status: DEGRADED PERFORMANCE]
* 009): Portal [ID: 142661pcf7h1, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 010): — Collaboration [ID: zxq967k6np07, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 011): Reporting [ID: xwqppk51m3mm, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 012): — Commons [ID: z5p8qvl1hj1y, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 013): — Conferences [ID: mtytktcmbk6p, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 014): — Document previewing [ID: ch8dsykb6hln, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 015): — Imports and exports [ID: qt6q9hfpbljc, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 016): — Media tools [ID: knh34j1129ft, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 017): — Notifications [ID: 6pnn3zwfyzxz, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 018): — Support: Phones/Chat [ID: d7cxm3fbff4h, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 019): — Support: Webform/email [ID: zlyh32dxbwjj, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 020): — ePortfolios [ID: tlhdyd68vb55, GroupID: 41wg86q5vc14, Status: OPERATIONAL]
* 021): — Gauge [ID: 3q12z77wvfjp, GroupID: 41wg86q5vc14, Status: OPERATIONAL]


Components (8) in a non-OK state:

* 001): Website [ID: j7jp6sq831c2, GroupID: 9c01dg04bfg5, Status: DEGRADED PERFORMANCE]
* 002): Canvas [ID: 9dlvqx1drp3d, GroupID: 41wg86q5vc14, Status: DEGRADED PERFORMANCE]
* 003): Assessments [ID: v6m5nhwgtshj, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 004): Benchmarks [ID: jt1kl5fj472f, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 005): Web Application [ID: 100xy482gkyf, GroupID: 9c01dg04bfg5, Status: DEGRADED PERFORMANCE]
* 006): EDU Platform [ID: c8zkn4rlhvw6, GroupID: 9c01dg04bfg5, Status: DEGRADED PERFORMANCE]
* 007): Portal [ID: 142661pcf7h1, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]
* 008): Reporting [ID: xwqppk51m3mm, GroupID: qw5j90r2w7k1, Status: DEGRADED PERFORMANCE]


Summary:

* Filtering applied to components set: false
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): false
* Number of total top-level components: 0
* Number of total component groups: 3
* Number of total subcomponents: 21
* Number of total problem components: 8
* Number of ignored problem components: 0
* Number of remaining problem components: 8
```

##### The `list` format

This mostly came in handy when I was crafting test cases, but left it as an
option in case others also might also find it useful.

```console
$ /usr/local/bin/lscs --filename testdata/components/instructure-components.json --output-format list

Page: Instructure (https://status.instructure.com)
Time Zone: America/Denver
Last Updated: 2021-12-07T11:07:15-07:00

"j7jp6sq831c2", "41wg86q5vc14", "9dlvqx1drp3d", "v6m5nhwgtshj",
"9c01dg04bfg5", "jt1kl5fj472f", "100xy482gkyf", "jw0fn0dnpcgn",
"57p1tjtk1yq0", "qw5j90r2w7k1", "c8zkn4rlhvw6", "142661pcf7h1",
"zxq967k6np07", "xwqppk51m3mm", "z5p8qvl1hj1y", "mtytktcmbk6p",
"ch8dsykb6hln", "qt6q9hfpbljc", "knh34j1129ft", "6pnn3zwfyzxz",
"d7cxm3fbff4h", "zlyh32dxbwjj", "tlhdyd68vb55", "3q12z77wvfjp",


Summary:

* Filtering applied to components set: false
* Evaluating all components in the set: false
* Omitting OK/operational components (if requested): true
* Number of total top-level components: 0
* Number of total component groups: 3
* Number of total subcomponents: 21
* Number of total problem components: 8
* Number of ignored problem components: 0
* Number of remaining problem components: 8
```

See the [configuration options](#configuration-options) section for all
command-line settings supported by this plugin along with descriptions of
each.

##### Other supported formats

The `debug` and `json` formats are also supported output formats, but are
*very* detailed. Give them a try if the other formats do not meet your needs.
Feedback is welcome.

## License

From the [LICENSE](LICENSE) file:

```license
MIT License

Copyright (c) 2021 Adam Chalkley

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
```

## References

- Related projects
  - <https://github.com/atc0005/send2teams>
  - <https://github.com/atc0005/check-cert>
  - <https://github.com/atc0005/check-vmware>
  - <https://github.com/atc0005/check-mail>
  - <https://github.com/atc0005/check-whois>
  - <https://github.com/atc0005/check-path>
  - <https://github.com/atc0005/nagios-debug>
  - <https://github.com/atc0005/go-nagios>

- Statuspage
  - [What is a component?](https://support.atlassian.com/statuspage/docs/what-is-a-component/)
  - [Show service status with components](https://support.atlassian.com/statuspage/docs/show-service-status-with-components/)
  - [Statuspage API documentation](https://developer.statuspage.io/)

- Logging
  - <https://github.com/rs/zerolog>

- Nagios
  - <https://github.com/atc0005/go-nagios>
  - <https://nagios-plugins.org/doc/guidelines.html>
  - <https://www.monitoring-plugins.org/doc/guidelines.html>
  - <https://icinga.com/docs/icinga-2/latest/doc/05-service-monitoring/>

<!-- Footnotes here  -->

[repo-url]: <https://github.com/atc0005/check-statuspage>  "This project's GitHub repo"

[go-docs-download]: <https://golang.org/dl>  "Download Go"

[go-docs-install]: <https://golang.org/doc/install>  "Install Go"

[go-supported-releases]: <https://go.dev/doc/devel/release#policy> "Go Release Policy"

<!-- []: PLACEHOLDER "DESCRIPTION_HERE" -->
