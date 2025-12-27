export interface CertificateRequest {
  id: number;
  owner_iss: string;
  owner_sub: string;
  owner_username: string;
  owner_display_name: string;
  message: string;
  events: CertificateEvent[];
  common_name: string;
  dns_names: string[];
  organizational_units: string[];
  validity_days: number;
  status: CertificateRequestStatus;
  requested_at: string;
  issued_at: string | null;
  expires_at: string | null;
  serial_number: string | null;
}

export interface CertificateEvent {
  id: number;
  certificate_request_id: number;
  requester_iss: string;
  requester_sub: string;
  requester_username: string;
  requester_display_name: string;
  reviewer_iss: string;
  reviewer_sub: string;
  reviewer_username: string;
  reviewer_display_name: string;
  new_status: CertificateRequestStatus;
  review_notes: string;
  created_at: string;
}

export type CertificateRequestStatus =
  | 'awaiting_review'
  | 'approved'
  | 'rejected'
  | 'pending'
  | 'issued'
  | 'failed'
  | 'completed';

export function parseCertificateRequest(
  raw: CertificateRequest
): CertificateRequestWithDates {
  return {
    ...raw,
    requested_at: new Date(raw.requested_at),
    issued_at: raw.issued_at ? new Date(raw.issued_at) : null,
    expires_at: raw.expires_at ? new Date(raw.expires_at) : null,
    events: raw.events.map((e) => ({
      ...e,
      created_at: new Date(e.created_at),
    })),
  };
}

// Typed version with Date objects
export interface CertificateRequestWithDates
  extends Omit<
    CertificateRequest,
    'requested_at' | 'issued_at' | 'expires_at' | 'events'
  > {
  requested_at: Date;
  issued_at: Date | null;
  expires_at: Date | null;
  events: CertificateEventWithDates[];
}

export interface CertificateEventWithDates
  extends Omit<CertificateEvent, 'created_at'> {
  created_at: Date;
}
