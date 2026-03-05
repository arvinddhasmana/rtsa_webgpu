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
  const [showToast, setShowToast] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  // Close on Escape
  useEffect(() => {
    const handleKey = (e: KeyboardEvent) => {
      // Don't close if showing toast
      if (e.key === "Escape" && !showToast) onClose();
    };
    window.addEventListener("keydown", handleKey);
    return () => window.removeEventListener("keydown", handleKey);
  }, [onClose, showToast]);

  // Close on outside click
  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node) && !showToast) {
        onClose();
      }
    };
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [onClose, showToast]);

  const handleAssign = () => {
    if (!selectedOp) return;

    // Show toast, then actually assign and close
    setShowToast(true);

    setTimeout(() => {
      onAssign(selectedOp);
      onClose();
    }, 1500);
  };

  const opInfo = AVAILABLE_OPERATORS.find(o => o.id === selectedOp);

  return (
    <div
      ref={ref}
      data-testid="alert-assign-popover"
      style={{
        position: "absolute",
        right: "16px",
        top: "50%",
        transform: "translateY(-50%)",
        width: "280px",
        backgroundColor: "rgba(30, 41, 59, 0.95)",
        backdropFilter: "blur(12px)",
        border: "1px solid rgba(255,255,255,0.1)",
        borderRadius: "8px",
        padding: "16px",
        zIndex: 50,
        boxShadow: "0 8px 32px rgba(0,0,0,0.6)",
      }}
    >
      {showToast ? (
         <div style={{
           display: "flex",
           flexDirection: "column",
           alignItems: "center",
           justifyContent: "center",
           height: "200px",
           gap: "12px",
           animation: "fadeIn 0.3s ease-in"
         }}>
           <div style={{ fontSize: "2rem" }}>✅</div>
           <div style={{ color: "#F1F5F9", fontWeight: "bold", fontSize: "0.85rem" }}>
             Alert Assigned
           </div>
           <div style={{ color: "#94A3B8", fontSize: "0.75rem", textAlign: "center" }}>
             Alert {alertId.slice(0, 8)}... handed off to {opInfo?.name}
           </div>
         </div>
      ) : (
        <>
          <div
            style={{
              display: "flex",
              justifyContent: "space-between",
              alignItems: "center",
              marginBottom: "16px",
            }}
          >
            <span
              style={{
                fontSize: "0.8rem",
                fontWeight: "bold",
                color: "#EAB308",
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
                color: "#94A3B8",
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
              color: "#94A3B8",
              marginBottom: "12px",
              fontFamily: "monospace",
            }}
          >
            Alert ID: {alertId}
          </div>

          {/* Operator list */}
          <div
            style={{ display: "flex", flexDirection: "column", gap: "6px", marginBottom: "16px" }}
          >
            {AVAILABLE_OPERATORS.map((op) => (
              <label
                key={op.id}
                data-testid={`assign-op-${op.id}`}
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: "12px",
                  padding: "8px 10px",
                  borderRadius: "6px",
                  cursor: "pointer",
                  backgroundColor:
                    selectedOp === op.id
                      ? "rgba(59, 130, 246, 0.2)"
                      : "rgba(255,255,255,0.03)",
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
                  style={{ accentColor: "#3B82F6", margin: 0 }}
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
              height: "60px",
              backgroundColor: "rgba(15, 23, 42, 0.6)",
              color: "#CBD5E1",
              border: "1px solid rgba(255,255,255,0.1)",
              borderRadius: "4px",
              padding: "8px",
              fontSize: "0.7rem",
              resize: "none",
              marginBottom: "16px",
              boxSizing: "border-box",
              outline: "none",
            }}
          />

          <button
            data-testid="assign-confirm-btn"
            onClick={handleAssign}
            disabled={!selectedOp}
            style={{
              width: "100%",
              padding: "10px",
              backgroundColor: selectedOp ? "#2563EB" : "rgba(255,255,255,0.05)",
              color: selectedOp ? "#F1F5F9" : "#64748B",
              border: "none",
              borderRadius: "4px",
              cursor: selectedOp ? "pointer" : "not-allowed",
              fontSize: "0.75rem",
              fontWeight: "bold",
              transition: "background-color 0.15s ease",
            }}
          >
            Confirm Assignment →
          </button>
        </>
      )}
    </div>
  );
};
