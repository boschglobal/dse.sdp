package main

import (
	"context"
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
	_, err = chromedp.Run(ctx, chromedp.Steps(
		chromedp.Navigate("file://"+tmpHtmlfile.Name()),
		chromedp.Sleep(time.Second),
	))
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	height, err = chromedp.Run(ctx, chromedp.Evaluate[int64](`Math.max(document.body.scrollHeight, document.documentElement.scrollHeight)`))
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	_, err = chromedp.Run(ctx, chromedp.Steps(
		chromedp.EmulateViewport(1000, height, chromedp.EmulateScale(2)),
		chromedp.Sleep(1*time.Second),
	))
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	buf, err = chromedp.Run(ctx, chromedp.FullScreenshot(100))
	if err != nil {
		log.Fatalf("Failed to capture screenshot: %v", err)
	}

	err = os.WriteFile(outputImg, buf, 0644)
	if err != nil {
		log.Fatalf("Failed to save screenshot: %v", err)
	}
}
