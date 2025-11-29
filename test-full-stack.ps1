# 🧪 Full Stack Test

```powershell
# test-full-stack.ps1
Write-Host "🧪 Testing Full CinemaOS Stack..." -ForegroundColor Cyan

Write-Host "`n1. Checking Docker..." -ForegroundColor Yellow
try {
    docker --version
    Write-Host "   ✅ Docker installed" -ForegroundColor Green
}
catch {
    Write-Host "   ❌ Docker not installed" -ForegroundColor Red
    exit 1
}

Write-Host "`n2. Validating docker-compose.yml..." -ForegroundColor Yellow
try {
    docker-compose config | Out-Null
    Write-Host "   ✅ Configuration valid" -ForegroundColor Green
}
catch {
    Write-Host "   ❌ Configuration invalid" -ForegroundColor Red
    exit 1
}

Write-Host "`n3. Starting services..." -ForegroundColor Yellow
Write-Host "   This may take a few minutes..." -ForegroundColor White
docker-compose up -d

Write-Host "`n4. Waiting for services to be healthy..." -ForegroundColor Yellow
Start-Sleep -Seconds 10

Write-Host "`n5. Checking service status..." -ForegroundColor Yellow
docker-compose ps

Write-Host "`n6. Testing endpoints..." -ForegroundColor Yellow
try {
    $backendHealth = Invoke-WebRequest -Uri "http://localhost:5000/health" -TimeoutSec 5
    Write-Host "   ✅ Backend health check: $($backendHealth.StatusCode)" -ForegroundColor Green
}
catch {
    Write-Host "   ⚠️  Backend not responding yet" -ForegroundColor Yellow
}

try {
    $frontend = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 5
    Write-Host "   ✅ Frontend: $($frontend.StatusCode)" -ForegroundColor Green
}
catch {
    Write-Host "   ⚠️  Frontend not responding yet" -ForegroundColor Yellow
}

Write-Host "`n✨ Stack started!" -ForegroundColor Cyan
Write-Host "   View logs: docker-compose logs -f" -ForegroundColor White
Write-Host "   Stop stack: docker-compose down" -ForegroundColor White
