#Requires -Version 5.1
<#
.SYNOPSIS
  从 GitHub Releases 安装 NexCtl Agent（Windows），参考 nezhahq/scripts/agent 的体验：
  TLS 1.2、地域镜像、GitHub API 解析版本、下载重试、重装前清理。

.EXAMPLE
  .\install.ps1 -ServerUrl "http://控制面:8080" -EnrollmentToken "节点令牌"

.EXAMPLE
  .\install.ps1 -Uninstall

  环境变量：NEXCTL_AGENT_REPO、CN=1、NEXCTL_NO_MIRROR=1、NEXCTL_GH_PROXY
#>

[CmdletBinding(DefaultParameterSetName = 'Install')]
param(
  [Parameter(ParameterSetName = 'Install', Mandatory = $true, Position = 0)]
  [string] $ServerUrl,

  [Parameter(ParameterSetName = 'Install', Mandatory = $true, Position = 1)]
  [string] $EnrollmentToken,

  [Parameter(ParameterSetName = 'Install', Mandatory = $false, Position = 2)]
  [string] $ReleaseTag = '',

  [Parameter(ParameterSetName = 'Uninstall', Mandatory = $true)]
  [switch] $Uninstall
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

  $yaml = @"
agent:
  server_url: "$(Escape-YamlDouble $ServerUrl)"
  enrollment_token: "$(Escape-YamlDouble $EnrollmentToken)"
  node_name: "$(Escape-YamlDouble $NodeName)"
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

  $StdoutLog = Join-Path $DataRoot 'data\logs\install-stdout.log'
  $StderrLog = Join-Path $DataRoot 'data\logs\install-stderr.log'
  $proc = Start-Process -FilePath $BinPath -ArgumentList @('-config', $ConfigFile) -PassThru -WindowStyle Hidden `
    -RedirectStandardOutput $StdoutLog -RedirectStandardError $StderrLog

  Start-Sleep -Seconds 1
  $still = Get-Process -Id $proc.Id -ErrorAction SilentlyContinue
  if (-not $still) {
    Write-Warning "agent 可能已退出，请查看 $StdoutLog、$StderrLog 与 $(Join-Path $DataRoot 'data\logs\agent.log')"
  }

  Write-Info '安装完成。'
  Write-Host "  二进制: $BinPath"
  Write-Host "  配置: $ConfigFile"
  Write-Host "  PID: $($proc.Id)  日志: $(Join-Path $DataRoot 'data\logs')"
  Write-Host ''
  Write-Host '如需前台调试，可先结束进程再运行:'
  Write-Host ("  Stop-Process -Id {0} -ErrorAction SilentlyContinue" -f $proc.Id)
  Write-Host ("  & '{0}' -config '{1}'" -f $BinPath, $ConfigFile)
  Write-Host ''

  if (-not $IsAdmin) {
    $binDirNorm = [System.IO.Path]::GetFullPath($BinDir).TrimEnd('\')
    $pathEnv = [Environment]::GetEnvironmentVariable('Path', 'User')
    if ($pathEnv -notlike "*$binDirNorm*") {
      Write-Host '若找不到 nexctl-agent.exe，可将目录加入用户 PATH:'
      Write-Host "  [Environment]::SetEnvironmentVariable('Path', `"$BinDir;`$env:Path`", 'User')"
    }
  }
} finally {
  Remove-Item -LiteralPath $TmpRoot -Recurse -Force -ErrorAction SilentlyContinue
}
