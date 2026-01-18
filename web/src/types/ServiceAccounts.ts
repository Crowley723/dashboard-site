export interface ServiceAccount {
  sub: string;
  iss: string;
  name: string;
  token?: string;
  expires_at: string;
  is_disabled: boolean;
  deleted_at?: string | null;
  scopes: string[];
  created_by_iss: string;
  created_by_sub: string;
  created_at: string;
}

export interface CreateServiceAccountInput {
  name: string;
  token_expires_at: string;
  scopes: string[];
}

export interface UserScopes {
  scopes: string[];
}
