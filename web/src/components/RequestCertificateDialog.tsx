import { Dialog, DialogContent, DialogTrigger } from '@/components/ui/dialog';
import { RequestCertificateForm } from '@/components/RequestCertificateForm.tsx';
import { useCreateCertificateRequest } from '@/api/Certificates.tsx';
import React from 'react';

interface RequestCertificateDialogProps {
  children: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  onSuccess?: () => void;
}

export function RequestCertificateDialog({
  children,
  open,
  onOpenChange,
  onSuccess,
}: RequestCertificateDialogProps) {
  const [successMessage, setSuccessMessage] = React.useState<string | null>(
    null
  );

  const createMutation = useCreateCertificateRequest();

  const handleSubmit = async (data: {
    message: string;
    validity_days: number;
  }) => {
    setSuccessMessage(null);
    try {
      await createMutation.mutateAsync(data);
      setSuccessMessage(
        'Certificate request submitted successfully! You will be notified once it is reviewed.'
      );
      onSuccess?.();
      // Auto-close the dialog after 2 seconds on success
      setTimeout(() => {
        onOpenChange?.(false);
        setSuccessMessage(null);
      }, 2000);
    } catch (error) {
      // Error is handled by the mutation's error state
    }
  };

  // Reset success message when dialog closes
  React.useEffect(() => {
    if (!open) {
      setSuccessMessage(null);
      createMutation.reset();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open]);

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent className="sm:max-w-md p-0 bg-transparent border-0 shadow-none">
        <div className="flex w-full items-center justify-center md:p-10 bg-transparent">
          <div className="w-full max-w-sm">
            <RequestCertificateForm
              onSubmit={handleSubmit}
              isLoading={createMutation.isPending}
              errorMessage={
                createMutation.error
                  ? createMutation.error.message
                  : undefined
              }
              successMessage={successMessage || undefined}
            />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
