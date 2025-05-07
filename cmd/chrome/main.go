package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/chromedp/cdproto/network"
)

func main() {
	// Create context
	ctx, cancel := chromedp.NewExecAllocator(
		context.Background(),
		chromedp.DefaultExecAllocatorOptions[:]...,
	)
	defer cancel()

	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	// Capture all network events
	var networkEvents []string

	chromedp.ListenTarget(ctx, func(ev interface{}) {
		if reqEvent, ok := ev.(*network.EventRequestWillBeSent); ok {
			networkEvents = append(networkEvents,
				fmt.Sprintf("Request: %s %s", reqEvent.Request.Method, reqEvent.Request.URL))
		}
		if respEvent, ok := ev.(*network.EventResponseReceived); ok {
			networkEvents = append(networkEvents,
				fmt.Sprintf("Response: %d %s", int(respEvent.Response.Status), respEvent.Response.URL))
		}
	})

	// Enable network tracking
	if err := chromedp.Run(ctx, network.Enable()); err != nil {
		log.Fatal(err)
	}

	// Visit the LKML archive
	url := "https://lore.kernel.org/lkml/"
	err := chromedp.Run(ctx,
		chromedp.Navigate(url),
		chromedp.Sleep(5*time.Second), // Wait for page load and assets
	)
	if err != nil {
		log.Fatal(err)
	}

	// Output all captured network transactions
	fmt.Println("=== Network Transactions ===")
	for _, event := range networkEvents {
		fmt.Println(event)
	}
}

