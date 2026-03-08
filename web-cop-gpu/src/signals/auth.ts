// CLASSIFICATION: UNCLASSIFIED
// src/signals/auth.ts — Authenticated operator identity signal
//
// Holds the operator identity parsed from the JWT payload received from
// the auth service. The token is assumed to have been validated by the
// server (pkg/webtransport/auth.go); this module only reads the payload
// claims without re-verifying the signature.
//
// Components read `operatorId()` to identify the current operator; the
// App root sets it when a token is successfully acquired.
//
// Reference: docs/sdlc_guidelines/04_coding_standards/solidjs_standards.md §3.3
//            pkg/webtransport/auth.go — JWT claims: operator_id, clearance_level

import { createSignal } from "solid-js";

/**
 * Current authenticated operator identity.
 * Defaults to "anonymous" when no JWT has been acquired yet.
 * Set by App.tsx after fetchAuthToken() returns a valid token.
 *
 * IMPORTANT: Never log this value — it is a user identifier. (SDLC Rule 5)
 */
export const [operatorId, setOperatorId] = createSignal<string>("anonymous");

/**
 * Parse the `operator_id` claim from a JWT payload section.
 *
 * JWTs use base64url encoding (RFC 4648 §5): "-" and "_" replace "+" and "/",
 * and padding ("=") is omitted. This helper normalises the segment to standard
 * base64 before decoding so that tokens produced by real issuers are handled
 * correctly.
 *
 * The signature is NOT verified here. The token is assumed to have been
 * issued and validated by the RTSA auth service before being passed to the
 * frontend. Returns "anonymous" if the claim is absent or unparseable.
 *
 * IMPORTANT: Do not log the token or the extracted identity. (SDLC Rule 5)
 */
export function operatorIdFromToken(token: string | undefined): string {
  if (!token) return "anonymous";
  try {
    const parts = token.split(".");
    if (parts.length !== 3 || !parts[1]) return "anonymous";
    // Normalise base64url → standard base64 (RFC 4648 §5 → §4)
    const b64 = parts[1].replace(/-/g, "+").replace(/_/g, "/");
    // Pad to the next multiple of 4: 0 bytes if already aligned, otherwise 1–3.
    const padded = b64 + "=".repeat((4 - (b64.length % 4)) % 4);
    const payload = JSON.parse(atob(padded)) as { operator_id?: string };
    const id = payload.operator_id;
    if (typeof id === "string" && id.length > 0) return id;
    return "anonymous";
  } catch {
    return "anonymous";
  }
}
