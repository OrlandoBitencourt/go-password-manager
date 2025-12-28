export interface PasswordOptions {
  length: number;
  uppercase: boolean;
  lowercase: boolean;
  numbers: boolean;
  symbols: boolean;
}

const CHAR_SETS = {
  uppercase: 'ABCDEFGHIJKLMNOPQRSTUVWXYZ',
  lowercase: 'abcdefghijklmnopqrstuvwxyz',
  numbers: '0123456789',
  symbols: '!@#$%^&*()_+-=[]{}|;:,.<>?',
};

export function generatePassword(options: PasswordOptions): string {
  let charset = '';
  let password = '';

  if (options.uppercase) charset += CHAR_SETS.uppercase;
  if (options.lowercase) charset += CHAR_SETS.lowercase;
  if (options.numbers) charset += CHAR_SETS.numbers;
  if (options.symbols) charset += CHAR_SETS.symbols;

  if (charset === '') charset = CHAR_SETS.lowercase;

  const values = new Uint32Array(options.length);
  crypto.getRandomValues(values);

  for (let i = 0; i < options.length; i++) {
    password += charset[values[i] % charset.length];
  }

  return password;
}

export function calculatePasswordStrength(password: string): {
  score: number;
  label: string;
  color: string;
} {
  let score = 0;

  if (password.length >= 8) score += 1;
  if (password.length >= 12) score += 1;
  if (password.length >= 16) score += 1;
  if (/[a-z]/.test(password)) score += 1;
  if (/[A-Z]/.test(password)) score += 1;
  if (/[0-9]/.test(password)) score += 1;
  if (/[^a-zA-Z0-9]/.test(password)) score += 1;

  if (score <= 2) return { score: 1, label: 'Weak', color: 'bg-red-500' };
  if (score <= 4) return { score: 2, label: 'Fair', color: 'bg-orange-500' };
  if (score <= 5) return { score: 3, label: 'Good', color: 'bg-yellow-500' };
  if (score <= 6) return { score: 4, label: 'Strong', color: 'bg-green-500' };
  return { score: 5, label: 'Very Strong', color: 'bg-green-600' };
}
