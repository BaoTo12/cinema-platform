# 🧪 CinemaOS Test Scripts

## Quick Test Scripts

### 1️⃣ Test Frontend (Windows PowerShell)

```powershell
# test-frontend.ps1
Write-Host "🧪 Testing CinemaOS Frontend..." -ForegroundColor Cyan

# Check if dev server is running
Write-Host "`n1. Checking dev server..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:3000" -TimeoutSec 5
    Write-Host "✅ Frontend is running!" -ForegroundColor Green
    Write-Host "   Status Code: $($response.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "❌ Frontend not responding" -ForegroundColor Red
    Write-Host "   Run: cd frontend; npm run dev" -ForegroundColor Yellow
    exit 1
}

# Test health of pages
Write-Host "`n2. Testing pages..." -ForegroundColor Yellow
$pages = @("/", "/movies", "/login", "/register")

foreach ($page in $pages) {
    try {
        $response = Invoke-WebRequest -Uri "http://localhost:3000$page" -TimeoutSec 5
        Write-Host "   ✅ $page - OK" -ForegroundColor Green
    } catch {
        Write-Host "   ❌ $page - FAILED" -ForegroundColor Red
    }
}

Write-Host "`n✨ Frontend tests complete!" -ForegroundColor Cyan
Write-Host "   Open in browser: http://localhost:3000" -ForegroundColor White
