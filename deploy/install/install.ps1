#Requires -Version 5.1
<#
.SYNOPSIS
  从 GitHub Releases 安装 NexCtl Agent（Windows），参考 nezhahq/scripts/agent 的体验：
  TLS 1.2、地域镜像、GitHub API 解析版本、下载重试、重装前清理。

.EXAMPLE
  .\install.ps1 -ServerUrl "http://控制面:8080" -AgentId "..." -AgentSecret "..." -NodeKey "..."

.EXAMPLE
  .\install.ps1 -Uninstall

  环境变量：NEXCTL_AGENT_REPO、CN=1、NEXCTL_NO_MIRROR=1、NEXCTL_GH_PROXY、NEXCTL_NO_SERVICE=1（管理员也不注册 Windows 服务）、NEXCTL_NO_ELEVATE=1（不弹出 UAC，仅当前用户目录安装，无服务）

  默认行为：非管理员时会弹出 UAC，同意后以管理员安装到 C:\nexctl 并注册 Windows 服务；可用 -SkipElevation 跳过提权。

  远程一键：请把脚本保存到 %TEMP% 或用户目录（例如 -OutFile "$env:TEMP\nexctl-install.ps1"），
  勿在 C:\ 根目录写入 install.ps1，否则可能「对路径的访问被拒绝」。
#>

[CmdletBinding(DefaultParameterSetName = 'Install')]
param(
  [Parameter(ParameterSetName = 'Install', Mandatory = $true, Position = 0)]
  [string] $ServerUrl,

  [Parameter(ParameterSetName = 'Install', Mandatory = $true, Position = 1)]
  [string] $AgentId,

  [Parameter(ParameterSetName = 'Install', Mandatory = $true, Position = 2)]
  [string] $AgentSecret,

  [Parameter(ParameterSetName = 'Install', Mandatory = $true, Position = 3)]
  [string] $NodeKey,

  [Parameter(ParameterSetName = 'Install', Mandatory = $false)]
  [string] $NodeId = '',

  [Parameter(ParameterSetName = 'Install', Mandatory = $false, Position = 4)]
  [string] $ReleaseTag = '',

  [Parameter(ParameterSetName = 'Uninstall', Mandatory = $true)]
  [switch] $Uninstall,

  # 不请求 UAC 提权：保持当前用户权限安装（用户目录 + 进程运行，不注册 Windows 服务）
  [Parameter(ParameterSetName = 'Install')]
  [switch] $SkipElevation
)

$ErrorActionPreference = 'Stop'

# ---------- 与 nezha 类似的控制台高亮 ----------
function Write-Info([string] $Message) {
  Write-Host $Message -BackgroundColor DarkGreen -ForegroundColor White
}
function Write-WarnLine([string] $Message) {
  Write-Host $Message -BackgroundColor DarkRed -ForegroundColor Green
}

