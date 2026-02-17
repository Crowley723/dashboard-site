import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState } from 'react';
import { AddIPWhitelistForm } from '@/components/AddIPWhitelistForm';
import { useAvailableAliases, useAddIPWhitelistEntry } from '@/api/Firewall';

export const Route = createFileRoute('/settings/firewall/add')({
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
  const navigate = useNavigate();
  const { data: aliases, isLoading: aliasesLoading } = useAvailableAliases();
  const addIPMutation = useAddIPWhitelistEntry();
  const [errorMessage, setErrorMessage] = useState<string>('');
  const [successMessage, setSuccessMessage] = useState<string>('');

  const handleSubmit = async (data: {
    alias_name: string;
    ip_address: string;
    description?: string;
    ttl?: string;
  }) => {
    setErrorMessage('');
    setSuccessMessage('');

    try {
      const result = await addIPMutation.mutateAsync(data);
      setSuccessMessage(result.message || 'IP address added successfully!');

      // Clear form and redirect after a short delay
      setTimeout(() => {
        navigate({ to: '/settings/firewall' });
      }, 2000);
    } catch (error) {
      setErrorMessage(
        error instanceof Error ? error.message : 'Failed to add IP address'
      );
    }
  };

  if (aliasesLoading) {
    return (
      <div className="container mx-auto p-6 max-w-2xl">
        <div className="text-center py-12">Loading...</div>
      </div>
    );
  }

  if (!aliases || aliases.length === 0) {
    return (
      <div className="container mx-auto p-6 max-w-2xl">
        <div className="text-center py-12 text-muted-foreground">
          No firewall aliases available. Contact your administrator.
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-6 max-w-2xl">
      <AddIPWhitelistForm
        aliases={aliases}
        onSubmit={handleSubmit}
        isLoading={addIPMutation.isPending}
        errorMessage={errorMessage}
        successMessage={successMessage}
      />
    </div>
  );
}
