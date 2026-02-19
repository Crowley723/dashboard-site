import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type {
  FirewallAlias,
  FirewallIPWhitelistEntry,
  AddIPWhitelistRequest,
  AddIPWhitelistResponse,
} from '@/types/Firewall.ts';

export const firewallKeys = {
  all: ['firewall'] as const,
  aliases: () => [...firewallKeys.all, 'aliases'] as const,
  entries: () => [...firewallKeys.all, 'entries'] as const,
  entry: (id: number) => [...firewallKeys.entries(), id] as const,
};

async function fetchAvailableAliases(): Promise<FirewallAlias[]> {
  const response = await fetch('/api/firewall/aliases', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch aliases: ${response.statusText}`);
  }

  return response.json();
}

async function fetchUserEntries(): Promise<FirewallIPWhitelistEntry[]> {
  const response = await fetch('/api/firewall/entries', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch entries: ${response.statusText}`);
  }

  return response.json();
}

async function addIPWhitelistEntry(
  input: AddIPWhitelistRequest
): Promise<AddIPWhitelistResponse> {
  const response = await fetch('/api/firewall/entries', {
    method: 'POST',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(input),
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: response.statusText }));
    throw new Error(error.error || 'Failed to add IP to whitelist');
  }

  return response.json();
}

async function removeIPWhitelistEntry(id: number): Promise<void> {
  const response = await fetch(`/api/firewall/entries/${id}`, {
    method: 'DELETE',
    credentials: 'include',
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: response.statusText }));
    throw new Error(error.error || 'Failed to remove IP from whitelist');
  }
}

async function blacklistIPEntry(
  id: number,
  reason?: string
): Promise<void> {
  const response = await fetch(`/api/firewall/entries/${id}/blacklist`, {
    method: 'DELETE',
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ reason }),
  });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: response.statusText }));
    throw new Error(error.error || 'Failed to blacklist IP');
  }
}

async function fetchAllEntries(): Promise<FirewallIPWhitelistEntry[]> {
  const response = await fetch('/api/firewall/entries?all_users=1', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch all entries: ${response.statusText}`);
  }

  return response.json();
}

export function useAvailableAliases() {
  return useQuery({
    queryKey: firewallKeys.aliases(),
    queryFn: fetchAvailableAliases,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

export function useUserEntries() {
  return useQuery({
    queryKey: firewallKeys.entries(),
    queryFn: fetchUserEntries,
    staleTime: 1000 * 60, // 1 minute
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  });
}

export function useAllFirewallEntries() {
  return useQuery({
    queryKey: [...firewallKeys.entries(), 'all'],
    queryFn: fetchAllEntries,
    staleTime: 1000 * 60, // 1 minute
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  });
}

export function useAddIPWhitelistEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: addIPWhitelistEntry,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: firewallKeys.entries() });
    },
  });
}

export function useRemoveIPWhitelistEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: removeIPWhitelistEntry,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: firewallKeys.entries() });
    },
  });
}

export function useBlacklistIPEntry() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, reason }: { id: number; reason?: string }) =>
      blacklistIPEntry(id, reason),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: firewallKeys.entries() });
    },
  });
}
