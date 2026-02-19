import { createFileRoute, Navigate } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState, useMemo } from 'react';
import { useAuth } from '@/hooks/useAuth';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Input } from '@/components/ui/input';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@/components/ui/table';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { Textarea } from '@/components/ui/textarea';
import { Label } from '@/components/ui/label';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import {
  useAllFirewallEntries,
  useRemoveIPWhitelistEntry,
  useBlacklistIPEntry,
} from '@/api/Firewall';
import type {
  FirewallIPStatus,
  FirewallIPWhitelistEntry,
} from '@/types/Firewall';
import { UserDisplay } from '@/components/UserDisplay.tsx';
import { RefreshCw, Check } from 'lucide-react';
import { getRelativeTimeString } from '@/hooks/RelativeTimeString.tsx';

export const Route = createFileRoute('/settings/firewall/admin/')({
  component: RouteComponent,
  beforeLoad: async ({ location }) => {
    await requireAuth(
      location,
      'You must login to access the settings page.',
      true
    );
  },
});

function RouteComponent() {
  const { isFirewallAdmin, isLoading: authLoading } = useAuth();
  const {
    data: entries,
    isLoading,
    isError,
    error,
    refetch,
  } = useAllFirewallEntries();
  const removeIPMutation = useRemoveIPWhitelistEntry();
  const blacklistMutation = useBlacklistIPEntry();

  // State management
  const [searchQuery, setSearchQuery] = useState('');
  const [aliasFilter, setAliasFilter] = useState<string>('all');
  const [userFilter, setUserFilter] = useState('');
  const [lastManualRefresh, setLastManualRefresh] = useState<number>(0);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);

  // Action state
  const [entryToRemove, setEntryToRemove] = useState<number | null>(null);
  const [entryToBlacklist, setEntryToBlacklist] =
    useState<FirewallIPWhitelistEntry | null>(null);
  const [blacklistReason, setBlacklistReason] = useState('');

  // Filtering logic - MUST be before early returns to avoid hooks rule violation
  const filteredEntries = useMemo(() => {
    if (!Array.isArray(entries)) return [];

    return entries.filter((entry) => {
      const query = searchQuery.toLowerCase();
      const matchesSearch =
        entry.ip_address.toLowerCase().includes(query) ||
        entry.owner_username.toLowerCase().includes(query) ||
        entry.owner_display_name.toLowerCase().includes(query) ||
        entry.alias_name.toLowerCase().includes(query) ||
        entry.description?.toLowerCase().includes(query) ||
        false;

      const matchesAlias =
        aliasFilter === 'all' || entry.alias_name === aliasFilter;

      const matchesUser =
        !userFilter ||
        entry.owner_username.toLowerCase().includes(userFilter.toLowerCase()) ||
        entry.owner_display_name
          .toLowerCase()
          .includes(userFilter.toLowerCase());

      return matchesSearch && matchesAlias && matchesUser;
    });
  }, [entries, searchQuery, aliasFilter, userFilter]);

  // Group by status
  const activeEntries = filteredEntries.filter(
    (e) => e.status === 'requested' || e.status === 'added'
  );
  const inactiveEntries = filteredEntries.filter(
    (e) =>
      e.status === 'removed' ||
      e.status === 'removed_by_admin' ||
      e.status === 'blacklisted_by_admin'
  );

  // Get unique aliases for filter dropdown
  const uniqueAliases = useMemo(() => {
    if (!Array.isArray(entries)) return [];
    const aliasSet = new Set(entries.map((e) => e.alias_name));
    return Array.from(aliasSet).sort();
  }, [entries]);

  // Manual refresh handler
  const handleManualRefresh = async () => {
    const now = Date.now();
    const timeSinceLastRefresh = now - lastManualRefresh;

    setIsRefreshing(true);
    setShowSuccess(false);

    if (timeSinceLastRefresh < 10000) {
      setTimeout(() => {
        setIsRefreshing(false);
      }, 1500);
      return;
    }

    setLastManualRefresh(now);
    await refetch();

    setTimeout(() => {
      setIsRefreshing(false);
      setShowSuccess(true);
      setTimeout(() => {
        setShowSuccess(false);
      }, 2000);
    }, 1500);
  };

  // Authorization check - AFTER all hooks
  if (authLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading...</div>
      </div>
    );
  }

  if (!isFirewallAdmin()) {
    return <Navigate to="/settings/firewall" />;
  }

  // Helper functions
  const getStatusBadge = (status: FirewallIPStatus) => {
    const variants: Record<
      FirewallIPStatus,
      | 'default'
      | 'outline'
      | 'secondary'
      | 'destructive'
      | 'warning'
      | 'success'
      | 'info'
    > = {
      requested: 'warning',
      added: 'success',
      removed: 'secondary',
      removed_by_admin: 'destructive',
      blacklisted_by_admin: 'destructive',
    };

    const labels: Record<FirewallIPStatus, string> = {
      requested: 'Pending',
      added: 'Active',
      removed: 'Removed',
      removed_by_admin: 'Removed by Admin',
      blacklisted_by_admin: 'Blacklisted',
    };

    return <Badge variant={variants[status]}>{labels[status]}</Badge>;
  };

  const formatDate = (dateString: string | null) => {
    if (!dateString) return 'N/A';
    const date = new Date(dateString);
    return `${date.toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    })} at ${date.toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
      hour12: true,
      timeZoneName: 'short',
    })}`;
  };

  // Action handlers
  const handleRemoveIP = async (id: number) => {
    try {
      await removeIPMutation.mutateAsync(id);
      setEntryToRemove(null);
    } catch (error) {
      console.error('Failed to remove IP:', error);
    }
  };

  const handleBlacklistIP = async () => {
    if (!entryToBlacklist) return;
    try {
      await blacklistMutation.mutateAsync({
        id: entryToBlacklist.id,
        reason: blacklistReason.trim() || undefined,
      });
      setEntryToBlacklist(null);
      setBlacklistReason('');
    } catch (error) {
      console.error('Failed to blacklist IP:', error);
    }
  };

  if (isLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading firewall entries...</div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12 text-destructive">
          Error loading firewall entries: {error?.message}
        </div>
      </div>
    );
  }

  const renderEntry = (entry: FirewallIPWhitelistEntry) => {
    const canRemove = entry.status === 'requested' || entry.status === 'added';
    const canBlacklist = entry.status !== 'blacklisted_by_admin';

    return (
      <AccordionItem
        key={entry.id}
        value={entry.id.toString()}
        className="border border-border rounded-lg px-4 !border-b"
      >
        <AccordionTrigger className="hover:no-underline">
          <div className="flex items-center justify-between w-full pr-4">
            <div className="flex items-center gap-4">
              <div className="text-left">
                <div className="font-medium font-mono">{entry.ip_address}</div>
                <div className="text-sm text-muted-foreground">
                  <UserDisplay
                    displayName={entry.owner_display_name}
                    username={entry.owner_username}
                    sub={entry.owner_sub}
                    iss={entry.owner_iss}
                  />
                  {' • '}
                  {entry.alias_name}
                </div>
              </div>
            </div>
            <div className="flex items-center gap-4">
              {getStatusBadge(entry.status)}
            </div>
          </div>
        </AccordionTrigger>
        <AccordionContent>
          <div className="pt-4 space-y-6">
            <Table>
              <TableBody>
                <TableRow>
                  <TableHead className="w-1/3">IP Address</TableHead>
                  <TableCell className="font-mono">
                    {entry.ip_address}
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableHead>IP Version</TableHead>
                  <TableCell>IPv{entry.ip_version}</TableCell>
                </TableRow>

                <TableRow>
                  <TableHead>Alias</TableHead>
                  <TableCell>{entry.alias_name}</TableCell>
                </TableRow>

                {entry.description && (
                  <TableRow>
                    <TableHead>Description</TableHead>
                    <TableCell>{entry.description}</TableCell>
                  </TableRow>
                )}

                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableCell>{getStatusBadge(entry.status)}</TableCell>
                </TableRow>

                <TableRow>
                  <TableHead>Owner</TableHead>
                  <TableCell>
                    <UserDisplay
                      displayName={entry.owner_display_name}
                      username={entry.owner_username}
                      sub={entry.owner_sub}
                      iss={entry.owner_iss}
                    />
                  </TableCell>
                </TableRow>

                <TableRow>
                  <TableHead>Requested At</TableHead>
                  <TableCell>{formatDate(entry.requested_at)}</TableCell>
                </TableRow>

                {entry.added_at && (
                  <TableRow>
                    <TableHead>Added At</TableHead>
                    <TableCell>{formatDate(entry.added_at)}</TableCell>
                  </TableRow>
                )}

                {entry.removed_at && (
                  <TableRow>
                    <TableHead>Removed At</TableHead>
                    <TableCell>{formatDate(entry.removed_at)}</TableCell>
                  </TableRow>
                )}

                {entry.expires_at && (
                  <TableRow>
                    <TableHead>Expires At</TableHead>
                    <TableCell>{formatDate(entry.expires_at)}</TableCell>
                  </TableRow>
                )}

                {entry.removed_by_username && (
                  <TableRow>
                    <TableHead>Removed By</TableHead>
                    <TableCell>
                      <UserDisplay
                        displayName={entry.removed_by_display_name || ''}
                        username={entry.removed_by_username}
                        sub={entry.removed_by_sub || ''}
                        iss={entry.removed_by_iss || ''}
                      />
                    </TableCell>
                  </TableRow>
                )}

                {entry.removal_reason && (
                  <TableRow>
                    <TableHead>Removal Reason</TableHead>
                    <TableCell>{entry.removal_reason}</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>

            {entry.events && entry.events.length > 0 && (
              <div>
                <h3 className="font-semibold mb-2">Event History</h3>
                <Table>
                  <TableBody>
                    {entry.events.map((event) => (
                      <TableRow key={event.id}>
                        <TableCell className="font-medium">
                          {event.event_type}
                        </TableCell>
                        <TableCell>
                          <UserDisplay
                            displayName={event.actor_display_name}
                            username={event.actor_username}
                            sub={event.actor_sub}
                            iss={event.actor_iss}
                          />
                        </TableCell>
                        <TableCell>
                          {event.notes && (
                            <div className="text-sm">{event.notes}</div>
                          )}
                          {event.client_ip && (
                            <div className="text-xs text-muted-foreground">
                              IP: {event.client_ip}
                            </div>
                          )}
                          {event.user_agent && (
                            <div className="text-xs text-muted-foreground">
                              UA: {event.user_agent}
                            </div>
                          )}
                        </TableCell>
                        <TableCell className="text-sm text-muted-foreground">
                          {formatDate(event.created_at)}
                        </TableCell>
                      </TableRow>
                    ))}
                  </TableBody>
                </Table>
              </div>
            )}

            <div className="flex gap-2 pt-2">
              {canRemove && (
                <Button
                  variant="destructive"
                  onClick={() => setEntryToRemove(entry.id)}
                  disabled={removeIPMutation.isPending}
                >
                  Remove IP
                </Button>
              )}
              {canBlacklist && (
                <Button
                  variant="destructive"
                  onClick={() => setEntryToBlacklist(entry)}
                  disabled={blacklistMutation.isPending}
                >
                  Blacklist IP
                </Button>
              )}
            </div>
          </div>
        </AccordionContent>
      </AccordionItem>
    );
  };

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">
            Admin - Firewall Whitelist
          </h1>
          <p className="text-muted-foreground">
            Manage all firewall whitelist entries across users
            {lastManualRefresh > 0 && (
              <span className="ml-2">
                • Last synced{' '}
                {getRelativeTimeString(new Date(lastManualRefresh))}
              </span>
            )}
          </p>
        </div>
        <Button
          variant="outline"
          size="icon"
          onClick={handleManualRefresh}
          disabled={isRefreshing}
          className={showSuccess ? 'text-green-600' : ''}
        >
          {showSuccess ? (
            <Check />
          ) : (
            <RefreshCw className={isRefreshing ? 'animate-spin' : ''} />
          )}
        </Button>
      </div>

      <div className="mb-4 flex gap-4">
        <Input
          placeholder="Search by IP, user, alias, description..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="flex-1"
        />
        <Select value={aliasFilter} onValueChange={setAliasFilter}>
          <SelectTrigger className="w-[200px]">
            <SelectValue placeholder="Filter by alias" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Aliases</SelectItem>
            {uniqueAliases.map((alias) => (
              <SelectItem key={alias} value={alias}>
                {alias}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <Input
          placeholder="Filter by user..."
          value={userFilter}
          onChange={(e) => setUserFilter(e.target.value)}
          className="w-[200px]"
        />
      </div>

      {activeEntries.length > 0 && (
        <div className="mb-8">
          <h2 className="text-xl font-semibold mb-4">
            Active Entries ({activeEntries.length})
          </h2>
          <Accordion type="single" collapsible className="space-y-4">
            {activeEntries.map(renderEntry)}
          </Accordion>
        </div>
      )}

      {inactiveEntries.length > 0 && (
        <div>
          <h2 className="text-xl font-semibold mb-4">
            Inactive Entries ({inactiveEntries.length})
          </h2>
          <Accordion type="single" collapsible className="space-y-4">
            {inactiveEntries.map(renderEntry)}
          </Accordion>
        </div>
      )}

      {filteredEntries.length === 0 && entries && entries.length > 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No entries found matching your filters
        </div>
      )}

      {(!entries || entries.length === 0) && (
        <div className="text-center py-12 text-muted-foreground">
          No firewall whitelist entries found
        </div>
      )}

      {/* Remove IP Dialog */}
      <AlertDialog
        open={entryToRemove !== null}
        onOpenChange={() => setEntryToRemove(null)}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Remove IP from Whitelist?</AlertDialogTitle>
            <AlertDialogDescription>
              This will remove the IP address from the firewall whitelist. This
              action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={() => entryToRemove && handleRemoveIP(entryToRemove)}
              disabled={removeIPMutation.isPending}
            >
              {removeIPMutation.isPending ? 'Removing...' : 'Remove'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>

      {/* Blacklist IP Dialog */}
      <AlertDialog
        open={entryToBlacklist !== null}
        onOpenChange={() => {
          setEntryToBlacklist(null);
          setBlacklistReason('');
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Blacklist IP Address?</AlertDialogTitle>
            <AlertDialogDescription>
              This will blacklist the IP address{' '}
              <span className="font-mono font-semibold">
                {entryToBlacklist?.ip_address}
              </span>
              , preventing it from being re-added. This action cannot be
              undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="py-4">
            <Label htmlFor="blacklist-reason">Reason (optional)</Label>
            <Textarea
              id="blacklist-reason"
              placeholder="Explain why this IP is being blacklisted..."
              value={blacklistReason}
              onChange={(e) => setBlacklistReason(e.target.value)}
              rows={3}
              className="mt-2"
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleBlacklistIP}
              disabled={blacklistMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {blacklistMutation.isPending ? 'Blacklisting...' : 'Blacklist'}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
