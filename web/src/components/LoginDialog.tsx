import { Dialog, DialogContent, DialogTrigger } from '@/components/ui/dialog';
import { LoginForm } from '@/components/LoginForm';

interface LoginDialogProps {
  children: React.ReactNode;
  login: (redirectTo?: string) => void;
  isLoggingIn: boolean;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  message?: string;
  error?: boolean;
  redirectTo?: string;
}

export function LoginDialog({
  children,
  login,
  isLoggingIn,
  open,
  onOpenChange,
  message,
  error,
  redirectTo,
}: LoginDialogProps) {
  const handleLogin = () => {
    login(redirectTo);
    onOpenChange?.(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent
        showCloseButton={false}
        className="sm:max-w-md p-0 bg-transparent border-0 shadow-none"
      >
        <div className="flex w-full items-center justify-center md:p-10 bg-transparent">
          <div className="w-full max-w-sm">
            <LoginForm
              onLogin={handleLogin}
              isLoading={isLoggingIn}
              message={message}
              error={error}
            />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
