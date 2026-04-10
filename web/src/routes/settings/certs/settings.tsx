import { createFileRoute } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useAuth } from '@/hooks/useAuth';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { InfoIcon } from 'lucide-react';

export const Route = createFileRoute('/settings/certs/settings')({
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
  const { config } = useAuth();
  const mtlsConfig = config?.mtls;

  if (!mtlsConfig?.enabled) {
    return (
      <div className="container mx-auto p-6 max-w-4xl">
        <Alert>
          <InfoIcon className="h-4 w-4" />
          <AlertDescription>
            mTLS management is not enabled on this server.
          </AlertDescription>
        </Alert>
      </div>
    );
  }

  const subject = mtlsConfig.certificate_subject;
  const isDatabaseProvider = mtlsConfig.provider_type === 'database';

  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">Certificate Configuration</h1>
        <p className="text-muted-foreground">
          Configured Certificate Issuer Settings
        </p>
      </div>

      <Alert className="mb-6">
        <InfoIcon className="h-4 w-4" />
        <AlertDescription>
          These settings are currently managed by the server and cannot be
          modified through the UI. Contact your system administrator to make
          changes.
        </AlertDescription>
      </Alert>

      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Provider Information</CardTitle>
            <CardDescription>
              Certificate provider configuration
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="grid grid-cols-[200px_1fr] gap-4">
              <dt className="text-sm font-medium text-muted-foreground">
                Provider Type
              </dt>
              <dd className="text-sm">
                <Badge variant={isDatabaseProvider ? 'default' : 'secondary'}>
                  {mtlsConfig.provider_type || 'unknown'}
                </Badge>
              </dd>

              {isDatabaseProvider && mtlsConfig.key_algorithm && (
                <>
                  <dt className="text-sm font-medium text-muted-foreground">
                    Key Algorithm
                  </dt>
                  <dd className="text-sm font-mono">
                    {mtlsConfig.key_algorithm}
                  </dd>
                </>
              )}
            </div>
          </CardContent>
        </Card>

        {subject && (
          <Card>
            <CardHeader>
              <CardTitle>Certificate Subject</CardTitle>
              <CardDescription>
                Default subject information included in all issued certificates
              </CardDescription>
            </CardHeader>
            <CardContent>
              <dl className="grid grid-cols-[200px_1fr] gap-4">
                {subject.organization && (
                  <>
                    <dt className="text-sm font-medium text-muted-foreground">
                      Organization (O)
                    </dt>
                    <dd className="text-sm">{subject.organization}</dd>
                  </>
                )}

                {subject.country && (
                  <>
                    <dt className="text-sm font-medium text-muted-foreground">
                      Country (C)
                    </dt>
                    <dd className="text-sm">{subject.country}</dd>
                  </>
                )}

                {subject.locality && (
                  <>
                    <dt className="text-sm font-medium text-muted-foreground">
                      Locality (L)
                    </dt>
                    <dd className="text-sm">{subject.locality}</dd>
                  </>
                )}

                {subject.province && (
                  <>
                    <dt className="text-sm font-medium text-muted-foreground">
                      Province/State (ST)
                    </dt>
                    <dd className="text-sm">{subject.province}</dd>
                  </>
                )}

                {!subject.organization &&
                  !subject.country &&
                  !subject.locality &&
                  !subject.province && (
                    <dd className="col-span-2 text-sm text-muted-foreground">
                      No certificate subject information configured
                    </dd>
                  )}
              </dl>
            </CardContent>
          </Card>
        )}

        {isDatabaseProvider && (
          <Card>
            <CardHeader>
              <CardTitle>Database Provider</CardTitle>
              <CardDescription>
                Certificates are issued and stored directly in the database
              </CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              <p>
                The database certificate provider generates and stores all
                certificates internally. Certificate private keys are encrypted
                at rest using the configured encryption key.
              </p>
            </CardContent>
          </Card>
        )}

        {mtlsConfig.provider_type === 'kubernetes' && (
          <Card>
            <CardHeader>
              <CardTitle>Kubernetes Provider</CardTitle>
              <CardDescription>
                Certificates are issued via Kubernetes cert-manager
              </CardDescription>
            </CardHeader>
            <CardContent className="text-sm text-muted-foreground">
              <p>
                The Kubernetes certificate provider uses cert-manager to issue
                and manage certificates. Certificate resources are created in
                the Kubernetes cluster and synchronized with the database.
              </p>
            </CardContent>
          </Card>
        )}
      </div>
    </div>
  );
}
