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
	"time"
	
	"golang.org/x/text/language"
	"golang.org/x/text/language/display"
	
	"github.com/getlantern/systray"
	"gopkg.in/toast.v1"
)

var (
  langID int
  l10n [][]string
)

func initLang(){

    var supportedLang = []language.Tag{
      language.AmericanEnglish, //first language is fallback
      language.French,
    }
    var matched = language.NewMatcher(supportedLang)
    var lang = display.English.Tags().Name(matched)
    
    l10n = [][]string{}
    
    english := []string{"Windows Security Center", "Enable Windows Firewall", "Disable Windows Firewall", "Exit", "Windows Firewall", "Windows Firewall is turned on.", "Loading ...", "Windows Firewall is Enabled !", "Windows Firewall is Disabled !"}
    french := []string{"Centre de sécurité Windows", "Activer le pare-feu Windows", "Désactiver le pare-feu Windows", "Quitter", "Pare-feu Windows", "Le pare-feu Windows est activé.", "Chargement ...", "Le pare-feu Windows est activé !", "Le pare-feu Windows est désactivé !"}
    
    l10n = append(l10n, english)
    l10n = append(l10n, french)
    
    switch lang {
      case "English":
          langID = 1
      case "French":
          langID = 2
      default:
          langID = 1
    }

}

func main() {
    initLang()
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

func updateTrayIcon(cd string, showLoading bool) {

    if showLoading {
      systray.SetIcon(getIcon(filepath.Join(cd,"icon/default.ico")))
      systray.SetTooltip(l10n[langID][6])
    }
    
    if IswinFirewallEnabled() {
      systray.SetIcon(getIcon(filepath.Join(cd,"icon/enabled.ico")))
      systray.SetTooltip(l10n[langID][7])
    } else {
      systray.SetIcon(getIcon(filepath.Join(cd,"icon/disabled.ico")))
      systray.SetTooltip(l10n[langID][8])
    }
}

func onReady() {

    ex, err := os.Executable()
    if err != nil {
        log.Fatal(err)
    }
    exPath := filepath.Dir(ex)
    
    systray.SetTitle("Windows Firewall Tray Control")
    menu_open_security_center := systray.AddMenuItem(l10n[langID][0], l10n[langID][0])
    systray.AddSeparator()
    menu_enable := systray.AddMenuItem(l10n[langID][1], l10n[langID][1])
    menu_disable := systray.AddMenuItem(l10n[langID][2], l10n[langID][2])
    systray.AddSeparator()
    mQuit := systray.AddMenuItem(l10n[langID][3], l10n[langID][3])
    
    updateTrayIcon(exPath,true)
    ms := 60000
    ticker := time.NewTicker(time.Millisecond * time.Duration(ms))
    
    go func() {
        for {
            select {
            case <-menu_open_security_center.ClickedCh:
                openSecurityCenter()
            case <-menu_enable.ClickedCh:
                setFirewall(true)
                    notification := toast.Notification{
                        //Get-StartApps in powershell to list installed AppID
                        AppID: "Microsoft.Windows.SecHealthUI_cw5n1h2txyewy!SecHealthUI",
                        Title: l10n[langID][4],
                        Message: l10n[langID][5],
                        Icon: filepath.Join(exPath,"icon/toast_ok.png"),
                    }
                    err := notification.Push()
                    if err != nil {
                        log.Fatalln(err)
                    }
                updateTrayIcon(exPath,true)
            case <-menu_disable.ClickedCh:
                setFirewall(false)
                updateTrayIcon(exPath,true)
            case <-mQuit.ClickedCh:
                systray.Quit()
                return
            }
        }
    }()
    
    go func() {
        for range ticker.C {
            updateTrayIcon(exPath,false)
        }
    }()

}

func onExit() {
    
    dir := os.TempDir();
    
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