# TLS 1.2（旧版 .NET 默认行为）
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Test-Administrator {
  $p = [Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()
  return $p.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Escape-YamlDouble {
  param([string] $Value)
  if ($null -eq $Value) { return '' }
  return $Value.Replace('\', '\\').Replace('"', '\"')
}

function Get-GitHubLatestTag {
  param([string] $Repo)
  $api = "https://api.github.com/repos/$Repo/releases/latest"
  $headers = @{ 'User-Agent' = 'nexctl-install-script' }
  try {
    $rel = Invoke-RestMethod -Uri $api -Headers $headers -TimeoutSec 15
    return [string]$rel.tag_name
  } catch {
    return $null
  }
}

function Test-ChinaRegion {
  foreach ($url in @(
      'https://dash.cloudflare.com/cdn-cgi/trace',
      'https://developers.cloudflare.com/cdn-cgi/trace',
      'https://1.0.0.1/cdn-cgi/trace'
    )) {
    try {
      $t = Invoke-WebRequest -Uri $url -TimeoutSec 8 -UseBasicParsing
      if ($t.Content -match 'loc=CN') { return $true }
    } catch { }
  }
  return $false
}

function Wrap-GitHubMirror {
  param(
    [string] $Url,
    [bool] $IsCnRegion
  )
  if ($env:NEXCTL_NO_MIRROR -eq '1') { return $Url }
  if (-not $IsCnRegion) { return $Url }
  if ($Url -notmatch '^https://github\.com/') { return $Url }
  $prefix = if ($env:NEXCTL_GH_PROXY) { $env:NEXCTL_GH_PROXY } else { 'https://ghproxy.net/https://' }
  $rest = $Url.Substring('https://'.Length)
  return $prefix + $rest
}

function Download-FileRetry {
  param([string] $Uri, [string] $OutFile)
  $last = $null
  for ($i = 1; $i -le 3; $i++) {
    try {
      Invoke-WebRequest -Uri $Uri -OutFile $OutFile -UseBasicParsing -TimeoutSec 120
      return
    } catch {
      $last = $_
      Write-WarnLine "下载失败 ($i/3)，重试..."
      Start-Sleep -Seconds 2
    }
  }
  throw $last
}

# ---------- 卸载 ----------
if ($Uninstall) {
  Write-Info '正在卸载 NexCtl Agent...'
  $svcPairs = @(
    @('C:\nexctl\nexctl-agent.exe', 'C:\nexctl\agent.yaml'),
    @((Join-Path $env:ProgramFiles 'nexctl\nexctl-agent.exe'), (Join-Path $env:ProgramFiles 'nexctl\agent.yaml'))
  )
  foreach ($pair in $svcPairs) {
    $exe, $cfg = $pair[0], $pair[1]
    if ((Test-Path -LiteralPath $exe) -and (Test-Path -LiteralPath $cfg)) {
      & $exe @('service', 'stop', '-config', $cfg) 2>$null
      & $exe @('service', 'uninstall', '-config', $cfg) 2>$null
    }
  }
  Get-Process -Name 'nexctl-agent' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  Start-Sleep -Seconds 1
  $paths = @('C:\nexctl', (Join-Path $env:ProgramFiles 'nexctl'))
  foreach ($p in $paths) {
    if (Test-Path -LiteralPath $p) {
      Remove-Item -LiteralPath $p -Recurse -Force -ErrorAction SilentlyContinue
      Write-Info "已移除: $p"
    }
  }
  $uc = Join-Path $env:USERPROFILE '.config\nexctl'
  if (Test-Path -LiteralPath $uc) {
    Remove-Item -LiteralPath $uc -Recurse -Force -ErrorAction SilentlyContinue
  }
  $ud = Join-Path $env:USERPROFILE '.local\share\nexctl'
  if (Test-Path -LiteralPath $ud) {
    Remove-Item -LiteralPath $ud -Recurse -Force -ErrorAction SilentlyContinue
  }
  $ub = Join-Path $env:USERPROFILE '.local\bin\nexctl-agent.exe'
  if (Test-Path -LiteralPath $ub) {
    Remove-Item -LiteralPath $ub -Force -ErrorAction SilentlyContinue
  }
  Write-Info '卸载完成。'
  exit 0
}

# ---------- 默认提权：以管理员安装到 C:\nexctl 并注册 Windows 服务 ----------
if (-not (Test-Administrator)) {
  if ($SkipElevation -or $env:NEXCTL_NO_ELEVATE -eq '1') {
    Write-WarnLine '当前非管理员：将安装到用户目录并以进程运行（未注册 Windows 服务）。'
  } else {
    $scriptPath = $MyInvocation.MyCommand.Path
    if ([string]::IsNullOrWhiteSpace($scriptPath)) {
      Write-Error '无法确定脚本路径，无法自动提权。请将脚本保存为 .ps1 后执行，或以管理员身份打开 PowerShell 再运行。'
      exit 1
    }
    Write-Info '请求管理员权限：安装到 C:\nexctl 并注册 Windows 服务（请在 UAC 中确认）...'
    $shellExe = $null
    try {
      $shellExe = (Get-Process -Id $PID -ErrorAction Stop).Path
    } catch {
      $shellExe = $null
    }
    if ([string]::IsNullOrWhiteSpace($shellExe) -or -not (Test-Path -LiteralPath $shellExe)) {
      if ($PSVersionTable.PSEdition -eq 'Core') {
        $cmd = Get-Command -Name 'pwsh.exe' -ErrorAction SilentlyContinue
        $shellExe = if ($cmd) { $cmd.Source } else { 'pwsh.exe' }
      } else {
        $shellExe = Join-Path $env:SystemRoot 'System32\WindowsPowerShell\v1.0\powershell.exe'
      }
    }
    $argList = @(
      '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', $scriptPath,
      '-ServerUrl', $ServerUrl,
      '-AgentId', $AgentId,
      '-AgentSecret', $AgentSecret,
      '-NodeKey', $NodeKey
    )
    if (-not [string]::IsNullOrWhiteSpace($NodeId)) {
      $argList += @('-NodeId', $NodeId)
    }
    if (-not [string]::IsNullOrWhiteSpace($ReleaseTag)) {
      $argList += @('-ReleaseTag', $ReleaseTag)
    }
    try {
      $elevated = Start-Process -FilePath $shellExe -Verb RunAs -ArgumentList $argList -PassThru -Wait
      $code = $elevated.ExitCode
      if ($null -eq $code) { $code = 0 }
      exit $code
    } catch {
      Write-Error @"
提升权限失败或 UAC 已取消。需要管理员权限才能将 Agent 注册为 Windows 服务。
请以管理员身份运行 PowerShell 后重新执行，或使用 -SkipElevation / 设置 NEXCTL_NO_ELEVATE=1 进行仅当前用户安装（无服务）。
$_
"@
      exit 1
    }
  }
}

$GithubRepo = if ($env:NEXCTL_AGENT_REPO) { $env:NEXCTL_AGENT_REPO } else { 'nexctl/agent' }

# ---------- 架构（64 位系统优先 amd64，与 nezha 一致）----------
if ([System.Environment]::Is64BitOperatingSystem) {
  if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') {
    Write-Error '当前仓库未提供 Windows ARM64 构建，请使用 x64 或自行编译。'
    exit 1
  }
  $GoArch = 'amd64'
} else {
  $GoArch = '386'
}

$ZipName = "nexctl_windows_$GoArch.zip"

# ---------- 安装目录（管理员：C:\nexctl，对齐 nezha 的 C:\nezha 单目录习惯）----------
$IsAdmin = Test-Administrator
if ($IsAdmin) {
  $BaseDir = 'C:\nexctl'
  $BinDir = $BaseDir
  $ConfigDir = $BaseDir
  $DataRoot = $BaseDir
} else {
  $BinDir = Join-Path $env:USERPROFILE '.local\bin'
  $ConfigDir = Join-Path $env:USERPROFILE '.config\nexctl'
  $DataRoot = Join-Path $env:USERPROFILE '.local\share\nexctl'
}

$ConfigFile = Join-Path $ConfigDir 'agent.yaml'
$BinName = 'nexctl-agent.exe'
$BinPath = Join-Path $BinDir $BinName

# ---------- 重装：结束旧进程并清理同路径（参考 nezha）----------
if (Test-Path -LiteralPath $BinPath) {
  Write-Info '检测到已安装的 nexctl-agent，先结束进程并覆盖...'
  Get-Process -Name 'nexctl-agent' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
  Start-Sleep -Seconds 1
}

try {
  $NodeName = [System.Net.Dns]::GetHostName()
} catch {
  $NodeName = 'nexctl-node'
}

# ---------- 解析下载 URL ----------
if ([string]::IsNullOrWhiteSpace($ReleaseTag)) {
  Write-Info '正在解析最新版本...'
  $tag = Get-GitHubLatestTag -Repo $GithubRepo
  if ([string]::IsNullOrWhiteSpace($tag)) {
    $ZipUrl = "https://github.com/$GithubRepo/releases/latest/download/$ZipName"
    Write-WarnLine 'GitHub API 不可用，使用 latest 直链。'
  } else {
    $ZipUrl = "https://github.com/$GithubRepo/releases/download/$tag/$ZipName"
    Write-Info "将安装版本: $tag"
  }
} else {
  $tag = $ReleaseTag
  if (-not $tag.StartsWith('v')) { $tag = "v$tag" }
  $ZipUrl = "https://github.com/$GithubRepo/releases/download/$tag/$ZipName"
}

$isCnRegion = ($env:CN -eq '1')
if (-not $isCnRegion) { $isCnRegion = Test-ChinaRegion }
$ZipUrl = Wrap-GitHubMirror -Url $ZipUrl -IsCnRegion $isCnRegion
if ($isCnRegion) {
  Write-WarnLine '当前可能位于中国大陆网络，已启用 GitHub 下载代理（NEXCTL_NO_MIRROR=1 可禁用）。'
}

$TmpRoot = Join-Path ([System.IO.Path]::GetTempPath()) ("nexctl-install-" + [Guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $TmpRoot -Force | Out-Null

try {
  Write-Info "正在下载: $ZipUrl"
  $ZipLocal = Join-Path $TmpRoot $ZipName
  Download-FileRetry -Uri $ZipUrl -OutFile $ZipLocal

  $ExtractDir = Join-Path $TmpRoot 'extract'
  New-Item -ItemType Directory -Path $ExtractDir -Force | Out-Null
  Expand-Archive -Path $ZipLocal -DestinationPath $ExtractDir -Force

  $ExtractedBin = Join-Path $ExtractDir $BinName
  if (-not (Test-Path -LiteralPath $ExtractedBin)) {
    $Alt = Get-ChildItem -Path $ExtractDir -Filter 'nexctl-agent.exe' -Recurse -ErrorAction SilentlyContinue | Select-Object -First 1
    if ($Alt) { $ExtractedBin = $Alt.FullName }
  }
  if (-not (Test-Path -LiteralPath $ExtractedBin)) {
    throw "压缩包内未找到 nexctl-agent.exe。目录: $ExtractDir"
  }

  $null = New-Item -ItemType Directory -Path $BinDir -Force
  $null = New-Item -ItemType Directory -Path $ConfigDir -Force
  $null = New-Item -ItemType Directory -Path (Join-Path $DataRoot 'data\config') -Force
  $null = New-Item -ItemType Directory -Path (Join-Path $DataRoot 'data\credentials') -Force
  $null = New-Item -ItemType Directory -Path (Join-Path $DataRoot 'data\logs') -Force

  Copy-Item -LiteralPath $ExtractedBin -Destination $BinPath -Force

  $DataDirFwd = ($DataRoot + '\data').Replace('\', '/')
  $CfgDirFwd = ($DataRoot + '\data\config').Replace('\', '/')
  $CredDirFwd = ($DataRoot + '\data\credentials').Replace('\', '/')
  $LogDirFwd = ($DataRoot + '\data\logs').Replace('\', '/')

  $nodeIdLine = ''
  if (-not [string]::IsNullOrWhiteSpace($NodeId)) {
    $nodeIdLine = "  node_id: $NodeId`n"
  }
  $yaml = @"
agent:
  server_url: "$(Escape-YamlDouble $ServerUrl)"
  agent_id: "$(Escape-YamlDouble $AgentId)"
  agent_secret: "$(Escape-YamlDouble $AgentSecret)"
  node_key: "$(Escape-YamlDouble $NodeKey)"
$nodeIdLine  node_name: "$(Escape-YamlDouble $NodeName)"
  data_dir: "$DataDirFwd"
  config_dir: "$CfgDirFwd"
  credential_dir: "$CredDirFwd"
  log_dir: "$LogDirFwd"
  github_repo: "$(Escape-YamlDouble $GithubRepo)"
  disable_auto_update: false
  self_update_period_minutes: 0
"@
  $utf8NoBom = New-Object System.Text.UTF8Encoding $false
  [System.IO.File]::WriteAllText($ConfigFile, $yaml, $utf8NoBom)

  if ($IsAdmin) {
    Set-Content -LiteralPath (Join-Path $BaseDir '.nexctl-install') -Value 'nexctl-install-script' -Encoding ascii
  }

  $ConfigFileAbs = [System.IO.Path]::GetFullPath($ConfigFile)
  $useService = $false
  if ($IsAdmin -and $env:NEXCTL_NO_SERVICE -ne '1') {
    & $BinPath @('service', 'install', '-config', $ConfigFileAbs)
    if ($LASTEXITCODE -eq 0) {
      & $BinPath @('service', 'start', '-config', $ConfigFileAbs)
      if ($LASTEXITCODE -eq 0) {
        $useService = $true
      } else {
        & $BinPath @('service', 'uninstall', '-config', $ConfigFileAbs) 2>$null
        Write-WarnLine '服务启动失败，已卸载服务并改用后台进程。'
      }
    } else {
      Write-WarnLine '注册 Windows 服务失败，改用后台进程启动（可设置 NEXCTL_NO_SERVICE=1 跳过服务步骤）。'
    }
  }

  $StdoutLog = Join-Path $DataRoot 'data\logs\install-stdout.log'
  $StderrLog = Join-Path $DataRoot 'data\logs\install-stderr.log'
  $proc = $null
  if (-not $useService) {
    $proc = Start-Process -FilePath $BinPath -ArgumentList @('-config', $ConfigFile) -PassThru -WindowStyle Hidden `
      -RedirectStandardOutput $StdoutLog -RedirectStandardError $StderrLog
    Start-Sleep -Seconds 1
    $still = Get-Process -Id $proc.Id -ErrorAction SilentlyContinue
    if (-not $still) {
      Write-Warning "agent 可能已退出，请查看 $StdoutLog、$StderrLog 与 $(Join-Path $DataRoot 'data\logs\agent.log')"
    }
  }

  Write-Info '安装完成。'
  Write-Host "  二进制: $BinPath"
  Write-Host "  配置: $ConfigFile"
  if ($useService) {
    Write-Host ('  已安装并启动 Windows 服务，日志: ' + (Join-Path $DataRoot 'data\logs'))
    Write-Host ("  管理示例: & '{0}' service status -config '{1}'" -f $BinPath, $ConfigFileAbs)
  } else {
    Write-Host "  PID: $($proc.Id)  日志: $(Join-Path $DataRoot 'data\logs')"
    Write-Host ''
    Write-Host '如需前台调试，可先结束进程再运行:'
    Write-Host ("  Stop-Process -Id {0} -ErrorAction SilentlyContinue" -f $proc.Id)
    Write-Host ("  & '{0}' -config '{1}'" -f $BinPath, $ConfigFile)
  }
  Write-Host ''

  if (-not $IsAdmin) {
    $binDirNorm = [System.IO.Path]::GetFullPath($BinDir).TrimEnd('\')
    $pathEnv = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($pathEnv -notlike "*$binDirNorm*") {
      Write-Host '若找不到 nexctl-agent.exe，可将目录加入用户 PATH:'
      $pathHint = [string]::Format('  [Environment]::SetEnvironmentVariable({0}Path{0}, "{1};{2}", {0}User{0})', [char]39, $BinDir, '$env:Path')
      Write-Host $pathHint
    }
  }
} finally {
  Remove-Item -LiteralPath $TmpRoot -Recurse -Force -ErrorAction SilentlyContinue
}
