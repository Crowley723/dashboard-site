'use client';

import * as React from 'react';
import { NavMain } from '@/components/ui/nav-main.tsx';
import {
  Sidebar,
  SidebarContent,
  SidebarHeader,
  SidebarRail,
} from '@/components/ui/sidebar.tsx';
import type { LucideIcon } from 'lucide-react';

interface AppSidebarProps extends React.ComponentProps<typeof Sidebar> {
  navItems: Array<{
    title: string;
    url: string;
    icon?: LucideIcon;
    isActive?: boolean;
    items?: Array<{
      title: string;
      url: string;
    }>;
  }>;
  title?: string;
}

export function AppSidebar({ navItems, title, ...props }: AppSidebarProps) {
  return (
    <Sidebar collapsible="icon" {...props}>
      {title && (
        <SidebarHeader>
          <div className="group-data-[collapsible=icon]:hidden">
            <h2 className="text-lg font-semibold">{title}</h2>
          </div>
        </SidebarHeader>
      )}
      <SidebarContent>
        <NavMain items={navItems} />
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
