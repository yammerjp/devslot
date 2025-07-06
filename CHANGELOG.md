# Changelog

## [v0.2.5](https://github.com/yammerjp/devslot/compare/v0.2.4...v0.2.5) - 2025-07-06
- Implement automatic branch creation and fix environment variables by @yammerjp in https://github.com/yammerjp/devslot/pull/29

## [v0.2.4](https://github.com/yammerjp/devslot/compare/v0.2.3...v0.2.4) - 2025-07-06
- Fix GoReleaser build failure after project refactoring by @yammerjp in https://github.com/yammerjp/devslot/pull/27

## [v0.2.3](https://github.com/yammerjp/devslot/compare/v0.2.2...v0.2.3) - 2025-07-06
- Add E2E tests to CI workflow by @yammerjp in https://github.com/yammerjp/devslot/pull/19
- Refactor: Implement package structure following golang-standards/project-layout by @yammerjp in https://github.com/yammerjp/devslot/pull/21
- Fix help display and improve CLI usability by @yammerjp in https://github.com/yammerjp/devslot/pull/23
- Improve logging and exit handling by @yammerjp in https://github.com/yammerjp/devslot/pull/22
- Enhance create command implementation by @yammerjp in https://github.com/yammerjp/devslot/pull/25
- Add tests for list command by @yammerjp in https://github.com/yammerjp/devslot/pull/24
- Add executable hook scripts to boilerplate command by @yammerjp in https://github.com/yammerjp/devslot/pull/26

## [v0.2.2](https://github.com/yammerjp/devslot/compare/v0.2.1...v0.2.2) - 2025-07-05
- Implement init command for syncing bare repositories by @yammerjp in https://github.com/yammerjp/devslot/pull/7
- Integrate GoReleaser into tagpr workflow by @yammerjp in https://github.com/yammerjp/devslot/pull/18

## [v0.2.1](https://github.com/yammerjp/devslot/compare/v0.2.0...v0.2.1) - 2025-07-05
- Fix version embedding for GoReleaser builds by @yammerjp in https://github.com/yammerjp/devslot/pull/8
- Add tagpr for PR-based release management by @yammerjp in https://github.com/yammerjp/devslot/pull/9
- Fix: APP_IDをシークレットから読み取るように修正 by @yammerjp in https://github.com/yammerjp/devslot/pull/10
- Add clarifying comment about version override by @yammerjp in https://github.com/yammerjp/devslot/pull/12
- Configure tagpr to use git tags only by @yammerjp in https://github.com/yammerjp/devslot/pull/15

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
