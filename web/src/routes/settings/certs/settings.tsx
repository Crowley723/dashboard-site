import { createFileRoute } from '@tanstack/react-router';
import { requireAuth } from '@/utils/Auth.ts';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Badge } from '@/components/ui/badge';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';

export const Route = createFileRoute('/settings/certs/settings')({
  component: RouteComponent,
  beforeLoad: async ({ location }) => {
    await requireAuth(
      location,
      'You must login to access the settings page.',
      true
    );
  },
});

interface CertificateSettings {
  defaultOrganization: string;
  defaultOrganizationalUnits: string[];
  issuerName: string;
  defaultValidityDays: number;
  allowedDnsPatterns: string[];
  allowedCommonNamePatterns: string[];
}

// Mock initial settings - replace with actual API call
const mockInitialSettings: CertificateSettings = {
  defaultOrganization: 'Crowley Labs',
  defaultOrganizationalUnits: ['Engineering', 'Security'],
  issuerName: 'Crowley Labs Internal CA',
  defaultValidityDays: 365,
  allowedDnsPatterns: [
    '*.example.com',
    '*.crowley.example.com',
    '*.internal.example.com',
  ],
  allowedCommonNamePatterns: ['*.example.com', '*.crowley.example.com'],
};

function RouteComponent() {
  const [settings, setSettings] =
    useState<CertificateSettings>(mockInitialSettings);
  const [isSaving, setIsSaving] = useState(false);

  // Temporary input states for adding new items
  const [newOU, setNewOU] = useState('');
  const [newDnsPattern, setNewDnsPattern] = useState('');
  const [newCNPattern, setNewCNPattern] = useState('');

  const handleSave = async () => {
    setIsSaving(true);
    // TODO: Implement actual save API call
    console.log('Saving settings:', settings);
    setTimeout(() => {
      setIsSaving(false);
      alert('Settings saved successfully!');
    }, 1000);
  };

  const handleAddOU = () => {
    if (
      newOU.trim() &&
      !settings.defaultOrganizationalUnits.includes(newOU.trim())
    ) {
      setSettings({
        ...settings,
        defaultOrganizationalUnits: [
          ...settings.defaultOrganizationalUnits,
          newOU.trim(),
        ],
      });
      setNewOU('');
    }
  };

  const handleRemoveOU = (ou: string) => {
    setSettings({
      ...settings,
      defaultOrganizationalUnits: settings.defaultOrganizationalUnits.filter(
        (u) => u !== ou
      ),
    });
  };

  const handleAddDnsPattern = () => {
    if (
      newDnsPattern.trim() &&
      !settings.allowedDnsPatterns.includes(newDnsPattern.trim())
    ) {
      setSettings({
        ...settings,
        allowedDnsPatterns: [
          ...settings.allowedDnsPatterns,
          newDnsPattern.trim(),
        ],
      });
      setNewDnsPattern('');
    }
  };

  const handleRemoveDnsPattern = (pattern: string) => {
    setSettings({
      ...settings,
      allowedDnsPatterns: settings.allowedDnsPatterns.filter(
        (p) => p !== pattern
      ),
    });
  };

  const handleAddCNPattern = () => {
    if (
      newCNPattern.trim() &&
      !settings.allowedCommonNamePatterns.includes(newCNPattern.trim())
    ) {
      setSettings({
        ...settings,
        allowedCommonNamePatterns: [
          ...settings.allowedCommonNamePatterns,
          newCNPattern.trim(),
        ],
      });
      setNewCNPattern('');
    }
  };

  const handleRemoveCNPattern = (pattern: string) => {
    setSettings({
      ...settings,
      allowedCommonNamePatterns: settings.allowedCommonNamePatterns.filter(
        (p) => p !== pattern
      ),
    });
  };

  return (
    <div className="container mx-auto p-6 max-w-4xl">
      <div className="mb-6">
        <h1 className="text-3xl font-bold mb-2">Certificate Settings</h1>
        <p className="text-muted-foreground">
          Configure default values and policies for certificate issuance
        </p>
      </div>

      <div className="space-y-6">
        <Card>
          <CardHeader>
            <CardTitle>Default Values</CardTitle>
            <CardDescription>
              Default values applied to new certificate requests
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="organization">Default Organization</Label>
              <Input
                id="organization"
                value={settings.defaultOrganization}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    defaultOrganization: e.target.value,
                  })
                }
                placeholder="e.g., Crowley Labs"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="issuer">Issuer Name</Label>
              <Input
                id="issuer"
                value={settings.issuerName}
                onChange={(e) =>
                  setSettings({ ...settings, issuerName: e.target.value })
                }
                placeholder="e.g., Crowley Labs Internal CA"
              />
            </div>

            <div className="space-y-2">
              <Label htmlFor="validity">Default Validity Period (days)</Label>
              <Input
                id="validity"
                type="number"
                min="1"
                max="3650"
                value={settings.defaultValidityDays}
                onChange={(e) =>
                  setSettings({
                    ...settings,
                    defaultValidityDays: parseInt(e.target.value) || 365,
                  })
                }
              />
              <p className="text-sm text-muted-foreground">
                Current: {settings.defaultValidityDays} days (~
                {Math.round(settings.defaultValidityDays / 30)} months)
              </p>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Default Organizational Units</CardTitle>
            <CardDescription>
              Organizational units that will be included in certificates by
              default
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              {settings.defaultOrganizationalUnits.map((ou) => (
                <Badge key={ou} variant="secondary" className="text-sm">
                  {ou}
                  <button
                    onClick={() => handleRemoveOU(ou)}
                    className="ml-2 hover:text-destructive"
                  >
                    ×
                  </button>
                </Badge>
              ))}
              {settings.defaultOrganizationalUnits.length === 0 && (
                <span className="text-sm text-muted-foreground">
                  No organizational units defined
                </span>
              )}
            </div>

            <div className="flex gap-2">
              <Input
                placeholder="Add organizational unit..."
                value={newOU}
                onChange={(e) => setNewOU(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddOU()}
              />
              <Button onClick={handleAddOU} variant="outline">
                Add
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Allowed DNS Name Patterns</CardTitle>
            <CardDescription>
              Permitted DNS name patterns for certificate requests. Use
              wildcards (*) for pattern matching.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              {settings.allowedDnsPatterns.map((pattern) => (
                <Badge
                  key={pattern}
                  variant="outline"
                  className="text-sm font-mono"
                >
                  {pattern}
                  <button
                    onClick={() => handleRemoveDnsPattern(pattern)}
                    className="ml-2 hover:text-destructive"
                  >
                    ×
                  </button>
                </Badge>
              ))}
              {settings.allowedDnsPatterns.length === 0 && (
                <span className="text-sm text-muted-foreground">
                  No DNS patterns defined - all patterns will be allowed
                </span>
              )}
            </div>

            <div className="flex gap-2">
              <Input
                placeholder="Add DNS pattern (e.g., *.example.com)..."
                value={newDnsPattern}
                onChange={(e) => setNewDnsPattern(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddDnsPattern()}
                className="font-mono"
              />
              <Button onClick={handleAddDnsPattern} variant="outline">
                Add
              </Button>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Allowed Common Name Patterns</CardTitle>
            <CardDescription>
              Permitted common name patterns for certificate requests. Use
              wildcards (*) for pattern matching.
            </CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex flex-wrap gap-2">
              {settings.allowedCommonNamePatterns.map((pattern) => (
                <Badge
                  key={pattern}
                  variant="outline"
                  className="text-sm font-mono"
                >
                  {pattern}
                  <button
                    onClick={() => handleRemoveCNPattern(pattern)}
                    className="ml-2 hover:text-destructive"
                  >
                    ×
                  </button>
                </Badge>
              ))}
              {settings.allowedCommonNamePatterns.length === 0 && (
                <span className="text-sm text-muted-foreground">
                  No common name patterns defined - all patterns will be allowed
                </span>
              )}
            </div>

            <div className="flex gap-2">
              <Input
                placeholder="Add CN pattern (e.g., *.example.com)..."
                value={newCNPattern}
                onChange={(e) => setNewCNPattern(e.target.value)}
                onKeyDown={(e) => e.key === 'Enter' && handleAddCNPattern()}
                className="font-mono"
              />
              <Button onClick={handleAddCNPattern} variant="outline">
                Add
              </Button>
            </div>
          </CardContent>
        </Card>

        {/* Save Button */}
        <div className="flex justify-end gap-2 pt-4">
          <Button
            variant="outline"
            onClick={() => setSettings(mockInitialSettings)}
          >
            Reset to Defaults
          </Button>
          <Button onClick={handleSave} disabled={isSaving}>
            {isSaving ? 'Saving...' : 'Save Settings'}
          </Button>
        </div>
      </div>
    </div>
  );
}
