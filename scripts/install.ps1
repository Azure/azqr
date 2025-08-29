if ($env:PROCESSOR_ARCHITECTURE -eq "AMD64") {
    $arch = "amd64"
} elseif ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $arch = "arm64"
} else {
    Write-Host "Unsupported architecture: $($env:PROCESSOR_ARCHITECTURE)"
    exit
}

$latest_azqr=$(iwr https://api.github.com/repos/Azure/azqr/releases/latest).content | convertfrom-json | Select-Object -ExpandProperty tag_name
iwr https://github.com/Azure/azqr/releases/download/$latest_azqr/azqr-win-$arch.zip -OutFile azqr.zip
Expand-Archive -Path azqr.zip -DestinationPath ./azqr_bin
Get-ChildItem -Path ./azqr_bin -Recurse -File | ForEach-Object { Move-Item -Path $_.FullName -Destination . -Force }
Remove-Item -Path ./azqr_bin -Recurse -Force
Remove-Item -Path azqr.zip
.\azqr.exe --version