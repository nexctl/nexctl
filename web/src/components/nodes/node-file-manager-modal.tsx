'use client';

import {
  ArrowUpOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EyeOutlined,
  FolderAddOutlined,
  ReloadOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import { App, Button, Input, Modal, Popconfirm, Space, Table, Typography } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useT } from '@/i18n';
import { nodeFileOp } from '@/services/node';
import type { NodeFileEntry } from '@/services/node';
import { defaultRootPath, joinRemotePath, parentRemotePath } from '@/components/nodes/node-file-path';

function arrayBufferToBase64(buffer: ArrayBuffer): string {
  let binary = '';
  const bytes = new Uint8Array(buffer);
  for (let i = 0; i < bytes.byteLength; i++) {
    binary += String.fromCharCode(bytes[i]!);
  }
  return btoa(binary);
}

function base64ToUtf8(b64: string): string {
  try {
    const bin = atob(b64);
    const bytes = new Uint8Array(bin.length);
    for (let i = 0; i < bin.length; i++) {
      bytes[i] = bin.charCodeAt(i);
    }
    return new TextDecoder('utf-8', { fatal: false }).decode(bytes);
  } catch {
    return '';
  }
}

type NodeFileManagerModalProps = {
  open: boolean;
  onClose: () => void;
  nodeId: number;
  nodeName: string;
  /** 来自节点 platform 字段，用于默认根路径 */
  platform?: string;
};

