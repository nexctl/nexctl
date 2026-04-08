import 'antd/dist/reset.css';
import './globals.css';
import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { AntdRegistry } from '@ant-design/nextjs-registry';
import { AppProviders } from '@/store/providers';

export const metadata: Metadata = {
  title: 'NexCtl Console',
  description: 'NexCtl operations console — node and fleet management UI.',
};

export default function RootLayout({ children }: { children: ReactNode }) {
  return (
    <html lang="zh-CN" suppressHydrationWarning data-google-analytics-opt-out="">
      <body suppressHydrationWarning>
        <AntdRegistry>
          <AppProviders>{children}</AppProviders>
        </AntdRegistry>
      </body>
    </html>
  );
}
