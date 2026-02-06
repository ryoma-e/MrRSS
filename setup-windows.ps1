# MrRSS Windows 安装脚本
# 请以管理员权限运行此脚本

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "MrRSS Windows 环境安装脚本" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

# 检查是否以管理员权限运行
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "警告: 建议以管理员权限运行此脚本以避免安装问题" -ForegroundColor Yellow
    Write-Host "右键点击此脚本 -> 以管理员身份运行" -ForegroundColor Yellow
    Write-Host ""
    $continue = Read-Host "是否继续? (y/n)"
    if ($continue -ne "y") {
        exit
    }
}

# 1. 检查 Go 安装
Write-Host "[1/5] 检查 Go 安装..." -ForegroundColor Green
$goInstalled = Get-Command go -ErrorAction SilentlyContinue
if ($goInstalled) {
    $goVersion = go version
    Write-Host "  ✓ Go 已安装: $goVersion" -ForegroundColor Green
} else {
    Write-Host "  ✗ Go 未安装" -ForegroundColor Red
    Write-Host ""
    Write-Host "请按照以下步骤安装 Go:" -ForegroundColor Yellow
    Write-Host "  1. 访问: https://go.dev/dl/" -ForegroundColor Yellow
    Write-Host "  2. 下载 go1.25.x.windows-amd64.msi (或最新版本)" -ForegroundColor Yellow
    Write-Host "  3. 运行安装程序，使用默认设置" -ForegroundColor Yellow
    Write-Host "  4. 安装完成后，重启 PowerShell 窗口" -ForegroundColor Yellow
    Write-Host "  5. 重新运行此脚本" -ForegroundColor Yellow
    Write-Host ""
    
    # 尝试使用 winget 安装（如果可用）
    $wingetInstalled = Get-Command winget -ErrorAction SilentlyContinue
    if ($wingetInstalled) {
        $installGo = Read-Host "检测到 winget，是否使用 winget 自动安装 Go? (y/n)"
        if ($installGo -eq "y") {
            Write-Host "  正在使用 winget 安装 Go..." -ForegroundColor Cyan
            winget install GoLang.Go.1.25
            Write-Host "  安装完成，请重启 PowerShell 并重新运行此脚本" -ForegroundColor Green
        }
    }
    
    exit
}

# 2. 检查 Node.js 安装
Write-Host "[2/5] 检查 Node.js 安装..." -ForegroundColor Green
$nodeInstalled = Get-Command node -ErrorAction SilentlyContinue
if ($nodeInstalled) {
    $nodeVersion = node --version
    Write-Host "  ✓ Node.js 已安装: $nodeVersion" -ForegroundColor Green
} else {
    Write-Host "  ✗ Node.js 未安装，请访问 https://nodejs.org/ 下载安装" -ForegroundColor Red
    exit
}

# 3. 安装 Wails v3 CLI
Write-Host "[3/5] 检查 Wails v3 CLI..." -ForegroundColor Green
$wailsInstalled = Get-Command wails3 -ErrorAction SilentlyContinue
if ($wailsInstalled) {
    $wailsVersion = wails3 version
    Write-Host "  ✓ Wails v3 已安装: $wailsVersion" -ForegroundColor Green
} else {
    Write-Host "  正在安装 Wails v3 CLI..." -ForegroundColor Cyan
    go install github.com/wailsapp/wails/v3/cmd/wails3@latest
    
    # 将 GOPATH\bin 添加到 PATH（如果尚未添加）
    $goPath = go env GOPATH
    $goBin = Join-Path $goPath "bin"
    
    if ($env:PATH -notlike "*$goBin*") {
        Write-Host "  正在将 $goBin 添加到 PATH..." -ForegroundColor Cyan
        [Environment]::SetEnvironmentVariable("Path", $env:PATH + ";$goBin", [EnvironmentVariableTarget]::User)
        $env:PATH += ";$goBin"
    }
    
    Write-Host "  ✓ Wails v3 安装完成" -ForegroundColor Green
}

# 4. 安装前端依赖
Write-Host "[4/5] 安装前端依赖..." -ForegroundColor Green
if (Test-Path "frontend/node_modules") {
    Write-Host "  ✓ 前端依赖已安装" -ForegroundColor Green
} else {
    Write-Host "  正在安装前端依赖..." -ForegroundColor Cyan
    Push-Location frontend
    npm install
    Pop-Location
    Write-Host "  ✓ 前端依赖安装完成" -ForegroundColor Green
}

# 5. 构建应用
Write-Host "[5/5] 准备构建..." -ForegroundColor Green
Write-Host ""
Write-Host "环境准备完成！" -ForegroundColor Green
Write-Host ""
Write-Host "下一步操作:" -ForegroundColor Cyan
Write-Host "  1. 开发模式（热重载）: wails3 dev" -ForegroundColor Yellow
Write-Host "  2. 构建应用: wails3 build" -ForegroundColor Yellow
Write-Host "  3. 构建后的应用位于: build\bin\MrRSS.exe" -ForegroundColor Yellow
Write-Host ""

$build = Read-Host "是否立即构建应用? (y/n)"
if ($build -eq "y") {
    Write-Host ""
    Write-Host "正在构建 MrRSS..." -ForegroundColor Cyan
    wails3 build
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✓ 构建成功！" -ForegroundColor Green
        Write-Host "可执行文件位于: build\bin\MrRSS.exe" -ForegroundColor Green
    } else {
        Write-Host ""
        Write-Host "✗ 构建失败，请检查错误信息" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "安装脚本执行完成" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
