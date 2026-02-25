// CLASSIFICATION: UNCLASSIFIED
// src/utils/coordinates.ts

/**
 * Converts decimal degrees to DMS (Degrees, Minutes, Seconds) format.
 */
export function toDMS(
  decimal: number,
  isLatitude: boolean
): string {
  const direction = isLatitude
    ? decimal >= 0
      ? "N"
      : "S"
    : decimal >= 0
    ? "E"
    : "W";
  const abs = Math.abs(decimal);
  const degrees = Math.floor(abs);
  const minutesFloat = (abs - degrees) * 60;
  const minutes = Math.floor(minutesFloat);
  const seconds = ((minutesFloat - minutes) * 60).toFixed(1);
  return `${degrees}°${minutes}'${seconds}"${direction}`;
}

/**
 * Formats a position as DMS string pair.
 */
export function formatPosition(lat: number, lon: number): string {
  return `${toDMS(lat, true)} ${toDMS(lon, false)}`;
}

/**
 * Converts knots to km/h.
 */
export function knotsToKmh(knots: number): number {
  return knots * 1.852;
}
