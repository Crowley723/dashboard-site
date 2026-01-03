import { Settings, ScrollText as Certificate } from 'lucide-react';
import { AppSidebar } from '@/components/app-sidebar';
import { createFileRoute, Outlet } from '@tanstack/react-router';
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '@/components/ui/sidebar';
import { requireAuth } from '@/utils/Auth.ts';
import { useAuth } from '@/hooks/useAuth';

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

export default function SettingsLayout() {
  const { isMTLSUser, isMTLSAdmin } = useAuth();

  // Build certificate menu items based on user permissions
  const certificateItems = [];
  if (isMTLSUser() || isMTLSAdmin()) {
    certificateItems.push({ title: 'Certificates', url: '/settings/certs' });
    certificateItems.push({
      title: 'Requests',
      url: '/settings/certs/requests',
    });
  }
  if (isMTLSAdmin()) {
    certificateItems.push({
      title: 'Admin Requests',
      url: '/settings/certs/admin/requests',
    });
  }
  if (isMTLSAdmin()) {
    certificateItems.push({
      title: 'Settings',
      url: '/settings/certs/settings',
    });
  }

  const settingsNavItems = [];

  // Only show certificates section if user has access
  if (certificateItems.length > 0) {
    settingsNavItems.push({
      title: 'Certificates',
      url: '/settings/certs',
      icon: Certificate,
      isActive: true,
      items: certificateItems,
    });
  }

  // General settings (always visible)
  settingsNavItems.push({
    title: 'General',
    url: '/settings',
    icon: Settings,
    items: [{ title: 'Profile', url: '/settings/profile' }],
  });

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
