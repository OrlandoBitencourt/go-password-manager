export interface PasswordRecord {
  id: string;
  name: string;
  username: string;
  password: string;
  created_at: string;
  updated_at: string;
}

export interface Vault {
  name: string;
  isUnlocked: boolean;
}

export interface CreateVaultRequest {
  name: string;
  master_password: string;
}

export interface UnlockVaultRequest {
  name: string;
  master_password: string;
}

export interface AddRecordRequest {
  vault_name: string;
  name: string;
  username: string;
  password: string;
}

export interface UpdateRecordRequest {
  vault_name: string;
  name: string;
  username?: string;
  password?: string;
}

export interface APIResponse<T = any> {
  data?: T;
  error?: string;
  message?: string;
}
