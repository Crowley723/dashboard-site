import { redirect } from '@tanstack/react-router';
import { authApi } from '@/api/auth';

export async function requireAuth(
  location: { pathname: string },
  customMessage?: string,
  isError?: boolean
) {
  // Use .catch() so network errors return null without swallowing the redirect throw
  const authResponse = await authApi.getCurrentUser().catch(() => null);
  if (!authResponse?.authenticated) {
    throw redirect({
      to: '/',
      search: {
        showLogin: true,
        message: customMessage || 'You must login to access that page',
        rd: location.pathname,
        error: isError || false,
      },
    });
  }
}
