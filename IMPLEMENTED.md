# Implemented fields

Tokens that exist today (`[x]`) and fields that could be added (`[ ]`). Fields
needing an extra API resource have it noted; everything else is on the object
we already fetch.

## Story

- [x] `{story.id}`
- [x] `{story.name}`
- [x] `{story.url}` — `app_url`
- [x] `{story.state}` — via `/workflows` lookup
- [ ] `{story.type}` — `story_type` ("feature" / "bug" / "chore")
- [ ] `{story.deadline}`
- [ ] `{story.estimate}`
- [ ] `{story.started}` / `{story.started_at}`
- [ ] `{story.completed}` / `{story.completed_at}`
- [ ] `{story.created_at}` / `{story.updated_at}` / `{story.moved_at}`
- [ ] `{story.cycle_time}` / `{story.lead_time}`
- [ ] `{story.comments}` — count
- [ ] `{story.tasks}` — count, or open/total
- [ ] `{story.branches}` — count
- [ ] `{story.pull_requests}` — count
- [ ] `{story.blocked}` / `{story.blocker}` — booleans
- [ ] `{story.archived}` — boolean
- [ ] `{story.owner}` — needs `/members` lookup (UUIDs → mention_name)
- [ ] `{story.requestor}` — needs `/members` lookup
- [ ] `{story.team}` — `group_id`, needs `/groups` lookup
- [ ] `{story.iteration}` — `iteration_id`, needs `/iterations` lookup
- [ ] `{story.labels}` — `labels[].name`, joined
- [ ] `{story.priority}` / other custom fields — `custom_fields`
- [ ] `{story.description}` — usually too long for a statusline; skip?

## Epic

- [x] `{epic.id}`
- [x] `{epic.name}`
- [x] `{epic.url}` — `app_url`
- [x] `{epic.state}` — via `/epic-workflow` lookup
- [ ] `{epic.deadline}`
- [ ] `{epic.planned_start_date}`
- [ ] `{epic.health}` — "onTrack" / "atRisk" / "offTrack"
- [ ] `{epic.started}` / `{epic.started_at}`
- [ ] `{epic.completed}` / `{epic.completed_at}`
- [ ] `{epic.created_at}` / `{epic.updated_at}`
- [ ] `{epic.archived}`
- [ ] `{epic.owner}` — `owner_ids`, needs `/members` lookup
- [ ] `{epic.team}` — `group_id`, needs `/groups` lookup
- [ ] `{epic.labels}` — `labels[].name`
- [ ] `{epic.description}`

## Objective

- [x] `{objective.id}`
- [x] `{objective.name}`
- [x] `{objective.url}` — `app_url`
- [x] `{objective.state}` — direct field, no lookup
- [ ] `{objective.started}` / `{objective.started_at}`
- [ ] `{objective.completed}` / `{objective.completed_at}`
- [ ] `{objective.created_at}` / `{objective.updated_at}`
- [ ] `{objective.archived}`
- [ ] `{objective.categories}` — `categories[].name`
- [ ] `{objective.key_results}` — count (`len(key_result_ids)`)
- [ ] `{objective.description}`

## Other object types worth considering

These aren't statusline subjects on their own, but unlock fields above:

- **Iteration** (`/iterations/{id}`) — current sprint. Would expose
  `{story.iteration}` and possibly `{iteration.name|state|end_date}`. Lookup is
  per-iteration but very cacheable (iterations span weeks).
- **Member** (`/members`) — workspace-wide list. One workspace-scoped fetch
  unlocks `{story.owner}`, `{story.requestor}`, `{epic.owner}`. Same pattern as
  the workflow-state lookup we already have: long TTL, separate cache file.
- **Group / Team** (`/groups`) — workspace-wide list. Unlocks
  `{story.team}` / `{epic.team}`. Same cacheable-once pattern as Member.
- **Workflow / Epic Workflow** — already implemented (used for state name).
- **Label** — already on stories/epics inline (`labels[].name`); no separate
  fetch needed.

## What would *not* be worth implementing

- `entity_type`, `global_id`, `external_id`, `position` — internal/structural.
- `*_mention_ids`, `follower_ids`, `*_template_id` — IDs without obvious
  statusline value.
- `stats` blob — multi-field, hard to render in one token.
- Full `description` — invariably too long for a prompt.
- `linked_files`, `files`, `external_links` — too noisy.
