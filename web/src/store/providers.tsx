'use client';

import { App as AntdApp, ConfigProvider, theme } from 'antd';
import enUS from 'antd/locale/en_US';
import zhCN from 'antd/locale/zh_CN';
import type { ReactNode } from 'react';
import { useMemo } from 'react';
import { I18nProvider, useI18n } from '@/i18n';
import { AuthProvider } from '@/store/auth-store';

function AntdLocaleBridge({ children }: { children: ReactNode }) {
  const { locale } = useI18n();
  const antdLocale = locale === 'zh-CN' ? zhCN : enUS;

  const themeConfig = useMemo(
    () => ({
      algorithm: theme.defaultAlgorithm,
      token: {
        colorPrimary: '#1677ff',
        borderRadius: 12,
        colorBgLayout: '#f5f7fa',
      },
    }),
    [],
  );

  return (
    <ConfigProvider locale={antdLocale} theme={themeConfig}>
      <AntdApp>{children}</AntdApp>
    </ConfigProvider>
  );
}

export function AppProviders({ children }: { children: ReactNode }) {
  return (
    <I18nProvider>
      <AntdLocaleBridge>
        <AuthProvider>{children}</AuthProvider>
      </AntdLocaleBridge>
    </I18nProvider>
  );
}
