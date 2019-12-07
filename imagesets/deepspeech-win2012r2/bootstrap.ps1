# use TLS 1.2 (see bug 1443595)
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# capture env
Get-ChildItem Env: | Out-File "C:\install_env.txt"

# needed for making http requests
$client = New-Object system.net.WebClient
$shell = new-object -com shell.application

# utility function to download a zip file and extract it
function Expand-ZIPFile($file, $destination, $url)
{
    $client.DownloadFile($url, $file)
    $zip = $shell.NameSpace($file)
    foreach($item in $zip.items())
    {
        $shell.Namespace($destination).copyhere($item)
    }
}

# allow powershell scripts to run
Set-ExecutionPolicy Unrestricted -Force -Scope Process

# Disable AV for IO speed
Set-Service "WinDefend" -StartupType Disabled -Status Stopped

# Disable disk indexing
Set-Service "WSearch" -StartupType Disabled -Status Stopped

# install chocolatey package manager
Invoke-Expression ($client.DownloadString('https://chocolatey.org/install.ps1'))

# install Windows 10 SDK
choco install -y windows-sdk-10.0

# install NodeJS v8
choco install -y nodejs --version 8.15.0

# install git
choco install -y git --version 2.21.0

# install python2 as well for node-gyp later
choco install -y python2 --version 2.7.16

# install python3.6
choco install -y python --version 3.6.8

# install 7zip, since msys2 p7zip behaves erratically
choco install -y 7zip --version 19.0

# install VisualStudio 2017 Community
choco install -y visualstudio2017community --version 15.9.7.0 --package-parameters "--add Microsoft.VisualStudio.Workload.MSBuildTools;Microsoft.VisualStudio.Component.VC.140 --passive --locale en-US"
choco install -y visualstudio2017buildtools --version 15.9.7.0 --package-parameters "--add Microsoft.VisualStudio.Workload.VCTools;includeRecommended --add Microsoft.VisualStudio.Component.VC.140 --add Microsoft.VisualStudio.Component.NuGet.BuildTools  --add Microsoft.Net.Component.4.5.TargetingPack --add Microsoft.Net.Component.4.6.TargetingPack --add Microsoft.Net.Component.4.7.TargetingPack --passive --locale en-US"

# vcredist140 required at least for bazel
choco install -y vcredist140 --version 14.16.27027.1

# .Net Framework v4.5.2
choco install -y netfx-4.5.2-devpack --version 4.5.5165101.20180721

# .Net Framework v4.6.2
choco install -y netfx-4.6.2-devpack --version 4.6.01590.20170129

# .Net Framework v4.7.2
choco install -y netfx-4.7.2-devpack --version 4.7.2.20190225

# NuGet
choco install -y nuget.commandline --version 4.9.3

# Carbon for later
choco install -y carbon --version 2.5.0

# Prepare CUDA v9.0
#$client.DownloadFile("https://developer.nvidia.com/compute/cuda/9.0/Prod/local_installers/cuda_9.0.176_win10-exe", "C:\cuda_9.0.176_win10.exe")
#Start-Process -FilePath "C:\cuda_9.0.176_win10.exe" -ArgumentList "-s compiler_9.0 command_line_tools_9.0 cublas_dev_9.0 cudart_9.0 cufft_dev_9.0 curand_dev_9.0 cusolver_dev_9.0 cusparse_dev_9.0" -Wait -NoNewWindow
$client.DownloadFile("https://developer.nvidia.com/compute/cuda/9.0/Prod/local_installers/cuda_9.0.176_windows-exe", "C:\cuda_9.0.176_windows.exe")
Start-Process -FilePath "C:\cuda_9.0.176_windows.exe" -ArgumentList "-s compiler_9.0 command_line_tools_9.0 cublas_dev_9.0 cudart_9.0 cufft_dev_9.0 curand_dev_9.0 cusolver_dev_9.0 cusparse_dev_9.0" -Wait -NoNewWindow

# CuDNN v7.3.1 for CUDA 9.0
#Expand-ZIPFile -File "C:\cudnn-9.0-windows10-x64-v7.3.1.20.zip" -Destination "C:\CUDNN-9.0\" -Url "http://developer.download.nvidia.com/compute/redist/cudnn/v7.3.1/cudnn-9.0-windows10-x64-v7.3.1.20.zip"
md "C:\CUDNN-9.0"
Expand-ZIPFile -File "C:\cudnn-9.0-windows7-x64-v7.3.1.20.zip" -Destination "C:\CUDNN-9.0\" -Url "http://developer.download.nvidia.com/compute/redist/cudnn/v7.3.1/cudnn-9.0-windows7-x64-v7.3.1.20.zip"
cp "C:\CUDNN-9.0\cuda\include\*" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v9.0\include\"
cp "C:\CUDNN-9.0\cuda\lib\x64\*" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v9.0\lib\x64\"
cp "C:\CUDNN-9.0\cuda\bin\*" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v9.0\bin\"

