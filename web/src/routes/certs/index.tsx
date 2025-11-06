import { createFileRoute } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';

export const Route = createFileRoute('/certs/')({
  component: RouteComponent,
  beforeLoad: async ({ location }) => {
    await requireAuth(
      location,
      'You must login to access the certs page.',
      true
    );
  },
});

function RouteComponent() {
  return <div>Hello "/certs/"!</div>;
}
