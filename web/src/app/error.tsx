'use client';

import { Button, Result, Space, Typography } from 'antd';
import Link from 'next/link';
import { useEffect } from 'react';
import { useT } from '@/i18n';

export default function AppError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  const t = useT();

  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div style={{ padding: 48, maxWidth: 560, margin: '0 auto' }}>
      <Result
        status="error"
        title={t('error.title')}
        subTitle={error.message || t('error.subtitle')}
        extra={
          <Space>
            <Button type="primary" onClick={() => reset()}>
              {t('common.retry')}
            </Button>
            <Link href="/">
              <Button>{t('common.backHome')}</Button>
            </Link>
          </Space>
        }
      />
      {process.env.NODE_ENV === 'development' && error.digest ? (
        <Typography.Paragraph type="secondary" copyable style={{ marginTop: 16 }}>
          digest: {error.digest}
        </Typography.Paragraph>
      ) : null}
    </div>
  );
}
