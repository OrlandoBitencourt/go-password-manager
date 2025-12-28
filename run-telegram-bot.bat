@echo off
REM Batch script to run Telegram bot with .env file

if not exist .env (
    echo Error: .env file not found!
    echo Please copy .env.example to .env and add your bot token
    exit /b 1
)

echo Loading environment variables from .env...

REM Load .env file
for /f "usebackq tokens=1,* delims==" %%a in (".env") do (
    set %%a=%%b
    echo Loaded: %%a
)

if "%TELEGRAM_BOT_TOKEN%"=="" (
    echo Error: TELEGRAM_BOT_TOKEN not found in .env file!
    exit /b 1
)

echo.
echo Starting Telegram Bot...
go run cmd/telegram-bot/main.go
