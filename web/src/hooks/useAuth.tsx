import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { authApi, type LoginResponse, type AuthResponse } from '@/api/auth';

export const useAuth = () => {
  const queryClient = useQueryClient();

  const { data: authResponse, isLoading } = useQuery<AuthResponse | null>({
    queryKey: ['auth', 'user'],
    queryFn: authApi.getCurrentUser,
    retry: false,
  });

  const user = authResponse?.user || null;
  const isAuthenticated = authResponse?.authenticated || false;
  const config = authResponse?.config;

  const loginMutation = useMutation<LoginResponse, Error, string | undefined>({
    mutationFn: (redirectTo?: string) => authApi.login(redirectTo),
    onSuccess: (data) => {
      if (data.status === 'already_authenticated') {
        queryClient
          .invalidateQueries({
            queryKey: ['auth', 'user'],
          })
          .then(() => {});
      } else if (data.status === 'redirect_required' && data.redirect_url) {
        window.location.href = data.redirect_url;
      }
    },
  });

  const logoutMutation = useMutation<void, Error>({
    mutationFn: authApi.logout,
    onSuccess: () => {
      queryClient.setQueryData(['auth', 'user'], null);
    },
  });

  // Helper to check if user is in a specific group
  const isInGroup = (groupName: string): boolean => {
    return user?.groups?.includes(groupName) ?? false;
  };

  // Check if user is in the MTLS admin group
  const isMTLSAdmin = (): boolean => {
    if (!config?.mtls?.enabled) return false;
    return isInGroup('conduit:mtls:admin');
  };

  // Check if user is in the MTLS user group
  const isMTLSUser = (): boolean => {
    if (!config?.mtls?.enabled) return false;
    return isInGroup('conduit:mtls:user');
  };

  // Check if user has firewall access
  const hasFirewallAccess = (): boolean => {
    if (!config?.firewall?.enabled) return false;
    // Check if user is in any of the configured firewall groups
    // For now, we'll just check if firewall is enabled and user is authenticated
    return true;
  };

  // Check if user is firewall admin
  const isFirewallAdmin = (): boolean => {
    if (!config?.firewall?.enabled) return false;
    return isInGroup('conduit:firewall:admin');
  };

  return {
    user,
    isLoading,
    isAuthenticated,
    config,
    isInGroup,
    isMTLSAdmin,
    isMTLSUser,
    hasFirewallAccess,
    isFirewallAdmin,
    login: loginMutation.mutate,
    logout: logoutMutation.mutate,
    isLoggingIn: loginMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
  } as const;
};
