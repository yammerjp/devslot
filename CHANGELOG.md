# Changelog

## [Unreleased]

### Added
- Add tagpr for automated release management via pull requests
- Integrate GoReleaser with tagpr workflow for automated binary releases

### Changed
- Replace manual tag-based release workflow with tagpr-based PR workflow

### Fixed
- Fix version embedding for GoReleaser builds (change const to var)