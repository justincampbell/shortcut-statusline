# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2026-05-20

### Added

- `{story.type}` token rendered title-cased (`Feature` / `Bug` / `Chore`) with
  per-type color (cyan / red / yellow).
- `{story.typeChar}` token — one-letter form (`F` / `B` / `C`) for tight prompts,
  same per-type colors.
- `{story.owner}` / `{story.requestor}` / `{story.team}` tokens, plus
  `…Mention` (`@`-prefixed) and `…Name` (full display name) variants, backed by
  `/members` and `/groups` lookups cached at the workflow TTL.
- `{epic.owner}` / `{epic.team}` (and the same `…Mention` / `…Name` variants).
- OSC8 hyperlink wrapping on `name` and `id` tokens for story, epic, and
  objective — clickable in supporting terminals, plain text elsewhere.
  Disable with `--no-links`, `SHORTCUT_STATUSLINE_NO_LINKS=1`, or `NO_COLOR`.
- State-based ANSI color on name / id / idName / state tokens, matching
  Shortcut's web UI (gray for backlog/unstarted, magenta for started, green for
  done).
- Story-ID resolution for branches that don't embed `sc-NNNNN` (e.g.
  `chore/sc-new-story/foo`) via Shortcut's `/search/stories?query=branch:<name>`
  endpoint. Results cached per branch in `branches.json`.

### Changed

- Default format is now `{story.idName} ({epic.name})` (was `{story.id} …`).
  Drops the redundant epic ID; story still shows `id: name`.
- Cached bundles missing owner / team / requestor data are refetched, so older
  caches pick up the new tokens without a manual `--refresh`.

## [0.1.0] - 2026-05-11

Initial release.

[0.2.0]: https://github.com/justincampbell/shortcut-statusline/compare/v0.1.0...v0.2.0
[0.1.0]: https://github.com/justincampbell/shortcut-statusline/releases/tag/v0.1.0
