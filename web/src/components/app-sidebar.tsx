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

interface NavGroup {
  groupLabel?: string;
  items: Array<{
    title: string;
    url: string;
    icon?: LucideIcon;
    isActive?: boolean;
    items?: Array<{
      title: string;
      url: string;
    }>;
  }>;
}

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
  }> | NavGroup[];
  title?: string;
}

export function AppSidebar({ navItems, title, ...props }: AppSidebarProps) {
  // Check if navItems is grouped or flat
  const isGrouped = navItems.length > 0 && 'groupLabel' in navItems[0];

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
        {isGrouped ? (
          (navItems as NavGroup[]).map((group, index) => (
            <NavMain key={index} items={group.items} groupLabel={group.groupLabel} />
          ))
        ) : (
          <NavMain items={navItems as NavGroup['items']} />
        )}
      </SidebarContent>
      <SidebarRail />
    </Sidebar>
  );
}
