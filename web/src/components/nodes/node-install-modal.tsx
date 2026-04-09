'use client';

import { Alert, Modal, Spin } from 'antd';
import { useEffect, useState } from 'react';
import { useT } from '@/i18n';
import { getNodeAgentCredentials } from '@/services/node';
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
  const [credResult, setCredResult] = useState<{
    agentId: string;
    agentSecret: string;
    nodeKey: string;
    name: string;
  } | null>(null);

  useEffect(() => {
    if (!open || !node) {
      return;
    }
    setError(null);
    setCredResult(null);
    if (node.status !== 'pending') {
      return;
    }
    setLoading(true);
    void getNodeAgentCredentials(node.id)
      .then((res) => {
        setCredResult({
          agentId: res.agent_id,
          agentSecret: res.agent_secret,
          nodeKey: res.node_key,
          name: node.name,
        });
      })
      .catch((e: unknown) => {
        setError(e instanceof Error ? e.message : t('nodes.fetchCredentialsFailed'));
      })
      .finally(() => {
        setLoading(false);
      });
  }, [open, node, t]);

  const handleClose = () => {
    setCredResult(null);
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
        <Alert type="error" showIcon title={t('nodes.fetchCredentialsFailed')} description={error} />
      ) : credResult ? (
        <EnrollmentDeployCommands
          nodeName={credResult.name}
          agentId={credResult.agentId}
          agentSecret={credResult.agentSecret}
          nodeKey={credResult.nodeKey}
          nodeId={typeof node.id === 'number' ? node.id : Number(node.id)}
        />
      ) : null}
    </Modal>
  );
}
