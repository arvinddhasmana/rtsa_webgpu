// CLASSIFICATION: UNCLASSIFIED
// src/services/operators.ts
//
// Mock service for operator retrieval and search.
// Used for the Assignment Quick-Action flow in the Alert Sidebar.


export type OperatorStatus = "online" | "offline" | "busy";

export interface Operator {
  id: string;
  name: string;
  status: OperatorStatus;
  avatarUrl?: string; // Optional for future UI enhancement
}

/** Seed data for testing and demo. */
const MOCK_OPERATORS: Operator[] = [
  { id: "OP-103", name: "Sarah Chen", status: "online" },
  { id: "OP-215", name: "Mike Ross", status: "online" },
  { id: "OP-409", name: "Jessica Pearson", status: "busy" },
  { id: "OP-056", name: "Harvey Specter", status: "online" },
  { id: "OP-772", name: "Louis Litt", status: "online" },
  { id: "OP-331", name: "Donna Paulsen", status: "online" },
  { id: "OP-992", name: "Rachel Zane", status: "offline" },
];

/**
 * Searches for operators based on query string (name or ID).
 * Returns a subset of mock data.
 */
export async function searchOperators(query: string): Promise<Operator[]> {
  // Simulate network latency
  await new Promise((resolve) => setTimeout(resolve, 150));

  const q = query.toLowerCase().trim();
  if (!q) return MOCK_OPERATORS;

  return MOCK_OPERATORS.filter(
    (op) => op.id.toLowerCase().includes(q) || op.name.toLowerCase().includes(q)
  );
}

/**
 * Fetches an operator by ID.
 */
export async function getOperatorById(id: string): Promise<Operator | undefined> {
  await new Promise((resolve) => setTimeout(resolve, 50));
  return MOCK_OPERATORS.find((op) => op.id === id);
}

/**
 * Provides access to the full list of operators for UI pickers.
 */
export function getAvailableOperators(): Operator[] {
  return [...MOCK_OPERATORS];
}
