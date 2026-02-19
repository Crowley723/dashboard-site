import { ScrollText as Certificate, Key, Shield, User } from 'lucide-react';
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
  const { isMTLSUser, isMTLSAdmin, hasFirewallAccess, isFirewallAdmin } =
    useAuth();

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

  // Build firewall menu items based on user permissions
  const firewallItems = [];
  if (hasFirewallAccess() || isFirewallAdmin()) {
    firewallItems.push({ title: 'Whitelist', url: '/settings/firewall' });
  }
  if (isFirewallAdmin()) {
    firewallItems.push({
      title: 'Admin',
      url: '/settings/firewall/admin',
    });
  }

  const settingsNavItems = [];

  settingsNavItems.push({
    groupLabel: 'Account',
    items: [
      {
        title: 'Profile',
        url: '/settings/profile',
        icon: User,
      },
    ],
  });

  settingsNavItems.push({
    groupLabel: 'API & Access',
    items: [
      {
        title: 'Service Accounts',
        url: '/settings/service-accounts',
        icon: Key,
      },
    ],
  });

  const securityItems = [];
  if (certificateItems.length > 0) {
    securityItems.push({
      title: 'Certificates',
      url: '/settings/certs',
      icon: Certificate,
      isActive: true,
      items: certificateItems,
    });
  }
  if (firewallItems.length > 0) {
    securityItems.push({
      title: 'Firewall',
      url: '/settings/firewall',
      icon: Shield,
      items: firewallItems.length > 1 ? firewallItems : undefined,
    });
  }

  if (securityItems.length > 0) {
    settingsNavItems.push({
      groupLabel: 'Security',
      items: securityItems,
    });
  }

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
