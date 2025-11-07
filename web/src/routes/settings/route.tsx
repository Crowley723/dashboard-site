import { Settings, ScrollText as Certificate } from 'lucide-react';
import { AppSidebar } from '@/components/app-sidebar';
import { createFileRoute, Outlet } from '@tanstack/react-router';
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar';
import { requireAuth } from '@/utils/Auth.ts';

export const Route = createFileRoute('/settings')({
  component: SettingsLayout,
  beforeLoad: async ({ location }) => {
    await requireAuth(
      location,
      'You must login to access the settings page.',
      true
    );
  },
});

const settingsNavItems = [
  {
    title: 'Certificates',
    url: '/settings/certs',
    icon: Certificate,
    isActive: true,
    items: [
      { title: 'Certificates', url: '/settings/certs' },
      { title: 'Requests', url: '/settings/certs/requests' },
      { title: 'Settings', url: '/settings/certs/settings' },
    ],
  },
  {
    title: 'General',
    url: '/settings',
    icon: Settings,
    items: [{ title: 'Profile', url: '/settings/profile' }],
  },
];

export default function SettingsLayout() {
  return (
    <SidebarProvider className="flex flex-col">
      <div className="flex flex-1">
        <AppSidebar navItems={settingsNavItems} title="Settings" />
        <SidebarInset>
          <div className="flex flex-1 flex-col gap-4 p-4">
            <div className="mb-4">
              <SidebarTrigger />
            </div>
            <Outlet />
          </div>
        </SidebarInset>
      </div>
    </SidebarProvider>
  );
}
