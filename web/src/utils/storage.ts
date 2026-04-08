const AUTH_KEY = 'nexctl_console_auth';

export function readAuthStorage(): { username: string; token: string } | null {
  if (typeof window === 'undefined') {
    return null;
  }
  const raw = window.localStorage.getItem(AUTH_KEY);
  if (!raw) {
    return null;
  }
  try {
    return JSON.parse(raw) as { username: string; token: string };
  } catch {
    return null;
  }
}

export function writeAuthStorage(value: { username: string; token: string }) {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.setItem(AUTH_KEY, JSON.stringify(value));
}

export function clearAuthStorage() {
  if (typeof window === 'undefined') {
    return;
  }
  window.localStorage.removeItem(AUTH_KEY);
}

