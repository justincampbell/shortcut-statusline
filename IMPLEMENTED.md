# Implemented fields

Tokens that exist today (`[x]`) and fields that could be added (`[ ]`). Fields
needing an extra API resource have it noted; everything else is on the object
we already fetch.

## Story

- [x] `{story.id}`
- [x] `{story.name}`
- [x] `{story.idName}` — `id: name`
- [x] `{story.url}` — `app_url`
- [x] `{story.state}` — via `/workflows` lookup
- [x] `{story.type}` — `story_type`, rendered title-cased (`Feature`/`Bug`/`Chore`) and colored cyan/red/yellow
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
- [x] `{story.owner}` — first owner's `mention_name` via `/members` lookup
- [x] `{story.requestor}` — `mention_name` via `/members` lookup
- [x] `{story.team}` — group `mention_name` via `/groups` lookup
- [ ] `{story.iteration}` — `iteration_id`, needs `/iterations` lookup
- [ ] `{story.labels}` — `labels[].name`, joined
- [ ] `{story.priority}` / other custom fields — `custom_fields`
- [ ] `{story.description}` — usually too long for a statusline; skip?

## Epic

- [x] `{epic.id}`
- [x] `{epic.name}`
- [x] `{epic.idName}` — `id: name`
- [x] `{epic.url}` — `app_url`
- [x] `{epic.state}` — via `/epic-workflow` lookup
- [ ] `{epic.deadline}`
- [ ] `{epic.planned_start_date}`
- [ ] `{epic.health}` — "onTrack" / "atRisk" / "offTrack"
- [ ] `{epic.started}` / `{epic.started_at}`
- [ ] `{epic.completed}` / `{epic.completed_at}`
- [ ] `{epic.created_at}` / `{epic.updated_at}`
- [ ] `{epic.archived}`
- [x] `{epic.owner}` — first owner's `mention_name` via `/members` lookup
- [x] `{epic.team}` — group `mention_name` via `/groups` lookup
- [ ] `{epic.labels}` — `labels[].name`
- [ ] `{epic.description}`

## Objective

- [x] `{objective.id}`
- [x] `{objective.name}`
- [x] `{objective.idName}` — `id: name`
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
- **Member** (`/members`) — already implemented; cached in `members.json` at
  the workflow TTL.
- **Group / Team** (`/groups`) — already implemented; cached in
  `groups.json` at the workflow TTL.
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
