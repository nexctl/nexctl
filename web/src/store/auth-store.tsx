'use client';

import { createContext, useEffect, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import type { AuthUser } from '@/types/auth';
import { clearAuthStorage, readAuthStorage, writeAuthStorage } from '@/utils/storage';

interface AuthContextValue {
  user: AuthUser | null;
  initialized: boolean;
  login: (username: string, token: string) => void;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [initialized, setInitialized] = useState(false);

  useEffect(() => {
    const stored = readAuthStorage();
    if (stored) {
      setUser({ username: stored.username, token: stored.token });
    }
    setInitialized(true);
  }, []);

  const value = useMemo<AuthContextValue>(
    () => ({
      user,
      initialized,
      login: (username: string, token: string) => {
        writeAuthStorage({ username, token });
        setUser({ username, token });
      },
      logout: () => {
        clearAuthStorage();
        setUser(null);
      },
    }),
    [initialized, user],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

