# Changelog

## Overview

All notable changes to this project will be documented in this file.

The format is based on [Keep a
Changelog](https://keepachangelog.com/en/1.0.0/), and this project adheres to
[Semantic Versioning](https://semver.org/spec/v2.0.0.html).

Please [open an issue](https://github.com/atc0005/check-statuspage/issues) for
any deviations that you spot; I'm still learning!.

## Types of changes

The following types of changes will be recorded in this file:

- `Added` for new features.
- `Changed` for changes in existing functionality.
- `Deprecated` for soon-to-be removed features.
- `Removed` for now removed features.
- `Fixed` for any bug fixes.
- `Security` in case of vulnerabilities.

## [Unreleased]

- placeholder

## [v0.1.10] - 2022-05-11

### Overview

- Dependency updates
- built using Go 1.17.10
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `Go`
    - `1.17.9` to `1.17.10`
  - `github/codeql-action`
    - `v2.1.9` to `v2.1.10`

## [v0.1.9] - 2022-04-29

### Overview

- Bugfixes
- Dependency updates
- built using Go 1.17.9
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `Go`
    - `1.17.7` to `1.17.9`
  - `sanity-io/litter`
    - `v1.5.2` to `v1.5.5`
  - `google/go-cmp`
    - `v0.5.7` to `v0.5.8`
  - `github/codeql-action`
    - `v1.1.3` to `v2.1.9`
  - `actions/checkout`
    - `v2.4.0` to `v3`

### Fixed

- (GH-107) (*NullString).UnmarshalJSON - result 0 (error) is always nil
  (unparam)
- (GH-109) cmd/check_statuspage_components/main_test.go:108:28:
  shouldIgnoreCmpResult - t is unused (unparam)

## [v0.1.8] - 2022-03-01

### Overview

- Bugfixes
- Dependency updates
- built using Go 1.17.7
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `github/codeql-action`
    - `v1.0.32` to `v1.1.3`
  - `actions/setup-node`
    - `v2.5.1` to `v3`

### Fixed

- (GH-86) internal/reports/reports.go:917:23: nilness: nil dereference in
  index operation (govet)
- (GH-89) `ServiceOutput` (i.e., "one liner" output) is too vague when
  failures to read or decode JSON data occur

## [v0.1.7] - 2022-02-11

### Overview

- Bugfix
- Dependency updates
- built using Go 1.17.7
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `Go`
    - `1.17.6` to `1.17.7`
  - `github/codeql-action`
    - `v1.0.27` to `v1.0.32`
  - `google/go-cmp`
    - `v0.5.6` to `v0.5.7`

- (GH-76) Switch Docker image source from Docker Hub to GitHub Container
  Registry (GHCR)
- (GH-77) Expand linting GitHub Actions Workflow to include `oldstable`,
  `unstable` container images

## [v0.1.6] - 2022-01-13

### Overview

- Bugfix
- built using Go 1.17.6
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Fixed

- (GH-66) Incorrect "success" HTTP status code range identification

## [v0.1.5] - 2022-01-13

### Overview

- Bugfixes
- Dependency updates
- built using Go 1.17.6
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `Go`
    - `1.17.5` to `1.17.6`
  - `github/codeql-action`
    - `v1.0.26` to `v1.0.27`
  - `sanity-io/litter`
    - `v1.5.1` to `v1.5.2`

### Fixed

- (GH-58) Fix typos, update advice for plugin errors
- (GH-60) components plugin emits incomplete one-line summary for failure to
  parse JSON feed from URL
- (GH-61) components plugin emits duplicate (wrapped) error when for failure
  to parse JSON feed

## [v0.1.4] - 2022-01-06

### Overview

- Bugfixes
- Download links filename tweak
- built using Go 1.17.5
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- (GH-48) Add project name to generated download links file
- (GH-49) Review & update `timeout-minutes` setting (if needed) for all GitHub
  Actions Workflows

### Fixed

- (GH-52) components plugin timeout reported as OK status

## [v0.1.3] - 2022-01-02

### Overview

- Formatting tweak
- built using Go 1.17.5
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- (GH-44) Using tabs instead of spaces does not appear to improve
  `LongServiceOutput` formatting

## [v0.1.2] - 2022-01-01

### Overview

- Dependency updates
- built using Go 1.17.5
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `atc0005/go-nagios`
    - `v0.8.1` to `v0.8.2`

- (GH-40) Attempt to use tabs to improve web UI formatting

## [v0.1.1] - 2021-12-31

### Overview

- Dependency updates
- built using Go 1.17.5
  - Statically linked
  - Windows (x86, x64)
  - Linux (x86, x64)

### Changed

- Dependencies
  - `Go`
    - (GH-35) Update go.mod file, canary Dockerfile to reflect Go 1.17
    - `1.16.12` to `1.17.5`
  - `actions/setup-node`
    - `v2.5.0` to `v2.5.1`

- (GH-33) Drop `Group ID`, `Component ID`, `Evaluated` columns from
  `reports.ComponentsReport()` output

- (GH-30) Update Nagios command definition examples to use `--omit-ok` flag

## [v0.1.0] - 2021-12-27

Initial release!

This release provides early versions of tooling used to evaluate `components`
of Statuspage powered sites. This evaluation is performed using official APIs.

### Added

- `lscs`, a CLI app to list Components in multiple output formats.

- `check_statuspage_components`, a Nagios plugin to monitor one, many or all
  components.

[Unreleased]: https://github.com/atc0005/check-statuspage/compare/v0.1.10...HEAD
[v0.1.10]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.10
[v0.1.9]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.9
[v0.1.8]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.8
[v0.1.7]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.7
[v0.1.6]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.6
[v0.1.5]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.5
[v0.1.4]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.4
[v0.1.3]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.3
[v0.1.2]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.2
[v0.1.1]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.1
[v0.1.0]: https://github.com/atc0005/check-statuspage/releases/tag/v0.1.0
