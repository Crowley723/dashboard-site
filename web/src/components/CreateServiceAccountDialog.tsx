import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { Copy, Check, AlertCircle } from 'lucide-react';
import { useCreateServiceAccount, useUserScopes } from '@/api/ServiceAccounts';

interface CreateServiceAccountDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function CreateServiceAccountDialog({
  open,
  onOpenChange,
}: CreateServiceAccountDialogProps) {
  const [name, setName] = useState('');
  const [expiryDays, setExpiryDays] = useState('365');
  const [selectedScopes, setSelectedScopes] = useState<string[]>([]);
  const [createdToken, setCreatedToken] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const { data: userScopes, isLoading: scopesLoading } = useUserScopes();
  const createMutation = useCreateServiceAccount();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!name.trim() || selectedScopes.length === 0) {
      return;
    }
    const days = Number.parseInt(expiryDays, 10);
    if (!Number.isFinite(days) || days < 1) {
      return;
    }
    const expiresAt = new Date();
    expiresAt.setDate(expiresAt.getDate() + days);
    try {
      const result = await createMutation.mutateAsync({
        name: name.trim(),
        token_expires_at: expiresAt.toISOString(),
        scopes: selectedScopes,
      });
      setCreatedToken(result.token || null);
    } catch (error) {
      console.error('Failed to create service account:', error);
    }
  };

  const handleCopy = async () => {
    if (!createdToken) return;
    try {
      if (!navigator.clipboard) {
        console.error('Clipboard API unavailable');
        return;
      }
      await navigator.clipboard.writeText(createdToken);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (error) {
      console.error('Failed to copy token:', error);
    }
  };

  const handleClose = () => {
    setName('');
    setExpiryDays('365');
    setSelectedScopes([]);
    setCreatedToken(null);
    setCopied(false);
    createMutation.reset();
    onOpenChange(false);
  };

  const toggleScope = (scope: string) => {
    setSelectedScopes((prev) =>
      prev.includes(scope) ? prev.filter((s) => s !== scope) : [...prev, scope]
    );
  };

  // Show token display if service account was created
  if (createdToken) {
    return (
      <Dialog open={open} onOpenChange={handleClose}>
        <DialogContent className="sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>Service Account Created</DialogTitle>
            <DialogDescription>
              Save this token now - it won't be shown again!
            </DialogDescription>
          </DialogHeader>

          <Alert variant="destructive">
            <AlertCircle className="h-4 w-4" />
            <AlertDescription>
              This is the only time you'll see this token. Copy it now and store
              it securely.
            </AlertDescription>
          </Alert>

          <div className="space-y-4">
            <div className="space-y-2">
              <Label>API Token</Label>
              <div className="flex gap-2">
                <Input
                  value={createdToken}
                  readOnly
                  className="font-mono text-sm"
                />
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  onClick={handleCopy}
                >
                  {copied ? (
                    <Check className="h-4 w-4" />
                  ) : (
                    <Copy className="h-4 w-4" />
                  )}
                </Button>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button onClick={handleClose}>Done</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  // Show creation form
  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[600px]">
        <form onSubmit={handleSubmit}>
          <DialogHeader>
            <DialogTitle>Create Service Account</DialogTitle>
            <DialogDescription>
              Create a service account with specific scopes for API access.
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4 py-4">
            <div className="space-y-2">
              <Label htmlFor="name">Name</Label>
              <Input
                id="name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="My API Service"
                required
                maxLength={255}
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="expiry">Token Expiry (days)</Label>
              <Input
                id="expiry"
                type="number"
                value={expiryDays}
                onChange={(e) => setExpiryDays(e.target.value)}
                min="1"
                max="3650"
                required
              />
            </div>

            <div className="space-y-2">
              <Label>Scopes</Label>
              <p className="text-sm text-muted-foreground">
                Select the permissions for this service account. You can only
                grant scopes that you have.
              </p>
              {scopesLoading ? (
                <div className="text-sm text-muted-foreground">
                  Loading available scopes...
                </div>
              ) : userScopes?.scopes && userScopes.scopes.length > 0 ? (
                <div className="space-y-2 border rounded-md p-4 max-h-[200px] overflow-y-auto">
                  {userScopes.scopes.map((scope) => (
                    <div key={scope} className="flex items-center space-x-2">
                      <Checkbox
                        id={`scope-${scope}`}
                        checked={selectedScopes.includes(scope)}
                        onCheckedChange={() => toggleScope(scope)}
                      />
                      <Label
                        htmlFor={`scope-${scope}`}
                        className="text-sm font-normal cursor-pointer"
                      >
                        {scope}
                      </Label>
                    </div>
                  ))}
                </div>
              ) : (
                <Alert>
                  <AlertDescription>
                    You don't have any scopes available to grant.
                  </AlertDescription>
                </Alert>
              )}
            </div>

            {createMutation.isError && (
              <Alert variant="destructive">
                <AlertCircle className="h-4 w-4" />
                <AlertDescription>
                  {createMutation.error?.message ||
                    'Failed to create service account'}
                </AlertDescription>
              </Alert>
            )}
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={handleClose}
              disabled={createMutation.isPending}
            >
              Cancel
            </Button>
            <Button
              type="submit"
              disabled={
                !name.trim() ||
                selectedScopes.length === 0 ||
                createMutation.isPending
              }
            >
              {createMutation.isPending ? 'Creating...' : 'Create'}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
