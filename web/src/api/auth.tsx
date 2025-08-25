export interface User {
  sub: string;
  iss: string;
  username: string;
  display_name: string;
  email: string;
}

export interface LoginResponse {
  status: 'already_authenticated' | 'redirect_required';
  redirect_url?: string;
}

export interface AuthResponse {
  authenticated: boolean;
  user: User;
}

export interface ApiError {
  error: string;
  message?: string;
}

export const authApi = {
  login: async (): Promise<LoginResponse> => {
    const response = await fetch('/api/auth/login', {
      method: 'GET',
      credentials: 'include',
    });

    if (!response.ok) {
      const error: ApiError = await response.json();
      throw new Error(error.message || 'Login failed');
    }

    return response.json();
  },

  logout: async (): Promise<void> => {
    const response = await fetch('/api/auth/logout', {
      method: 'POST',
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error('Logout failed');
    }
  },

  getCurrentUser: async (): Promise<AuthResponse | null> => {
    const response = await fetch('/api/auth/status', {
      credentials: 'include',
    });

    if (!response.ok) {
      if (response.status === 401) {
        return null;
      }
      throw new Error('Failed to get user');
    }

    return response.json();
  },
};
