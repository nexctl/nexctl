'use client';

import Link from 'next/link';
import { useT } from '@/i18n';

export default function NotFound() {
  const t = useT();
  return (
    <div style={{ padding: 48, maxWidth: 480, margin: '15vh auto 0', textAlign: 'center' }}>
      <h1 style={{ fontSize: 48, margin: '0 0 8px', color: '#94a3b8' }}>404</h1>
      <p style={{ color: '#64748b', marginBottom: 24 }}>{t('notFound.message')}</p>
      <Link href="/" style={{ color: '#1677ff' }}>
        {t('common.backHome')}
      </Link>
    </div>
  );
}