export function NodeFileManagerModal({ open, onClose, nodeId, nodeName, platform }: NodeFileManagerModalProps) {
  const t = useT();
  const { message } = App.useApp();
  const [path, setPath] = useState('');
  const [pathInput, setPathInput] = useState('');
  const [entries, setEntries] = useState<NodeFileEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [viewerOpen, setViewerOpen] = useState(false);
  const [viewerTitle, setViewerTitle] = useState('');
  const [viewerText, setViewerText] = useState('');
  const [mkdirOpen, setMkdirOpen] = useState(false);
  const [mkdirName, setMkdirName] = useState('');
  const uploadRef = useRef<HTMLInputElement>(null);

  const root = defaultRootPath(platform);

  const loadList = useCallback(async () => {
    const p = pathInput.trim() || path.trim() || root;
    setLoading(true);
    try {
      const res = await nodeFileOp(nodeId, { op: 'list', path: p });
      if (!res.ok) {
        message.error(res.error || t('nodes.fileOpFailed'));
        return;
      }
      setPath(p);
      setPathInput(p);
      setEntries(res.entries ?? []);
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
    } finally {
      setLoading(false);
    }
  }, [nodeId, path, pathInput, root, message, t]);

  useEffect(() => {
    if (!open) {
      return;
    }
    const initial = root;
    setPath(initial);
    setPathInput(initial);
    setEntries([]);
    void (async () => {
      setLoading(true);
      try {
        const res = await nodeFileOp(nodeId, { op: 'list', path: initial });
        if (res.ok) {
          setPath(initial);
          setPathInput(initial);
          setEntries(res.entries ?? []);
        } else {
          message.error(res.error || t('nodes.fileOpFailed'));
        }
      } catch (e) {
        message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
      } finally {
        setLoading(false);
      }
    })();
  }, [open, nodeId, root, message, t]);

  const enterDir = (name: string) => {
    const next = joinRemotePath(path || root, name);
    setPathInput(next);
    setPath(next);
    void (async () => {
      setLoading(true);
      try {
        const res = await nodeFileOp(nodeId, { op: 'list', path: next });
        if (!res.ok) {
          message.error(res.error || t('nodes.fileOpFailed'));
          return;
        }
        setPath(next);
        setPathInput(next);
        setEntries(res.entries ?? []);
      } catch (e) {
        message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
      } finally {
        setLoading(false);
      }
    })();
  };

  const goUp = () => {
    const next = parentRemotePath(path || root);
    setPathInput(next);
    setPath(next);
    void (async () => {
      setLoading(true);
      try {
        const res = await nodeFileOp(nodeId, { op: 'list', path: next });
        if (!res.ok) {
          message.error(res.error || t('nodes.fileOpFailed'));
          return;
        }
        setPath(next);
        setPathInput(next);
        setEntries(res.entries ?? []);
      } catch (e) {
        message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
      } finally {
        setLoading(false);
      }
    })();
  };

  const viewFile = async (name: string) => {
    const fp = joinRemotePath(path || root, name);
    setLoading(true);
    try {
      const res = await nodeFileOp(nodeId, { op: 'read', path: fp, max_bytes: 4 * 1024 * 1024 });
      if (!res.ok) {
        message.error(res.error || t('nodes.fileOpFailed'));
        return;
      }
      const text = base64ToUtf8(res.content_b64 ?? '');
      setViewerTitle(name);
      setViewerText(text);
      setViewerOpen(true);
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
    } finally {
      setLoading(false);
    }
  };

  const downloadFile = async (name: string) => {
    const fp = joinRemotePath(path || root, name);
    setLoading(true);
    try {
      const res = await nodeFileOp(nodeId, { op: 'read', path: fp, max_bytes: 4 * 1024 * 1024 });
      if (!res.ok) {
        message.error(res.error || t('nodes.fileOpFailed'));
        return;
      }
      const bin = atob(res.content_b64 ?? '');
      const bytes = new Uint8Array(bin.length);
      for (let i = 0; i < bin.length; i++) {
        bytes[i] = bin.charCodeAt(i);
      }
      const blob = new Blob([bytes]);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = name;
      a.click();
      URL.revokeObjectURL(url);
      message.success(t('nodes.fileDownloadStarted'));
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
    } finally {
      setLoading(false);
    }
  };

  const removeEntry = async (name: string, isDir: boolean) => {
    const fp = joinRemotePath(path || root, name);
    setLoading(true);
    try {
      const res = await nodeFileOp(nodeId, { op: 'remove', path: fp, recursive: isDir });
      if (!res.ok) {
        message.error(res.error || t('nodes.fileOpFailed'));
        return;
      }
      message.success(t('nodes.fileDeleted'));
      await loadList();
    } catch (e) {
      message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
    } finally {
      setLoading(false);
    }
  };

  const onUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file) {
      return;
    }
    void (async () => {
      const buf = await file.arrayBuffer();
      const b64 = arrayBufferToBase64(buf);
      const fp = joinRemotePath(path || root, file.name);
      setLoading(true);
      try {
        const res = await nodeFileOp(nodeId, { op: 'write', path: fp, content_b64: b64 });
        if (!res.ok) {
          message.error(res.error || t('nodes.fileOpFailed'));
          return;
        }
        message.success(t('nodes.fileUploaded'));
        await loadList();
      } catch (err) {
        message.error(err instanceof Error ? err.message : t('nodes.fileOpFailed'));
      } finally {
        setLoading(false);
      }
    })();
  };

  const doMkdir = () => {
    const name = mkdirName.trim();
    if (!name) {
      return;
    }
    const fp = joinRemotePath(path || root, name);
    void (async () => {
      setLoading(true);
      try {
        const res = await nodeFileOp(nodeId, { op: 'mkdir', path: fp });
        if (!res.ok) {
          message.error(res.error || t('nodes.fileOpFailed'));
          return;
        }
        message.success(t('nodes.fileMkdirOk'));
        setMkdirOpen(false);
        setMkdirName('');
        await loadList();
      } catch (e) {
        message.error(e instanceof Error ? e.message : t('nodes.fileOpFailed'));
      } finally {
        setLoading(false);
      }
    })();
  };

  const columns: ColumnsType<NodeFileEntry> = [
    {
      title: t('nodes.fileColName'),
      dataIndex: 'name',
      key: 'name',
      render: (name: string, row) =>
        row.is_dir ? (
          <Button type="link" onClick={() => enterDir(name)} style={{ padding: 0, height: 'auto' }}>
            {name}
          </Button>
        ) : (
          <Typography.Text>{name}</Typography.Text>
        ),
    },
    {
      title: t('nodes.fileColType'),
      key: 'type',
      width: 80,
      render: (_, row) => (row.is_dir ? t('nodes.fileTypeDir') : t('nodes.fileTypeFile')),
    },
    {
      title: t('nodes.fileColSize'),
      dataIndex: 'size',
      key: 'size',
      width: 110,
      render: (s: number, row) => (row.is_dir ? '—' : formatSize(s)),
    },
    {
      title: t('nodes.fileColMtime'),
      dataIndex: 'mod_time',
      key: 'mod_time',
      width: 200,
    },
    {
      title: t('nodes.fileColAction'),
      key: 'action',
      width: 220,
      render: (_, row) => (
        <Space size="small">
          {row.is_dir ? null : (
            <>
              <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => void viewFile(row.name)}>
                {t('nodes.fileView')}
              </Button>
              <Button type="link" size="small" icon={<DownloadOutlined />} onClick={() => void downloadFile(row.name)}>
                {t('nodes.fileDownload')}
              </Button>
            </>
          )}
          <Popconfirm title={t('nodes.fileDeleteConfirm', { name: row.name })} onConfirm={() => void removeEntry(row.name, row.is_dir)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              {t('common.delete')}
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <>
      <Modal
        title={`${t('nodes.fileManagerTitle')} — ${nodeName}`}
        open={open}
        onCancel={onClose}
        width={960}
        footer={null}
        destroyOnClose
      >
        <Space orientation="vertical" size="middle" style={{ width: '100%' }}>
          <Typography.Text type="secondary">{t('nodes.fileManagerHint')}</Typography.Text>
          <Space wrap>
            <Input
              style={{ minWidth: 320 }}
              value={pathInput}
              onChange={(e) => setPathInput(e.target.value)}
              onPressEnter={() => void loadList()}
              placeholder={t('nodes.filePathPlaceholder')}
            />
            <Button icon={<ReloadOutlined />} onClick={() => void loadList()} loading={loading}>
              {t('nodes.fileGo')}
            </Button>
            <Button icon={<ArrowUpOutlined />} onClick={() => goUp()} disabled={loading}>
              {t('nodes.fileUp')}
            </Button>
            <Button icon={<FolderAddOutlined />} onClick={() => setMkdirOpen(true)}>
              {t('nodes.fileNewFolder')}
            </Button>
            <input ref={uploadRef} type="file" hidden disabled={loading} onChange={onUpload} />
            <Button icon={<UploadOutlined />} disabled={loading} onClick={() => uploadRef.current?.click()}>
              {t('nodes.fileUpload')}
            </Button>
          </Space>
          <Table<NodeFileEntry>
            rowKey={(r) => `${r.is_dir ? 'd' : 'f'}:${r.name}`}
            loading={loading}
            columns={columns}
            dataSource={entries}
            pagination={false}
            size="small"
            scroll={{ y: 360 }}
          />
        </Space>
      </Modal>

      <Modal title={viewerTitle} open={viewerOpen} onCancel={() => setViewerOpen(false)} footer={null} width={720}>
        <Input.TextArea value={viewerText} readOnly rows={18} style={{ fontFamily: 'monospace', fontSize: 12 }} />
      </Modal>

      <Modal
        title={t('nodes.fileNewFolder')}
        open={mkdirOpen}
        onCancel={() => setMkdirOpen(false)}
        onOk={() => doMkdir()}
        okText={t('common.confirm')}
      >
        <Input value={mkdirName} onChange={(e) => setMkdirName(e.target.value)} placeholder={t('nodes.fileFolderName')} onPressEnter={() => doMkdir()} />
      </Modal>
    </>
  );
}

function formatSize(n: number): string {
  if (n < 1024) {
    return `${n} B`;
  }
  if (n < 1024 * 1024) {
    return `${(n / 1024).toFixed(1)} KB`;
  }
  return `${(n / (1024 * 1024)).toFixed(1)} MB`;
}
