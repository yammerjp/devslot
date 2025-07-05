# Changelog

## [v0.2.1](https://github.com/yammerjp/devslot/compare/v0.2.0...v0.2.1) - 2025-07-05
- Fix version embedding for GoReleaser builds by @yammerjp in https://github.com/yammerjp/devslot/pull/8
- Add tagpr for PR-based release management by @yammerjp in https://github.com/yammerjp/devslot/pull/9
- Fix: APP_IDをシークレットから読み取るように修正 by @yammerjp in https://github.com/yammerjp/devslot/pull/10
- Add clarifying comment about version override by @yammerjp in https://github.com/yammerjp/devslot/pull/12

## [Unreleased]

### Added
- Add tagpr for automated release management via pull requests
- Integrate GoReleaser with tagpr workflow for automated binary releases

### Changed
- Replace manual tag-based release workflow with tagpr-based PR workflow

### Fixed
- Fix version embedding for GoReleaser builds (change const to var)
