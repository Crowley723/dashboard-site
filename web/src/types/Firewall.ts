export interface FirewallAlias {
  name: string;
  uuid: string;
  description: string;
  max_ips_per_user: number;
  max_total_ips: number;
  default_ttl: string | null; // Duration string like "720h" or null for no expiration
  auth_group: string;
}

export interface FirewallIPWhitelistEntry {
  id: number;
  owner_iss: string;
  owner_sub: string;
  owner_username: string;
  owner_display_name: string;
  alias_name: string;
  alias_uuid: string;
  ip_address: string;
  ip_version: number; // 4 or 6
  description: string | null;
  status: FirewallIPStatus;
  requested_at: string;
  added_at: string | null;
  removed_at: string | null;
  expires_at: string | null;
  removed_by_iss: string | null;
  removed_by_sub: string | null;
  removed_by_username: string | null;
  removed_by_display_name: string | null;
  removal_reason: string | null;
  events: FirewallIPWhitelistEvent[];
}

export interface FirewallIPWhitelistEvent {
  id: number;
  whitelist_id: number;
  actor_iss: string;
  actor_sub: string;
  actor_username: string;
  actor_display_name: string;
  event_type: FirewallEventType;
  notes: string | null;
  client_ip: string | null;
  user_agent: string | null;
  created_at: string;
}

export type FirewallIPStatus =
  | 'requested'
  | 'added'
  | 'removed'
  | 'removed_by_admin'
  | 'blacklisted_by_admin';

export type FirewallEventType =
  | 'requested'
  | 'added'
  | 'removed'
  | 'removed_by_admin'
  | 'blacklisted_by_admin'
  | 'expired'
  | 'sync_failed';

export interface AddIPWhitelistRequest {
  alias_name: string;
  ip_address: string;
  description?: string;
  ttl?: string; // Duration string like "24h", "7d", etc.
}

export interface AddIPWhitelistResponse {
  entry: FirewallIPWhitelistEntry;
  message: string;
}
