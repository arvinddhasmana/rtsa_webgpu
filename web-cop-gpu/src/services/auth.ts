// CLASSIFICATION: UNCLASSIFIED
// src/services/auth.ts — JWT token acquisition for WebTransport authentication
//
// Fetches a short-lived JWT from the RTSA auth endpoint via the gRPC-Web
// cold-path gateway. The token is passed to the Data Worker so that it can
// authenticate the WebTransport QUIC connection.
//
// Security: tokens are never stored in localStorage or logged.
// Reference: docs/sdlc_guidelines/08_tech_specific/webtransport_guidelines.md §7.1

/**
 * POST /api/v1/auth/token response body.
 * In production this is issued by the gRPC auth service via the Envoy gateway.
 */
interface AuthTokenResponse {
  token: string;
}

/**
 * Fetch a short-lived JWT for the current operator session.
 *
 * The auth endpoint URL is read from VITE_API_GATEWAY_URL (shared with the
 * gRPC-Web cold path) and suffixed with /api/v1/auth/token.
 *
 * Returns `undefined` when:
 *   - No gateway URL is configured (local dev without a backend).
 *   - The auth endpoint returns an error status.
 *   - The network request fails.
 *
 * IMPORTANT: Never log the returned token value. (SDLC Rule 5)
 */
export async function fetchAuthToken(): Promise<string | undefined> {
  const gatewayUrl = import.meta.env.VITE_API_GATEWAY_URL as string | undefined;
  if (!gatewayUrl) {
    return undefined;
  }

  const authUrl = `${gatewayUrl}/api/v1/auth/token`;

  try {
    const response = await fetch(authUrl, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include",
    });
    if (!response.ok) {
      return undefined;
    }
    const body = (await response.json()) as AuthTokenResponse;
    return body.token ?? undefined;
  } catch {
    return undefined;
  }
}
