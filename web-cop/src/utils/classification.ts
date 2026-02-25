// CLASSIFICATION: UNCLASSIFIED
// src/utils/classification.ts

import { ClassificationLevel } from "../types/common";

export interface ClassificationStyle {
  backgroundColor: string;
  textColor: string;
  label: string;
}

const CLASSIFICATION_STYLES: Record<ClassificationLevel, ClassificationStyle> =
  {
    UNCLASSIFIED: {
      backgroundColor: "#008000",
      textColor: "#FFFFFF",
      label: "UNCLASSIFIED",
    },
    PROTECTED_A: {
      backgroundColor: "#0000FF",
      textColor: "#FFFFFF",
      label: "PROTECTED A",
    },
    PROTECTED_B: {
      backgroundColor: "#0000FF",
      textColor: "#FFFFFF",
      label: "PROTECTED B",
    },
    PROTECTED_C: {
      backgroundColor: "#FF0000",
      textColor: "#FFFFFF",
      label: "PROTECTED C",
    },
    SECRET: {
      backgroundColor: "#FF0000",
      textColor: "#FFFFFF",
      label: "SECRET",
    },
  };

export function getClassificationStyle(
  level: ClassificationLevel
): ClassificationStyle {
  return CLASSIFICATION_STYLES[level];
}

const CLASSIFICATION_ORDER: ClassificationLevel[] = [
  "UNCLASSIFIED",
  "PROTECTED_A",
  "PROTECTED_B",
  "PROTECTED_C",
  "SECRET",
];

export function getHighestClassification(
  levels: ClassificationLevel[]
): ClassificationLevel {
  if (levels.length === 0) return "UNCLASSIFIED";
  return levels.reduce((highest, current) => {
    return CLASSIFICATION_ORDER.indexOf(current) >
      CLASSIFICATION_ORDER.indexOf(highest)
      ? current
      : highest;
  });
}

export function isAccessible(
  operatorClearance: ClassificationLevel,
  dataClassification: ClassificationLevel
): boolean {
  return (
    CLASSIFICATION_ORDER.indexOf(operatorClearance) >=
    CLASSIFICATION_ORDER.indexOf(dataClassification)
  );
}
