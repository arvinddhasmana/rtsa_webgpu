// CLASSIFICATION: UNCLASSIFIED
// src/signals/auth.ts — Authenticated operator identity signal
//
// Holds the operator identity extracted from the validated JWT claim.
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
 * Decode the `operator_id` claim from a JWT payload without verifying
 * the signature. Returns "anonymous" if the claim is absent or unparseable.
 *
 * IMPORTANT: Do not log the token or the extracted identity. (SDLC Rule 5)
 */
export function operatorIdFromToken(token: string | undefined): string {
  if (!token) return "anonymous";
  try {
    const parts = token.split(".");
    if (parts.length !== 3 || !parts[1]) return "anonymous";
    const payload = JSON.parse(atob(parts[1])) as { operator_id?: string };
    const id = payload.operator_id;
    if (typeof id === "string" && id.length > 0) return id;
    return "anonymous";
  } catch {
    return "anonymous";
  }
}
