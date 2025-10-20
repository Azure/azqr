if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
    $arch = "amd64"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
} else {
    Write-Host "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
    exit
}

$latest_azqr=$(iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | convertfrom-json | Select-Object -ExpandProperty tag_name
iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-windows-$arch.exe -OutFile azqr.exe
.\azqr.exe --version