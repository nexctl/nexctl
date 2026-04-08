'use client';

/**
 * 根布局内抛错时使用；必须自带 html/body，且不能使用 AntdRegistry（尚未挂载）。
 */
export default function GlobalError({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  return (
    <html lang="zh-CN">
      <body style={{ fontFamily: 'system-ui, sans-serif', padding: 48, margin: 0, background: '#f5f7fa' }}>
        <h1 style={{ fontSize: 20 }}>NexCtl console error / 控制台异常</h1>
        <p style={{ color: '#64748b' }}>{error.message || 'Root layout failed / 根布局加载失败'}</p>
        <button
          type="button"
          onClick={() => reset()}
          style={{
            marginTop: 16,
            padding: '8px 16px',
            cursor: 'pointer',
            borderRadius: 8,
            border: '1px solid #1677ff',
            background: '#1677ff',
            color: '#fff',
          }}
        >
          Retry / 重试
        </button>
      </body>
    </html>
  );
}
