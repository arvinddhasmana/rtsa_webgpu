<!-- CLASSIFICATION: UNCLASSIFIED -->
# Phase 4: Security Officer & NATO Liaison — Detail Implementation Plan

**Status**: ⬜ Planned
**Roles**: Security Officer, NATO Liaison
**Use Cases**: UC010, UC011, UC014, UC015

---

## Scope

### 4.1 Security Officer — Audit & Feedback Dashboard
- **Feedback Log**: Table of all operator feedback submissions, trust scores, anti-poisoning decisions
- **Audit Log**: Immutable event stream from ClickHouse `audit_log` table
- **Trust Score Distribution**: Histogram of trust scores by operator
- **Alert Assignment History**: Who assigned what, when
- Filters: operator ID, feedback type, time range, trust score threshold

### 4.2 NATO Liaison — NATO Exchange Dashboard
- **Link Connectivity Header**: Link 16, NFFI, MIP status indicators
- **Track Nomination Queue**: Outbound tracks awaiting review with `[Nominate]` / `[Revoke]` buttons
- **Inbound Allied Tracks**: NATO-shared tracks displayed on map with NATO icon and `REL TO` labels
- **Classification Guard**: Visual indicator when outbound track is blocked by classification ceiling
- Audit trail for all nomination/revocation actions

---

## Components to Build

| Component | File | Notes |
|---|---|---|
| `AuditDashboard.tsx` | `src/components/layout/` | Security Officer default view |
| `FeedbackLogTable.tsx` | `src/components/audit/` | Sortable/filterable feedback log |
| `TrustScoreHistogram.tsx` | `src/components/audit/` | SVG histogram |
| `NatoExchangeDashboard.tsx` | `src/components/layout/` | NATO Liaison default view |
| `NominationQueue.tsx` | `src/components/nato/` | Track nomination + revocation |
| `InboundTracksPanel.tsx` | `src/components/nato/` | Allied track list |
| `LinkStatusHeader.tsx` | `src/components/nato/` | Link 16 / NFFI connectivity |

## Reused from Phase 2-3
- `EntityDetailPanel.tsx`, `EventTimeline.tsx`, `FeedbackForm.tsx`

---

## Key APIs

| API | Usage |
|---|---|
| `AuditService.GetAuditLog` | Immutable event log |
| `FeedbackService.GetFeedbackHistory` | Historical feedback submissions |
| `NatoAdapterService.NominateTrack` | Outbound nomination |
| `NatoAdapterService.RevokeNomination` | Revoke outbound track |
| `NatoAdapterService.ListInboundTracks` | Allied tracks |
