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

func Kiosk(cfg *Config) {
	dir, err := os.MkdirTemp(os.TempDir(), "chromedp-kiosk")
	if err != nil {
		panic(err)
	}

	log.Println("Using temp dir:", dir)
	defer os.RemoveAll(dir)

	opts := generateExecutorOptions(dir, cfg)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	taskCtx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(log.Printf))
	defer cancel()

	/*
		chromedp.ListenBrowser(taskCtx, func(ev interface{}) {

			fmt.Println("Event Fired")
			if ev, ok := ev.(*target.EventTargetDestroyed); ok {
				if c := chromedp.FromContext(taskCtx); c != nil {
					if c.Target.TargetID == ev.TargetID {
						log.Println("Window closed")
						cancel()
					}
				}
			}
		})
	*/

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

func generateExecutorOptions(dir string, c *Config) []chromedp.ExecAllocatorOption {
	return []chromedp.ExecAllocatorOption{
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("user-agent", "Mozilla/5.0 (X11; CrOS armv7l 13597.84.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
		chromedp.Flag("noerrdialogs", true),
		chromedp.Flag("kiosk", true),
		chromedp.Flag("bwsi", true),
		chromedp.Flag("incognito", true),
		chromedp.Flag("disable-sync", true),
		chromedp.Flag("disable-notifications", true),
		chromedp.Flag("disable-overlay-scrollbar", true),
		chromedp.Flag("window-position", c.WindowPosition),
		chromedp.Flag("check-for-update-interval", "31536000"),
		chromedp.Flag("ignore-certificate-errors", c.IgnoreCertificateErrors),
		chromedp.Flag("test-type", c.IgnoreCertificateErrors),
		chromedp.Flag("autoplay-policy", "no-user-gesture-required"),
		//chromedp.Flag("window-size", cfg.WindowSize),
		chromedp.UserDataDir(dir),
	}
}
