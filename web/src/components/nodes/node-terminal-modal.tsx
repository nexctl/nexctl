'use client';

import '@xterm/xterm/css/xterm.css';

import { FitAddon } from '@xterm/addon-fit';
import { Terminal } from '@xterm/xterm';
import { App, Modal } from 'antd';
import { useCallback, useEffect, useRef } from 'react';
import { useAuth } from '@/hooks/use-auth';
import { useT } from '@/i18n';

type TerminalWsPayload = {
  session_id?: string;
  data?: string;
  message?: string;
  code?: number;
};

type TerminalWsMessage = {
  type: string;
  request_id?: string;
  timestamp?: string;
  payload?: TerminalWsPayload;
};

function buildTerminalWsURL(nodeId: number, token: string, cols: number, rows: number): string {
  const proto = typeof window !== 'undefined' && window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = typeof window !== 'undefined' ? window.location.host : '';
  const q = new URLSearchParams({
    token,
    cols: String(cols),
    rows: String(rows),
  });
  return `${proto}//${host}/api/v1/nodes/${nodeId}/terminal/ws?${q.toString()}`;
}

function utf8ToBase64(s: string): string {
  return btoa(unescape(encodeURIComponent(s)));
}

function base64ToUint8Array(b64: string): Uint8Array {
  const bin = atob(b64);
  const out = new Uint8Array(bin.length);
  for (let i = 0; i < bin.length; i++) {
    out[i] = bin.charCodeAt(i);
  }
  return out;
}

function sendWs(ws: WebSocket, type: string, payload: Record<string, unknown>) {
  ws.send(
    JSON.stringify({
      type,
      request_id: crypto.randomUUID(),
      timestamp: new Date().toISOString(),
      payload,
    }),
  );
}

type NodeTerminalModalProps = {
  open: boolean;
  onClose: () => void;
  nodeId: number;
  nodeName: string;
};

export function NodeTerminalModal({ open, onClose, nodeId, nodeName }: NodeTerminalModalProps) {
  const t = useT();
  const { message } = App.useApp();
  const { user } = useAuth();
  const containerRef = useRef<HTMLDivElement | null>(null);
  const termRef = useRef<Terminal | null>(null);
  const fitRef = useRef<FitAddon | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const roRef = useRef<ResizeObserver | null>(null);
  const teardownRef = useRef<(() => void) | null>(null);

  const cleanup = useCallback(() => {
    roRef.current?.disconnect();
    roRef.current = null;
    const ws = wsRef.current;
    wsRef.current = null;
    if (ws && ws.readyState === WebSocket.OPEN) {
      try {
        sendWs(ws, 'terminal_close', {});
      } catch {
        /* ignore */
      }
      ws.close();
    }
    termRef.current?.dispose();
    termRef.current = null;
    fitRef.current = null;
  }, []);

  useEffect(() => {
    if (!open || !user?.token || nodeId <= 0) {
      if (!open) {
        cleanup();
      }
      return;
    }

    let disposed = false;

    const timer = window.setTimeout(() => {
      if (disposed) {
        return;
      }
      const el = containerRef.current;
      if (!el) {
        return;
      }

      const term = new Terminal({
        cursorBlink: true,
        fontSize: 14,
        fontFamily: 'Consolas, "Cascadia Code", "Courier New", monospace',
      });
      const fit = new FitAddon();
      term.loadAddon(fit);
      term.open(el);
      fit.fit();
      termRef.current = term;
      fitRef.current = fit;

      const cols = term.cols;
      const rows = term.rows;
      const url = buildTerminalWsURL(nodeId, user.token, cols, rows);
      let ws: WebSocket;
      try {
        ws = new WebSocket(url);
      } catch (e) {
        message.error(e instanceof Error ? e.message : t('nodes.terminalConnectFailed'));
        cleanup();
        return;
      }
      wsRef.current = ws;

      ws.onopen = () => {
        const f = fitRef.current;
        const tt = termRef.current;
        if (f && tt) {
          f.fit();
          sendWs(ws, 'terminal_resize', { cols: tt.cols, rows: tt.rows });
        }
      };

      ws.onmessage = (ev) => {
        let data: TerminalWsMessage;
        try {
          data = JSON.parse(ev.data as string) as TerminalWsMessage;
        } catch {
          return;
        }
        const tt = termRef.current;
        if (!tt) {
          return;
        }
        switch (data.type) {
          case 'terminal_output': {
            const raw = data.payload?.data;
            if (raw) {
              try {
                tt.write(base64ToUint8Array(raw));
              } catch {
                /* ignore corrupt chunk */
              }
            }
            break;
          }
          case 'terminal_exit':
            tt.write(`\r\n\x1b[33m[exit ${data.payload?.code ?? 0}]\x1b[0m\r\n`);
            break;
          case 'terminal_error':
            message.error(data.payload?.message ?? t('nodes.terminalConnectFailed'));
            break;
          default:
            break;
        }
      };

      ws.onerror = () => {
        message.error(t('nodes.terminalConnectFailed'));
      };

      ws.onclose = () => {
        wsRef.current = null;
      };

      const sub = term.onData((d) => {
        const socket = wsRef.current;
        if (socket && socket.readyState === WebSocket.OPEN) {
          sendWs(socket, 'terminal_input', { data: utf8ToBase64(d) });
        }
      });

      const onResize = () => {
        fitRef.current?.fit();
        const socket = wsRef.current;
        const tt = termRef.current;
        if (socket && socket.readyState === WebSocket.OPEN && tt) {
          sendWs(socket, 'terminal_resize', { cols: tt.cols, rows: tt.rows });
        }
      };

      const ro = new ResizeObserver(() => {
        onResize();
      });
      ro.observe(el);
      roRef.current = ro;
      window.addEventListener('resize', onResize);

      teardownRef.current = () => {
        window.removeEventListener('resize', onResize);
        sub.dispose();
        ro.disconnect();
        roRef.current = null;
      };
    }, 0);

    return () => {
      disposed = true;
      window.clearTimeout(timer);
      teardownRef.current?.();
      teardownRef.current = null;
      cleanup();
    };
  }, [open, user?.token, nodeId, cleanup, message, t]);

  const handleClose = () => {
    cleanup();
    onClose();
  };

  return (
    <Modal
      title={t('nodes.terminalModalTitle', { name: nodeName || `#${nodeId}` })}
      open={open}
      onCancel={handleClose}
      footer={null}
      width="min(960px, 96vw)"
      destroyOnHidden
      styles={{ body: { padding: 0 } }}
    >
      {!user?.token ? (
        <div style={{ padding: 24 }}>{t('nodes.terminalNeedLogin')}</div>
      ) : (
        <div ref={containerRef} style={{ width: '100%', height: 'min(480px, 70vh)', padding: 8 }} />
      )}
    </Modal>
  );
}
