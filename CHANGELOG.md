# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.8.9] - 2019-10-03
### Added
- Adds support for filtering on indicator type in grafana controller

## [0.8.8] - 2019-10-02
### Fixed
- The verification tool no longer errors on v0 documents when overriding metadata.

## [0.8.7] - 2019-09-30
### Added
- Support for metadata interpolation anywhere, for example, in the value of a threshold.
### Updated
- Bumped Golang version to 1.13

## [0.8.6] - 2019-09-20
### Added
- Schemas for Indicator Documents (and kubernetes objects) are provided in `schemas.yml`.
- Support for `${variableName}` metadata interpolation.
- Ability to query indicator registry for all documents with specific metadata value.


## [0.8.5] - 2019-08-20
### Fixed
- verification tool did not interpolate a metadata key in a promql query correctly if it was followed by an underscore.

