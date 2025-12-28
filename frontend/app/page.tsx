'use client';

import { useState, useEffect } from 'react';
import { Lock, Plus, LogOut, Search, Key, Shield, Moon, Sun } from 'lucide-react';
import { Button } from './components/Button';
import { Card, CardContent, CardHeader, CardTitle } from './components/Card';
import { Modal } from './components/Modal';
import { Input } from './components/Input';
import { PasswordInput } from './components/PasswordInput';
import { vaultAPI, recordAPI } from './lib/api';
import type { PasswordRecord } from './types';
import { VaultSelector } from './components/VaultSelector';
import { PasswordRecordCard } from './components/PasswordRecordCard';
import { AddPasswordModal } from './components/AddPasswordModal';
import { PasswordGeneratorModal } from './components/PasswordGeneratorModal';

export default function Home() {
  const [vaults, setVaults] = useState<string[]>([]);
  const [currentVault, setCurrentVault] = useState<string | null>(null);
  const [records, setRecords] = useState<PasswordRecord[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [darkMode, setDarkMode] = useState(false);

  // Modals
  const [showCreateVault, setShowCreateVault] = useState(false);
  const [showUnlockVault, setShowUnlockVault] = useState(false);
  const [showAddPassword, setShowAddPassword] = useState(false);
  const [showPasswordGenerator, setShowPasswordGenerator] = useState(false);

  // Load vaults on mount
  useEffect(() => {
    loadVaults();
    const isDark = localStorage.getItem('darkMode') === 'true';
    setDarkMode(isDark);
    if (isDark) {
      document.documentElement.classList.add('dark');
    }
  }, []);

  // Load records when vault changes
  useEffect(() => {
    if (currentVault) {
      loadRecords();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentVault]);

  // Toggle dark mode
  const toggleDarkMode = () => {
    const newDarkMode = !darkMode;
    setDarkMode(newDarkMode);
    localStorage.setItem('darkMode', String(newDarkMode));
    if (newDarkMode) {
      document.documentElement.classList.add('dark');
    } else {
      document.documentElement.classList.remove('dark');
    }
  };

  const loadVaults = async () => {
    try {
      const vaultList = await vaultAPI.list();
      setVaults(vaultList);
    } catch (err: any) {
      setError('Failed to load vaults');
    }
  };

  const loadRecords = async () => {
    if (!currentVault) return;
    setIsLoading(true);
    setError(null);
    try {
      const recordList = await recordAPI.list(currentVault);
      setRecords(recordList);
    } catch (err: any) {
      setError('Failed to load password records');
      setRecords([]);
    } finally {
      setIsLoading(false);
    }
  };

  const handleUnlockVault = async (vaultName: string, masterPassword: string) => {
    setIsLoading(true);
    setError(null);
    try {
      await vaultAPI.unlock({ name: vaultName, master_password: masterPassword });
      setCurrentVault(vaultName);
      setShowUnlockVault(false);
      await loadRecords();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to unlock vault');
      throw err;
    } finally {
      setIsLoading(false);
    }
  };

  const handleLockVault = async () => {
    if (!currentVault) return;
    try {
      await vaultAPI.lock(currentVault);
      setCurrentVault(null);
      setRecords([]);
    } catch (err: any) {
      setError('Failed to lock vault');
    }
  };

  const filteredRecords = records.filter(record =>
    record.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
    record.username.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="min-h-screen bg-gradient-to-br from-primary-50 via-background to-secondary dark:from-gray-900 dark:via-background dark:to-gray-800">
      {/* Header */}
      <header className="border-b border-border bg-card/50 backdrop-blur-sm sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-4">
          <div className="flex items-center justify-between">
            <div className="flex items-center gap-3">
              <div className="p-2 bg-primary-600 rounded-lg">
                <Shield className="text-white" size={24} />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-foreground">Password Manager</h1>
                <p className="text-sm text-muted-foreground">Secure your digital life</p>
              </div>
            </div>
            <div className="flex items-center gap-2">
              <button
                onClick={toggleDarkMode}
                className="p-2 rounded-lg hover:bg-secondary transition-colors"
                title={darkMode ? 'Light mode' : 'Dark mode'}
              >
                {darkMode ? <Sun size={20} /> : <Moon size={20} />}
              </button>
              {currentVault && (
                <Button variant="ghost" onClick={handleLockVault}>
                  <LogOut size={18} className="mr-2" />
                  Lock Vault
                </Button>
              )}
            </div>
          </div>
        </div>
      </header>

      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {error && (
          <div className="mb-6 p-4 bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 rounded-lg text-red-800 dark:text-red-200">
            {error}
          </div>
        )}

        {!currentVault ? (
          <VaultSelector
            vaults={vaults}
            onUnlock={() => setShowUnlockVault(true)}
            onCreate={() => setShowCreateVault(true)}
            onRefresh={loadVaults}
          />
        ) : (
          <div className="space-y-6">
            {/* Vault Info & Actions */}
            <Card>
              <CardHeader>
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <div className="p-2 bg-green-100 dark:bg-green-900/30 rounded-lg">
                      <Lock className="text-green-600 dark:text-green-400" size={20} />
                    </div>
                    <div>
                      <CardTitle>{currentVault}</CardTitle>
                      <p className="text-sm text-muted-foreground mt-1">
                        {records.length} password{records.length !== 1 ? 's' : ''} stored
                      </p>
                    </div>
                  </div>
                  <div className="flex gap-2">
                    <Button onClick={() => setShowPasswordGenerator(true)} variant="secondary">
                      <Key size={18} className="mr-2" />
                      Generate Password
                    </Button>
                    <Button onClick={() => setShowAddPassword(true)}>
                      <Plus size={18} className="mr-2" />
                      Add Password
                    </Button>
                  </div>
                </div>
              </CardHeader>
            </Card>

            {/* Search */}
            {records.length > 0 && (
              <div className="relative">
                <Input
                  placeholder="Search passwords..."
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  icon={<Search size={18} />}
                />
              </div>
            )}

            {/* Password List */}
            {isLoading ? (
              <div className="flex items-center justify-center py-12">
                <div className="text-center">
                  <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
                  <p className="mt-4 text-muted-foreground">Loading passwords...</p>
                </div>
              </div>
            ) : filteredRecords.length === 0 ? (
              <Card>
                <CardContent className="py-12 text-center">
                  <Key className="mx-auto text-muted-foreground mb-4" size={48} />
                  <h3 className="text-lg font-semibold text-foreground mb-2">
                    {searchTerm ? 'No passwords found' : 'No passwords yet'}
                  </h3>
                  <p className="text-muted-foreground mb-6">
                    {searchTerm
                      ? 'Try a different search term'
                      : 'Add your first password to get started'}
                  </p>
                  {!searchTerm && (
                    <Button onClick={() => setShowAddPassword(true)}>
                      <Plus size={18} className="mr-2" />
                      Add Your First Password
                    </Button>
                  )}
                </CardContent>
              </Card>
            ) : (
              <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
                {filteredRecords.map((record) => (
                  <PasswordRecordCard
                    key={record.id}
                    record={record}
                    vaultName={currentVault}
                    onUpdate={loadRecords}
                  />
                ))}
              </div>
            )}
          </div>
        )}
      </main>

      {/* Modals */}
      <CreateVaultModal
        isOpen={showCreateVault}
        onClose={() => setShowCreateVault(false)}
        onSuccess={() => {
          loadVaults();
          setShowCreateVault(false);
        }}
      />

      <UnlockVaultModal
        isOpen={showUnlockVault}
        onClose={() => setShowUnlockVault(false)}
        vaults={vaults}
        onUnlock={handleUnlockVault}
      />

      {currentVault && (
        <>
          <AddPasswordModal
            isOpen={showAddPassword}
            onClose={() => setShowAddPassword(false)}
            vaultName={currentVault}
            onSuccess={() => {
              loadRecords();
              setShowAddPassword(false);
            }}
          />

          <PasswordGeneratorModal
            isOpen={showPasswordGenerator}
            onClose={() => setShowPasswordGenerator(false)}
          />
        </>
      )}
    </div>
  );
}

// Create Vault Modal Component
function CreateVaultModal({ isOpen, onClose, onSuccess }: any) {
  const [name, setName] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (password !== confirmPassword) {
      setError('Passwords do not match');
      return;
    }
    setIsLoading(true);
    setError('');
    try {
      await vaultAPI.create({ name, master_password: password });
      setName('');
      setPassword('');
      setConfirmPassword('');
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to create vault');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Create New Vault">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Vault Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="My Personal Vault"
          required
        />
        <PasswordInput
          label="Master Password"
          value={password}
          onChange={setPassword}
          placeholder="Enter a strong master password"
          showStrength
        />
        <PasswordInput
          label="Confirm Master Password"
          value={confirmPassword}
          onChange={setConfirmPassword}
          placeholder="Re-enter your master password"
          error={error}
        />
        <div className="flex gap-3 pt-4">
          <Button type="button" variant="ghost" onClick={onClose} className="flex-1">
            Cancel
          </Button>
          <Button type="submit" isLoading={isLoading} className="flex-1">
            Create Vault
          </Button>
        </div>
      </form>
    </Modal>
  );
}

// Unlock Vault Modal Component
function UnlockVaultModal({ isOpen, onClose, vaults, onUnlock }: any) {
  const [selectedVault, setSelectedVault] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    try {
      await onUnlock(selectedVault, password);
      setPassword('');
    } catch (err) {
      setError('Invalid master password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Unlock Vault">
      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label className="block text-sm font-medium text-foreground mb-1.5">
            Select Vault
          </label>
          <select
            value={selectedVault}
            onChange={(e) => setSelectedVault(e.target.value)}
            className="w-full px-3 py-2 bg-background border border-input rounded-lg text-foreground focus:outline-none focus:ring-2 focus:ring-ring"
            required
          >
            <option value="">-- Select a vault --</option>
            {vaults.map((vault: string) => (
              <option key={vault} value={vault}>
                {vault}
              </option>
            ))}
          </select>
        </div>
        <PasswordInput
          label="Master Password"
          value={password}
          onChange={setPassword}
          placeholder="Enter your master password"
          error={error}
        />
        <div className="flex gap-3 pt-4">
          <Button type="button" variant="ghost" onClick={onClose} className="flex-1">
            Cancel
          </Button>
          <Button type="submit" isLoading={isLoading} className="flex-1">
            Unlock
          </Button>
        </div>
      </form>
    </Modal>
  );
}
