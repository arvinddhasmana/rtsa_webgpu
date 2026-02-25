// CLASSIFICATION: UNCLASSIFIED
/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_GRPC_WEB_URL: string;
  readonly VITE_MAP_TILE_URL: string;
  readonly VITE_APP_TITLE: string;
  readonly VITE_CLASSIFICATION_CEILING: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
