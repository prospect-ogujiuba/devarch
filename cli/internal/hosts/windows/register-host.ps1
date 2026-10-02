[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)+$')]
    [string]$HostName,

    [Parameter(Mandatory = $false)]
    [string]$Address = '127.0.0.1',

    [Parameter(Mandatory = $false)]
    [string]$HostsPath,

    [switch]$Elevated
)

$ErrorActionPreference = 'Stop'
$parsedAddress = $null
if (-not [System.Net.IPAddress]::TryParse($Address, [ref]$parsedAddress) -or
    $parsedAddress.AddressFamily -ne [System.Net.Sockets.AddressFamily]::InterNetwork) {
    throw 'Address must be an IPv4 address.'
}

$systemHostsPath = Join-Path $env:SystemRoot 'System32\drivers\etc\hosts'
$hostsPath = if ($HostsPath) { $HostsPath } else { $systemHostsPath }
$editingSystemHosts = -not $HostsPath -or
    [System.IO.Path]::GetFullPath($hostsPath) -eq [System.IO.Path]::GetFullPath($systemHostsPath)

$bytes = [System.IO.File]::ReadAllBytes($hostsPath)
$preambleLength = 0
if ($bytes.Length -ge 3 -and $bytes[0] -eq 0xEF -and $bytes[1] -eq 0xBB -and $bytes[2] -eq 0xBF) {
    $encoding = [System.Text.UTF8Encoding]::new($true)
    $preambleLength = 3
}
elseif ($bytes.Length -ge 2 -and $bytes[0] -eq 0xFF -and $bytes[1] -eq 0xFE) {
    $encoding = [System.Text.Encoding]::Unicode
    $preambleLength = 2
}
elseif ($bytes.Length -ge 2 -and $bytes[0] -eq 0xFE -and $bytes[1] -eq 0xFF) {
    $encoding = [System.Text.Encoding]::BigEndianUnicode
    $preambleLength = 2
}
else {
    $encoding = [System.Text.Encoding]::Default
}
$content = $encoding.GetString($bytes, $preambleLength, $bytes.Length - $preambleLength)
$newline = if ($content.Contains("`r`n")) { "`r`n" } elseif ($content.Contains("`n")) { "`n" } else { [Environment]::NewLine }
$hadTrailingNewline = $content.EndsWith("`r`n") -or $content.EndsWith("`n") -or $content.EndsWith("`r")
$lines = [System.Collections.Generic.List[string]]::new()
foreach ($line in [regex]::Split($content, '\r\n|\n|\r')) { $lines.Add($line) }
if ($hadTrailingNewline -and $lines.Count -gt 0 -and $lines[$lines.Count - 1] -eq '') {
    $lines.RemoveAt($lines.Count - 1)
}

$matchingLines = [System.Collections.Generic.List[object]]::new()
for ($index = 0; $index -lt $lines.Count; $index++) {
    $line = $lines[$index]
    if ($line -match '^\s*(#|$)') { continue }
    $body = ($line -split '#', 2)[0]
    $fields = @($body -split '\s+' | Where-Object { $_ })
    if ($fields.Count -lt 2) { continue }
    $aliases = @($fields | Select-Object -Skip 1)
    if (@($aliases | Where-Object { $_ -ieq $HostName }).Count -gt 0) {
        $matchingLines.Add([pscustomobject]@{ Index = $index; Address = $fields[0]; Aliases = $aliases })
    }
}

if ($matchingLines.Count -eq 1 -and
    $matchingLines[0].Address -eq $Address -and
    $matchingLines[0].Aliases.Count -eq 1) {
    Write-Host "[hosts] already registered $Address $HostName in $hostsPath"
    exit 0
}

$identity = [Security.Principal.WindowsIdentity]::GetCurrent()
$principal = [Security.Principal.WindowsPrincipal]::new($identity)
$isAdministrator = $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if ($editingSystemHosts -and -not $isAdministrator) {
    if ($Elevated) {
        throw 'Administrator access was requested but is not available.'
    }

    $arguments = @(
        '-NoProfile',
        '-ExecutionPolicy', 'Bypass',
        '-File', ('"{0}"' -f $PSCommandPath),
        '-HostName', $HostName,
        '-Address', $Address,
        '-Elevated'
    )
    $process = Start-Process -FilePath 'powershell.exe' -Verb RunAs -Wait -PassThru -ArgumentList $arguments
    if ($process.ExitCode -ne 0) {
        throw "Elevated hosts update failed with exit code $($process.ExitCode)."
    }
    exit 0
}

$escapedHostName = [regex]::Escape($HostName)
$updated = [System.Collections.Generic.List[string]]::new()
for ($index = 0; $index -lt $lines.Count; $index++) {
    $line = $lines[$index]
    $match = @($matchingLines | Where-Object { $_.Index -eq $index })
    if ($match.Count -eq 0) {
        $updated.Add($line)
        continue
    }

    $commentIndex = $line.IndexOf('#')
    $comment = if ($commentIndex -ge 0) { $line.Substring($commentIndex) } else { '' }
    $body = if ($commentIndex -ge 0) { $line.Substring(0, $commentIndex) } else { $line }
    $body = [regex]::Replace($body, "(?i)(?<!\S)$escapedHostName[ `t]+", '')
    $body = [regex]::Replace($body, "(?i)[ `t]+$escapedHostName(?=[ `t]*$)", '')
    $remainingFields = @($body -split '\s+' | Where-Object { $_ })
    if ($remainingFields.Count -gt 1) {
        $updated.Add($body + $comment)
    }
    elseif ($comment) {
        $updated.Add($comment)
    }
}
$updated.Add("$Address`t$HostName")

$newContent = [string]::Join($newline, $updated)
if ($hadTrailingNewline) { $newContent += $newline }
$bodyBytes = $encoding.GetBytes($newContent)
if ($preambleLength -gt 0) {
    $preamble = $encoding.GetPreamble()
    $outputBytes = [byte[]]::new($preamble.Length + $bodyBytes.Length)
    [Array]::Copy($preamble, 0, $outputBytes, 0, $preamble.Length)
    [Array]::Copy($bodyBytes, 0, $outputBytes, $preamble.Length, $bodyBytes.Length)
}
else {
    $outputBytes = $bodyBytes
}
[System.IO.File]::WriteAllBytes($hostsPath, $outputBytes)
Write-Host "[hosts] registered $Address $HostName in $hostsPath"
