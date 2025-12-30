import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import type { CertificateRequest } from '@/types/Certificates.ts';

export const certificateKeys = {
  all: ['certificates'] as const,
  lists: () => [...certificateKeys.all, 'list'] as const,
  details: () => [...certificateKeys.all, 'detail'] as const,
  detail: (id: number) => [...certificateKeys.details(), id] as const,
  myRequests: () => [...certificateKeys.all, 'my-requests'] as const,
};

async function fetchAllCertificateRequests(): Promise<CertificateRequest[]> {
  const response = await fetch('/api/certificates/requests', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch certificates: ${response.statusText}`);
  }

  return response.json();
}

async function fetchMyCertificateRequests(): Promise<CertificateRequest[]> {
  const response = await fetch('/api/certificates/my-requests', {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch my certificates: ${response.statusText}`);
  }

  return response.json();
}

async function fetchCertificateRequest(
  id: number
): Promise<CertificateRequest> {
  const response = await fetch(`/api/certificates/request/${id}`, {
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error(`Failed to fetch certificate: ${response.statusText}`);
  }

  return response.json();
}

interface CreateCertificateRequestInput {
  message: string;
  validity_days?: number;
}

async function createCertificateRequest(
  input: CreateCertificateRequestInput
): Promise<CertificateRequest> {
  const response = await fetch('/api/certificates/request', {
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
    throw new Error(error.message || 'Failed to create certificate request');
  }

  return response.json();
}

interface ReviewCertificateInput {
  new_status: 'approved' | 'rejected';
  review_notes: string;
}

async function reviewCertificateRequest(
  id: number,
  input: ReviewCertificateInput
): Promise<CertificateRequest> {
  const response = await fetch(`/api/certificates/requests/${id}/review`, {
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
    throw new Error(error.message || 'Failed to review certificate request');
  }

  return response.json();
}

interface UnlockCertificateInput {
  passphrase: string;
}

interface UnlockCertificateResponse {
  unlocked: boolean;
  error?: string;
  download_token?: string;
  expires_in?: number;
}

async function unlockCertificate(
  id: number,
  input: UnlockCertificateInput
): Promise<UnlockCertificateResponse> {
  const response = await fetch(`/api/certificates/${id}/unlock`, {
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
    throw new Error(
      error.message || error.error || 'Failed to unlock certificate'
    );
  }

  return response.json();
}

async function downloadCertificate(id: number, token: string): Promise<Blob> {
  const response = await fetch(
    `/api/certificates/${id}/download?token=${encodeURIComponent(token)}`,
    {
      credentials: 'include',
    }
  );

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ message: response.statusText }));
    throw new Error(error.message || 'Failed to download certificate');
  }

  return response.blob();
}

export function useCertificateRequests() {
  return useQuery({
    queryKey: certificateKeys.lists(),
    queryFn: fetchAllCertificateRequests,
    staleTime: 1000 * 60 * 5, // 5 minutes
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  });
}

export function useMyCertificateRequests() {
  return useQuery({
    queryKey: certificateKeys.myRequests(),
    queryFn: fetchMyCertificateRequests,
    staleTime: 1000 * 60 * 5,
    refetchInterval: 30000, // Auto-refresh every 30 seconds
  });
}

export function useCertificateRequest(id: number) {
  return useQuery({
    queryKey: certificateKeys.detail(id),
    queryFn: () => fetchCertificateRequest(id),
    staleTime: 1000 * 60 * 5,
    enabled: !!id, // Only fetch if id is provided
  });
}

export function useCreateCertificateRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createCertificateRequest,
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: certificateKeys.myRequests() });
    },
  });
}

export function useReviewCertificateRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, ...input }: ReviewCertificateInput & { id: number }) =>
      reviewCertificateRequest(id, input),
    onSuccess: (data) => {
      queryClient.invalidateQueries({ queryKey: certificateKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: certificateKeys.detail(data.id),
      });
    },
  });
}

export function useUnlockCertificate() {
  return useMutation({
    mutationFn: ({ id, ...input }: UnlockCertificateInput & { id: number }) =>
      unlockCertificate(id, input),
  });
}

export function useDownloadCertificate() {
  return useMutation({
    mutationFn: ({ id, token }: { id: number; token: string }) =>
      downloadCertificate(id, token),
  });
}
