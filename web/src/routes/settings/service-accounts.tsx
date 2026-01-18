import { createFileRoute } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import {
  Accordion,
  AccordionContent,
  AccordionItem,
  AccordionTrigger,
} from '@/components/ui/accordion';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
} from '@/components/ui/table';
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
import { CreateServiceAccountDialog } from '@/components/CreateServiceAccountDialog';
import {
  useServiceAccounts,
  useDeleteServiceAccount,
  usePauseServiceAccount,
  useUnpauseServiceAccount,
} from '@/api/ServiceAccounts';
import { Plus, Trash2, Pause, Play } from 'lucide-react';
import type { ServiceAccount } from '@/types/ServiceAccounts';
import { getRelativeTimeString } from '@/hooks/RelativeTimeString.tsx';

export const Route = createFileRoute('/settings/service-accounts')({
  component: ServiceAccountsPage,
  beforeLoad: async ({ location }) => {
    await requireAuth(
      location,
      'You must login to access the settings page.',
      true
    );
  },
});

function ServiceAccountsPage() {
  const {
    data: serviceAccounts,
    isLoading,
    isError,
    error,
  } = useServiceAccounts();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedAccount, setSelectedAccount] =
    useState<ServiceAccount | null>(null);

  const deleteMutation = useDeleteServiceAccount();
  const pauseMutation = usePauseServiceAccount();
  const unpauseMutation = useUnpauseServiceAccount();

  const handleDelete = async () => {
    if (!selectedAccount) return;

    try {
      await deleteMutation.mutateAsync({
        iss: selectedAccount.iss,
        sub: selectedAccount.sub,
      });
      setDeleteDialogOpen(false);
      setSelectedAccount(null);
    } catch (error) {
      console.error('Failed to delete service account:', error);
    }
  };

  const handlePause = async (account: ServiceAccount) => {
    try {
      await pauseMutation.mutateAsync({
        iss: account.iss,
        sub: account.sub,
      });
    } catch (error) {
      console.error('Failed to pause service account:', error);
    }
  };

  const handleUnpause = async (account: ServiceAccount) => {
    try {
      await unpauseMutation.mutateAsync({
        iss: account.iss,
        sub: account.sub,
      });
    } catch (error) {
      console.error('Failed to unpause service account:', error);
    }
  };

  const openDeleteDialog = (account: ServiceAccount) => {
    setSelectedAccount(account);
    setDeleteDialogOpen(true);
  };

  const formatDate = (dateString: string) => {
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

  const isExpired = (expiresAt: string) => {
    return new Date(expiresAt) < new Date();
  };

  const getStatusBadge = (account: ServiceAccount) => {
    if (account.deleted_at) {
      return <Badge variant="destructive">Deleted</Badge>;
    }
    if (account.is_disabled) {
      return <Badge variant="outline">Paused</Badge>;
    }
    if (isExpired(account.expires_at)) {
      return <Badge variant="destructive">Expired</Badge>;
    }
    return <Badge variant="success">Active</Badge>;
  };

  if (isLoading) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12">Loading service accounts...</div>
      </div>
    );
  }

  if (isError) {
    return (
      <div className="container mx-auto p-6 max-w-6xl">
        <div className="text-center py-12 text-destructive">
          Error loading service accounts: {error?.message}
        </div>
      </div>
    );
  }

  const accounts = Array.isArray(serviceAccounts) ? serviceAccounts : [];

  return (
    <div className="container mx-auto p-6 max-w-6xl">
      <div className="mb-6 flex items-start justify-between">
        <div>
          <h1 className="text-3xl font-bold mb-2">Service Accounts</h1>
          <p className="text-muted-foreground">
            Manage API service accounts for automated access
          </p>
        </div>
        <Button onClick={() => setDialogOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Create Service Account
        </Button>
      </div>

      <Accordion type="single" collapsible className="space-y-4">
        {accounts.map((account) => (
          <AccordionItem
            key={`${account.iss}:${account.sub}`}
            value={`${account.iss}:${account.sub}`}
            className="border border-border rounded-lg px-4 !border-b"
          >
            <AccordionTrigger className="hover:no-underline">
              <div className="flex items-center justify-between w-full pr-4">
                <div className="flex items-center gap-4">
                  <div className="text-left">
                    <div className="font-medium">{account.name}</div>
                    <div className="text-sm text-muted-foreground">
                      Created {getRelativeTimeString(new Date(account.created_at))}
                    </div>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  {getStatusBadge(account)}
                </div>
              </div>
            </AccordionTrigger>
            <AccordionContent>
              <div className="pt-4 space-y-6">
                <Table>
                  <TableBody>
                    <TableRow>
                      <TableHead className="w-1/3">Name</TableHead>
                      <TableCell>{account.name}</TableCell>
                    </TableRow>

                    <TableRow>
                      <TableHead>Scopes</TableHead>
                      <TableCell>
                        <div className="flex flex-wrap gap-1">
                          {account.scopes.map((scope) => (
                            <Badge key={scope} variant="secondary">
                              {scope}
                            </Badge>
                          ))}
                        </div>
                      </TableCell>
                    </TableRow>

                    <TableRow>
                      <TableHead>Status</TableHead>
                      <TableCell>{getStatusBadge(account)}</TableCell>
                    </TableRow>

                    <TableRow>
                      <TableHead>Expires At</TableHead>
                      <TableCell>{formatDate(account.expires_at)}</TableCell>
                    </TableRow>

                    <TableRow>
                      <TableHead>Created At</TableHead>
                      <TableCell>{formatDate(account.created_at)}</TableCell>
                    </TableRow>

                    {account.deleted_at && (
                      <TableRow>
                        <TableHead>Deleted At</TableHead>
                        <TableCell>{formatDate(account.deleted_at)}</TableCell>
                      </TableRow>
                    )}

                    <TableRow>
                      <TableHead>Issuer</TableHead>
                      <TableCell className="font-mono text-sm">
                        {account.iss}
                      </TableCell>
                    </TableRow>

                    <TableRow>
                      <TableHead>Subject</TableHead>
                      <TableCell className="font-mono text-sm">
                        {account.sub}
                      </TableCell>
                    </TableRow>
                  </TableBody>
                </Table>

                {!account.deleted_at && (
                  <div className="flex gap-2 pt-2">
                    {account.is_disabled ? (
                      <Button
                        onClick={() => handleUnpause(account)}
                        disabled={unpauseMutation.isPending}
                      >
                        <Play className="mr-2 h-4 w-4" />
                        Unpause
                      </Button>
                    ) : (
                      <Button
                        variant="outline"
                        onClick={() => handlePause(account)}
                        disabled={pauseMutation.isPending}
                      >
                        <Pause className="mr-2 h-4 w-4" />
                        Pause
                      </Button>
                    )}
                    <Button
                      variant="destructive"
                      onClick={() => openDeleteDialog(account)}
                      disabled={deleteMutation.isPending}
                    >
                      <Trash2 className="mr-2 h-4 w-4" />
                      Delete
                    </Button>
                  </div>
                )}
              </div>
            </AccordionContent>
          </AccordionItem>
        ))}
      </Accordion>

      {accounts.length === 0 && (
        <div className="text-center py-12 text-muted-foreground">
          <p className="mb-4">No service accounts found</p>
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            Create Your First Service Account
          </Button>
        </div>
      )}

      <CreateServiceAccountDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
      />

      <AlertDialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Service Account</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete "{selectedAccount?.name}"? This
              action cannot be undone and the service account will be
              permanently disabled.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDelete}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              Delete
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
