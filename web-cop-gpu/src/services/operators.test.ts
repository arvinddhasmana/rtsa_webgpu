// CLASSIFICATION: UNCLASSIFIED
// src/services/operators.test.ts

import { describe, expect, it } from "vitest";
import {
  getAvailableOperators,
  getOperatorById,
  searchOperators,
} from "./operators";

describe("Operators Service", () => {
  it("should return the full list of available operators", () => {
    const operators = getAvailableOperators();
    expect(operators.length).toBeGreaterThan(5);
    expect(operators[0]).toHaveProperty("id");
    expect(operators[0]).toHaveProperty("name");
    expect(operators[0]).toHaveProperty("status");
  });

  it("should find an operator by ID", async () => {
    const op = await getOperatorById("OP-103");
    expect(op).toBeDefined();
    expect(op?.name).toBe("Sarah Chen");
  });

  it("should return undefined for non-existent operator ID", async () => {
    const op = await getOperatorById("OP-999");
    expect(op).toBeUndefined();
  });

  it("should search operators by name (case-insensitive)", async () => {
    const results = await searchOperators("sarah");
    expect(results.length).toBe(1);
    expect(results[0].name).toBe("Sarah Chen");
  });

  it("should search operators by ID", async () => {
    const results = await searchOperators("OP-215");
    expect(results.length).toBe(1);
    expect(results[0].id).toBe("OP-215");
  });

  it("should return all operators for empty search query", async () => {
    const results = await searchOperators("");
    expect(results.length).toBe(getAvailableOperators().length);
  });

  it("should return empty list for no matches", async () => {
    const results = await searchOperators("NONEXISTENT");
    expect(results.length).toBe(0);
  });
});
