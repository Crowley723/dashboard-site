import { createFileRoute } from '@tanstack/react-router';
import { z } from 'zod';
import { Card } from '@/components/ui/card.tsx';

const oidcErrorSearchSchema = z.object({
  error: z.string().optional(),
  error_description: z.string().optional(),
  error_uri: z.string().optional(),
  state: z.string().optional(),
});

export const Route = createFileRoute('/error')({
  component: ErrorRoute,
  validateSearch: oidcErrorSearchSchema,
});

function ErrorRoute() {
  const { error, error_description, error_uri } = Route.useSearch();

  return (
    <div
      className={'flex flex-col items-center justify-center pt-8 lg:pt-[20vh]'}
    >
      <Card className={'p-6'}>
        <div className={'text-center'}>
          <h1>Authentication Failed</h1>
          {error && (
            <p>
              <strong>Error:</strong> {error}
            </p>
          )}
          {error_description && (
            <p>
              <strong>Description:</strong> {error_description}
            </p>
          )}
          {error_uri && (
            <p>
              <a href={error_uri}>More info</a>
            </p>
          )}
        </div>
      </Card>
    </div>
  );
}
