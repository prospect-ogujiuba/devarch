[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [string]$BlockBase64,

    [Parameter(Mandatory = $false)]
    [string]$HostsPath,

    [switch]$Elevated
)

$ErrorActionPreference = 'Stop'
$beginMarker = '# BEGIN DEVARCH HOSTS'
$endMarker = '# END DEVARCH HOSTS'
$block = [Text.Encoding]::UTF8.GetString([Convert]::FromBase64String($BlockBase64)).TrimEnd("`r", "`n")
if (-not $block.StartsWith($beginMarker) -or -not $block.EndsWith($endMarker)) {
    throw 'Generated block does not contain valid DevArch markers.'
}

$systemHostsPath = Join-Path $env:SystemRoot 'System32\drivers\etc\hosts'
$hostsPath = if ($HostsPath) { $HostsPath } else { $systemHostsPath }
$editingSystemHosts = -not $HostsPath -or
    [IO.Path]::GetFullPath($hostsPath) -eq [IO.Path]::GetFullPath($systemHostsPath)

$bytes = [IO.File]::ReadAllBytes($hostsPath)
$preambleLength = 0
if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
    $encoding = [Text.UTF8Encoding]::new($true)
    $preambleLength = 3
}
elseif ($bytes.Length -ge 2 -and $bytes[0] -eq 0xFF -and $bytes[1] -eq 0xFE) {
    $encoding = [Text.Encoding]::Unicode
    $preambleLength = 2
}
elseif ($bytes.Length -ge 2 -and $bytes[0] -eq 0xFE -and $bytes[1] -eq 0xFF) {
    $encoding = [Text.Encoding]::BigEndianUnicode
    $preambleLength = 2
}
else {
    $encoding = [Text.Encoding]::Default
}
$content = $encoding.GetString($bytes, $preambleLength, $bytes.Length - $preambleLength)
$newline = if ($content.Contains("`r`n")) { "`r`n" } elseif ($content.Contains("`n")) { "`n" } else { [Environment]::NewLine }
$hadTrailingNewline = $content.EndsWith("`r`n") -or $content.EndsWith("`n") -or $content.EndsWith("`r")
$lines = [Collections.Generic.List[string]]::new()
foreach ($line in [regex]::Split($content, '\r\n|\n|\r')) { $lines.Add($line) }
if ($hadTrailingNewline -and $lines.Count -gt 0 -and $lines[$lines.Count - 1] -eq '') { $lines.RemoveAt($lines.Count - 1) }

$beginIndexes = @()
$endIndexes = @()
for ($index = 0; $index -lt $lines.Count; $index++) {
    if ($lines[$index] -eq $beginMarker) { $beginIndexes += $index }
    if ($lines[$index] -eq $endMarker) { $endIndexes += $index }
}
if ($beginIndexes.Count -ne $endIndexes.Count -or $beginIndexes.Count -gt 1 -or
    ($beginIndexes.Count -eq 1 -and $beginIndexes[0] -ge $endIndexes[0])) {
    throw 'Hosts file contains malformed or duplicate DevArch managed-block markers.'
}

$blockLines = @([regex]::Split($block, '\r\n|\n|\r'))
$updated = [Collections.Generic.List[string]]::new()
if ($beginIndexes.Count -eq 1) {
    for ($index = 0; $index -lt $beginIndexes[0]; $index++) { $updated.Add($lines[$index]) }
    foreach ($line in $blockLines) { $updated.Add($line) }
    for ($index = $endIndexes[0] + 1; $index -lt $lines.Count; $index++) { $updated.Add($lines[$index]) }
}
else {
    foreach ($line in $lines) { $updated.Add($line) }
    if ($updated.Count -gt 0 -and $updated[$updated.Count - 1] -ne '') { $updated.Add('') }
    foreach ($line in $blockLines) { $updated.Add($line) }
}
$newContent = [string]::Join($newline, $updated)
if ($hadTrailingNewline -or $content.Length -eq 0) { $newContent += $newline }
if ($newContent -eq $content) {
    Write-Host "[hosts] managed block already current in $hostsPath"
    exit 0
}

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]::new($identity)
$isAdministrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($editingSystemHosts -and -not $isAdministrator) {
    if ($Elevated) { throw 'Administrator access was requested but is not available.' }
    $arguments = @(
        '-NoProfile', '-ExecutionPolicy', 'Bypass', '-File', ('"{0}"' -f $PSCommandPath),
        '-BlockBase64', ('"{0}"' -f $BlockBase64), '-Elevated'
    )
    $process = Start-Process -FilePath 'powershell.exe' -Verb RunAs -Wait -PassThru -ArgumentList $arguments
    if ($process.ExitCode -ne 0) { throw "Elevated hosts update failed with exit code $($process.ExitCode)." }
    exit 0
}

$bodyBytes = $encoding.GetBytes($newContent)
if ($preambleLength -gt 0) {
    $preamble = $encoding.GetPreamble()
    $outputBytes = [byte[]]::new($preamble.Length + $bodyBytes.Length)
    [Array]::Copy($preamble, 0, $outputBytes, 0, $preamble.Length)
    [Array]::Copy($bodyBytes, 0, $outputBytes, $preamble.Length, $bodyBytes.Length)
}
else { $outputBytes = $bodyBytes }
[IO.File]::WriteAllBytes($hostsPath, $outputBytes)
Write-Host "[hosts] synchronized the DevArch managed block in $hostsPath"
