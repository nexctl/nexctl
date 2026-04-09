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
 * Agent 的 server_url（无 /api/v1 后缀）。
 * 浏览器内默认使用当前控制台页面的 origin（与地址栏同源，即前端域名）；经反向代理或 Next rewrites 时 Agent 与浏览器访问同一入口。
 * 构建时可设 NEXT_PUBLIC_AGENT_SERVER_URL 覆盖；无 window 时（SSR）回退到内部 API origin。
 */
export function resolveAgentServerUrl(): string {
  const explicit = process.env.NEXT_PUBLIC_AGENT_SERVER_URL?.trim();
  if (explicit) {
    return stripTrailingSlash(explicit);
  }

  if (typeof window !== 'undefined') {
    return stripTrailingSlash(window.location.origin);
  }

  const fromInternal = process.env.NEXT_PUBLIC_INTERNAL_API_ORIGIN?.trim();
  if (fromInternal) {
    return stripTrailingSlash(fromInternal);
  }

  const api = process.env.NEXT_PUBLIC_API_BASE_URL ?? '/api/v1';
  if (api.startsWith('http://') || api.startsWith('https://')) {
    try {
      return stripTrailingSlash(new URL(api).origin);
    } catch {
      /* fall through */
    }
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

export function buildLinuxInstallCommand(
  serverUrl: string,
  agentId: string,
  agentSecret: string,
  nodeKey: string,
  nodeId?: number,
): string {
  const base = getInstallScriptBaseUrl();
  const url = escapeBashSingleQuoted(serverUrl);
  const id = escapeBashSingleQuoted(agentId);
  const sec = escapeBashSingleQuoted(agentSecret);
  const nk = escapeBashSingleQuoted(nodeKey);
  const nid =
    nodeId != null && !Number.isNaN(nodeId) ? escapeBashSingleQuoted(String(nodeId)) : escapeBashSingleQuoted('');
  return `curl -fsSL ${base}/install.sh | sh -s -- ${url} ${id} ${sec} ${nk} ${nid}`;
}

/** Windows：下载到用户可写目录（%TEMP%），避免在 C:\\ 等根目录写入 install.ps1 遭拒绝 */
export function buildWindowsInstallLines(
  serverUrl: string,
  agentId: string,
  agentSecret: string,
  nodeKey: string,
  nodeId?: number,
): string {
  const base = getInstallScriptBaseUrl();
  const su = escapePowerShellSingleQuoted(serverUrl);
  const id = escapePowerShellSingleQuoted(agentId);
  const sec = escapePowerShellSingleQuoted(agentSecret);
  const nk = escapePowerShellSingleQuoted(nodeKey);
  const nidArg =
    nodeId != null && !Number.isNaN(nodeId) ? ` -NodeId ${escapePowerShellSingleQuoted(String(nodeId))}` : '';
  return (
    `Invoke-WebRequest -Uri "${base}/install.ps1" -OutFile "$env:TEMP\\nexctl-install.ps1" -UseBasicParsing
powershell -ExecutionPolicy Bypass -File "$env:TEMP\\nexctl-install.ps1" -ServerUrl ${su} -AgentId ${id} -AgentSecret ${sec} -NodeKey ${nk}${nidArg}`
  );
}
