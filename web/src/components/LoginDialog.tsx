import { Dialog, DialogContent, DialogTrigger } from '@/components/ui/dialog';
import { LoginForm } from '@/components/login-form';
import { useState } from 'react';

interface LoginDialogProps {
  children: React.ReactNode;
  login: () => void;
  isLoggingIn: boolean;
}

export function LoginDialog({
  children,
  login,
  isLoggingIn,
}: LoginDialogProps) {
  const [open, setOpen] = useState(false);

  const handleLogin = () => {
    login();
    setOpen(false);
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>{children}</DialogTrigger>
      <DialogContent
        showCloseButton={false}
        className="sm:max-w-md p-0 bg-transparent border-0 shadow-none"
      >
        <div className="flex w-full items-center justify-center md:p-10 bg-transparent">
          <div className="w-full max-w-sm">
            <LoginForm onLogin={handleLogin} isLoading={isLoggingIn} />
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
