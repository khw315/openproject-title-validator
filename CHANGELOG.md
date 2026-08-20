# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] - 2026-08-21

### Fixed
- Added in-memory subject deduplication cache in webhook handler to prevent infinite comment feedback loops when OpenProject triggers `work_package:updated` events after a comment is posted.
- Sanitized log output to prevent logging unvalidated user input directly in server logs.

## [0.1.1] - 2026-08-21

### Fixed
- Added `X-Forwarded-Proto: https` and `X-Forwarded-Ssl: on` headers when communicating with OpenProject API to prevent Rails SSL enforcement redirect loops inside container bridge networks.

## [0.1.0] - 2026-08-21

### Added
- OpenProject webhook receiver service to validate work package (ticket) titles automatically upon creation and updates.
- Default 4-segment bracket format rule: `[TC-NNN][NAMA FEATURE][SUB FEATURE][PIC TESTER]`.
- Detailed structural violation feedback pinpointing invalid test case numbers, missing brackets, empty fields, and extra characters.
- Automated comment notifications mentioning ticket authors with exponential retry logic on transient API failures.
- Configurable regex validation patterns via `TITLE_PATTERN` and format descriptions via `TITLE_CRITERIA_DESC`.
- Customizable comment template via `COMMENT_TEMPLATE` with support for `{author}`, `{criteria}`, and `{violations}` placeholders.
- Optional HMAC-SHA1 webhook signature verification via `WEBHOOK_SECRET`.
- Service health check endpoint (`/health`) and HTTP request logging middleware.
- Multi-stage `Dockerfile` and `docker-compose.yml` container deployment configuration with internal bridge networking.
- Unit test suite with table-driven tests for configuration loading, regex parsing, comment formatting, and webhook handlers.
- Contribution guidelines and Git PR workflow policy in `AGENTS.md`.

[0.1.2]: https://github.com/khw315/openproject-title-validator/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/khw315/openproject-title-validator/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/khw315/openproject-title-validator/releases/tag/v0.1.0
