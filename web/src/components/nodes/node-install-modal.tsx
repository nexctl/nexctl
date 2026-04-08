'use client';

import { Alert, Modal, Spin } from 'antd';
import { useEffect, useState } from 'react';
import { useT } from '@/i18n';
import { issueNodeEnrollmentToken } from '@/services/node';
import type { NodeItem } from '@/types/node';
import { EnrollmentDeployCommands } from '@/components/nodes/enrollment-deploy-commands';

type Props = {
  open: boolean;
  onClose: () => void;
  node: NodeItem | null;
};

export function NodeInstallModal({ open, onClose, node }: Props) {
  const t = useT();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [tokenResult, setTokenResult] = useState<{
    token: string;
    expiresAt?: string;
    name: string;
  } | null>(null);

  useEffect(() => {
    if (!open || !node) {
      return;
    }
    setError(null);
    setTokenResult(null);
    if (node.status !== 'pending') {
      return;
    }
    setLoading(true);
    void issueNodeEnrollmentToken(node.id)
      .then((res) => {
        setTokenResult({
          token: res.enrollment_token,
          expiresAt: res.enrollment_expires_at,
          name: res.name,
        });
      })
      .catch((e: unknown) => {
        setError(e instanceof Error ? e.message : t('nodes.issueTokenFailed'));
      })
      .finally(() => {
        setLoading(false);
      });
  }, [open, node, t]);

  const handleClose = () => {
    setTokenResult(null);
    setError(null);
    onClose();
  };

  return (
    <Modal
      title={t('nodes.modalInstallTitle')}
      open={open}
      onCancel={handleClose}
      footer={null}
      destroyOnHidden
      width={720}
    >
      {!node ? null : node.status !== 'pending' ? (
        <Alert type="info" showIcon title={t('nodes.installNotPendingTitle')} description={t('nodes.installNotPendingDesc')} />
      ) : loading ? (
        <div style={{ textAlign: 'center', padding: 48 }}>
          <Spin description={t('nodes.installLoading')} />
        </div>
      ) : error ? (
        <Alert type="error" showIcon title={t('nodes.issueTokenFailed')} description={error} />
      ) : tokenResult ? (
        <EnrollmentDeployCommands
          nodeName={tokenResult.name}
          token={tokenResult.token}
          expiresAt={tokenResult.expiresAt}
        />
      ) : null}
    </Modal>
  );
}
