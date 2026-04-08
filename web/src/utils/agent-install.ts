/**
 * 与仓库 deploy/install 脚本一致，默认使用 GitHub raw（分支可通过 NEXT_PUBLIC_NEXCTL_INSTALL_REF 覆盖）。
 * @see https://github.com/nexctl/nexctl
 */
const DEFAULT_REF = 'master';

export function getInstallScriptRef(): string {
  return (process.env.NEXT_PUBLIC_NEXCTL_INSTALL_REF || DEFAULT_REF).replace(/^\/+|\/+$/g, '');
}

export function getInstallScriptBaseUrl(): string {
  const ref = getInstallScriptRef();
  return `https://raw.githubusercontent.com/nexctl/nexctl/${ref}/deploy/install`;
}

/**
 * Agent 注册使用的控制面 HTTP 根地址（无 /api/v1 后缀），与 agent.yaml 的 server_url 一致。
 * 优先 NEXT_PUBLIC_AGENT_SERVER_URL；否则从 NEXT_PUBLIC_API_BASE_URL 推导；
 * 本地 dev 常见为控制台 :3000 + API 反代，则默认回退到 http://127.0.0.1:8080。
 */
export function resolveAgentServerUrl(): string {
  if (typeof window === 'undefined') {
    return 'http://127.0.0.1:8080';
  }
  const explicit = process.env.NEXT_PUBLIC_AGENT_SERVER_URL?.trim();
  if (explicit) {
    return explicit.replace(/\/$/, '');
  }
  const api = process.env.NEXT_PUBLIC_API_BASE_URL ?? '/api/v1';
  if (api.startsWith('http://') || api.startsWith('https://')) {
    try {
      const u = new URL(api);
      return u.origin;
    } catch {
      /* fall through */
    }
  }
  const { protocol, hostname, port } = window.location;
  if (
    (hostname === 'localhost' || hostname === '127.0.0.1') &&
    (port === '3000' || port === '')
  ) {
    return 'http://127.0.0.1:8080';
  }
  return window.location.origin;
}

/** Bash 单引号字符串转义 */
export function escapeBashSingleQuoted(s: string): string {
  return `'${s.replace(/'/g, `'\\''`)}'`;
}

/** PowerShell 单引号字符串内单引号加倍 */
export function escapePowerShellSingleQuoted(s: string): string {
  return `'${s.replace(/'/g, "''")}'`;
}

export function buildLinuxInstallCommand(serverUrl: string, enrollmentToken: string): string {
  const base = getInstallScriptBaseUrl();
  const url = escapeBashSingleQuoted(serverUrl);
  const tok = escapeBashSingleQuoted(enrollmentToken);
  return `curl -fsSL ${base}/install.sh | sh -s -- ${url} ${tok}`;
}

/** Windows：下载到用户可写目录（%TEMP%），避免在 C:\\ 等根目录写入 install.ps1 遭拒绝 */
export function buildWindowsInstallLines(serverUrl: string, enrollmentToken: string): string {
  const base = getInstallScriptBaseUrl();
  const su = escapePowerShellSingleQuoted(serverUrl);
  const tok = escapePowerShellSingleQuoted(enrollmentToken);
  return (
    `Invoke-WebRequest -Uri "${base}/install.ps1" -OutFile "$env:TEMP\\nexctl-install.ps1" -UseBasicParsing
powershell -ExecutionPolicy Bypass -File "$env:TEMP\\nexctl-install.ps1" -ServerUrl ${su} -EnrollmentToken ${tok}`
  );
}
