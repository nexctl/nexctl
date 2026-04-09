/** 远程路径拼接（Windows 用 \\，否则用 /）。 */
export function joinRemotePath(dir: string, name: string): string {
  const isWin = /[A-Za-z]:/.test(dir) || dir.includes('\\');
  const sep = isWin ? '\\' : '/';
  const d = dir.replace(/[/\\]+$/, '');
  return `${d}${sep}${name}`;
}

/** 上级目录；已在根目录时返回原路径。 */
export function parentRemotePath(p: string): string {
  const norm = p.replace(/[/\\]+$/, '');
  const isWin = /^[A-Za-z]:/.test(norm) || (norm.includes('\\') && !norm.startsWith('/'));
  if (isWin) {
    const m = /^([A-Za-z]:)(\\.*)?$/.exec(norm);
    if (m && (!m[2] || m[2] === '\\')) {
      return `${m[1]}\\`;
    }
  }
  const idx = Math.max(norm.lastIndexOf('/'), norm.lastIndexOf('\\'));
  if (idx <= 0) {
    if (isWin && /^[A-Za-z]:\\?$/.test(norm)) {
      return norm.length === 2 ? `${norm}\\` : norm;
    }
    return '/';
  }
  return norm.slice(0, idx);
}

export function defaultRootPath(platform?: string): string {
  const p = (platform ?? '').toLowerCase();
  if (p.includes('windows') || p.includes('win')) {
    return 'C:\\';
  }
  return '/';
}
