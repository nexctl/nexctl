'use client';

import { GlobalOutlined } from '@ant-design/icons';
import { Select, Space } from 'antd';
import type { AppLocale } from '@/i18n/types';
import { useI18n } from '@/i18n';

const options: { value: AppLocale; labelKey: string }[] = [
  { value: 'zh-CN', labelKey: 'locale.zhCN' },
  { value: 'en', labelKey: 'locale.en' },
];

export function LocaleSwitcher({ size = 'small' }: { size?: 'small' | 'middle' | 'large' }) {
  const { locale, setLocale, t } = useI18n();

  return (
    <Space size={6}>
      <GlobalOutlined />
      <Select
        size={size}
        variant="borderless"
        value={locale}
        options={options.map((o) => ({
          value: o.value,
          label: t(o.labelKey),
        }))}
        onChange={(v) => setLocale(v as AppLocale)}
        aria-label={t('common.language')}
        popupMatchSelectWidth={false}
        style={{ minWidth: 120 }}
      />
    </Space>
  );
}
