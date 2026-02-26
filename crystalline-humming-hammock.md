# Scenario Modes — Implementation Plan

## Context

Currently every stage shows the question immediately — there's no proof-of-presence. The `lat`/`lng` fields on stages are stored but unused. We want to add **scenario modes** that control what happens at each stage: whether players must prove they're at the location before seeing/answering a question, and whether there's a question at all.

The mode is **scenario-level** (all stages share it) and gets **copied to the game** at creation time, same as stages.

## The 5 Modes

| Mode | At the location... | Stage fields used |
|------|-------------------|-------------------|
| `classic` | Question shown immediately (current behavior) | clue, question, correctAnswer |
| `qr_quiz` | Scan QR to unlock, then answer question | clue, question, correctAnswer, unlockCode |
| `qr_hunt` | Scan QR — that's it, stage done | clue, unlockCode |
| `math_puzzle` | Enter calculated code (teamSecret + locationNumber) — stage done | clue, locationNumber |
| `guided` | Supervisor taps "Unlock", then optionally answer question | clue, (question, correctAnswer if hasQuestions) |

## Data Model Changes

### `AdminStage` (handle_admin_scenarios.go) — add 2 optional fields:
```go
UnlockCode     string `json:"unlockCode,omitempty"`     // qr_quiz, qr_hunt
LocationNumber int    `json:"locationNumber,omitempty"` // math_puzzle
```

### `scenario` (store_docs.go) — add:
```go
Mode         string `json:"mode"`                  // "classic"|"qr_quiz"|"qr_hunt"|"math_puzzle"|"guided"
HasQuestions bool   `json:"hasQuestions,omitempty"` // only for "guided" mode
```

### `game` (store_docs.go) — add same fields (copied from scenario):
```go
Mode         string `json:"mode"`
HasQuestions bool   `json:"hasQuestions,omitempty"`
```

### `team` (store_docs.go) — add:
```go
UnlockedStages []int  `json:"unlockedStages,omitempty"` // stage numbers unlocked by this team
TeamSecret     int    `json:"teamSecret,omitempty"`     // random 3-digit number for math_puzzle
```

### `gameStateData` (store.go) — add:
```go
Mode           string
HasQuestions   bool
UnlockedStages []int
TeamSecret     int
```

### `AdminScenarioRequest`, `AdminScenarioSummary`, `AdminScenarioDetail` — add `Mode`, `HasQuestions`

### `AdminGameRequest` — add `Mode string`, `HasQuestions bool` (set by handler from scenario, not from JSON)

### `AdminGameSummary`, `AdminGameDetail`, `AdminGameStatus` — add `Mode string`

### Player-facing types (handle_game_state.go):

`GameInfo` — add:
```go
Mode         string `json:"mode"`
HasQuestions bool   `json:"hasQuestions,omitempty"`
```

`StageInfo` — add:
```go
Locked         bool `json:"locked"`                   // true = needs unlock before question
LocationNumber int  `json:"locationNumber,omitempty"` // math_puzzle only
```

`GameStateResponse` — add:
```go
TeamSecret int `json:"teamSecret,omitempty"` // math_puzzle only
```

### Frontend types (types.ts, adminTypes.ts) — mirror all the above

## New Endpoint: `POST /api/{client}/game/unlock`

New file: `handle_unlock.go`

**Request:** `{ "code": "string" }` — QR payload, math result, or empty for guided

**Response:**
```go
type UnlockResponse struct {
    StageNumber   int        `json:"stageNumber"`
    Unlocked      bool       `json:"unlocked"`
    StageComplete bool       `json:"stageComplete,omitempty"` // true for no-question modes
    NextStage     *StageInfo `json:"nextStage,omitempty"`     // only if stageComplete
    GameComplete  bool       `json:"gameComplete,omitempty"`
    Question      string     `json:"question,omitempty"`      // revealed after unlock for quiz modes
}
```

