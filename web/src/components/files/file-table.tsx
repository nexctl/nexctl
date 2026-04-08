'use client';

import { Button, Space, Table } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useMemo } from 'react';
import { useT } from '@/i18n';
import type { FileItem } from '@/types/file';

export function FileTable({ files }: { files: FileItem[] }) {
  const t = useT();

  const columns: ColumnsType<FileItem> = useMemo(
    () => [
      { title: t('files.colName'), dataIndex: 'file_name', key: 'file_name' },
      { title: t('files.colSize'), dataIndex: 'size_text', key: 'size_text' },
      { title: 'Checksum', dataIndex: 'checksum', key: 'checksum' },
      { title: t('files.colCreated'), dataIndex: 'created_at', key: 'created_at' },
      { title: t('files.colDistCount'), dataIndex: 'distribution_count', key: 'distribution_count' },
      {
        title: t('files.colAction'),
        key: 'actions',
        render: () => (
          <Space>
            <Button size="small">{t('files.distLog')}</Button>
            <Button size="small">{t('files.download')}</Button>
          </Space>
        ),
      },
    ],
    [t],
  );

  return <Table rowKey="id" dataSource={files} columns={columns} />;
}
