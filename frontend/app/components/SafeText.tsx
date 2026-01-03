'use client';

import { useEffect, useRef } from 'react';

interface SafeTextProps {
  value: string | null | undefined;
  className?: string;
  as?: 'span' | 'div' | 'p' | 'h1' | 'h2' | 'h3';
  title?: string;
}

/**
 * SafeText component that prevents XSS by using textContent instead of innerHTML.
 * All user-provided strings should be rendered through this component.
 */
export const SafeText: React.FC<SafeTextProps> = ({
  value,
  className = '',
  as: Element = 'span',
  title
}) => {
  const textRef = useRef<HTMLElement>(null);

  useEffect(() => {
    if (textRef.current) {
      // Use textContent instead of innerHTML to prevent XSS
      textRef.current.textContent = value || '';
    }
  }, [value]);

  return (
    <Element
      ref={textRef as any}
      className={className}
      title={title}
      suppressHydrationWarning
    />
  );
};
