'use client';

import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { App, Button, Card, Form, Input, Typography } from 'antd';
import { useRouter } from 'next/navigation';
import { LocaleSwitcher } from '@/components/i18n/locale-switcher';
import { useT } from '@/i18n';
import { login } from '@/services/auth';
import type { LoginPayload } from '@/types/auth';
import { useAuth } from '@/hooks/use-auth';

export function LoginForm() {
  const router = useRouter();
  const { message } = App.useApp();
  const { login: saveLogin } = useAuth();
  const t = useT();

  const onFinish = async (values: LoginPayload) => {
    try {
      const response = await login(values);
      saveLogin(values.username, response.access_token);
      message.success(t('login.success'));
      router.replace('/dashboard');
    } catch (error) {
      message.error(error instanceof Error ? error.message : t('login.failed'));
    }
  };

  return (
    <Card style={{ width: 420, borderRadius: 20, boxShadow: '0 12px 30px rgba(15, 23, 42, 0.08)' }}>
      <div style={{ display: 'flex', justifyContent: 'flex-end', marginBottom: 8 }}>
        <LocaleSwitcher size="middle" />
      </div>
      <Typography.Title level={2} style={{ marginBottom: 8 }}>
        NexCtl
      </Typography.Title>
      <Typography.Paragraph type="secondary">{t('login.subtitle')}</Typography.Paragraph>
      <Form layout="vertical" onFinish={onFinish} initialValues={{ username: 'admin' }}>
        <Form.Item
          name="username"
          label={t('login.username')}
          rules={[{ required: true, message: t('login.usernameRequired') }]}
        >
          <Input prefix={<UserOutlined />} size="large" />
        </Form.Item>
        <Form.Item
          name="password"
          label={t('login.password')}
          rules={[{ required: true, message: t('login.passwordRequired') }]}
        >
          <Input.Password prefix={<LockOutlined />} size="large" />
        </Form.Item>
        <Button htmlType="submit" type="primary" size="large" block>
          {t('login.submit')}
        </Button>
      </Form>
    </Card>
  );
}
