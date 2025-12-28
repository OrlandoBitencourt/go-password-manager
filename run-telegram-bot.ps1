# PowerShell script to run Telegram bot with .env file

# Check if .env exists
if (-not (Test-Path ".env")) {
    Write-Host "Error: .env file not found!" -ForegroundColor Red
    Write-Host "Please copy .env.example to .env and add your bot token" -ForegroundColor Yellow
    exit 1
}

# Load .env file
Get-Content .env | ForEach-Object {
    if ($_ -match '^([^=]+)=(.*)$') {
        $key = $matches[1].Trim()
        $value = $matches[2].Trim()

        # Remove quotes if present
        $value = $value -replace '^["'']|["'']$'

        # Set environment variable
        [Environment]::SetEnvironmentVariable($key, $value, "Process")
        Write-Host "Loaded: $key" -ForegroundColor Green
    }
}

# Check if TELEGRAM_BOT_TOKEN is set
if (-not $env:TELEGRAM_BOT_TOKEN) {
    Write-Host "Error: TELEGRAM_BOT_TOKEN not found in .env file!" -ForegroundColor Red
    exit 1
}

Write-Host "`nStarting Telegram Bot..." -ForegroundColor Cyan
go run cmd/telegram-bot/main.go
