// CLASSIFICATION: UNCLASSIFIED
// tests/signals/auth.test.ts — Unit tests for the auth signal and operatorIdFromToken helper.

import { describe, it, expect, afterEach } from "vitest";
import { operatorId, setOperatorId, operatorIdFromToken } from "../../src/signals/auth";

afterEach(() => {
  // Reset to default after each test
  setOperatorId("anonymous");
});

describe("operatorId signal", () => {
  it("defaults to anonymous", () => {
    expect(operatorId()).toBe("anonymous");
  });

  it("setOperatorId updates the signal", () => {
    setOperatorId("op-42");
    expect(operatorId()).toBe("op-42");
  });

  it("resets back to anonymous when set to anonymous", () => {
    setOperatorId("some-operator");
    setOperatorId("anonymous");
    expect(operatorId()).toBe("anonymous");
  });
});

describe("operatorIdFromToken", () => {
  /**
   * Build a minimal JWT using base64url encoding (RFC 4648 §5) — the same
   * encoding real JWT issuers produce: "-" and "_" replace "+" and "/",
   * and padding ("=") is omitted.
   */
  function makeJwt(payload: Record<string, unknown>): string {
    const toBase64Url = (str: string): string =>
      btoa(str).replace(/\+/g, "-").replace(/\//g, "_").replace(/=+$/, "");
    const header = toBase64Url(JSON.stringify({ alg: "HS256", typ: "JWT" }));
    const body = toBase64Url(JSON.stringify(payload));
    return `${header}.${body}.fakesig`;
  }

  it("returns operator_id from a valid JWT", () => {
    const token = makeJwt({ operator_id: "op-007", clearance_level: 1, exp: 9999999999 });
    expect(operatorIdFromToken(token)).toBe("op-007");
  });

  it("returns anonymous when operator_id claim is absent", () => {
    const token = makeJwt({ sub: "some-user", exp: 9999999999 });
    expect(operatorIdFromToken(token)).toBe("anonymous");
  });

  it("returns anonymous for undefined token", () => {
    expect(operatorIdFromToken(undefined)).toBe("anonymous");
  });

  it("returns anonymous for empty string token", () => {
    expect(operatorIdFromToken("")).toBe("anonymous");
  });

  it("returns anonymous when operator_id is not a string", () => {
    const token = makeJwt({ operator_id: 42 });
    expect(operatorIdFromToken(token)).toBe("anonymous");
  });

  it("returns anonymous for malformed JWT (missing parts)", () => {
    expect(operatorIdFromToken("not.a")).toBe("anonymous");
  });

  it("returns anonymous for invalid base64 payload", () => {
    expect(operatorIdFromToken("header.!!!invalid!!!.sig")).toBe("anonymous");
  });

  it("returns anonymous when operator_id is empty string", () => {
    const token = makeJwt({ operator_id: "" });
    expect(operatorIdFromToken(token)).toBe("anonymous");
  });
});
