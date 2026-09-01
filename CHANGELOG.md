# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) rules.


## [unreleased]

### Changed

- Selectors now return an error when `merge` or `update` matches no existing target.
- Documented idempotent unmatched removal, the prohibition on whole-root removal, valid root removal masks, and annotation-only EDMX behavior.

### Fixed

- Fixed root removal masks across all JSON-family processors while keeping absent mask fields as no-ops.
- Fixed unmatched EDMX removal to succeed as an idempotent no-op, consistent with the JSON-family processors.
- Fixed CSDL JSON enum-member removal so sibling annotation keys are removed with the member.
- Fixed EDMX annotation replacement and removal to use term-plus-qualifier identity.
- Fixed EDMX conversion errors to return through `Apply` instead of panicking.
- Fixed exact-arity EDMX operation signatures, including explicit zero-argument selectors and `FunctionImport` exclusion.

## [[0.0.1](https://github.com/open-resource-discovery/overlay-golang/releases/tag/v0.0.1)] - 2026-08-28

Initial implementation of ORD overlays in Golang
