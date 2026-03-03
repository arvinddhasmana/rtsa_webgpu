// CLASSIFICATION: UNCLASSIFIED
// src/components/alert/AlertAssignPopover.tsx
// Floating popover to assign an alert to an operator.

import React, { useEffect, useRef, useState } from "react";

interface AlertAssignPopoverProps {
  alertId: string;
  onAssign: (operatorId: string) => void;
  onClose: () => void;
}

const AVAILABLE_OPERATORS = [
  { id: "op-charlie-1", name: "Charlie-01", role: "Operator" },
  { id: "op-delta-2", name: "Delta-02", role: "Operator" },
  { id: "op-echo-3", name: "Echo-03", role: "Watch Officer" },
  { id: "op-foxtrot-4", name: "Foxtrot-04", role: "Supervisor" },
];

/**
 * AlertAssignPopover — floating panel to assign an alert to another operator.
 * Dismissed by clicking outside or pressing Escape.
 */
export const AlertAssignPopover: React.FC<AlertAssignPopoverProps> = ({
  alertId,
  onAssign,
  onClose,
}) => {
  const [selectedOp, setSelectedOp] = useState<string>("");
  const [note, setNote] = useState<string>("");
  const ref = useRef<HTMLDivElement>(null);

  // Close on Escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [onClose]);

  // Close on outside click
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [onClose]);

  const handleAssign = () => {
    if (!selectedOp) return;
    onAssign(selectedOp);
  };

  return (
    <div
      ref={ref}
      data-testid="alert-assign-popover"
      style={{
        position: "absolute",
        right: "16px",
        top: "50%",
        transform: "translateY(-50%)",
        width: "260px",
        backgroundColor: "#1E293B",
        border: "1px solid #334155",
        borderRadius: "8px",
        padding: "16px",
        zIndex: 50,
        boxShadow: "0 8px 32px rgba(0,0,0,0.6)",
      }}
    >
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: "12px",
        }}
      >
        <span
          style={{
            fontSize: "0.8rem",
            fontWeight: "bold",
            color: "#F59E0B",
            letterSpacing: "0.05em",
          }}
        >
          ASSIGN ALERT
        </span>
        <button
          onClick={onClose}
          style={{
            background: "none",
            border: "none",
            color: "#9CA3AF",
            cursor: "pointer",
            fontSize: "1rem",
          }}
        >
          ✕
        </button>
      </div>

      <div
        style={{
          fontSize: "0.65rem",
          color: "#64748B",
          marginBottom: "12px",
          fontFamily: "monospace",
        }}
      >
        Alert: {alertId}
      </div>

      {/* Operator list */}
      <div
        style={{ display: "flex", flexDirection: "column", gap: "6px", marginBottom: "12px" }}
      >
        {AVAILABLE_OPERATORS.map((op) => (
          <label
            key={op.id}
            data-testid={`assign-op-${op.id}`}
            style={{
              display: "flex",
              alignItems: "center",
              gap: "8px",
              padding: "6px 8px",
              borderRadius: "4px",
              cursor: "pointer",
              backgroundColor:
                selectedOp === op.id
                  ? "rgba(59, 130, 246, 0.2)"
                  : "rgba(255,255,255,0.04)",
              border:
                selectedOp === op.id
                  ? "1px solid #3B82F6"
                  : "1px solid transparent",
              transition: "all 0.15s ease",
            }}
          >
            <input
              type="radio"
              name="operator"
              value={op.id}
              checked={selectedOp === op.id}
              onChange={() => setSelectedOp(op.id)}
              style={{ accentColor: "#3B82F6" }}
            />
            <div>
              <div style={{ fontSize: "0.75rem", color: "#F1F5F9", fontWeight: "bold" }}>
                {op.name}
              </div>
              <div style={{ fontSize: "0.65rem", color: "#64748B" }}>{op.role}</div>
            </div>
          </label>
        ))}
      </div>

      {/* Note */}
      <textarea
        placeholder="Handoff note (optional)..."
        value={note}
        onChange={(e) => setNote(e.target.value)}
        style={{
          width: "100%",
          height: "52px",
          backgroundColor: "#0F172A",
          color: "#CBD5E1",
          border: "1px solid #334155",
          borderRadius: "4px",
          padding: "6px 8px",
          fontSize: "0.7rem",
          resize: "none",
          marginBottom: "12px",
          boxSizing: "border-box",
        }}
      />

      <button
        data-testid="assign-confirm-btn"
        onClick={handleAssign}
        disabled={!selectedOp}
        style={{
          width: "100%",
          padding: "8px",
          backgroundColor: selectedOp ? "#1D4ED8" : "#374151",
          color: selectedOp ? "#F1F5F9" : "#6B7280",
          border: "none",
          borderRadius: "4px",
          cursor: selectedOp ? "pointer" : "not-allowed",
          fontSize: "0.75rem",
          fontWeight: "bold",
          transition: "background-color 0.15s ease",
        }}
      >
        Assign →
      </button>
    </div>
  );
};
