import { createFileRoute } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState } from 'react';
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
  useUserEntries,
  useRemoveIPWhitelistEntry,
  useAvailableAliases,
  useAddIPWhitelistEntry,
} from '@/api/Firewall';
import type { FirewallIPStatus } from '@/types/Firewall';
import { UserDisplay } from '@/components/UserDisplay.tsx';
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
import { AddIPWhitelistDialog } from '@/components/AddIPWhitelistDialog';
import { RefreshCw, Check } from 'lucide-react';

export const Route = createFileRoute('/settings/firewall/')({
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
  const { isLoading: authLoading } = useAuth();
  const {
    data: entries,
    isLoading,
    isError,
    error,
    refetch,
  } = useUserEntries();
  const { data: aliases, isLoading: aliasesLoading } = useAvailableAliases();
  const removeIPMutation = useRemoveIPWhitelistEntry();
  const addIPMutation = useAddIPWhitelistEntry();
  const [searchQuery, setSearchQuery] = useState('');
  const [entryToRemove, setEntryToRemove] = useState<number | null>(null);
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [lastManualRefresh, setLastManualRefresh] = useState<number>(0);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [showSuccess, setShowSuccess] = useState(false);
  const [addErrorMessage, setAddErrorMessage] = useState<string>('');
  const [addSuccessMessage, setAddSuccessMessage] = useState<string>('');

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

  const handleAddIP = async (data: {
    alias_name: string;
    ip_address: string;
    description?: string;
    ttl?: string;
  }) => {
    setAddErrorMessage('');
    setAddSuccessMessage('');

    try {
      const result = await addIPMutation.mutateAsync(data);
      setAddSuccessMessage(result.message || 'IP address added successfully!');

      setTimeout(() => {
        setAddDialogOpen(false);
        setAddSuccessMessage('');
      }, 2000);
    } catch (error) {
      setAddErrorMessage(
        error instanceof Error ? error.message : 'Failed to add IP address'
      );
    }
  };

  // Authorization check
  if (authLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading...</div>
      </div>
    );
  }

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

  const getDaysUntilExpiry = (expiresAt: string | null) => {
    if (!expiresAt) return null;
    const expiry = new Date(expiresAt);
    const now = new Date();
    const diff = expiry.getTime() - now.getTime();
    const days = Math.ceil(diff / (1000 * 60 * 60 * 24));
    return days;
  };

  const handleRemoveIP = async (id: number) => {
    try {
      await removeIPMutation.mutateAsync(id);
      setEntryToRemove(null);
    } catch (error) {
      console.error('Failed to remove IP:', error);
    }
  };

  const filteredEntries = Array.isArray(entries)
    ? entries.filter((entry) => {
        const query = searchQuery.toLowerCase();
        return (
          entry.ip_address.toLowerCase().includes(query) ||
          entry.alias_name.toLowerCase().includes(query) ||
          entry.description?.toLowerCase().includes(query) ||
          entry.status.toLowerCase().includes(query)
        );
      })
    : [];

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

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">Firewall Whitelist</h1>
          <p className="text-muted-foreground">
            Manage your whitelisted IP addresses for firewall access
          </p>
        </div>
        <div className="flex gap-2">
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
          <AddIPWhitelistDialog
            aliases={aliases || []}
            entries={entries || []}
            open={addDialogOpen}
            onOpenChange={setAddDialogOpen}
            onSubmit={handleAddIP}
            isLoading={addIPMutation.isPending || aliasesLoading}
            errorMessage={addErrorMessage}
            successMessage={addSuccessMessage}
          >
            <Button>Add IP</Button>
          </AddIPWhitelistDialog>
        </div>
      </div>

      <div className="mb-4">
        <Input
          placeholder="Search by IP address, alias, description..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="max-w-xl"
        />
      </div>

      <Accordion type="single" collapsible className="space-y-4">
        {filteredEntries.map((entry) => {
          const daysUntilExpiry = getDaysUntilExpiry(entry.expires_at);
          const isExpiringSoon =
            daysUntilExpiry !== null &&
            daysUntilExpiry <= 7 &&
            daysUntilExpiry > 0;
          const canRemove =
            entry.status === 'requested' || entry.status === 'added';

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
                      <div className="font-medium font-mono">
                        {entry.ip_address}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {entry.alias_name}
                        {entry.description && ` • ${entry.description}`}
                      </div>
                    </div>
                  </div>
                  <div className="flex items-center gap-4">
                    {isExpiringSoon && (
                      <span className="text-sm text-orange-500">
                        Expires in {daysUntilExpiry} days
                      </span>
                    )}
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
                        <TableHead>Requested At</TableHead>
                        <TableCell>{formatDate(entry.requested_at)}</TableCell>
                      </TableRow>

                      {entry.added_at && (
                        <TableRow>
                          <TableHead>Added At</TableHead>
                          <TableCell>{formatDate(entry.added_at)}</TableCell>
                        </TableRow>
                      )}

                      {entry.expires_at && (
                        <TableRow>
                          <TableHead>Expires At</TableHead>
                          <TableCell>
                            <div className="flex items-center gap-2">
                              {formatDate(entry.expires_at)}
                              {daysUntilExpiry !== null &&
                                daysUntilExpiry > 0 && (
                                  <span className="text-sm text-muted-foreground">
                                    ({daysUntilExpiry} days remaining)
                                  </span>
                                )}
                            </div>
                          </TableCell>
                        </TableRow>
                      )}

                      {entry.removed_at && (
                        <TableRow>
                          <TableHead>Removed At</TableHead>
                          <TableCell>{formatDate(entry.removed_at)}</TableCell>
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
                      <div className="space-y-2">
                        {entry.events.map((event) => (
                          <div
                            key={event.id}
                            className="text-sm border-l-2 border-muted pl-3"
                          >
                            <div className="font-medium">
                              Event: {event.event_type}
                            </div>
                            <div className="text-muted-foreground">
                              By:{' '}
                              <UserDisplay
                                displayName={event.actor_display_name}
                                username={event.actor_username}
                                sub={event.actor_sub}
                                iss={event.actor_iss}
                              />
                            </div>
                            {event.notes && (
                              <div className="text-muted-foreground">
                                Notes: {event.notes}
                              </div>
                            )}
                            {event.client_ip && (
                              <div className="text-muted-foreground text-xs">
                                From: {event.client_ip}
                              </div>
                            )}
                            <div className="text-xs text-muted-foreground">
                              {formatDate(event.created_at)}
                            </div>
                          </div>
                        ))}
                      </div>
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
                    {entry.status === 'blacklisted_by_admin' && (
                      <Button variant="destructive" disabled>
                        Blacklisted
                      </Button>
                    )}
                  </div>
                </div>
              </AccordionContent>
            </AccordionItem>
          );
        })}
      </Accordion>

      {filteredEntries.length === 0 && entries && entries.length > 0 && (
        <div className="text-center py-12 text-muted-foreground">
          No entries found matching "{searchQuery}"
        </div>
      )}

      {(!entries || entries.length === 0) && (
        <div className="text-center py-12 text-muted-foreground">
          <p className="mb-4">You don't have any whitelisted IPs</p>
          <Button onClick={() => setAddDialogOpen(true)}>Add one here</Button>
        </div>
      )}

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
    </div>
  );
}
