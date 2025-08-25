import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { authApi, type LoginResponse, type AuthResponse } from '@/api/auth';

export const useAuth = () => {
  const queryClient = useQueryClient();

  const { data: authResponse, isLoading } = useQuery<AuthResponse | null>({
    queryKey: ['auth', 'user'],
    queryFn: authApi.getCurrentUser,
    retry: false,
  });

  // Extract the user and auth status
  const user = authResponse?.user || null;
  const isAuthenticated = authResponse?.authenticated || false;

  const loginMutation = useMutation<LoginResponse, Error>({
    mutationFn: authApi.login,
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

  return {
    user,
    isLoading,
    isAuthenticated,
    login: loginMutation.mutate,
    logout: logoutMutation.mutate,
    isLoggingIn: loginMutation.isPending,
    isLoggingOut: logoutMutation.isPending,
  } as const;
};
