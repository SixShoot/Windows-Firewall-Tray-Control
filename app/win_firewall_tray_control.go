//go:generate goversioninfo -icon=icon.ico -manifest=manifest.manifest
package main

import (
	"log"
	"os"
	"os/exec"
	"syscall"
	"io/ioutil"
	"path/filepath"
	"bytes"
	"strings"
	
	"github.com/getlantern/systray"
)

func main() {
    systray.Run(onReady, onExit)
}

func IswinFirewallEnabled() bool {

    cmd := exec.Command("powershell", "get-netfirewallprofile", "|", "select", "name,enabled")
    cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    
    var outb, errb bytes.Buffer
    cmd.Stdout = &outb
    cmd.Stderr = &errb
    
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
    
    if strings.Contains(outb.String(), "True") {
      return true
    } else {
      return false
    }

}

func setFirewall(state bool) {
    var s string
    if state {
      s = "on"
    } else {
      s = "off"
    }

    cmd := exec.Command("cmd", "/c", "netsh", "advfirewall", "set", "allprofiles", "state", s)
    cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    err := cmd.Run()
    if err != nil {
        log.Fatal(err)
    }
}

func openSecurityCenter() {
    cmd := exec.Command("explorer", "windowsdefender:")
    err := cmd.Start()
    if err != nil {
        log.Fatal(err)
    }
}

func updateTrayIcon(cd string) {

    systray.SetIcon(getIcon(filepath.Join(cd,"icon/default.ico")))
    systray.SetTooltip("Loading ...")
    
    if IswinFirewallEnabled() {
      systray.SetIcon(getIcon(filepath.Join(cd,"icon/enabled.ico")))
      systray.SetTooltip("Windows Firewall is Enabled !")
    } else {
      systray.SetIcon(getIcon(filepath.Join(cd,"icon/disabled.ico")))
      systray.SetTooltip("Windows Firewall is Disabled !")
    }
}

func onReady() {

    ex, err := os.Executable()
    if err != nil {
        log.Fatal(err)
    }
    exPath := filepath.Dir(ex)

    systray.SetTitle("Windows Firewall Tray Control")
    menu_open_security_center := systray.AddMenuItem("Windows Security Center", "Windows Security Center")
    systray.AddSeparator()
    menu_enable := systray.AddMenuItem("Enable Windows Firewall", "Enable Windows Firewall")
    menu_disable := systray.AddMenuItem("Disable Windows Firewall", "Disable Windows Firewall")
    systray.AddSeparator()
    mQuit := systray.AddMenuItem("Exit", "Exit")

    updateTrayIcon(exPath)
    
    go func() {
        for {
            select {
            case <-menu_open_security_center.ClickedCh:
                openSecurityCenter()
            case <-menu_enable.ClickedCh:
                setFirewall(true)
                updateTrayIcon(exPath)
            case <-menu_disable.ClickedCh:
                setFirewall(false)
                updateTrayIcon(exPath)
            case <-mQuit.ClickedCh:
                systray.Quit()
                return
            }
        }
    }()

}

func onExit() {
    
    dir := os.Getenv("TEMP");
    
    files, err := filepath.Glob(filepath.Join(dir,"/systray_temp_icon_*"))
    if err != nil {
        log.Fatal(err)
    }
    for _, f := range files {
        if err := os.Remove(f); err != nil {
            log.Fatal(err)
        }
    }
    
}

func getIcon(s string) []byte {
    b, err := ioutil.ReadFile(s)
    if err != nil {
        log.Fatal(err)
    }
    return b
}