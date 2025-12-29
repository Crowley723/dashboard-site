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
import { useMyCertificateRequests } from '@/api/Certificates';
import type { CertificateRequestStatus } from '@/types/Certificates';
import { UserDisplay } from '@/components/UserDisplay.tsx';
import { RequestCertificateDialog } from '@/components/RequestCertificateDialog.tsx';
import { DownloadCertificateDialog } from '@/components/DownloadCertificateDialog.tsx';
import { getRelativeTimeString } from '@/hooks/RelativeTimeString.tsx';
import { RefreshCw, Check } from 'lucide-react';

export const Route = createFileRoute('/settings/certs/requests')({
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
  const { isMTLSUser, isMTLSAdmin, isLoading: authLoading } = useAuth();
  const {
    data: requests,
    isLoading,
    isError,
    error,
    refetch,
  } = useMyCertificateRequests();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [downloadDialogOpen, setDownloadDialogOpen] = useState(false);
  const [selectedCertificateId, setSelectedCertificateId] = useState<
    number | null
  >(null);
  const [lastManualRefresh, setLastManualRefresh] = useState<number>(0);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);

  const handleManualRefresh = async () => {
    const now = Date.now();
    const timeSinceLastRefresh = now - lastManualRefresh;

    setIsRefreshing(true);
    setShowSuccess(false);

    if (timeSinceLastRefresh < 10000) {
      setTimeout(() => {
        setIsRefreshing(false);
      }, 1500);
      return;
    }

    setLastManualRefresh(now);
    await refetch();

    setTimeout(() => {
      setIsRefreshing(false);
      setShowSuccess(true);
      setTimeout(() => {
        setShowSuccess(false);
      }, 2000);
    }, 1500);
  };

  if (authLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading...</div>
      </div>
    );
  }

  if (!isMTLSUser() && !isMTLSAdmin()) {
    return <Navigate to="/settings" />;
  }

  const getStatusBadge = (status: CertificateRequestStatus) => {
    const variants: Record<
      CertificateRequestStatus,
      | 'default'
      | 'outline'
      | 'secondary'
      | 'destructive'
      | 'warning'
      | 'success'
      | 'info'
    > = {
      awaiting_review: 'warning', // Yellow
      approved: 'info', // Blue
      rejected: 'destructive', // Red
      pending: 'secondary', // Gray
      issued: 'success', // Green
      failed: 'destructive', // Red
      completed: 'secondary', // Gray
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

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">Certificate Requests</h1>
          <p className="text-muted-foreground">
            View your certificate requests and submit new ones
          </p>
        </div>
        <div className="flex gap-2">
          <Button
            variant="outline"
            size="icon"
            onClick={handleManualRefresh}
            disabled={isRefreshing}
            className={showSuccess ? 'text-green-600' : ''}
          >
            {showSuccess ? (
              <Check />
            ) : (
              <RefreshCw className={isRefreshing ? 'animate-spin' : ''} />
            )}
          </Button>
          <RequestCertificateDialog
            open={dialogOpen}
            onOpenChange={setDialogOpen}
          >
            <Button>Request Certificate</Button>
          </RequestCertificateDialog>
        </div>
      </div>

      <Accordion type="single" collapsible className="space-y-4">
        {certificateRequests.map((request) => {
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
                        {request.requested_at === undefined
                          ? ''
                          : ` Requested ${getRelativeTimeString(new Date(request.requested_at))}`}
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
                          <TableCell>{formatDate(request.issued_at)}</TableCell>
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
                        <TableCell>{getStatusBadge(request.status)}</TableCell>
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

                  <div className="flex gap-2 pt-2">
                    {request.status === 'issued' && (
                      <Button
                        onClick={() => {
                          setSelectedCertificateId(request.id);
                          setDownloadDialogOpen(true);
                        }}
                      >
                        Download Certificate
                      </Button>
                    )}
                    {request.status === 'awaiting_review' && (
                      <Button variant="outline" disabled>
                        Awaiting Admin Review
                      </Button>
                    )}
                    {request.status === 'rejected' && (
                      <Button variant="destructive" disabled>
                        Request Rejected
                      </Button>
                    )}
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>

      {certificateRequests.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          <p className="mb-4">No certificate requests found</p>
          <Button onClick={() => setDialogOpen(true)}>
            Request Your First Certificate
          </Button>
        </div>
      )}

      {selectedCertificateId && (
        <DownloadCertificateDialog
          certificateId={selectedCertificateId}
          open={downloadDialogOpen}
          onOpenChange={setDownloadDialogOpen}
        />
      )}
    </div>
  );
}
