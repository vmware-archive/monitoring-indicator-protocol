# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- Schemas for Indicator Documents (and kubernetes objects) are provided in `schemas.yml`.
- Support for `${variableName}` metadata interpolation.
- Ability to query indicator registry for all documents with specific metadata value.


## [0.8.5] - 2019-08-20
### Fixed
- verification tool did not interpolate a metadata key in a promql query correctly if it was followed by an underscore.
