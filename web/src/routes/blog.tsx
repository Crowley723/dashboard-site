import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/blog')({
  component: RouteComponent,
});

function RouteComponent() {
  return <div className={'p-4'}>Hello from /blog!</div>;
}
