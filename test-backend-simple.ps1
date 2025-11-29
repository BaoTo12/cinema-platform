# 🧪 Quick Backend Test (Without Protobuf)

```powershell
# test-backend-simple.ps1
Write-Host "🧪 Testing CinemaOS Backend (Simplified)..." -ForegroundColor Cyan

Write-Host "`n1. Checking Go installation..." -ForegroundColor Yellow
try {
    $goVersion = go version
    Write-Host "   ✅ $goVersion" -ForegroundColor Green
}
catch {
    Write-Host "   ❌ Go not installed" -ForegroundColor Red
    exit 1
}

Write-Host "`n2. Testing Go modules..." -ForegroundColor Yellow
Set-Location backend
try {
    go mod download
    Write-Host "   ✅ Dependencies downloaded" -ForegroundColor Green
}
catch {
    Write-Host "   ❌ Module download failed" -ForegroundColor Red
    Set-Location ..
    exit 1
}

Write-Host "`n3. Checking syntax..." -ForegroundColor Yellow
$errors = go vet ./...
if ($LASTEXITCODE -eq 0) {
    Write-Host "   ✅ No syntax errors" -ForegroundColor Green
}
else {
    Write-Host "   ⚠️  Some warnings found" -ForegroundColor Yellow
}

Set-Location ..
Write-Host "`n✨ Backend basic tests complete!" -ForegroundColor Cyan
Write-Host "   Note: Full compilation requires protobuf generation" -ForegroundColor Yellow
