package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/chromedp/chromedp"
)

func main() {
	if len(os.Args) != 3 {
		log.Fatalf("Usage: %s <html_File> <output.png>", os.Args[0])
	}
	htmlFile := os.Args[1]
	outputImg := os.Args[2]

	htmlContent, err := os.ReadFile(htmlFile)
	if err != nil {
		log.Fatalf("Failed to read HTML file: %v", err)
	}

	// for chromedp.Navigate() usage
	tmpHtmlfile, err := os.CreateTemp("", "*.html")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpHtmlfile.Name())

	if _, err := tmpHtmlfile.Write(htmlContent); err != nil {
		log.Fatal(err)
	}
	tmpHtmlfile.Close()

	ctx, cancel := chromedp.NewExecAllocator(context.Background(),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Headless,
		chromedp.DisableGPU,
		chromedp.Flag("incognito", true),
		chromedp.Flag("hide-scrollbars", true),
		chromedp.Flag("disable-extensions", true),
	)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var buf []byte
	var height int64
	err = chromedp.Run(ctx,
		chromedp.Navigate("file://"+tmpHtmlfile.Name()),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`Math.max(document.body.scrollHeight, document.documentElement.scrollHeight)`, &height),
		chromedp.EmulateViewport(int64(1000), height),
		chromedp.Evaluate(`window.devicePixelRatio = 2.0;`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.FullScreenshot(&buf, 100),
	)
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	err = ioutil.WriteFile(outputImg, buf, 0644)
	if err != nil {
		log.Fatalf("Failed to save screenshot: %v", err)
	}
}
