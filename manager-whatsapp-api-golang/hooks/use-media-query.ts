/**
 * Media Query Hook
 *
 * Reactive hook for media query matching with SSR safety.
 * Returns true if the media query matches the current viewport.
 *
 * @example
 * ```tsx
 * const isMobile = useMediaQuery('(max-width: 768px)');
 * const isDesktop = useMediaQuery('(min-width: 1024px)');
 * ```
 */

'use client';

import * as React from 'react';

/**
 * Hook to detect media query matches
 *
 * @param query - CSS media query string
 * @param defaultValue - Default value for SSR (default: false)
 * @returns true if the media query matches
 */
export function useMediaQuery(query: string, defaultValue = false): boolean {
  const [matches, setMatches] = React.useState(defaultValue);

  React.useEffect(() => {
    // Check if window is available (client-side)
    if (typeof window === 'undefined') {
      return;
    }

    const mediaQuery = window.matchMedia(query);

    // Set initial value
    setMatches(mediaQuery.matches);

    // Create event listener for changes
    const handleChange = (event: MediaQueryListEvent) => {
      setMatches(event.matches);
    };

    // Modern browsers
    if (mediaQuery.addEventListener) {
      mediaQuery.addEventListener('change', handleChange);
      return () => mediaQuery.removeEventListener('change', handleChange);
    }

    // Legacy browsers (Safari < 14)
    mediaQuery.addListener(handleChange);
    return () => mediaQuery.removeListener(handleChange);
  }, [query]);

  return matches;
}
