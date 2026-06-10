param(
    [switch]$Preview
)

if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
    $arch = "amd64"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
} else {
    Write-Host "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
    exit
}

if ($Preview) {
    $releases = (iwr https://api.github.com/repos/Azure/azqr/releases).content | ConvertFrom-Json
    $preRelease = $releases | Where-Object { $_.prerelease -eq $true } | Select-Object -First 1
    if (-not $preRelease) {
        Write-Host "No preview version available."
        exit 1
    }
    $latest_azqr = $preRelease.tag_name
    Write-Host "Installing preview version: $latest_azqr"
} else {
    $latest_azqr = (iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | ConvertFrom-Json | Select-Object -ExpandProperty tag_name
}

iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-win-$arch.zip -OutFile azqr.zip
Expand-Archive -Path azqr.zip -DestinationPath ./azqr_bin
Get-ChildItem -Path ./azqr_bin -Recurse -File | ForEach-Object { Move-Item -Path $_.FullName -Destination . -Force }
Remove-Item -Path ./azqr_bin -Recurse -Force
Remove-Item -Path azqr.zip
.\azqr.exe --version