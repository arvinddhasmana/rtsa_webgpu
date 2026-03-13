// CLASSIFICATION: UNCLASSIFIED
// src/types/vite-env.d.ts — Vite environment variable type declarations

/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_MOCK_SENSORS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
