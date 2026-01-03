/**
 * Sanitizes text to prevent XSS attacks by removing HTML tags and escaping special characters.
 */
export function sanitizeText(text: string | null | undefined): string {
  if (!text) return '';

  const str = String(text);

  // Remove all HTML tags
  const withoutTags = str.replace(/<[^>]*>/g, '');

  // Escape special HTML characters
  const escaped = withoutTags
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#x27;')
    .replace(/\//g, '&#x2F;');

  return escaped;
}

/**
 * Checks if text contains potentially dangerous patterns.
 */
export function isTextSafe(text: string | null | undefined): boolean {
  if (!text) return true;

  const str = String(text);
  const dangerousPatterns = [
    /<script/i,
    /javascript:/i,
    /on\w+\s*=/i,
    /<iframe/i,
    /<object/i,
    /<embed/i,
    /eval\(/i,
    /expression\(/i,
  ];

  return !dangerousPatterns.some(pattern => pattern.test(str));
}