# Prepare CUDA v10.0
#$client.DownloadFile("https://developer.nvidia.com/compute/cuda/10.0/Prod/local_installers/cuda_10.0.130_411.31_win10", "C:\cuda_10.0.130_411.31_win10.exe")
#Start-Process -FilePath "C:\cuda_10.0.130_411.31_win10.exe" -ArgumentList "-s nvcc_10.0 cublas_dev_10.0 cudart_10.0 cufft_dev_10.0 curand_dev_10.0 cusolver_dev_10.0 cusparse_dev_10.0" -Wait -NoNewWindow
$client.DownloadFile("https://developer.nvidia.com/compute/cuda/10.0/Prod/local_installers/cuda_10.0.130_411.31_windows", "C:\cuda_10.0.130_411.31_windows.exe")
Start-Process -FilePath "C:\cuda_10.0.130_411.31_windows.exe" -ArgumentList "-s nvcc_10.0 nvprune_10.0 cupti_10.0 gpu_library_advisor_10.0 memcheck_10.0 cublas_dev_10.0 cudart_10.0 cufft_dev_10.0 curand_dev_10.0 cusolver_dev_10.0 cusparse_dev_10.0" -Wait -NoNewWindow

# Install CUDA v10.1 as well so we can patch v10.0 cudafe++.exe
# https://github.com/tensorflow/tensorflow/issues/27576#issuecomment-504703397
$client.DownloadFile("https://developer.nvidia.com/compute/cuda/10.1/Prod/local_installers/cuda_10.1.168_425.25_win10.exe", "C:\cuda_10.1.168_425.25_win10.exe")
Start-Process -FilePath "C:\cuda_10.1.168_425.25_win10.exe" -ArgumentList "-s nvcc_10.1 nvprune_10.1 cupti_10.1 gpu_library_advisor_10.1 memcheck_10.1 cublas_dev_10.1 cudart_10.1 cufft_dev_10.1 curand_dev_10.1 cusolver_dev_10.1 cusparse_dev_10.1" -Wait -NoNewWindow
mv "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.0\bin\cudafe++.exe" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.0\bin\cudafe++.exe.v10.0"
cp "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.1\bin\cudafe++.exe" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.0\bin\cudafe++.exe"

# CuDNN v7.5.0 for CUDA 10.0
#Expand-ZIPFile -File "C:\cudnn-10.0-windows10-x64-v7.5.0.56.zip" -Destination "C:\CUDNN-10.0\" -Url "http://developer.download.nvidia.com/compute/redist/cudnn/v7.5.0/cudnn-10.0-windows10-x64-v7.5.0.56.zip"
md "C:\CUDNN-10.0"
Expand-ZIPFile -File "C:\cudnn-10.0-windows7-x64-v7.5.0.56.zip" -Destination "C:\CUDNN-10.0\" -Url "http://developer.download.nvidia.com/compute/redist/cudnn/v7.5.0/cudnn-10.0-windows7-x64-v7.5.0.56.zip"
cp "C:\CUDNN-10.0\cuda\include\*" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.0\include\"
cp "C:\CUDNN-10.0\cuda\lib\x64\*" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.0\lib\x64\"
cp "C:\CUDNN-10.0\cuda\bin\*" "C:\Program Files\NVIDIA GPU Computing Toolkit\CUDA\v10.0\bin\"

# Create C:\builds and give full access to all users (for hg-shared, tooltool_cache, etc)
md "C:\builds"
$acl = Get-Acl -Path "C:\builds"
$ace = New-Object System.Security.AccessControl.FileSystemAccessRule("Everyone","Full","ContainerInherit,ObjectInherit","None","Allow")
$acl.AddAccessRule($ace)
Set-Acl "C:\builds" $acl

# GrantEveryoneSeCreateSymbolicLinkPrivilege
Start-Process "powershell" -ArgumentList "-command `"& {&'Import-Module' Carbon}`"; `"& {&'Grant-Privilege' -Identity Everyone -Privilege SeCreateSymbolicLinkPrivilege}`"" -Wait -NoNewWindow

# Ensure proper PATH setup
[Environment]::SetEnvironmentVariable("PATH", $Env:Path + ";C:\tools\msys64\usr\bin;C:\Python36;C:\Program Files\Git\bin", "Machine")

