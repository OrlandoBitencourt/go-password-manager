import axios from 'axios';
import type {
  PasswordRecord,
  CreateVaultRequest,
  UnlockVaultRequest,
  AddRecordRequest,
  UpdateRecordRequest,
} from '@/app/types';

const API_BASE = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const api = axios.create({
  baseURL: `${API_BASE}/api`,
  headers: {
    'Content-Type': 'application/json',
  },
});

export const vaultAPI = {
  list: async (): Promise<string[]> => {
    const response = await api.get('/vaults');
    return response.data.vaults || [];
  },

  create: async (data: CreateVaultRequest): Promise<void> => {
    await api.post('/vaults/create', data);
  },

  unlock: async (data: UnlockVaultRequest): Promise<void> => {
    await api.post('/vaults/unlock', data);
  },

  lock: async (name: string): Promise<void> => {
    await api.post('/vaults/lock', { name });
  },
};

export const recordAPI = {
  list: async (vaultName: string): Promise<PasswordRecord[]> => {
    const response = await api.get(`/records?vault_name=${encodeURIComponent(vaultName)}`);
    return response.data.records || [];
  },

  get: async (vaultName: string, name: string): Promise<PasswordRecord> => {
    const response = await api.get(
      `/records/get?vault_name=${encodeURIComponent(vaultName)}&name=${encodeURIComponent(name)}`
    );
    return response.data;
  },

  add: async (data: AddRecordRequest): Promise<void> => {
    await api.post('/records/add', data);
  },

  update: async (data: UpdateRecordRequest): Promise<void> => {
    await api.put('/records/update', data);
  },

  delete: async (vaultName: string, name: string): Promise<void> => {
    await api.delete('/records/delete', {
      data: { vault_name: vaultName, name },
    });
  },
};
