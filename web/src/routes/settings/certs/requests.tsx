import { createFileRoute } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState } from 'react';
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

interface CertificateRequest {
  id: string;
  commonName: string;
  dnsNames: string[];
  organization: string;
  organizationalUnits: string[];
  requestedAt: string;
  requestedBy: string;
  status: 'pending' | 'approved' | 'rejected' | 'issued';
  reviewedAt?: string;
  reviewedBy?: string;
  rejectionReason?: string;
  notes?: string;
}

const mockRequests: CertificateRequest[] = [
  {
    id: '1',
    commonName: 'new-service.example.com',
    dnsNames: ['new-service.example.com', '*.new-service.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['Engineering', 'Platform'],
    requestedAt: '2024-11-05T14:30:00Z',
    requestedBy: 'john.doe@example.com',
    status: 'pending',
    notes: 'Certificate needed for new microservice deployment',
  },
  {
    id: '2',
    commonName: 'staging.api.example.com',
    dnsNames: ['staging.api.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['Engineering'],
    requestedAt: '2024-11-04T10:15:00Z',
    requestedBy: 'jane.smith@example.com',
    status: 'approved',
    reviewedAt: '2024-11-04T11:00:00Z',
    reviewedBy: 'admin@example.com',
    notes: 'Staging environment certificate',
  },
  {
    id: '3',
    commonName: 'test.internal.example.com',
    dnsNames: ['test.internal.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['QA'],
    requestedAt: '2024-11-03T16:45:00Z',
    requestedBy: 'bob.wilson@example.com',
    status: 'rejected',
    reviewedAt: '2024-11-03T17:30:00Z',
    reviewedBy: 'admin@example.com',
    rejectionReason: 'Invalid DNS name - internal domain not approved',
  },
  {
    id: '4',
    commonName: 'prod.api.example.com',
    dnsNames: ['prod.api.example.com', 'api.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['Engineering', 'Production'],
    requestedAt: '2024-11-01T09:00:00Z',
    requestedBy: 'alice.brown@example.com',
    status: 'issued',
    reviewedAt: '2024-11-01T10:00:00Z',
    reviewedBy: 'admin@example.com',
    notes: 'Production API certificate - approved and issued',
  },
];

function RouteComponent() {
  const [requests] = useState<CertificateRequest[]>(mockRequests);
  const [processingId, setProcessingId] = useState<string | null>(null);

  const handleApprove = async (requestId: string) => {
    setProcessingId(requestId);
    // TODO: Implement actual approve API call
    console.log('Approving request:', requestId);
    setTimeout(() => {
      setProcessingId(null);
      alert('Request approval functionality coming soon!');
    }, 1000);
  };

  const handleReject = async (requestId: string) => {
    setProcessingId(requestId);
    // TODO: Implement actual reject API call
    console.log('Rejecting request:', requestId);
    setTimeout(() => {
      setProcessingId(null);
      alert('Request rejection functionality coming soon!');
    }, 1000);
  };

  const handleIssue = async (requestId: string) => {
    setProcessingId(requestId);
    // TODO: Implement actual issue API call
    console.log('Issuing certificate for request:', requestId);
    setTimeout(() => {
      setProcessingId(null);
      alert('Certificate issuance functionality coming soon!');
    }, 1000);
  };

  const getStatusBadge = (status: CertificateRequest['status']) => {
    const config = {
      pending: { variant: 'outline' as const, label: 'Pending Review' },
      approved: { variant: 'default' as const, label: 'Approved' },
      rejected: { variant: 'destructive' as const, label: 'Rejected' },
      issued: { variant: 'secondary' as const, label: 'Issued' },
    };

    const { variant, label } = config[status];
    return <Badge variant={variant}>{label}</Badge>;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const getRelativeTime = (dateString: string) => {
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

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">Certificate Requests</h1>
        <p className="text-muted-foreground">
          Review and manage certificate signing requests
        </p>
      </div>

      <Accordion type="single" collapsible className="space-y-4">
        {requests.map((request) => {
          return (
            <AccordionItem
              key={request.id}
              value={request.id}
              className="border border-border rounded-lg px-4 !border-b"
            >
              <AccordionTrigger className="hover:no-underline">
                <div className="flex items-center justify-between w-full pr-4">
                  <div className="flex items-center gap-4">
                    <div className="text-left">
                      <div className="font-medium">{request.commonName}</div>
                      <div className="text-sm text-muted-foreground">
                        Requested by {request.requestedBy} •{' '}
                        {getRelativeTime(request.requestedAt)}
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
                        <TableCell>{request.commonName}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>DNS Names</TableHead>
                        <TableCell>
                          <div className="flex flex-wrap gap-2">
                            {request.dnsNames.map((dns, idx) => (
                              <Badge key={idx} variant="outline">
                                {dns}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Organization</TableHead>
                        <TableCell>{request.organization}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Organizational Units</TableHead>
                        <TableCell>
                          {request.organizationalUnits.join(', ')}
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Requested At</TableHead>
                        <TableCell>{formatDate(request.requestedAt)}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Requested By</TableHead>
                        <TableCell>{request.requestedBy}</TableCell>
                      </TableRow>
                      {request.reviewedAt && (
                        <>
                          <TableRow>
                            <TableHead>Reviewed At</TableHead>
                            <TableCell>
                              {formatDate(request.reviewedAt)}
                            </TableCell>
                          </TableRow>
                          <TableRow>
                            <TableHead>Reviewed By</TableHead>
                            <TableCell>{request.reviewedBy}</TableCell>
                          </TableRow>
                        </>
                      )}
                      {request.rejectionReason && (
                        <TableRow>
                          <TableHead>Rejection Reason</TableHead>
                          <TableCell className="text-destructive">
                            {request.rejectionReason}
                          </TableCell>
                        </TableRow>
                      )}
                      {request.notes && (
                        <TableRow>
                          <TableHead>Notes</TableHead>
                          <TableCell>{request.notes}</TableCell>
                        </TableRow>
                      )}
                      <TableRow>
                        <TableHead>Status</TableHead>
                        <TableCell>{getStatusBadge(request.status)}</TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>

                  <div className="flex gap-2 pt-2">
                    {request.status === 'pending' && (
                      <>
                        <Button
                          variant="default"
                          onClick={() => handleApprove(request.id)}
                          disabled={processingId === request.id}
                        >
                          {processingId === request.id
                            ? 'Approving...'
                            : 'Approve'}
                        </Button>
                        <Button
                          variant="destructive"
                          onClick={() => handleReject(request.id)}
                          disabled={processingId === request.id}
                        >
                          {processingId === request.id
                            ? 'Rejecting...'
                            : 'Reject'}
                        </Button>
                      </>
                    )}
                    {request.status === 'approved' && (
                      <Button
                        variant="default"
                        onClick={() => handleIssue(request.id)}
                        disabled={processingId === request.id}
                      >
                        {processingId === request.id
                          ? 'Issuing...'
                          : 'Issue Certificate'}
                      </Button>
                    )}
                    {request.status === 'issued' && (
                      <Button variant="outline">View Certificate</Button>
                    )}
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>

      {requests.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No certificate requests found
        </div>
      )}
    </div>
  );
}
