import { createFileRoute } from '@tanstack/react-router';

export const Route = createFileRoute('/settings/')({
  component: SettingsPage,
});

export default function SettingsPage() {
  return (
    <div>
      <h1 className="text-2xl font-bold mb-4">General Settings</h1>
      <p>
        Welcome to settings. Use the sidebar to navigate to different sections.
      </p>
    </div>
  );
}
