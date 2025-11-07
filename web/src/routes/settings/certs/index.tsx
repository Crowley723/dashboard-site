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
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@/components/ui/table';

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

interface Certificate {
  id: string;
  commonName: string;
  dnsNames: string[];
  organization: string;
  organizationalUnits: string[];
  serialNumber: string;
  issuer: string;
  notBefore: string;
  notAfter: string;
  status: 'active' | 'expiring' | 'expired' | 'revoked';
  owner?: string;
}

const mockCertificates: Certificate[] = [
  {
    id: '1',
    commonName: 'brynn.crowley.example.com',
    dnsNames: ['brynn.crowley.example.com', '*.brynn.crowley.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['Engineering', 'Security'],
    serialNumber: '4a:3f:5c:d2:1b:9e:8f:7a',
    issuer: 'Crowley Labs Internal CA',
    notBefore: '2024-01-15T00:00:00Z',
    notAfter: '2025-01-15T00:00:00Z',
    status: 'active',
    owner: 'brynn@example.com',
  },
  {
    id: '2',
    commonName: 'dashboard.example.com',
    dnsNames: ['dashboard.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['Infrastructure'],
    serialNumber: '8b:2e:1a:c4:6d:9f:3e:5b',
    issuer: 'Crowley Labs Internal CA',
    notBefore: '2024-11-01T00:00:00Z',
    notAfter: '2024-12-15T00:00:00Z',
    status: 'expiring',
    owner: 'john.doe@example.com',
  },
  {
    id: '3',
    commonName: 'api.example.com',
    dnsNames: ['api.example.com', 'api-v2.example.com'],
    organization: 'Crowley Labs',
    organizationalUnits: ['Backend'],
    serialNumber: '2c:7d:9a:e1:4f:8b:6c:3a',
    issuer: 'Crowley Labs Internal CA',
    notBefore: '2023-06-01T00:00:00Z',
    notAfter: '2024-06-01T00:00:00Z',
    status: 'expired',
    owner: 'jane.smith@example.com',
  },
];

function RouteComponent() {
  const [certificates] = useState<Certificate[]>(mockCertificates);
  const [revokingId, setRevokingId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const handleRevoke = async (certId: string) => {
    setRevokingId(certId);
    // TODO: Implement actual revoke API call
    console.log('Revoking certificate:', certId);
    setTimeout(() => {
      setRevokingId(null);
      alert('Certificate revoke functionality coming soon!');
    }, 1000);
  };

  const getStatusBadge = (status: Certificate['status']) => {
    const variants = {
      active: 'default',
      expiring: 'outline',
      expired: 'secondary',
      revoked: 'destructive',
    } as const;

    return (
      <Badge variant={variants[status]}>
        {status.charAt(0).toUpperCase() + status.slice(1)}
      </Badge>
    );
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getDaysUntilExpiry = (notAfter: string) => {
    const expiry = new Date(notAfter);
    const now = new Date();
    const diff = expiry.getTime() - now.getTime();
    const days = Math.ceil(diff / (1000 * 60 * 60 * 24));
    return days;
  };

  const filteredCertificates = certificates.filter((cert) => {
    const query = searchQuery.toLowerCase();
    return (
      cert.commonName.toLowerCase().includes(query) ||
      cert.dnsNames.some((dns) => dns.toLowerCase().includes(query)) ||
      cert.organization.toLowerCase().includes(query) ||
      cert.organizationalUnits.some((ou) => ou.toLowerCase().includes(query)) ||
      cert.serialNumber.toLowerCase().includes(query) ||
      cert.issuer.toLowerCase().includes(query) ||
      cert.owner?.toLowerCase().includes(query) ||
      cert.status.toLowerCase().includes(query)
    );
  });

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">mTLS Certificates</h1>
          <p className="text-muted-foreground">
            Manage your mutual TLS certificates for secure authentication
          </p>
        </div>
        <Button onClick={() => alert('Certificate request form coming soon!')}>
          Request Certificate
        </Button>
      </div>

      <div className="mb-4">
        <Input
          placeholder="Search certificates by name, DNS, owner, serial, organization..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-xl"
        />
      </div>

      <Accordion type="single" collapsible className="space-y-4">
        {filteredCertificates.map((cert) => {
          const daysUntilExpiry = getDaysUntilExpiry(cert.notAfter);
          const isExpiringSoon = daysUntilExpiry <= 30 && daysUntilExpiry > 0;

          return (
            <AccordionItem
              key={cert.id}
              value={cert.id}
              className="border border-border rounded-lg px-4 !border-b"
            >
              <AccordionTrigger className="hover:no-underline">
                <div className="flex items-center justify-between w-full pr-4">
                  <div className="flex items-center gap-4">
                    <div className="text-left">
                      <div className="font-medium">{cert.commonName}</div>
                      <div className="text-sm text-muted-foreground">
                        Serial: {cert.serialNumber}
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
                        <TableCell>{cert.commonName}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>DNS Names</TableHead>
                        <TableCell>
                          <div className="flex flex-wrap gap-2">
                            {cert.dnsNames.map((dns, idx) => (
                              <Badge key={idx} variant="outline">
                                {dns}
                              </Badge>
                            ))}
                          </div>
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Organization</TableHead>
                        <TableCell>{cert.organization}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Organizational Units</TableHead>
                        <TableCell>
                          {cert.organizationalUnits.join(', ')}
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Serial Number</TableHead>
                        <TableCell className="font-mono text-sm">
                          {cert.serialNumber}
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Issuer</TableHead>
                        <TableCell>{cert.issuer}</TableCell>
                      </TableRow>
                      {cert.owner && (
                        <TableRow>
                          <TableHead>Owner</TableHead>
                          <TableCell>{cert.owner}</TableCell>
                        </TableRow>
                      )}
                      <TableRow>
                        <TableHead>Valid From</TableHead>
                        <TableCell>{formatDate(cert.notBefore)}</TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Valid Until</TableHead>
                        <TableCell>
                          <div className="flex items-center gap-2">
                            {formatDate(cert.notAfter)}
                            {daysUntilExpiry > 0 && (
                              <span className="text-sm text-muted-foreground">
                                ({daysUntilExpiry} days remaining)
                              </span>
                            )}
                          </div>
                        </TableCell>
                      </TableRow>
                      <TableRow>
                        <TableHead>Status</TableHead>
                        <TableCell>{getStatusBadge(cert.status)}</TableCell>
                      </TableRow>
                    </TableBody>
                  </Table>

                  <div className="flex gap-2 pt-2">
                    <Button
                      variant="destructive"
                      onClick={() => handleRevoke(cert.id)}
                      disabled={
                        cert.status === 'revoked' ||
                        cert.status === 'expired' ||
                        revokingId === cert.id
                      }
                    >
                      {revokingId === cert.id
                        ? 'Revoking...'
                        : 'Revoke Certificate'}
                    </Button>
                    <Button variant="outline">Download Certificate</Button>
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>

      {filteredCertificates.length === 0 && certificates.length > 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No certificates found matching "{searchQuery}"
        </div>
      )}

      {certificates.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No certificates found
        </div>
      )}
    </div>
  );
}