# install nssm, neded for generic-worker
Expand-ZIPFile -File "C:\nssm-2.24.zip" -Destination "C:\" -Url "http://www.nssm.cc/release/nssm-2.24.zip"

# download generic-worker
md C:\generic-worker
$client.DownloadFile("https://github.com/taskcluster/generic-worker/releases/download/v16.5.5/generic-worker-multiuser-windows-amd64.exe", "C:\generic-worker\generic-worker.exe")

# download livelog
$client.DownloadFile("https://github.com/taskcluster/livelog/releases/download/v1.1.0/livelog-windows-amd64.exe", "C:\generic-worker\livelog.exe")

# download taskcluster-proxy
$client.DownloadFile("https://github.com/taskcluster/taskcluster-proxy/releases/download/v5.1.0/taskcluster-proxy-windows-amd64.exe", "C:\generic-worker\taskcluster-proxy.exe")

# configure hosts file for taskcluster-proxy access via http://taskcluster
$HostsFile_Base64 = "IyBDb3B5cmlnaHQgKGMpIDE5OTMtMjAwOSBNaWNyb3NvZnQgQ29ycC4NCiMNCiMgVGhpcyBpcyBhIHNhbXBsZSBIT1NUUyBmaWxlIHVzZWQgYnkgTWljcm9zb2Z0IFRDUC9JUCBmb3IgV2luZG93cy4NCiMNCiMgVGhpcyBmaWxlIGNvbnRhaW5zIHRoZSBtYXBwaW5ncyBvZiBJUCBhZGRyZXNzZXMgdG8gaG9zdCBuYW1lcy4gRWFjaA0KIyBlbnRyeSBzaG91bGQgYmUga2VwdCBvbiBhbiBpbmRpdmlkdWFsIGxpbmUuIFRoZSBJUCBhZGRyZXNzIHNob3VsZA0KIyBiZSBwbGFjZWQgaW4gdGhlIGZpcnN0IGNvbHVtbiBmb2xsb3dlZCBieSB0aGUgY29ycmVzcG9uZGluZyBob3N0IG5hbWUuDQojIFRoZSBJUCBhZGRyZXNzIGFuZCB0aGUgaG9zdCBuYW1lIHNob3VsZCBiZSBzZXBhcmF0ZWQgYnkgYXQgbGVhc3Qgb25lDQojIHNwYWNlLg0KIw0KIyBBZGRpdGlvbmFsbHksIGNvbW1lbnRzIChzdWNoIGFzIHRoZXNlKSBtYXkgYmUgaW5zZXJ0ZWQgb24gaW5kaXZpZHVhbA0KIyBsaW5lcyBvciBmb2xsb3dpbmcgdGhlIG1hY2hpbmUgbmFtZSBkZW5vdGVkIGJ5IGEgJyMnIHN5bWJvbC4NCiMNCiMgRm9yIGV4YW1wbGU6DQojDQojICAgICAgMTAyLjU0Ljk0Ljk3ICAgICByaGluby5hY21lLmNvbSAgICAgICAgICAjIHNvdXJjZSBzZXJ2ZXINCiMgICAgICAgMzguMjUuNjMuMTAgICAgIHguYWNtZS5jb20gICAgICAgICAgICAgICMgeCBjbGllbnQgaG9zdA0KDQojIGxvY2FsaG9zdCBuYW1lIHJlc29sdXRpb24gaXMgaGFuZGxlZCB3aXRoaW4gRE5TIGl0c2VsZi4NCiMJMTI3LjAuMC4xICAgICAgIGxvY2FsaG9zdA0KIwk6OjEgICAgICAgICAgICAgbG9jYWxob3N0DQoNCiMgVXNlZnVsIGZvciBnZW5lcmljLXdvcmtlciB0YXNrY2x1c3Rlci1wcm94eSBpbnRlZ3JhdGlvbg0KIyBTZWUgaHR0cHM6Ly9idWd6aWxsYS5tb3ppbGxhLm9yZy9zaG93X2J1Zy5jZ2k/aWQ9MTQ0OTk4MSNjNg0KMTI3LjAuMC4xICAgICAgICB0YXNrY2x1c3RlciAgICANCg=="
$HostsFile_Content = [System.Convert]::FromBase64String($HostsFile_Base64)
Set-Content -Path "C:\Windows\System32\drivers\etc\hosts" -Value $HostsFile_Content -Encoding Byte

