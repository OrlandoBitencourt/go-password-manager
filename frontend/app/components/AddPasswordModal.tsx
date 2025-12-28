'use client';

import { useState } from 'react';
import { Modal } from './Modal';
import { Input } from './Input';
import { PasswordInput } from './PasswordInput';
import { Button } from './Button';
import { recordAPI } from '@/app/lib/api';
import { Key } from 'lucide-react';

interface AddPasswordModalProps {
  isOpen: boolean;
  onClose: () => void;
  vaultName: string;
  onSuccess: () => void;
}

export const AddPasswordModal: React.FC<AddPasswordModalProps> = ({
  isOpen,
  onClose,
  vaultName,
  onSuccess,
}) => {
  const [name, setName] = useState('');
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError('');
    try {
      await recordAPI.add({
        vault_name: vaultName,
        name,
        username,
        password,
      });
      setName('');
      setUsername('');
      setPassword('');
      onSuccess();
    } catch (err: any) {
      setError(err.response?.data?.error || 'Failed to add password');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Add New Password">
      <form onSubmit={handleSubmit} className="space-y-4">
        <Input
          label="Name"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="e.g., GitHub, Gmail"
          required
          icon={<Key size={18} />}
        />
        <Input
          label="Username"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="username or email"
          required
        />
        <PasswordInput
          label="Password"
          value={password}
          onChange={setPassword}
          placeholder="Enter password"
          showStrength
          showCopy
        />
        {error && <p className="text-sm text-red-600 dark:text-red-400">{error}</p>}
        <div className="flex gap-3 pt-4">
          <Button type="button" variant="ghost" onClick={onClose} className="flex-1">
            Cancel
          </Button>
          <Button type="submit" isLoading={isLoading} className="flex-1">
            Add Password
          </Button>
        </div>
      </form>
    </Modal>
  );
};
