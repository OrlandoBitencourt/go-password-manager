'use client';

import { useState } from 'react';
import { Copy, Eye, EyeOff, Trash2, Check, User } from 'lucide-react';
import { Card, CardContent } from './Card';
import { Button } from './Button';
import { recordAPI } from '@/app/lib/api';
import type { PasswordRecord } from '@/app/types';

interface PasswordRecordCardProps {
  record: PasswordRecord;
  vaultName: string;
  onUpdate: () => void;
}

export const PasswordRecordCard: React.FC<PasswordRecordCardProps> = ({
  record,
  vaultName,
  onUpdate,
}) => {
  const [showPassword, setShowPassword] = useState(false);
  const [copied, setCopied] = useState<'username' | 'password' | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleCopy = async (text: string, type: 'username' | 'password') => {
    await navigator.clipboard.writeText(text);
    setCopied(type);
    setTimeout(() => setCopied(null), 2000);
  };

  const handleDelete = async () => {
    if (!confirm(`Are you sure you want to delete "${record.name}"?`)) return;
    setIsDeleting(true);
    try {
      await recordAPI.delete(vaultName, record.name);
      onUpdate();
    } catch (err) {
      alert('Failed to delete password');
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <Card hover className="group">
      <CardContent className="p-4">
        <div className="space-y-3">
          {/* Header */}
          <div className="flex items-start justify-between">
            <div>
              <h3 className="font-semibold text-foreground text-lg">{record.name}</h3>
              <p className="text-sm text-muted-foreground">
                Updated {new Date(record.updated_at).toLocaleDateString()}
              </p>
            </div>
            <button
              onClick={handleDelete}
              disabled={isDeleting}
              className="p-2 rounded-lg hover:bg-red-50 dark:hover:bg-red-900/20 text-muted-foreground hover:text-red-600 dark:hover:text-red-400 transition-colors opacity-0 group-hover:opacity-100"
              title="Delete password"
            >
              <Trash2 size={16} />
            </button>
          </div>

          {/* Username */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Username</label>
            <div className="flex items-center gap-2">
              <div className="flex-1 p-2 bg-secondary rounded-md flex items-center gap-2 min-w-0">
                <User size={14} className="text-muted-foreground flex-shrink-0" />
                <span className="text-sm truncate">{record.username}</span>
              </div>
              <button
                onClick={() => handleCopy(record.username, 'username')}
                className="p-2 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
                title="Copy username"
              >
                {copied === 'username' ? <Check size={16} className="text-green-600" /> : <Copy size={16} />}
              </button>
            </div>
          </div>

          {/* Password */}
          <div className="space-y-1">
            <label className="text-xs font-medium text-muted-foreground">Password</label>
            <div className="flex items-center gap-2">
              <div className="flex-1 p-2 bg-secondary rounded-md flex items-center gap-2 min-w-0">
                <span className="text-sm font-mono truncate">
                  {showPassword ? record.password : '••••••••••••'}
                </span>
              </div>
              <button
                onClick={() => setShowPassword(!showPassword)}
                className="p-2 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
                title={showPassword ? 'Hide password' : 'Show password'}
              >
                {showPassword ? <EyeOff size={16} /> : <Eye size={16} />}
              </button>
              <button
                onClick={() => handleCopy(record.password, 'password')}
                className="p-2 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors flex-shrink-0"
                title="Copy password"
              >
                {copied === 'password' ? <Check size={16} className="text-green-600" /> : <Copy size={16} />}
              </button>
            </div>
          </div>
        </div>
      </CardContent>
    </Card>
  );
};
