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

// Get all certificate requests (admin only)
export function useCertificateRequests() {
  return useQuery({
    queryKey: certificateKeys.lists(),
    queryFn: fetchAllCertificateRequests,
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}

// Get my certificate requests
export function useMyCertificateRequests() {
  return useQuery({
    queryKey: certificateKeys.myRequests(),
    queryFn: fetchMyCertificateRequests,
    staleTime: 1000 * 60 * 5,
  });
}

// Get single certificate request
export function useCertificateRequest(id: number) {
  return useQuery({
    queryKey: certificateKeys.detail(id),
    queryFn: () => fetchCertificateRequest(id),
    staleTime: 1000 * 60 * 5,
    enabled: !!id, // Only fetch if id is provided
  });
}

// Create certificate request
export function useCreateCertificateRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: createCertificateRequest,
    onSuccess: () => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: certificateKeys.myRequests() });
    },
  });
}

// Review certificate request (admin only)
export function useReviewCertificateRequest() {
  const queryClient = useQueryClient();

  return useMutation({
    mutationFn: ({ id, ...input }: ReviewCertificateInput & { id: number }) =>
      reviewCertificateRequest(id, input),
    onSuccess: (data) => {
      // Invalidate and refetch
      queryClient.invalidateQueries({ queryKey: certificateKeys.lists() });
      queryClient.invalidateQueries({
        queryKey: certificateKeys.detail(data.id),
      });
    },
  });
}
