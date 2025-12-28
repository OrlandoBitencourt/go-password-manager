'use client';

import React, { useState } from 'react';
import { Eye, EyeOff, Copy, Check } from 'lucide-react';
import { Input } from './Input';

interface PasswordInputProps {
  label?: string;
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  showCopy?: boolean;
  showStrength?: boolean;
  error?: string;
}

export const PasswordInput: React.FC<PasswordInputProps> = ({
  label,
  value,
  onChange,
  placeholder,
  showCopy = false,
  showStrength = false,
  error,
}) => {
  const [showPassword, setShowPassword] = useState(false);
  const [copied, setCopied] = useState(false);

  const handleCopy = async () => {
    await navigator.clipboard.writeText(value);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const calculateStrength = () => {
    let strength = 0;
    if (value.length >= 8) strength += 25;
    if (value.length >= 12) strength += 25;
    if (/[a-z]/.test(value) && /[A-Z]/.test(value)) strength += 25;
    if (/[0-9]/.test(value) && /[^a-zA-Z0-9]/.test(value)) strength += 25;
    return strength;
  };

  const strength = showStrength ? calculateStrength() : 0;
  const strengthColor = strength < 50 ? 'bg-red-500' : strength < 75 ? 'bg-yellow-500' : 'bg-green-500';

  return (
    <div className="w-full">
      {label && (
        <label className="block text-sm font-medium text-foreground mb-1.5">
          {label}
        </label>
      )}
      <div className="relative">
        <input
          type={showPassword ? 'text' : 'password'}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          placeholder={placeholder}
          className={`
            w-full px-3 py-2 pr-20 bg-background border border-input rounded-lg
            text-foreground placeholder:text-muted-foreground
            focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent
            transition-all duration-200
            ${error ? 'border-red-500 focus:ring-red-500' : ''}
          `}
        />
        <div className="absolute right-2 top-1/2 -translate-y-1/2 flex items-center gap-1">
          {showCopy && value && (
            <button
              type="button"
              onClick={handleCopy}
              className="p-1.5 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors"
              title="Copy password"
            >
              {copied ? <Check size={18} /> : <Copy size={18} />}
            </button>
          )}
          <button
            type="button"
            onClick={() => setShowPassword(!showPassword)}
            className="p-1.5 rounded-md hover:bg-secondary text-muted-foreground hover:text-foreground transition-colors"
            title={showPassword ? 'Hide password' : 'Show password'}
          >
            {showPassword ? <EyeOff size={18} /> : <Eye size={18} />}
          </button>
        </div>
      </div>
      {showStrength && value && (
        <div className="mt-2">
          <div className="h-1.5 bg-secondary rounded-full overflow-hidden">
            <div
              className={`h-full transition-all duration-300 ${strengthColor}`}
              style={{ width: `${strength}%` }}
            />
          </div>
          <p className="text-xs text-muted-foreground mt-1">
            Password strength: {strength < 50 ? 'Weak' : strength < 75 ? 'Medium' : 'Strong'}
          </p>
        </div>
      )}
      {error && (
        <p className="mt-1.5 text-sm text-red-600 dark:text-red-400">{error}</p>
      )}
    </div>
  );
};
