import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  useUnlockCertificate,
  useDownloadCertificate,
} from '@/api/Certificates';
import { useState, useEffect } from 'react';
import { Alert, AlertDescription } from '@/components/ui/alert';

interface DownloadCertificateDialogProps {
  certificateId: number;
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

export function DownloadCertificateDialog({
  certificateId,
  open,
  onOpenChange,
}: DownloadCertificateDialogProps) {
  const [passphrase, setPassphrase] = useState('');
  const [downloadToken, setDownloadToken] = useState<string | null>(null);
  const [expiresIn, setExpiresIn] = useState<number | null>(null);

  const unlockMutation = useUnlockCertificate();
  const downloadMutation = useDownloadCertificate();

  useEffect(() => {
    if (!open) {
      setPassphrase('');
      setDownloadToken(null);
      setExpiresIn(null);
      unlockMutation.reset();
      downloadMutation.reset();
    }
  }, [open]);

  const handleUnlock = async () => {
    try {
      const response = await unlockMutation.mutateAsync({
        id: certificateId,
        passphrase,
      });

      if (response.unlocked && response.download_token) {
        setDownloadToken(response.download_token);
        setExpiresIn(response.expires_in || null);
      }
    } catch (error) {
      console.error('Unlock failed:', error);
    }
  };

  const handleDownload = async () => {
    if (!downloadToken) return;

    try {
      const blob = await downloadMutation.mutateAsync({
        id: certificateId,
        token: downloadToken,
      });

      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `certificate-${certificateId}.p12`;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);

      onOpenChange(false);
    } catch (error) {
      console.error('Download failed:', error);
    }
  };

  const isUnlocking = unlockMutation.isPending;
  const isDownloading = downloadMutation.isPending;

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Download Certificate</DialogTitle>
          <DialogDescription>
            {!downloadToken
              ? 'Enter a passphrase to encrypt your certificate file (PKCS#12 format).'
              : 'Your certificate is ready to download.'}
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4">
          {!downloadToken ? (
            <>
              <div className="space-y-2">
                <Label htmlFor="passphrase">Passphrase</Label>
                <Input
                  id="passphrase"
                  type="password"
                  placeholder="Enter a strong passphrase"
                  value={passphrase}
                  onChange={(e) => setPassphrase(e.target.value)}
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && passphrase) {
                      handleUnlock();
                    }
                  }}
                  disabled={isUnlocking}
                />
                <p className="text-sm text-muted-foreground">
                  Minimum 12 characters. You'll need this to import the
                  certificate.
                </p>
              </div>

              {unlockMutation.isError && (
                <Alert variant="destructive">
                  <AlertDescription>
                    {unlockMutation.error?.message ||
                      'Failed to unlock certificate'}
                  </AlertDescription>
                </Alert>
              )}

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                  disabled={isUnlocking}
                >
                  Cancel
                </Button>
                <Button
                  onClick={handleUnlock}
                  disabled={!passphrase || isUnlocking}
                >
                  {isUnlocking ? 'Unlocking...' : 'Unlock Certificate'}
                </Button>
              </div>
            </>
          ) : (
            <>
              <Alert>
                <AlertDescription>
                  Certificate unlocked! The download token expires in{' '}
                  {expiresIn ? Math.floor(expiresIn / 60) : 5} minutes.
                </AlertDescription>
              </Alert>

              {downloadMutation.isError && (
                <Alert variant="destructive">
                  <AlertDescription>
                    {downloadMutation.error?.message ||
                      'Failed to download certificate'}
                  </AlertDescription>
                </Alert>
              )}

              <div className="flex justify-end gap-2">
                <Button
                  variant="outline"
                  onClick={() => onOpenChange(false)}
                  disabled={isDownloading}
                >
                  Cancel
                </Button>
                <Button onClick={handleDownload} disabled={isDownloading}>
                  {isDownloading ? 'Downloading...' : 'Download Certificate'}
                </Button>
              </div>
            </>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
