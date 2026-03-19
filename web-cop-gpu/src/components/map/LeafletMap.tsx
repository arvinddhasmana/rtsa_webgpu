// CLASSIFICATION: UNCLASSIFIED
// src/components/map/LeafletMap.tsx

import L from "leaflet";
import "leaflet/dist/leaflet.css";
import { createEffect, onCleanup, onMount } from "solid-js";
import { mapStyle, setViewport, viewport } from "../../signals/viewport";

interface LeafletMapProps {
  onMapClick?: (x: number, y: number) => void;
}

export function LeafletMap(props: LeafletMapProps) {
  let mapDiv!: HTMLDivElement;
  let map: L.Map;
  let standardLayer: L.TileLayer;
  let hdLayer: L.TileLayer;

  onMount(() => {
    // Initialize map
    map = L.map(mapDiv, {
      center: [viewport().centerLat, viewport().centerLon],
      zoom: viewport().zoom,
      zoomControl: false, // UI might prefer to use custom controls or leave to mouse
      attributionControl: false, // Clean look for tactical UI
      zoomAnimation: false, // Prevents WebGPU desync during CSS fractional zoom animation
      markerZoomAnimation: false,
      worldCopyJump: false,
      maxBounds: [
        [-90, -180],
        [90, 180],
      ], // Limit map panning to true world bounds
      maxBoundsViscosity: 1.0,
      minZoom: 2,
    });

    // CartoDB Dark Matter (Standard)
    standardLayer = L.tileLayer("https://{s}.basemaps.cartocdn.com/dark_all/{z}/{x}/{y}{r}.png", {
      subdomains: "abcd",
      maxZoom: 20,
      crossOrigin: true, // Required for COEP and SharedArrayBuffer
      noWrap: true,      // Prevent multiple earths
      bounds: [[-90, -180], [90, 180]] as L.LatLngBoundsLiteral,
    });

    // Esri World Imagery (HD Satellite)
    hdLayer = L.tileLayer("https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}", {
      maxZoom: 19,
      crossOrigin: true, // Required for COEP and SharedArrayBuffer
      noWrap: true,      // Prevent multiple earths
      bounds: [[-90, -180], [90, 180]] as L.LatLngBoundsLiteral,
    });

    // Add initial layer
    if (mapStyle() === 1) {
      hdLayer.addTo(map);
    } else {
      standardLayer.addTo(map);
    }

    // Sync Leaflet state to global viewport signal
    // 'move' fires continuously during drag/zoom
    map.on("move", () => {
      const center = map.getCenter();
      const zoom = map.getZoom();
      // Keep lng clamped between -180 and 180 since we disabled wrapping
      let lng = center.lng;
      if (lng > 180) lng = 180;
      if (lng < -180) lng = -180;
      setViewport({ centerLat: center.lat, centerLon: lng, zoom });
    });

    // Forward clicks to WebGPU picker
    map.on("click", (e: L.LeafletMouseEvent) => {
      if (props.onMapClick) {
        props.onMapClick(e.containerPoint.x, e.containerPoint.y);
      }
    });

    onCleanup(() => {
      map.remove();
    });
  });

  // Sync global mapStyle signal to Leaflet layers
  createEffect(() => {
    if (!map) return;
    const style = mapStyle();
    if (style === 1) {
      if (map.hasLayer(standardLayer)) map.removeLayer(standardLayer);
      if (!map.hasLayer(hdLayer)) hdLayer.addTo(map);
    } else {
      if (map.hasLayer(hdLayer)) map.removeLayer(hdLayer);
      if (!map.hasLayer(standardLayer)) standardLayer.addTo(map);
    }
  });

  return (
    <div
      ref={mapDiv}
      style={{
        position: "absolute",
        top: 0,
        left: 0,
        width: "100%",
        height: "100%",
        "z-index": 0, // Behind WebGPU canvas
        background: "#090f1a",
      }}
    />
  );
}
