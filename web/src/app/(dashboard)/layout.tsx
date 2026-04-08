'use client';

import type { ReactNode } from 'react';
import { AuthGuard } from '@/components/auth/auth-guard';
import { ConsoleLayout } from '@/layouts/console-layout';

export default function DashboardLayout({ children }: { children: ReactNode }) {
  return (
    <AuthGuard>
      <ConsoleLayout>{children}</ConsoleLayout>
    </AuthGuard>
  );
}

