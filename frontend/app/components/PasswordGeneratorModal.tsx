'use client';

import { useState } from 'react';
import { Modal } from './Modal';
import { Button } from './Button';
import { Copy, Check, RefreshCw } from 'lucide-react';
import { generatePassword, type PasswordOptions } from '@/app/lib/password-generator';

interface PasswordGeneratorModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const PasswordGeneratorModal: React.FC<PasswordGeneratorModalProps> = ({
  isOpen,
  onClose,
}) => {
  const [options, setOptions] = useState<PasswordOptions>({
    length: 16,
    uppercase: true,
    lowercase: true,
    numbers: true,
    symbols: true,
  });
  const [password, setPassword] = useState(() => generatePassword(options));
  const [copied, setCopied] = useState(false);

  const handleGenerate = () => {
    const newPassword = generatePassword(options);
    setPassword(newPassword);
  };

  const handleCopy = async () => {
    await navigator.clipboard.writeText(password);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const handleOptionChange = (key: keyof Omit<PasswordOptions, 'length'>) => {
    const newOptions = { ...options, [key]: !options[key] };
    setOptions(newOptions);
    setPassword(generatePassword(newOptions));
  };

  const handleLengthChange = (length: number) => {
    const newOptions = { ...options, length };
    setOptions(newOptions);
    setPassword(generatePassword(newOptions));
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="Password Generator">
      <div className="space-y-6">
        {/* Generated Password */}
        <div>
          <label className="block text-sm font-medium text-foreground mb-2">
            Generated Password
          </label>
          <div className="relative">
            <input
              type="text"
              value={password}
              readOnly
              className="w-full px-4 py-3 pr-24 bg-secondary border border-border rounded-lg font-mono text-lg"
            />
            <div className="absolute right-2 top-1/2 -translate-y-1/2 flex gap-1">
              <button
                onClick={handleGenerate}
                className="p-2 rounded-md hover:bg-background transition-colors"
                title="Generate new password"
              >
                <RefreshCw size={18} />
              </button>
              <button
                onClick={handleCopy}
                className="p-2 rounded-md hover:bg-background transition-colors"
                title="Copy password"
              >
                {copied ? <Check size={18} className="text-green-600" /> : <Copy size={18} />}
              </button>
            </div>
          </div>
        </div>

        {/* Length Slider */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <label className="text-sm font-medium text-foreground">
              Length
            </label>
            <span className="text-sm font-mono bg-secondary px-2 py-1 rounded">
              {options.length}
            </span>
          </div>
          <input
            type="range"
            min="8"
            max="64"
            value={options.length}
            onChange={(e) => handleLengthChange(parseInt(e.target.value))}
            className="w-full h-2 bg-secondary rounded-lg appearance-none cursor-pointer accent-primary-600"
          />
        </div>

        {/* Options */}
        <div className="space-y-3">
          <label className="text-sm font-medium text-foreground block mb-3">
            Character Types
          </label>

          {[
            { key: 'uppercase' as const, label: 'Uppercase (A-Z)', example: 'ABC' },
            { key: 'lowercase' as const, label: 'Lowercase (a-z)', example: 'abc' },
            { key: 'numbers' as const, label: 'Numbers (0-9)', example: '123' },
            { key: 'symbols' as const, label: 'Symbols (!@#$%)', example: '!@#' },
          ].map(({ key, label, example }) => (
            <label
              key={key}
              className="flex items-center justify-between p-3 bg-secondary rounded-lg cursor-pointer hover:bg-secondary/80 transition-colors"
            >
              <div className="flex items-center gap-3">
                <input
                  type="checkbox"
                  checked={options[key]}
                  onChange={() => handleOptionChange(key)}
                  className="w-4 h-4 text-primary-600 bg-background border-border rounded focus:ring-2 focus:ring-primary-500"
                />
                <span className="text-sm font-medium">{label}</span>
              </div>
              <span className="text-xs font-mono text-muted-foreground">{example}</span>
            </label>
          ))}
        </div>

        {/* Actions */}
        <div className="flex gap-3 pt-4">
          <Button variant="secondary" onClick={onClose} className="flex-1">
            Close
          </Button>
          <Button onClick={handleCopy} className="flex-1">
            {copied ? (
              <>
                <Check size={18} className="mr-2" />
                Copied!
              </>
            ) : (
              <>
                <Copy size={18} className="mr-2" />
                Copy Password
              </>
            )}
          </Button>
        </div>
      </div>
    </Modal>
  );
};
