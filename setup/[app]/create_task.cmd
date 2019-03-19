@echo off
net session>nul 2>&1
if %errorlevel%==0 (goto RUN) else ( goto MESSAGE )
:RUN
    SCHTASKS /Create /TN "Windows Firewall Tray Control" /SC ONLOGON /TR "\"%~dp0win_firewall_tray_control.exe\"" /RL HIGHEST /F
    goto end
:MESSAGE
    echo "You need to run this script with admin privileges"
    goto end
:end
    PAUSE