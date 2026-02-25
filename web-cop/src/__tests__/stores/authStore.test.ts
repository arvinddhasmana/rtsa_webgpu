// CLASSIFICATION: UNCLASSIFIED
// src/__tests__/stores/authStore.test.ts

import { describe, it, expect, beforeEach } from "vitest";
import { useAuthStore } from "../../stores/authStore";

describe("AuthStore", () => {
  beforeEach(() => {
    useAuthStore.getState().logout();
  });

  it("initial state is unauthenticated", () => {
    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().operatorId).toBeNull();
    expect(useAuthStore.getState().clearanceLevel).toBe("UNCLASSIFIED");
  });

  it("setOperator populates all fields and sets isAuthenticated", () => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Alice",
      unit: "1st RCR",
      clearance: "SECRET",
      roles: ["OPERATOR"],
    });

    const state = useAuthStore.getState();
    expect(state.isAuthenticated).toBe(true);
    expect(state.operatorId).toBe("op-001");
    expect(state.operatorName).toBe("Alice");
    expect(state.unit).toBe("1st RCR");
    expect(state.clearanceLevel).toBe("SECRET");
    expect(state.roles).toContain("OPERATOR");
  });

  it("logout resets all state", () => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Alice",
      unit: "UNIT",
      clearance: "SECRET",
      roles: ["OPERATOR"],
    });
    useAuthStore.getState().logout();

    expect(useAuthStore.getState().isAuthenticated).toBe(false);
    expect(useAuthStore.getState().operatorId).toBeNull();
    expect(useAuthStore.getState().clearanceLevel).toBe("UNCLASSIFIED");
  });

  it("T05: canAccess returns true when clearance >= data classification", () => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Alice",
      unit: "UNIT",
      clearance: "PROTECTED_B",
      roles: [],
    });
    expect(useAuthStore.getState().canAccess("UNCLASSIFIED")).toBe(true);
    expect(useAuthStore.getState().canAccess("PROTECTED_A")).toBe(true);
    expect(useAuthStore.getState().canAccess("PROTECTED_B")).toBe(true);
  });

  it("T06: canAccess returns false when clearance < data classification", () => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Alice",
      unit: "UNIT",
      clearance: "PROTECTED_A",
      roles: [],
    });
    expect(useAuthStore.getState().canAccess("PROTECTED_B")).toBe(false);
    expect(useAuthStore.getState().canAccess("PROTECTED_C")).toBe(false);
    expect(useAuthStore.getState().canAccess("SECRET")).toBe(false);
  });

  it("hasRole returns true for assigned roles", () => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Alice",
      unit: "UNIT",
      clearance: "UNCLASSIFIED",
      roles: ["OPERATOR", "ANALYST"],
    });
    expect(useAuthStore.getState().hasRole("OPERATOR")).toBe(true);
    expect(useAuthStore.getState().hasRole("ANALYST")).toBe(true);
  });

  it("hasRole returns false for unassigned roles", () => {
    useAuthStore.getState().setOperator({
      id: "op-001",
      name: "Alice",
      unit: "UNIT",
      clearance: "UNCLASSIFIED",
      roles: ["OPERATOR"],
    });
    expect(useAuthStore.getState().hasRole("ADMIN")).toBe(false);
  });
});