# install generic-worker
Start-Process C:\generic-worker\generic-worker.exe -ArgumentList "install service --configure-for-%MY_CLOUD% --nssm C:\nssm-2.24\win64\nssm.exe --config C:\generic-worker\generic-worker.config" -Wait -NoNewWindow -PassThru -RedirectStandardOutput C:\generic-worker\install.log -RedirectStandardError C:\generic-worker\install.err
# Start-Process C:\generic-worker\generic-worker.exe -ArgumentList "install startup --config C:\generic-worker\generic-worker.config" -Wait -NoNewWindow -PassThru -RedirectStandardOutput C:\generic-worker\install.log -RedirectStandardError C:\generic-worker\install.err

# download Windows Server 2003 Resource Kit Tools
$client.DownloadFile("https://download.microsoft.com/download/8/e/c/8ec3a7d8-05b4-440a-a71e-ca3ee25fe057/rktools.exe", "C:\rktools.exe")

# open up firewall for livelog (both PUT and GET interfaces)
New-NetFirewallRule -DisplayName "Allow livelog PUT requests" -Direction Inbound -LocalPort 60022 -Protocol TCP -Action Allow
New-NetFirewallRule -DisplayName "Allow livelog GET requests" -Direction Inbound -LocalPort 60023 -Protocol TCP -Action Allow

# install PSTools
# md "C:\PSTools"
# Expand-ZIPFile -File "C:\PSTools\PSTools.zip" -Destination "C:\PSTools" -Url "https://download.sysinternals.com/files/PSTools.zip"

# generate OpenPGP key
Start-Process C:\generic-worker\generic-worker.exe -ArgumentList "new-openpgp-keypair --file C:\generic-worker\generic-worker-gpg-signing-key.key" -Wait -NoNewWindow -PassThru -RedirectStandardOutput C:\generic-worker\generate-gpg-signing-key.log -RedirectStandardError C:\generic-worker\generate-gpg-signing-key.err

# generate ed25519 key
Start-Process C:\generic-worker\generic-worker.exe -ArgumentList "new-ed25519-keypair --file C:\generic-worker\generic-worker-ed25519-signing-key.key" -Wait -NoNewWindow -PassThru -RedirectStandardOutput C:\generic-worker\generate-signing-key.log -RedirectStandardError C:\generic-worker\generate-signing-key.err

# install dependencywalker (useful utility for troubleshooting, not required)
md "C:\DependencyWalker"
Expand-ZIPFile -File "C:\depends22_x64.zip" -Destination "C:\DependencyWalker" -Url "http://dependencywalker.com/depends22_x64.zip"

# install ProcessExplorer (useful utility for troubleshooting, not required)
md "C:\ProcessExplorer"
Expand-ZIPFile -File "C:\ProcessExplorer.zip" -Destination "C:\ProcessExplorer" -Url "https://download.sysinternals.com/files/ProcessExplorer.zip"

# install ProcessMonitor (useful utility for troubleshooting, not required)
md "C:\ProcessMonitor"
Expand-ZIPFile -File "C:\ProcessMonitor.zip" -Destination "C:\ProcessMonitor" -Url "https://download.sysinternals.com/files/ProcessMonitor.zip"

# install handle
md "C:\Handle"
Expand-ZIPFile -File "C:\Handle.zip" -Destination "C:\Handle" -Url "https://download.sysinternals.com/files/Handle.zip"

# Free some space
Start-Process "cmd.exe" -ArgumentList "/c del C:\cuda_*" -Wait -NoNewWindow
Start-Process "cmd.exe" -ArgumentList "/c del C:\cudnn*" -Wait -NoNewWindow
Start-Process "cmd.exe" -ArgumentList "/c del C:\CUDNN*" -Wait -NoNewWindow

$computersys = Get-WmiObject Win32_ComputerSystem -EnableAllPrivileges;
$computersys.AutomaticManagedPagefile = $False;
$computersys.Put();
$pagefile = Get-WmiObject -Query "Select * From Win32_PageFileSetting Where Name like '%pagefile.sys'";
$pagefile.InitialSize = 512;
$pagefile.MaximumSize = 2048;
$pagefile.Put();

# now shutdown, in preparation for creating an image
# Stop-Computer isn't working, also not when specifying -AsJob, so reverting to using `shutdown` command instead
#   * https://www.reddit.com/r/PowerShell/comments/65250s/windows_10_creators_update_stopcomputer_not/dgfofug/?st=j1o3oa29&sh=e0c29c6d
#   * https://support.microsoft.com/en-in/help/4014551/description-of-the-security-and-quality-rollup-for-the-net-framework-4
#   * https://support.microsoft.com/en-us/help/4020459
shutdown -s