**Logic — switch on mode:**
- `classic` → 409 error (classic doesn't use unlock)
- `qr_quiz` → validate code == stage.UnlockCode → record unlock → return question
- `qr_hunt` → validate code == stage.UnlockCode → record unlock + record stageResult (auto-complete) → return nextStage
- `math_puzzle` → validate code == strconv.Itoa(teamSecret + stage.LocationNumber) → record unlock + stageResult → return nextStage
- `guided` → require role == supervisor → record unlock → if hasQuestions: return question, else: record stageResult → return nextStage

Publishes SSE event `stage_unlocked` (or `stage_completed` for auto-complete modes).

## New Store Methods

Add to `Store` interface and `DocStore`:
```go
UnlockStage(ctx context.Context, gameID, teamID string, stageNumber int) error
```

Uses `modifyGame` to append stageNumber to `team.UnlockedStages`. No-op if already present.

No separate `IsStageUnlocked` method needed — the unlock status comes from `gameStateData.UnlockedStages` which is already loaded in `GameState()`.

## Modified: `handle_answer.go`

Add two guards near the top (after determining currentStageNum):

1. If mode doesn't have questions → 409 "this mode does not use questions"
2. If mode requires unlock and stage not unlocked → 409 "stage not unlocked"

Also: in the nextStage response, set `Locked: true` for non-classic modes (next stage starts locked).

Helper function:
```go
func modeHasQuestion(mode string, hasQuestions bool) bool {
    switch mode {
    case "classic", "qr_quiz":
        return true
    case "guided":
        return hasQuestions
    default:
        return false
    }
}
```

## Modified: `handle_game_state.go`

When building `StageInfo` for the current stage:
- `classic` → include question (current behavior), `Locked: false`
- Other modes → check if stageNumber is in UnlockedStages
  - Not unlocked: `Locked: true`, omit question
  - Unlocked + mode has questions: `Locked: false`, include question
  - Unlocked + no questions: shouldn't happen (stage would be completed)
- `math_puzzle`: include `LocationNumber` on StageInfo, include `TeamSecret` on response

## Modified: Scenario Validation (`AdminScenarioRequest.validate()`)

- Default mode to `"classic"` if empty
- Validate mode is one of 5 valid values
- `classic`, `qr_quiz`: require question + correctAnswer per stage
- `qr_quiz`, `qr_hunt`: auto-generate unlockCode per stage if empty
- `qr_hunt`: question/correctAnswer not required
- `math_puzzle`: require locationNumber per stage, question not required
- `guided` + hasQuestions: require question + correctAnswer
- `guided` without: question not required

## Modified: Game/Team Creation

**`CreateGame`** — copy mode + hasQuestions from scenario (via AdminGameRequest).

**`CreateTeam`** — if game mode is `math_puzzle`, generate `TeamSecret` (random 100-999).

**`AdminTeamItem`** — add `TeamSecret int` field so admin can see/share it with teams.

## Backward Compatibility

Existing scenarios/games have no `mode` in JSONB → deserializes as empty string. Backfill in `getGame()` and `GetScenario()`:
```go
if g.Mode == "" {
    g.Mode = "classic"
}
```

Classic mode with empty `UnlockedStages` skips all unlock checks → existing behavior unchanged.

## Frontend Changes

### api.ts — new function:
```typescript
export function unlockStage(client: string, code: string): Promise<UnlockResponse>
```

### GamePage.tsx — phase state machine:

Replace current `stageStartedAt` with `stagePhase: 'interstitial' | 'unlocking' | 'answering'`:

- **interstitial** — show clue + "Go to stage" button (all modes, current behavior)
- **unlocking** — mode-specific unlock UI:
  - `qr_quiz` / `qr_hunt`: text input for QR code + "Submit Code" button (camera QR scanning is future enhancement)
  - `math_puzzle`: show team secret, text input for calculated code
  - `guided` + supervisor: "Unlock Stage" button; non-supervisor: "Waiting for guide..."
- **answering** — question + answer form (classic, qr_quiz, guided-with-questions)

Phase transitions:
- Click "Go to stage" → `unlocking` (non-classic) or `answering` (classic)
- Successful unlock → `answering` (if mode has questions) or `fetchState()` (stage auto-completed)
- SSE `stage_unlocked` → transition teammates from `unlocking` to `answering`
- Stage number changes → reset to `interstitial`

### AdminScenarioEditorPage.tsx:

- Mode dropdown at scenario level (above stages)
- `guided` mode: checkbox "Include questions at each stage"
- Stage fields shown/hidden based on mode:
  - `classic`: question, correctAnswer (required)
  - `qr_quiz`: question, correctAnswer (required) + unlockCode (auto-generated, shown read-only)
  - `qr_hunt`: unlockCode only (auto-generated)
  - `math_puzzle`: locationNumber only (required)
  - `guided` + questions: question, correctAnswer
  - `guided` - questions: no extra fields

### AdminScenariosPage.tsx — show mode label in list

## Routes (routes.go)

Add inside the player `r.Route("/api/{client}", ...)` block:
```go
r.Post("/game/unlock", handleUnlock(broker))
```

## Seed

Set `Mode: "classic"` explicitly on demo scenario in `SeedDemoScenario`.

## File Change Summary

| File | Change |
|------|--------|
| `server/handle_admin_scenarios.go` | Add Mode/HasQuestions to request/response types, UnlockCode/LocationNumber to AdminStage, mode-dependent validation |
| `server/handle_admin_games.go` | Add Mode to AdminGameRequest/Summary/Detail/Status, pass mode through CreateGame |
| `server/store_docs.go` | Add Mode/HasQuestions to scenario+game structs, UnlockedStages/TeamSecret to team, backfill classic, generate team secrets, implement UnlockStage |
| `server/store.go` | Add fields to gameStateData, add UnlockStage to Store interface |
| `server/handle_unlock.go` | **New file** — handleUnlock handler with mode dispatch, UnlockRequest/UnlockResponse types, modeHasQuestion helper |
| `server/handle_answer.go` | Add mode guards (unlock check, no-question rejection), locked flag on nextStage |
| `server/handle_game_state.go` | Add Mode/HasQuestions to GameInfo, Locked/LocationNumber to StageInfo, TeamSecret to response, conditional question visibility |
| `server/store_admin.go` | Backfill Mode="classic" in GetScenario/ListScenarios |
| `server/routes.go` | Add `/game/unlock` route |
| `server/seed.go` | Explicit Mode on demo scenario |
| `web/src/types.ts` | Add mode/locked/teamSecret/UnlockResponse types |
| `web/src/api.ts` | Add unlockStage() |
| `web/src/GamePage.tsx` | Phase state machine, conditional unlock/answer UI per mode |
| `web/src/admin/adminTypes.ts` | Add mode/hasQuestions/unlockCode/locationNumber to types |
| `web/src/admin/AdminScenarioEditorPage.tsx` | Mode picker, conditional stage fields |
| `web/src/admin/AdminScenariosPage.tsx` | Mode label in list |

## Implementation Order

1. Backend data model (structs, store interface, backfill)
2. Scenario validation (mode-dependent field requirements)
3. Game/team creation (copy mode, generate team secrets)
4. Game state endpoint (mode-aware stage info)
5. Unlock endpoint (new handle_unlock.go)
6. Answer endpoint (mode guards)
7. Routes (wire unlock)
8. Frontend types + API
9. GamePage (phase state machine)
10. Admin scenario editor (mode picker, conditional fields)
11. Admin list pages (mode display)
12. Seed (explicit classic)

## Verification

1. `go build ./...` — compiles
2. `go test ./...` — existing tests pass (classic mode unchanged)
3. Manual test: create scenario in each mode via admin UI, create game, play through
4. Verify backward compat: existing demo game still works without migration
