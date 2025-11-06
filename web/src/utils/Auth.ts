import { redirect } from '@tanstack/react-router';
import { authApi } from '@/api/auth';

export async function requireAuth(
  location: { pathname: string },
  customMessage?: string,
  isError?: boolean
) {
  try {
    const authResponse = await authApi.getCurrentUser();
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
  } catch (error) {
    throw redirect({
      to: '/',
      search: {
        showLogin: true,
        message: customMessage || 'You must login to access that page',
        rd: location.pathname,
        error: true,
      },
    });
  }
}
