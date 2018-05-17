@echo off
net session>nul 2>&1
if %errorlevel%==0 (goto RUN) else ( goto MESSAGE )
:RUN
    SCHTASKS /Delete /TN "Windows Firewall Tray Control" /F
    goto end
:MESSAGE
    echo "You need to run this script with admin privileges"
    goto end
:end
    PAUSE