import { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { FirewallAlias, FirewallIPWhitelistEntry } from '@/types/Firewall';

interface AddIPWhitelistDialogProps {
  aliases: FirewallAlias[];
  entries: FirewallIPWhitelistEntry[];
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: {
    alias_name: string;
    ip_address: string;
    description?: string;
    ttl?: string;
  }) => void;
  isLoading?: boolean;
  errorMessage?: string;
  successMessage?: string;
  children?: React.ReactNode;
}

export function AddIPWhitelistDialog({
  aliases,
  entries,
  open,
  onOpenChange,
  onSubmit,
  isLoading,
  errorMessage,
  successMessage,
  children,
}: AddIPWhitelistDialogProps) {
  const [aliasName, setAliasName] = useState('');
  const [ipAddress, setIPAddress] = useState('');
  const [description, setDescription] = useState('');
  const [useCustomTTL, setUseCustomTTL] = useState(false);
  const [ttl, setTTL] = useState('');
  const [errors, setErrors] = useState<{
    alias_name?: string;
    ip_address?: string;
    ttl?: string;
  }>({});

  const selectedAlias = aliases.find((a) => a.name === aliasName);

  // Count active IPs for the selected alias
  const activeIPCount = selectedAlias
    ? entries.filter(
        (entry) =>
          entry.alias_name === selectedAlias.name &&
          (entry.status === 'requested' || entry.status === 'added')
      ).length
    : 0;

  const validateIP = (ip: string): boolean => {
    const ipv4Pattern = /^(\d{1,3}\.){3}\d{1,3}$/;
    const ipv6Pattern =
      /^(([0-9a-fA-F]{1,4}:){7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$/;

    return ipv4Pattern.test(ip) || ipv6Pattern.test(ip);
  };

  const validateForm = (): boolean => {
    const newErrors: {
      alias_name?: string;
      ip_address?: string;
      ttl?: string;
    } = {};

    if (!aliasName) {
      newErrors.alias_name = 'Alias is required';
    }

    if (!ipAddress.trim()) {
      newErrors.ip_address = 'IP address is required';
    } else if (!validateIP(ipAddress.trim())) {
      newErrors.ip_address = 'Invalid IP address format';
    }

    if (useCustomTTL && ttl && !/^\d+[hdm]$/.test(ttl)) {
      newErrors.ttl = 'TTL must be in format like 24h, 7d, or 30m';
    }

    setErrors(newErrors);
    return Object.keys(newErrors).length === 0;
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (validateForm()) {
      onSubmit({
        alias_name: aliasName,
        ip_address: ipAddress.trim(),
        description: description.trim() || undefined,
        ttl: useCustomTTL && ttl ? ttl : undefined,
      });
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      // Reset form when closing
      setAliasName('');
      setIPAddress('');
      setDescription('');
      setUseCustomTTL(false);
      setTTL('');
      setErrors({});
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="max-w-2xl max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>Add IP to Whitelist</DialogTitle>
          <DialogDescription>
            Add an IP address to the firewall whitelist for secure access
          </DialogDescription>
        </DialogHeader>

        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
          {errorMessage && (
            <div className="text-destructive text-sm p-3 rounded-md bg-destructive/10 border border-destructive/20">
              {errorMessage}
            </div>
          )}
          {successMessage && (
            <div className="text-green-600 dark:text-green-400 text-sm p-3 rounded-md bg-green-500/10 border border-green-500/20">
              {successMessage}
            </div>
          )}

          <div className="flex flex-col gap-2">
            <Label htmlFor="alias">
              Firewall Alias
              <span className="text-destructive ml-1">*</span>
            </Label>
            <Select
              value={aliasName}
              onValueChange={setAliasName}
              disabled={isLoading}
            >
              <SelectTrigger id="alias" aria-invalid={!!errors.alias_name}>
                <SelectValue placeholder="Select an alias" />
              </SelectTrigger>
              <SelectContent>
                {aliases.map((alias) => (
                  <SelectItem key={alias.name} value={alias.name}>
                    {alias.name} - {alias.description}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.alias_name && (
              <span className="text-destructive text-xs">
                {errors.alias_name}
              </span>
            )}
            {selectedAlias && (
              <span className="text-muted-foreground text-xs">
                {activeIPCount}/{selectedAlias.max_ips_per_user} IPs used
                {selectedAlias.default_ttl &&
                  ` • Default expiration: ${selectedAlias.default_ttl}`}
              </span>
            )}
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="ip_address">
              IP Address
              <span className="text-destructive ml-1">*</span>
            </Label>
            <Input
              id="ip_address"
              placeholder="192.168.1.1 or 2001:db8::1"
              value={ipAddress}
              onChange={(e) => setIPAddress(e.target.value)}
              disabled={isLoading}
              aria-invalid={!!errors.ip_address}
            />
            {errors.ip_address && (
              <span className="text-destructive text-xs">
                {errors.ip_address}
              </span>
            )}
            <span className="text-muted-foreground text-xs">
              Enter an IPv4 or IPv6 address
            </span>
          </div>

          <div className="flex flex-col gap-2">
            <Label htmlFor="description">Description</Label>
            <Textarea
              id="description"
              placeholder="Home network, office VPN, etc..."
              value={description}
              onChange={(e) => setDescription(e.target.value)}
              disabled={isLoading}
              rows={3}
            />
            <span className="text-muted-foreground text-xs">
              Optional description to help identify this IP
            </span>
          </div>

          <div className="flex flex-col gap-2">
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="use_custom_ttl"
                checked={useCustomTTL}
                onChange={(e) => setUseCustomTTL(e.target.checked)}
                disabled={isLoading}
                className="h-4 w-4"
              />
              <Label htmlFor="use_custom_ttl" className="cursor-pointer">
                Use custom expiration time
              </Label>
            </div>

            {useCustomTTL && (
              <>
                <Input
                  id="ttl"
                  placeholder="24h, 7d, 30d"
                  value={ttl}
                  onChange={(e) => setTTL(e.target.value)}
                  disabled={isLoading}
                  aria-invalid={!!errors.ttl}
                />
                {errors.ttl && (
                  <span className="text-destructive text-xs">{errors.ttl}</span>
                )}
                <span className="text-muted-foreground text-xs">
                  Format: number + unit (h=hours, d=days, m=minutes). Example:
                  24h, 7d, 30d
                </span>
              </>
            )}
          </div>

          <div className="flex justify-end gap-2 pt-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => handleOpenChange(false)}
              disabled={isLoading}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={isLoading}>
              {isLoading ? 'Adding IP...' : 'Add IP to Whitelist'}
            </Button>
          </div>
        </form>
      </DialogContent>
    </Dialog>
  );
}
