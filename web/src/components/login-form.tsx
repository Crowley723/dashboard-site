import { Button } from './ui/button';
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card.tsx';
import { cn } from '@/lib/utils.ts';
import React from 'react';

export function LoginForm({
  className,
  onLogin,
  isLoading,
  ...props
}: React.ComponentProps<'div'> & {
  onLogin?: () => void;
  isLoading?: boolean;
}) {
  return (
    <div className={cn('flex flex-col gap-6', className)} {...props}>
      <Card>
        <CardHeader>
          <CardTitle>Login to your account</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex flex-col gap-3">
            <Button
              type="button"
              variant="outline"
              className="w-full"
              onClick={onLogin}
              disabled={isLoading}
            >
              {isLoading ? 'Starting login...' : 'Login with Authelia'}
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  );
}
