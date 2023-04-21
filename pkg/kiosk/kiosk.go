package kiosk

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/chromedp/chromedp/kb"
)

// GrafanaKioskLocal creates a chrome-based kiosk using a local grafana-server account.
func Kiosk(cfg *Config) {
	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		panic(err)
	}

	log.Println("Using temp dir:", dir)
	defer os.RemoveAll(dir)

	opts := generateExecutorOptions(dir, cfg.WindowPosition, cfg.IgnoreCertificateErrors)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// also set up a custom logger
	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	//go listenChromeEvents(taskCtx, targetCrashed)

	// ensure that the browser process is started
	if err := chromedp.Run(taskCtx); err != nil {
		panic(err)
	}

	log.Println("Navigating to ", cfg.Url)
	// Wait Browser to open and set FullScreen
	time.Sleep(2000 * time.Millisecond)

	var tasks chromedp.Tasks

	switch cfg.LoginMethod {
	case "none":
		tasks = loginNoneTasks(cfg)
	case "local":
		tasks = loginLocalTasks(cfg)
	case "token":
		tasks = loginApiTasks(cfg)

	default:
		panic(fmt.Errorf("no login method found"))
	}

	err = chromedp.Run(taskCtx, tasks)
	if err != nil {
		panic(err)
	}
}

func loginNoneTasks(cfg *Config) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(cfg.Url),
		chromedp.WaitVisible(`notinputPassword`, chromedp.ByID),
	}
}

func loginLocalTasks(cfg *Config) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(cfg.Url),
		chromedp.WaitVisible(`//input[@name="user"]`, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="user"]`, cfg.Username, chromedp.BySearch),
		chromedp.SendKeys(`//input[@name="password"]`, cfg.Password+kb.Enter, chromedp.BySearch),
		chromedp.WaitVisible(`notinputPassword`, chromedp.ByID),
	}
}

func loginApiTasks(cfg *Config) chromedp.Tasks {
	headers := map[string]interface{}{
		"Authorization": "Bearer " + cfg.Token,
	}
	return chromedp.Tasks{
		network.Enable(),
		network.SetExtraHTTPHeaders(network.Headers(headers)),
		chromedp.Navigate(cfg.Url),
		chromedp.WaitVisible(`//div[@class="main-view"]`, chromedp.BySearch),
		// wait forever (for now)
		chromedp.WaitVisible("notinputPassword", chromedp.ByID),
	}
}

func generateExecutorOptions(dir string, windowPosition string, ignoreCertificateErrors bool) []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("window-position", windowPosition),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.Flag("ignore-certificate-errors", ignoreCertificateErrors),
		chromedp.Flag("test-type", ignoreCertificateErrors),
		chromedp.UserDataDir(dir),
	}
}
