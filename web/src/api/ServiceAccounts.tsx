import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type {
  ServiceAccount,
  CreateServiceAccountInput,
  UserScopes,
} from '@/types/ServiceAccounts';

export const serviceAccountKeys = {
  all: ['service-accounts'] as const,
  lists: () => [...serviceAccountKeys.all, 'list'] as const,
  scopes: () => [...serviceAccountKeys.all, 'scopes'] as const,
};

async function fetchServiceAccounts(): Promise<ServiceAccount[]> {
  const response = await fetch('/api/service-accounts', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(
      `Failed to fetch service accounts: ${response.statusText}`
    );
  }

  return response.json();
}

async function fetchUserScopes(): Promise<UserScopes> {
  const response = await fetch('/api/service-accounts/scopes', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch user scopes: ${response.statusText}`);
  }

  return response.json();
}

async function createServiceAccount(
  input: CreateServiceAccountInput
): Promise<ServiceAccount> {
  const response = await fetch('/api/service-accounts', {
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
      .catch(() => ({ message: response.statusText }));
    throw new Error(error.message || 'Failed to create service account');
  }

  return response.json();
}

async function deleteServiceAccount(
  iss: string,
  sub: string
): Promise<{ message: string }> {
  const response = await fetch(
    `/api/service-accounts?iss=${encodeURIComponent(iss)}&sub=${encodeURIComponent(sub)}`,
    {
      method: 'DELETE',
      credentials: 'include',
    }
  );

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ message: response.statusText }));
    throw new Error(error.message || 'Failed to delete service account');
  }

  return response.json();
}

async function pauseServiceAccount(
  iss: string,
  sub: string
): Promise<{ message: string }> {
  const response = await fetch(
    `/api/service-accounts/pause?iss=${encodeURIComponent(iss)}&sub=${encodeURIComponent(sub)}`,
    {
      method: 'PATCH',
      credentials: 'include',
    }
  );

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ message: response.statusText }));
    throw new Error(error.message || 'Failed to pause service account');
  }

  return response.json();
}

async function unpauseServiceAccount(
  iss: string,
  sub: string
): Promise<{ message: string }> {
  const response = await fetch(
    `/api/service-accounts/unpause?iss=${encodeURIComponent(iss)}&sub=${encodeURIComponent(sub)}`,
    {
      method: 'PATCH',
      credentials: 'include',
    }
  );

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ message: response.statusText }));
    throw new Error(error.message || 'Failed to unpause service account');
  }

  return response.json();
}

// React Query hooks
export function useServiceAccounts() {
  return useQuery({
    queryKey: serviceAccountKeys.lists(),
    queryFn: fetchServiceAccounts,
  });
}

export function useUserScopes() {
  return useQuery({
    queryKey: serviceAccountKeys.scopes(),
    queryFn: fetchUserScopes,
  });
}

export function useCreateServiceAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createServiceAccount,
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: serviceAccountKeys.lists(),
      });
    },
  });
}

export function useDeleteServiceAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ iss, sub }: { iss: string; sub: string }) =>
      deleteServiceAccount(iss, sub),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: serviceAccountKeys.lists(),
      });
    },
  });
}

export function usePauseServiceAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ iss, sub }: { iss: string; sub: string }) =>
      pauseServiceAccount(iss, sub),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: serviceAccountKeys.lists(),
      });
    },
  });
}

export function useUnpauseServiceAccount() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ iss, sub }: { iss: string; sub: string }) =>
      unpauseServiceAccount(iss, sub),
    onSuccess: () => {
      queryClient.invalidateQueries({
        queryKey: serviceAccountKeys.lists(),
      });
    },
  });
}
