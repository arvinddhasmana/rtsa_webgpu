<!-- CLASSIFICATION: UNCLASSIFIED -->
# UC014 — NATO Outbound Data Exchange

> **Use Case ID**: UC014
> **Feature**: FEAT-15 (NATO Interoperability)
> **Priority**: MUST
> **Actors**: System (automated), NATO Interop Officer
> **Classification**: UNCLASSIFIED
> **Last Updated**: 2026-02-23

---

## 1. Description

The RTSA system exports fused entity tracks and anomaly alerts to NATO-allied systems using STANAG 5516 (Link 16 J-Series messages), NFFI (NATO Friendly Force Information), and MIP (Multilateral Interoperability Programme) formats. Outbound data is classification-guarded, format-transformed, and published through NATO communication channels. Classification downgrade rules ensure no data above the authorized release level exits the system.

## 2. Preconditions

- Fused entity tracks available (UC008)
- NATO communication link established (Link 16, IP-based NFFI/MIP)
- Cross-domain guard configured with release policies
- System has active NATO exercise or operational authorization

## 3. Triggers

- New fused track created or updated (automatic)
- Anomaly alert generated that meets NATO release criteria
- Manual track nomination by NATO Interop Officer
- Scheduled bulk position report (every 60 seconds for Link 16)

## 4. Main Flow

```mermaid
sequenceDiagram
    participant RP as Redpanda<br/>tracks.fused.*
    participant EXP as NATO Export<br/>Service
    participant CDG as Cross-Domain<br/>Guard
    participant XLAT as Format<br/>Translator
    participant LINK as Link 16<br/>Terminal
    participant NFFI as NFFI/MIP<br/>Gateway
    participant AUDIT as Audit Trail

    RP->>EXP: Consume fused track event
    EXP->>EXP: Evaluate release policy:<br/>- Classification ≤ NATO SECRET<br/>- Releasable to alliance<br/>- No NOFORN restriction

    alt Track is releasable
        EXP->>CDG: Submit for cross-domain review
        CDG->>CDG: Content inspection:<br/>- Sanitize sensitive fields<br/>- Remove source attribution<br/>- Apply REL TO marking
        CDG-->>EXP: Sanitized track data

        par Link 16 Export
            EXP->>XLAT: Transform to J3.2 (Track Report)
            XLAT->>LINK: Transmit J-Series message
            LINK->>AUDIT: Audit: nato_link16_exported
        and NFFI Export
            EXP->>XLAT: Transform to NFFI XML
            XLAT->>NFFI: Publish via MIP gateway
            NFFI->>AUDIT: Audit: nato_nffi_exported
        end
    else Track is NOT releasable
        EXP->>AUDIT: Audit: nato_export_blocked
        EXP->>EXP: Skip track (no export)
    end
```

## 5. Format Mappings

### 5.1 Internal Track → STANAG 5516 (Link 16)

| Internal Field | J-Series Field | Message | Notes |
|---|---|---|---|
| track_id | Track Number (TN) | J3.2 | 5-digit octal |
| latitude | Latitude | J3.2 | WGS-84, ±90° |
| longitude | Longitude | J3.2 | WGS-84, ±180° |
| speed_knots | Speed | J3.2 | Integer knots |
| heading_deg | Course | J3.2 | 0–360° |
| altitude_meters | Altitude | J3.2 | Feet (converted) |
| entity_type | Identity (IFF) | J3.2 | HOSTILE/FRIENDLY/NEUTRAL/UNKNOWN |
| classification | Exercise Indicator | J3.2 | Mapped per exercise rules |

### 5.2 Internal Track → NFFI XML

| Internal Field | NFFI Element | Path |
|---|---|---|
| track_id | `<TrackId>` | `/FFI/Track/TrackId` |
| latitude, longitude | `<Position>` | `/FFI/Track/Position/Latitude`, `Longitude` |
| speed_knots | `<Speed>` | `/FFI/Track/Kinematics/Speed` |
| heading_deg | `<Course>` | `/FFI/Track/Kinematics/Course` |
| entity_type | `<Identity>` | `/FFI/Track/Identity` |
| confidence_score | `<Quality>` | `/FFI/Track/Quality` |
| timestamp | `<DateTime>` | `/FFI/Track/DateTime` |

### 5.3 Anomaly Alert → MIP Share

| Internal Field | MIP Element | Notes |
|---|---|---|
| alert_id | ObservationId | UUID |
| track_id | SubjectTrackId | Reference to shared track |
| anomaly_type | ObservationType | Mapped to MIP taxonomy |
| severity | ThreatLevel | NORMAL/WATCH/ELEVATED/CRITICAL |
| confidence_score | Confidence | 0.0–1.0 |
| explanation | NarrativeText | Sanitized natural language |

## 6. Classification Guard Rules

| Rule | Description |
|---|---|
| Maximum classification | Only tracks ≤ NATO SECRET exported |
| NOFORN check | Tracks marked NOFORN never exported |
| Source sanitization | Source sensor details stripped (only "MULTI-SOURCE" attribution) |
| Confidence threshold | Only tracks with confidence ≥ 0.6 exported |
| Caveats applied | REL TO NATO added to all exported data |
| Content inspection | Free-text fields scanned for sensitive content |
| Rate limiting | Max 1000 tracks per second to Link 16 |

## 7. Alternative Flows

### 7a. Link 16 Degraded
- Terminal reports degraded link quality
- Reduce export rate to critical tracks only (HOSTILE + ELEVATED)
- Buffer non-critical tracks for retry
- Alert NATO Interop Officer

### 7b. Manual Track Nomination
- NATO Interop Officer selects specific tracks for export
- Officer provides justification and release authority
- System logs nomination with officer identity
- Track exported with manual authorization flag

### 7c. Exercise Mode
- All exported data tagged with Exercise Indicator
- Exercise-specific track numbering ranges used
- No impact on live operational track sharing

## 8. Security Considerations

- Cross-domain guard is mandatory for all outbound data
- Source attribution (specific sensor type) never leaves the system
- Complete audit trail: track_id, destination, format, timestamp, classification
- Rate limiting prevents data exfiltration via export channel
- Link 16 terminal in TEMPEST-approved enclosure
- NFFI/MIP gateway secured via mutual TLS

## 9. Requirements Traced

| Requirement | Description |
|---|---|
| CR-NATO-001 | Export tracks in STANAG 5516 Link 16 format |
| CR-NATO-002 | Export tracks in NFFI XML format |
| CR-NATO-003 | Classification guard for outbound data |
| CR-NATO-004 | Source sanitization for allied sharing |
| CR-NATO-005 | Audit trail for all NATO data exchanges |
