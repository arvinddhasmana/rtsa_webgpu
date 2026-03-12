// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/DashboardSidebar.tsx — Sensor health filters
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { For } from "solid-js";
import { SensorStatus } from "../../services/sensor-health";
import {
    selectedStatuses,
    selectedTypes,
    setSidebarCollapsed,
    sidebarCollapsed,
    toggleStatusFilter,
    toggleTypeFilter
} from "../../signals/sensor-filters";

interface DashboardSidebarProps {
    sensors: SensorStatus[];
}

/**
 * Collapsable sidebar with sensor health metrics and type filters.
 */
export function DashboardSidebar(props: DashboardSidebarProps) {
    const getCount = (status: string) => props.sensors.filter(s => s.status === status).length;

    const statuses = [
        { id: "CONNECTED", label: "Online", color: "#4ade80" },
        { id: "STALE", label: "Stale", color: "#fbbf24" },
        { id: "OFFLINE", label: "Offline", color: "#f87171" },
    ];

    const types = ["RADAR", "EW/SIGINT", "ELINT/COMINT", "ISR", "AIS/BFT", "CYBER"];

    return (
        <div
            style={{
                width: sidebarCollapsed() ? "60px" : "260px",
                height: "100%",
                background: "rgba(13, 20, 36, 0.95)",
                "border-right": "1px solid rgba(255, 255, 255, 0.1)",
                transition: "width 0.3s cubic-bezier(0.4, 0, 0.2, 1)",
                display: "flex",
                "flex-direction": "column",
                overflow: "hidden",
                color: "#e2e8f0",
                "backdrop-filter": "blur(10px)",
                "z-index": 10,
            }}
        >
            {/* Collapse Toggle */}
            <div style={{ padding: "12px", display: "flex", "justify-content": sidebarCollapsed() ? "center" : "flex-end" }}>
                <button
                    onClick={() => setSidebarCollapsed(!sidebarCollapsed())}
                    style={{
                        background: "rgba(255,255,255,0.05)",
                        border: "1px solid rgba(255,255,255,0.1)",
                        color: "#94a3b8",
                        cursor: "pointer",
                        "border-radius": "4px",
                        padding: "4px 8px",
                        display: "flex",
                        "align-items": "center",
                        transition: "all 0.2s"
                    }}
                    class="sidebar-toggle-btn"
                >
                    {sidebarCollapsed() ? "»" : "«"}
                </button>
            </div>

            <div style={{
                padding: "20px",
                display: sidebarCollapsed() ? "none" : "flex",
                "flex-direction": "column",
                gap: "24px",
                opacity: sidebarCollapsed() ? 0 : 1,
                transition: "opacity 0.2s"
            }}>
                {/* Global Metrics Section */}
                <div>
                    <h3 style={{ "font-size": "0.75rem", "text-transform": "uppercase", "letter-spacing": "0.1em", color: "#64748b", "margin-bottom": "12px" }}>System Health</h3>
                    <div style={{ display: "grid", "grid-template-columns": "1fr 1fr", gap: "10px" }}>
                        <For each={statuses}>
                            {(item) => (
                                <div
                                    onClick={() => toggleStatusFilter(item.id)}
                                    style={{
                                        background: selectedStatuses().includes(item.id) ? "rgba(255,255,255,0.08)" : "transparent",
                                        border: `1px solid ${selectedStatuses().includes(item.id) ? item.color : "rgba(255,255,255,0.05)"}`,
                                        padding: "10px",
                                        "border-radius": "8px",
                                        cursor: "pointer",
                                        transition: "all 0.2s",
                                        display: "flex",
                                        "flex-direction": "column",
                                        gap: "4px"
                                    }}
                                    class="filter-card"
                                >
                                    <div style={{ "font-size": "0.65rem", color: "#94a3b8" }}>{item.label}</div>
                                    <div style={{ "font-size": "1.2rem", "font-weight": "bold", color: item.color }}>{getCount(item.id)}</div>
                                </div>
                            )}
                        </For>
                        <div style={{
                            background: "rgba(255,255,255,0.03)",
                            border: "1px solid rgba(255,255,255,0.05)",
                            padding: "10px",
                            "border-radius": "8px",
                            display: "flex",
                            "flex-direction": "column",
                            gap: "4px"
                        }}>
                             <div style={{ "font-size": "0.65rem", color: "#94a3b8" }}>Total</div>
                             <div style={{ "font-size": "1.2rem", "font-weight": "bold" }}>{props.sensors.length}</div>
                        </div>
                    </div>
                </div>

                {/* Sensor Types Section */}
                <div style={{ "border-top": "1px solid rgba(255,255,255,0.05)", "padding-top": "24px" }}>
                    <h3 style={{ "font-size": "0.75rem", "text-transform": "uppercase", "letter-spacing": "0.1em", color: "#64748b", "margin-bottom": "16px" }}>Sensor Types</h3>
                    <div style={{ display: "flex", "flex-direction": "column", gap: "12px" }}>
                        <For each={types}>
                            {(type) => (
                                <label style={{ display: "flex", "align-items": "center", gap: "10px", cursor: "pointer", "font-size": "0.85rem" }}>
                                    <input
                                        type="checkbox"
                                        checked={selectedTypes().includes(type)}
                                        onChange={() => toggleTypeFilter(type)}
                                        style={{ "accent-color": "#3b82f6", width: "16px", height: "16px" }}
                                    />
                                    <span style={{ color: selectedTypes().includes(type) ? "#f1f5f9" : "#64748b", transition: "color 0.2s" }}>
                                        {type}
                                    </span>
                                </label>
                            )}
                        </For>
                    </div>
                </div>
            </div>

            <style>{`
                .sidebar-toggle-btn:hover {
                    background: rgba(255,255,255,0.1);
                    color: white;
                }
                .filter-card:hover {
                    background: rgba(255,255,255,0.12);
                }
            `}</style>
        </div>
    );
}
