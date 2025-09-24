import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/about')({
  component: About,
});

function About() {
  return <div className={'p-4'}>Hello from /about!</div>;
}
