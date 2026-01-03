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
  withCredentials: true, // Important for CSRF cookies
});

// Request interceptor to add CSRF token to all non-GET requests
api.interceptors.request.use(
  async (config) => {
    // Skip for GET, HEAD, OPTIONS requests
    if (
      config.method?.toUpperCase() === 'GET' ||
      config.method?.toUpperCase() === 'HEAD' ||
      config.method?.toUpperCase() === 'OPTIONS'
    ) {
      return config;
    }

    // Get CSRF token from cookie
    const csrfToken = document.cookie
      .split('; ')
      .find(row => row.startsWith('csrf_token='))
      ?.split('=')[1];

    if (csrfToken) {
      config.headers['X-CSRF-Token'] = csrfToken;
    } else {
      // If no token exists, fetch one first
      try {
        const response = await axios.get(`${API_BASE}/api/csrf-token`, {
          withCredentials: true,
        });

        const newToken = response.data.token;
        if (newToken) {
          config.headers['X-CSRF-Token'] = newToken;
        }
      } catch (error) {
        console.error('Failed to fetch CSRF token:', error);
      }
    }

    return config;
  },
  (error) => Promise.reject(error)
);

// Response interceptor to handle CSRF token errors
api.interceptors.response.use(
  (response) => response,
  async (error) => {
    // If we get a 403 CSRF error, try to refresh the token and retry
    if (error.response?.status === 403 && error.response?.data?.error?.includes('CSRF')) {
      try {
        // Fetch a new CSRF token
        await axios.get(`${API_BASE}/api/csrf-token`, {
          withCredentials: true,
        });

        // Retry the original request
        return api.request(error.config);
      } catch (retryError) {
        return Promise.reject(retryError);
      }
    }

    return Promise.reject(error);
  }
);

/**
 * Initialize CSRF protection by fetching a token on app startup
 * Call this in your root layout or app initialization
 */
export async function initializeCSRF() {
  try {
    await axios.get(`${API_BASE}/api/csrf-token`, {
      withCredentials: true,
    });
    console.log('CSRF token initialized');
  } catch (error) {
    console.error('Failed to initialize CSRF token:', error);
  }
}

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
