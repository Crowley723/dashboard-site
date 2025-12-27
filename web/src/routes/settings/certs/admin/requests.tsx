import { createFileRoute, Navigate } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState } from 'react';
import { useAuth } from '@/hooks/useAuth';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@/components/ui/table';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  useCertificateRequests,
  useReviewCertificateRequest,
} from '@/api/Certificates';
import type { CertificateRequestStatus } from '@/types/Certificates';
import { UserDisplay } from '@/components/UserDisplay.tsx';

export const Route = createFileRoute('/settings/certs/admin/requests')({
  component: RouteComponent,
  beforeLoad: async ({ location }) => {
    await requireAuth(
      location,
      'You must login to access the settings page.',
      true
    );
  },
});

function RouteComponent() {
  const { isMTLSAdmin, isLoading: authLoading } = useAuth();
  const {
    data: requests,
    isLoading,
    isError,
    error,
  } = useCertificateRequests();
  const reviewMutation = useReviewCertificateRequest();
  const [expandedRequest, setExpandedRequest] = useState<number | null>(null);
  const [reviewNotes, setReviewNotes] = useState<Record<number, string>>({});

  // Authorization check - only MTLS admins can access this page
  if (authLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading...</div>
      </div>
    );
  }

  if (!isMTLSAdmin()) {
    return <Navigate to="/settings/certs" />;
  }

  const getStatusBadge = (status: CertificateRequestStatus) => {
    const variants: Record<
      CertificateRequestStatus,
      'default' | 'outline' | 'secondary' | 'destructive'
    > = {
      awaiting_review: 'outline',
      approved: 'default',
      rejected: 'destructive',
      pending: 'outline',
      issued: 'default',
      failed: 'destructive',
      completed: 'secondary',
    };

    const labels: Record<CertificateRequestStatus, string> = {
      awaiting_review: 'Awaiting Review',
      approved: 'Approved',
      rejected: 'Rejected',
      pending: 'Pending',
      issued: 'Issued',
      failed: 'Failed',
      completed: 'Completed',
    };

    return <Badge variant={variants[status]}>{labels[status]}</Badge>;
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return `${date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })} at ${date.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
      timeZoneName: 'short',
    })}`;
  };

  const getRelativeTime = (dateString: string | null) => {
    if (!dateString) return 'Unknown';
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffHours = Math.floor(diffMs / (1000 * 60 * 60));
    const diffDays = Math.floor(diffHours / 24);

    if (diffDays > 0) return `${diffDays} day${diffDays > 1 ? 's' : ''} ago`;
    if (diffHours > 0)
      return `${diffHours} hour${diffHours > 1 ? 's' : ''} ago`;
    return 'Just now';
  };

  const handleReview = async (
    requestId: number,
    newStatus: 'approved' | 'rejected'
  ) => {
    const notes = reviewNotes[requestId] || '';
    try {
      await reviewMutation.mutateAsync({
        id: requestId,
        new_status: newStatus,
        review_notes: notes,
      });
      // Clear the notes after successful review
      setReviewNotes((prev) => {
        const updated = { ...prev };
        delete updated[requestId];
        return updated;
      });
    } catch (err) {
      // Error is handled by the mutation
      console.error('Failed to review request:', err);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading requests...</div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12 text-destructive">
          Error loading requests: {error?.message}
        </div>
      </div>
    );
  }

  const certificateRequests = Array.isArray(requests) ? requests : [];

  // Separate requests by status
  const awaitingReview = certificateRequests.filter(
    (r) => r.status === 'awaiting_review'
  );
  const otherRequests = certificateRequests.filter(
    (r) => r.status !== 'awaiting_review'
  );

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">
          Admin - Certificate Requests
        </h1>
        <p className="text-muted-foreground">
          Review and manage all certificate requests from users
        </p>
      </div>

      {awaitingReview.length > 0 && (
        <div className="mb-8">
          <h2 className="text-xl font-semibold mb-4">
            Awaiting Review ({awaitingReview.length})
          </h2>
          <Accordion
            type="single"
            collapsible
            className="space-y-4"
            value={expandedRequest?.toString()}
            onValueChange={(value) =>
              setExpandedRequest(value ? parseInt(value) : null)
            }
          >
            {awaitingReview.map((request) => {
              return (
                <AccordionItem
                  key={request.id}
                  value={request.id.toString()}
                  className="border border-border rounded-lg px-4 !border-b"
                >
                  <AccordionTrigger className="hover:no-underline">
                    <div className="flex items-center justify-between w-full pr-4">
                      <div className="flex items-center gap-4">
                        <div className="text-left">
                          <div className="font-medium">
                            <UserDisplay
                              displayName={request.owner_display_name}
                              username={request.owner_username}
                              sub={request.owner_sub}
                              iss={request.owner_iss}
                            />
                          </div>
                          <div className="text-sm text-muted-foreground">
                            Requested {getRelativeTime(request.requested_at)}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        {getStatusBadge(request.status)}
                      </div>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent>
                    <div className="pt-4 space-y-6">
                      <Table>
                        <TableBody>
                          <TableRow>
                            <TableHead className="w-1/3">Common Name</TableHead>
                            <TableCell>{request.common_name}</TableCell>
                          </TableRow>

                          {request.dns_names?.length > 0 && (
                            <TableRow>
                              <TableHead>DNS Names</TableHead>
                              <TableCell>
                                <div className="flex flex-wrap gap-2">
                                  {request.dns_names.map((dns, idx) => (
                                    <Badge key={idx} variant="outline">
                                      {dns}
                                    </Badge>
                                  ))}
                                </div>
                              </TableCell>
                            </TableRow>
                          )}

                          <TableRow>
                            <TableHead>Organizational Units</TableHead>
                            <TableCell>
                              {request.organizational_units &&
                              request.organizational_units.length > 0
                                ? request.organizational_units.join(', ')
                                : 'None'}
                            </TableCell>
                          </TableRow>

                          <TableRow>
                            <TableHead>Validity Period</TableHead>
                            <TableCell>{request.validity_days} days</TableCell>
                          </TableRow>

                          <TableRow>
                            <TableHead>Requested At</TableHead>
                            <TableCell>
                              {formatDate(request.requested_at)}
                            </TableCell>
                          </TableRow>

                          {request.message && (
                            <TableRow>
                              <TableHead>Request Message</TableHead>
                              <TableCell>{request.message}</TableCell>
                            </TableRow>
                          )}

                          <TableRow>
                            <TableHead>Owner</TableHead>
                            <TableCell>
                              <UserDisplay
                                displayName={request.owner_display_name}
                                username={request.owner_username}
                                sub={request.owner_sub}
                                iss={request.owner_iss}
                              />
                            </TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>

                      <div className="space-y-4">
                        <div>
                          <Label htmlFor={`notes-${request.id}`}>
                            Review Notes (optional)
                          </Label>
                          <Textarea
                            id={`notes-${request.id}`}
                            placeholder="Add any notes about this review..."
                            value={reviewNotes[request.id] || ''}
                            onChange={(e) =>
                              setReviewNotes((prev) => ({
                                ...prev,
                                [request.id]: e.target.value,
                              }))
                            }
                            rows={3}
                            className="mt-2"
                          />
                        </div>

                        <div className="flex gap-2">
                          <Button
                            variant="default"
                            onClick={() => handleReview(request.id, 'approved')}
                            disabled={reviewMutation.isPending}
                          >
                            {reviewMutation.isPending &&
                            reviewMutation.variables?.id === request.id
                              ? 'Approving...'
                              : 'Approve'}
                          </Button>
                          <Button
                            variant="destructive"
                            onClick={() => handleReview(request.id, 'rejected')}
                            disabled={reviewMutation.isPending}
                          >
                            {reviewMutation.isPending &&
                            reviewMutation.variables?.id === request.id
                              ? 'Rejecting...'
                              : 'Reject'}
                          </Button>
                        </div>

                        {reviewMutation.isError &&
                          reviewMutation.variables?.id === request.id && (
                            <div className="text-destructive text-sm">
                              Error: {reviewMutation.error?.message}
                            </div>
                          )}
                      </div>
                    </div>
                  </AccordionContent>
                </AccordionItem>
              );
            })}
          </Accordion>
        </div>
      )}

      {otherRequests.length > 0 && (
        <div>
          <h2 className="text-xl font-semibold mb-4">
            All Requests ({otherRequests.length})
          </h2>
          <Accordion type="single" collapsible className="space-y-4">
            {otherRequests.map((request) => {
              return (
                <AccordionItem
                  key={request.id}
                  value={request.id.toString()}
                  className="border border-border rounded-lg px-4 !border-b"
                >
                  <AccordionTrigger className="hover:no-underline">
                    <div className="flex items-center justify-between w-full pr-4">
                      <div className="flex items-center gap-4">
                        <div className="text-left">
                          <div className="font-medium">
                            <UserDisplay
                              displayName={request.owner_display_name}
                              username={request.owner_username}
                              sub={request.owner_sub}
                              iss={request.owner_iss}
                            />
                          </div>
                          <div className="text-sm text-muted-foreground">
                            Requested {getRelativeTime(request.requested_at)}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center gap-4">
                        {getStatusBadge(request.status)}
                      </div>
                    </div>
                  </AccordionTrigger>
                  <AccordionContent>
                    <div className="pt-4 space-y-6">
                      <Table>
                        <TableBody>
                          <TableRow>
                            <TableHead className="w-1/3">Common Name</TableHead>
                            <TableCell>{request.common_name}</TableCell>
                          </TableRow>

                          {request.dns_names?.length > 0 && (
                            <TableRow>
                              <TableHead>DNS Names</TableHead>
                              <TableCell>
                                <div className="flex flex-wrap gap-2">
                                  {request.dns_names.map((dns, idx) => (
                                    <Badge key={idx} variant="outline">
                                      {dns}
                                    </Badge>
                                  ))}
                                </div>
                              </TableCell>
                            </TableRow>
                          )}

                          <TableRow>
                            <TableHead>Organizational Units</TableHead>
                            <TableCell>
                              {request.organizational_units &&
                              request.organizational_units.length > 0
                                ? request.organizational_units.join(', ')
                                : 'None'}
                            </TableCell>
                          </TableRow>

                          {request.serial_number && (
                            <TableRow>
                              <TableHead>Serial Number</TableHead>
                              <TableCell className="font-mono text-sm">
                                {request.serial_number}
                              </TableCell>
                            </TableRow>
                          )}

                          <TableRow>
                            <TableHead>Validity Period</TableHead>
                            <TableCell>{request.validity_days} days</TableCell>
                          </TableRow>

                          <TableRow>
                            <TableHead>Requested At</TableHead>
                            <TableCell>
                              {formatDate(request.requested_at)}
                            </TableCell>
                          </TableRow>

                          {request.issued_at && (
                            <TableRow>
                              <TableHead>Issued At</TableHead>
                              <TableCell>
                                {formatDate(request.issued_at)}
                              </TableCell>
                            </TableRow>
                          )}

                          {request.expires_at && (
                            <TableRow>
                              <TableHead>Expires At</TableHead>
                              <TableCell>
                                {formatDate(request.expires_at)}
                              </TableCell>
                            </TableRow>
                          )}

                          {request.message && (
                            <TableRow>
                              <TableHead>Request Message</TableHead>
                              <TableCell>{request.message}</TableCell>
                            </TableRow>
                          )}

                          <TableRow>
                            <TableHead>Status</TableHead>
                            <TableCell>
                              {getStatusBadge(request.status)}
                            </TableCell>
                          </TableRow>

                          <TableRow>
                            <TableHead>Owner</TableHead>
                            <TableCell>
                              <UserDisplay
                                displayName={request.owner_display_name}
                                username={request.owner_username}
                                sub={request.owner_sub}
                                iss={request.owner_iss}
                              />
                            </TableCell>
                          </TableRow>
                        </TableBody>
                      </Table>

                      {request.events && request.events.length > 0 && (
                        <div>
                          <h3 className="font-semibold mb-2">Review History</h3>
                          <div className="space-y-2">
                            {request.events.map((event) => (
                              <div
                                key={event.id}
                                className="text-sm border-l-2 border-muted pl-3"
                              >
                                <div className="font-medium">
                                  Status changed to:{' '}
                                  {getStatusBadge(event.new_status)}
                                </div>
                                <div className="text-muted-foreground">
                                  Reviewed by:{' '}
                                  <UserDisplay
                                    displayName={event.reviewer_display_name}
                                    username={event.reviewer_username}
                                    sub={event.reviewer_sub}
                                    iss={event.reviewer_iss}
                                  />
                                </div>
                                {event.review_notes && (
                                  <div className="text-muted-foreground">
                                    Notes: {event.review_notes}
                                  </div>
                                )}
                                <div className="text-xs text-muted-foreground">
                                  {formatDate(event.created_at)}
                                </div>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </div>
                  </AccordionContent>
                </AccordionItem>
              );
            })}
          </Accordion>
        </div>
      )}

      {certificateRequests.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No certificate requests found
        </div>
      )}
    </div>
  );
}
