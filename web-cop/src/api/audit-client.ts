// CLASSIFICATION: UNCLASSIFIED
// src/api/audit-client.ts

/**
 * AuditClient wraps the AuditService gRPC-Web endpoint.
 * Audit events are emitted server-side; this client is for querying audit logs.
 */
export class AuditClient {
  async queryAuditLog(_req: {
    operatorId?: string;
    startTime?: Date;
    endTime?: Date;
  }): Promise<{ events: Array<{ eventId: string; description: string; timestamp: Date }> }> {
    return { events: [] };
  }
}

export const auditClient = new AuditClient();
