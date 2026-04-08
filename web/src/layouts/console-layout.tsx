'use client';

import {
  AlertOutlined,
  AppstoreOutlined,
  AuditOutlined,
  DeploymentUnitOutlined,
  FileOutlined,
  HomeOutlined,
  LogoutOutlined,
  RocketOutlined,
} from '@ant-design/icons';
import { Breadcrumb, Button, Layout, Menu, Space, Typography } from 'antd';
import { usePathname, useRouter } from 'next/navigation';
import type { ReactNode } from 'react';
import { useMemo } from 'react';
import { LocaleSwitcher } from '@/components/i18n/locale-switcher';
import { useAuth } from '@/hooks/use-auth';
import { useT } from '@/i18n';

const { Header, Sider, Content } = Layout;

const pathSegmentLabelKeys: Record<string, string> = {
  dashboard: 'console.menu.dashboard',
  nodes: 'console.menu.nodes',
  tasks: 'console.menu.tasks',
  files: 'console.menu.files',
  upgrades: 'console.menu.upgrades',
  alerts: 'console.menu.alerts',
  'audit-logs': 'console.menu.auditLogs',
};

export function ConsoleLayout({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const router = useRouter();
  const { user, logout } = useAuth();
  const t = useT();

  const menuItems = useMemo(
    () => [
      { key: '/dashboard', icon: <HomeOutlined />, label: t('console.menu.dashboard') },
      { key: '/nodes', icon: <DeploymentUnitOutlined />, label: t('console.menu.nodes') },
      { key: '/tasks', icon: <AppstoreOutlined />, label: t('console.menu.tasks') },
      { key: '/files', icon: <FileOutlined />, label: t('console.menu.files') },
      { key: '/upgrades', icon: <RocketOutlined />, label: t('console.menu.upgrades') },
      { key: '/alerts', icon: <AlertOutlined />, label: t('console.menu.alerts') },
      { key: '/audit-logs', icon: <AuditOutlined />, label: t('console.menu.auditLogs') },
    ],
    [t],
  );

  const selectedKey = useMemo(() => {
    const matched = menuItems.find((item) => pathname.startsWith(item.key));
    return matched?.key ?? '/dashboard';
  }, [pathname, menuItems]);

  const breadcrumbs = useMemo(() => {
    const segments = pathname.split('/').filter(Boolean);
    return segments.map((segment, index, arr) => {
      const labelKey = pathSegmentLabelKeys[segment];
      const title = labelKey ? t(labelKey) : segment;
      return {
        title,
        key: `${segment}-${index}`,
        href: `/${arr.slice(0, index + 1).join('/')}`,
      };
    });
  }, [pathname, t]);

  return (
    <Layout style={{ minHeight: '100vh' }}>
      <Sider width={240} theme="light" style={{ borderRight: '1px solid #e5e7eb' }}>
        <div style={{ padding: 24 }}>
          <Typography.Title level={4} style={{ margin: 0 }}>
            NexCtl
          </Typography.Title>
          <Typography.Text type="secondary">{t('console.brandSubtitle')}</Typography.Text>
        </div>
        <Menu
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={({ key }) => router.push(key)}
        />
      </Sider>
      <Layout>
        <Header
          style={{
            height: 'auto',
            padding: '16px 24px',
            background: '#fff',
            borderBottom: '1px solid #e5e7eb',
          }}
        >
          <Space style={{ width: '100%', justifyContent: 'space-between' }}>
            <Space orientation="vertical" size={4}>
              <Breadcrumb items={breadcrumbs.length ? breadcrumbs : [{ title: t('console.menu.dashboard') }]} />
              <Typography.Text type="secondary">
                {t('console.currentUser')}
                {user?.username ?? t('console.notLoggedIn')}
              </Typography.Text>
            </Space>
            <Space>
              <LocaleSwitcher />
              <Button
                icon={<LogoutOutlined />}
                onClick={() => {
                  logout();
                  router.replace('/login');
                }}
              >
                {t('console.logout')}
              </Button>
            </Space>
          </Space>
        </Header>
        <Content style={{ padding: 24 }}>{children}</Content>
      </Layout>
    </Layout>
  );
}
