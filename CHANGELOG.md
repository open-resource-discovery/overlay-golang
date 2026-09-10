# CHANGELOG

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).
This project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) rules.

## [unreleased]

## [[0.0.3](https://github.com/open-resource-discovery/overlay-golang/releases/tag/v0.0.3)] - 2026-09-11

- Ensure that 'remove of missing' is no-op by @mlakov in https://github.com/open-resource-discovery/overlay-golang/pull/27
- Update golang.org/x/exp digest to 85c1c22 by @renovate https://github.com/open-resource-discovery/overlay-golang/pull/28
- Update module github.com/ohler55/ojg to v1.28.6 by @renovate https://github.com/open-resource-discovery/overlay-golang/pull/30
- Add support for CSDL syntaxed enum type member annoations by @mlakov in https://github.com/open-resource-discovery/overlay-golang/pull/30

## [[0.0.2](https://github.com/open-resource-discovery/overlay-golang/releases/tag/v0.0.2)] - 2026-09-08

- Reject removing the document root by @Fannon in https://github.com/open-resource-discovery/overlay-golang/pull/11
- Remove enum members without deleting the enum type by @Fannon and @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/12
- Preserve root removal masks by @Fannon and @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/13
- Add 'to string' for models by @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/22
- Aggregate overlay errors by @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/19
- Apply JSONPath patches to nodes sorted in reverse by @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/23
- Distinguish zero-argument operation signatures by @Fannon and @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/16
- Reject OData v2 overlay targets by @Fannon and @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/20
- Remove complete annotation set by @Fannon and @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/21
- Correctly handle CSDL update for semantic selectors by @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/15
- Improve patch decomposition for CSDL semantic selectors by @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/24
- Add support for qualified EDMX annotations by @mlakov https://github.com/open-resource-discovery/overlay-golang/pull/24

## [[0.0.1](https://github.com/open-resource-discovery/overlay-golang/releases/tag/v0.0.1)] - 2026-08-28

- Initial implementation of ORD overlays in Golang by @mlakov

