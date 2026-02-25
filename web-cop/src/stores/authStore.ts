// CLASSIFICATION: UNCLASSIFIED
// src/stores/authStore.ts

import { create } from "zustand";
import { ClassificationLevel } from "../types/common";

const CLASSIFICATION_ORDER: ClassificationLevel[] = [
  "UNCLASSIFIED",
  "PROTECTED_A",
  "PROTECTED_B",
  "PROTECTED_C",
  "SECRET",
];

interface AuthState {
  operatorId: string | null;
  operatorName: string | null;
  unit: string | null;
  clearanceLevel: ClassificationLevel;
  roles: string[];
  isAuthenticated: boolean;

  setOperator: (operator: {
    id: string;
    name: string;
    unit: string;
    clearance: ClassificationLevel;
    roles: string[];
  }) => void;
  logout: () => void;

  canAccess: (dataClassification: ClassificationLevel) => boolean;
  hasRole: (role: string) => boolean;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  operatorId: null,
  operatorName: null,
  unit: null,
  clearanceLevel: "UNCLASSIFIED",
  roles: [],
  isAuthenticated: false,

  setOperator: (op) =>
    set({
      operatorId: op.id,
      operatorName: op.name,
      unit: op.unit,
      clearanceLevel: op.clearance,
      roles: op.roles,
      isAuthenticated: true,
    }),

  logout: () =>
    set({
      operatorId: null,
      operatorName: null,
      unit: null,
      clearanceLevel: "UNCLASSIFIED",
      roles: [],
      isAuthenticated: false,
    }),

  canAccess: (dataClassification) =>
    CLASSIFICATION_ORDER.indexOf(get().clearanceLevel) >=
    CLASSIFICATION_ORDER.indexOf(dataClassification),

  hasRole: (role) => get().roles.includes(role),
}));
