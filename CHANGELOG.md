# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html);
while on `v0.x`, breaking changes bump the minor version.

## [Unreleased]

## [0.3.0] - 2026-09-27

### Added

- Fuzz targets for base URL normalization, guild storage and forgotten
  technology decoding, and item input normalization.
- Runnable godoc examples covering client construction and common calls.
- Documented REST contract tests backed by fixtures matching the PalDefender
  wiki revision `67883072` (2026-09-17).
- `govulncheck` and Go tip jobs, a coverage summary and action SHA pinning in
  CI; automated release validation and publication on version tags.
- Dependabot updates for GitHub Actions and Go modules.

### Changed

- Move the Go module path from `github.com/tagamer-net/pdrest-go` to
  `github.com/TagamerStudio/pdrest-go` and update import examples.
- Raise the minimum supported Go version from 1.25 to 1.26.
- `make check` now runs `go vet`; add `make vet` and `make fuzz` targets.
- Raise `MaxIdleConnsPerHost` of the internal transport to 16 to support
  concurrent requests.

### Fixed

- Reject base URLs with a force-query marker (for example `http://host?`).
- Sanitize response snippets in decode errors so they are always valid UTF-8.

## [0.2.0] - 2026-09-27

### Added

- `SummonPal` and `SummonNPC` methods with typed request and response models
  for the documented `/summon/pal` and `/summon/npc` routes.

### Fixed

- Avoid a panic when `http.DefaultTransport` was replaced by a
  non-`*http.Transport` value.
- Reject hostnames with an empty trailing label, trim surrounding whitespace
  from path parameters, and reject control characters in the bearer token,
  display address and origin at construction time.

### Changed

- CI gains least-privilege permissions, a job timeout and actions pinned to
  commit SHAs.

## [0.1.2] - 2026-09-08

### Added

- Configurable `Origin` header via `WithOrigin`.

## [0.1.1] - 2026-08-04

### Fixed

- Reject invalid, whitespace-only and dot-segment identifiers in request
  paths, and validate base camp identifiers as GUIDs before sending.
- Validate relic types against the documented whitelist, reject non-positive
  timeouts and invalid banlist `active` values.
- Reject duplicate `UserIDs`, whitespace-only `UserID` values, empty Pal egg
  selectors and string-numeric tuple counts.
- Preserve exact integer counts for `json.Number` inputs and reject
  fractional or out-of-range counts.
- Normalize the `All` technology token case-insensitively and ignore null or
  empty forgotten technology strings.
- Clear `GuildStorage` state before re-decoding and keep the `*APIError` when
  reading the error body fails.

### Changed

- Raise the error body cap to 64 KiB and document the 10 MiB success body cap.
- Trim surrounding whitespace from bearer tokens and merge repeated item
  grants across inputs, keeping the first-seen order.
- Test handlers no longer call `t.Fatal` from server goroutines.

## [0.1.0] - 2026-08-02

### Added

- Initial release: typed client for the PalDefender REST API (`/v1/pdapi`)
  covering version, guilds, players, pals, items, techs, progression, bans,
  IP bans, kicks, broadcasts, alerts, config reload, item/Pal/egg/progression
  grants, technology granting and the ban list, with context support,
  configurable timeouts, response size caps, custom HTTP client injection and
  structured `APIError` values.
