import { createFileRoute, useNavigate, Navigate } from '@tanstack/react-router';
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
import { Input } from '@/components/ui/input';
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
import { DownloadCertificateDialog } from '@/components/DownloadCertificateDialog.tsx';

export const Route = createFileRoute('/settings/certs/')({
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
  const navigate = useNavigate();
  const { isMTLSUser, isMTLSAdmin, isLoading: authLoading } = useAuth();
  const {
    data: certificates,
    isLoading,
    isError,
    error,
  } = useMyCertificateRequests();
  const [searchQuery, setSearchQuery] = useState('');
  const [downloadDialogOpen, setDownloadDialogOpen] = useState(false);
  const [selectedCertificateId, setSelectedCertificateId] = useState<
    number | null
  >(null);

  // Authorization check - only MTLS users can access this page
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

  const getDaysUntilExpiry = (expiresAt: string | null) => {
    if (!expiresAt) return null;
    const expiry = new Date(expiresAt);
    const now = new Date();
    const diff = expiry.getTime() - now.getTime();
    const days = Math.ceil(diff / (1000 * 60 * 60 * 24));
    return days;
  };

  console.log('certificates:', certificates);
  console.log('type:', typeof certificates);
  console.log('isArray:', Array.isArray(certificates));

  const filteredCertificates = Array.isArray(certificates)
    ? certificates.filter((cert) => {
        const query = searchQuery.toLowerCase();
        return (
          cert.common_name.toLowerCase().includes(query) ||
          cert.dns_names?.some((dns) => dns.toLowerCase().includes(query)) ||
          cert.organizational_units?.some((ou) =>
            ou.toLowerCase().includes(query)
          ) ||
          cert.serial_number?.toLowerCase().includes(query) ||
          cert.status.toLowerCase().includes(query) ||
          cert.message?.toLowerCase().includes(query)
        );
      })
    : [];

  if (isLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading certificates...</div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12 text-destructive">
          Error loading certificates: {error?.message}
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">mTLS Certificates</h1>
        <p className="text-muted-foreground">
          Manage your mutual TLS certificates for secure authentication
        </p>
      </div>

      <div className="mb-4">
        <Input
          placeholder="Search certificates by name, DNS, organizational units, serial..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-xl"
        />
      </div>

      <Accordion type="single" collapsible className="space-y-4">
        {filteredCertificates.map((cert) => {
          const daysUntilExpiry = getDaysUntilExpiry(cert.expires_at);
          const isExpiringSoon =
            daysUntilExpiry !== null &&
            daysUntilExpiry <= 30 &&
            daysUntilExpiry > 0;

          return (
            <AccordionItem
              key={cert.id}
              value={cert.id.toString()}
              className="border border-border rounded-lg px-4 !border-b"
            >
              <AccordionTrigger className="hover:no-underline">
                <div className="flex items-center justify-between w-full pr-4">
                  <div className="flex items-center gap-4">
                    <div className="text-left">
                      <div className="font-medium">
                        <UserDisplay
                          displayName={cert.owner_display_name}
                          username={cert.owner_username}
                          sub={cert.owner_sub}
                          iss={cert.owner_iss}
                        />
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {cert.serial_number
                          ? `Serial: ${cert.serial_number}`
                          : ''}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    {isExpiringSoon && (
                      <span className="text-sm text-orange-500">
                        Expires in {daysUntilExpiry} days
                      </span>
                    )}
                    {getStatusBadge(cert.status)}
                  </div>
                </div>
              </AccordionTrigger>
              <AccordionContent>
                <div className="pt-4 space-y-6">
                  <Table>
                    <TableBody>
                      <TableRow>
                        <TableHead className="w-1/3">Common Name</TableHead>
                        <TableCell>{cert.common_name}</TableCell>
                      </TableRow>

                      {cert.dns_names?.length > 0 && (
                        <TableRow>
                          <TableHead>DNS Names</TableHead>
                          <TableCell>
                            <div className="flex flex-wrap gap-2">
                              {cert.dns_names.map((dns, idx) => (
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
                          {cert.organizational_units &&
                          cert.organizational_units.length > 0
                            ? cert.organizational_units.join(', ')
                            : 'None'}
                        </TableCell>
                      </TableRow>

                      {cert.serial_number && (
                        <TableRow>
                          <TableHead>Serial Number</TableHead>
                          <TableCell className="font-mono text-sm">
                            {cert.serial_number}
                          </TableCell>
                        </TableRow>
                      )}

                      <TableRow>
                        <TableHead>Validity Period</TableHead>
                        <TableCell>{cert.validity_days} days</TableCell>
                      </TableRow>

                      <TableRow>
                        <TableHead>Requested At</TableHead>
                        <TableCell>{formatDate(cert.requested_at)}</TableCell>
                      </TableRow>

                      {cert.issued_at && (
                        <TableRow>
                          <TableHead>Issued At</TableHead>
                          <TableCell>{formatDate(cert.issued_at)}</TableCell>
                        </TableRow>
                      )}

                      {cert.expires_at && (
                        <TableRow>
                          <TableHead>Expires At</TableHead>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              {formatDate(cert.expires_at)}
                              {daysUntilExpiry !== null &&
                                daysUntilExpiry > 0 && (
                                  <span className="text-sm text-muted-foreground">
                                    ({daysUntilExpiry} days remaining)
                                  </span>
                                )}
                            </div>
                          </TableCell>
                        </TableRow>
                      )}

                      {cert.message && (
                        <TableRow>
                          <TableHead>Request Message</TableHead>
                          <TableCell>{cert.message}</TableCell>
                        </TableRow>
                      )}

                      <TableRow>
                        <TableHead>Status</TableHead>
                        <TableCell>{getStatusBadge(cert.status)}</TableCell>
                      </TableRow>

                      <TableRow>
                        <TableHead>Owner</TableHead>
                        <TableCell className="font-mono text-xs">
                          <UserDisplay
                            displayName={cert.owner_display_name}
                            username={cert.owner_username}
                            sub={cert.owner_sub}
                            iss={cert.owner_iss}
                          />
                        </TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>

                  {cert.events && cert.events.length > 0 && (
                    <div>
                      <h3 className="font-semibold mb-2">Review History</h3>
                      <div className="space-y-2">
                        {cert.events.map((event) => (
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
                    {cert.status === 'issued' && (
                      <Button
                        onClick={() => {
                          setSelectedCertificateId(cert.id);
                          setDownloadDialogOpen(true);
                        }}
                      >
                        Download Certificate
                      </Button>
                    )}
                    {cert.status === 'awaiting_review' && (
                      <Button variant="outline" disabled>
                        Awaiting Admin Review
                      </Button>
                    )}
                    {cert.status === 'rejected' && (
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

      {filteredCertificates.length === 0 &&
        certificates &&
        certificates.length > 0 && (
          <div className="text-center py-12 text-muted-foreground">
            No certificates found matching "{searchQuery}"
          </div>
        )}

      {(!certificates || certificates.length === 0) && (
        <div className="text-center py-12 text-muted-foreground">
          <p className="mb-4">You have no certificates</p>
          <Button onClick={() => navigate({ to: '/settings/certs/requests' })}>
            Request one here
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
