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

function stripTrailingSlash(s: string): string {
  return s.replace(/\/$/, '');
}

/**
 * Agent 注册使用的控制面 HTTP 根地址（无 /api/v1 后缀），与 agent.yaml 的 server_url 一致。
 *
 * 优先级：
 * 1. NEXT_PUBLIC_AGENT_SERVER_URL（构建/运行时在 .env 中配置）
 * 2. 浏览器：常见「控制台 :3000、API :8080 同主机」→ 用当前 hostname + 8080（局域网访问控制台时避免错用 3000）
 * 3. NEXT_PUBLIC_INTERNAL_API_ORIGIN（next.config 由 INTERNAL_API_BASE_URL 注入，与 rewrites 同源）
 * 4. NEXT_PUBLIC_API_BASE_URL 为绝对 URL 时取其 origin
 * 5. localhost:3000 → http://127.0.0.1:8080
 * 6. window.location.origin
 */
export function resolveAgentServerUrl(): string {
  const explicit = process.env.NEXT_PUBLIC_AGENT_SERVER_URL?.trim();
  if (explicit) {
    return stripTrailingSlash(explicit);
  }

  if (typeof window !== 'undefined') {
    const { protocol, hostname, port } = window.location;
    if (port === '3000' && hostname && hostname !== 'localhost' && hostname !== '127.0.0.1') {
      return stripTrailingSlash(`${protocol}//${hostname}:8080`);
    }
  }

  const fromInternal = process.env.NEXT_PUBLIC_INTERNAL_API_ORIGIN?.trim();
  if (fromInternal) {
    return stripTrailingSlash(fromInternal);
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

  if (typeof window !== 'undefined') {
    const { protocol, hostname, port } = window.location;
    if (
      (hostname === 'localhost' || hostname === '127.0.0.1') &&
      (port === '3000' || port === '')
    ) {
      return 'http://127.0.0.1:8080';
    }
    return window.location.origin;
  }

  return 'http://127.0.0.1:8080';
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
