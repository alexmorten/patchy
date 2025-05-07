package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

func main() {
	// Seed random number generator
	rand.Seed(time.Now().UnixNano())

	// Create context with custom options to appear more like a real browser
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Set a realistic user agent - using Firefox instead of Chrome
		chromedp.UserAgent("Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/119.0"),
		// Disable automation flags
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("disable-features", "AutomationControlled"),
		chromedp.Flag("enable-automation", false),
		// Additional flags to make the browser appear more normal
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		// Set window size to a common resolution
		chromedp.WindowSize(1366, 768),
		// Enable images and JavaScript
		chromedp.Flag("disable-images", false),
		chromedp.Flag("disable-javascript", false),
		// Add plugins and extensions flags
		chromedp.Flag("enable-plugins", true),
		// Set language preferences
		chromedp.Flag("lang", "en-US,en;q=0.9"),
		// Add hardware concurrency
		chromedp.Flag("js-flags", "--expose_gc"),
		// Add WebGL support
		chromedp.Flag("enable-webgl", true),
		// Add WebRTC support
		chromedp.Flag("enable-webrtc", true),
	)
	
	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel = chromedp.NewContext(
		ctx,
		chromedp.WithLogf(log.Printf),
	)
	defer cancel()

	// Set a longer timeout for the entire context
	ctx, cancel = context.WithTimeout(ctx, 180*time.Second)
	defer cancel()

	// Store challenge-related transactions and response headers
	var challengeRequests []string
	var challengeResponses []string
	var responseHeaders []string

	// Track if we've seen the pass-challenge request
	passChallengeRequested := false
	passChallengeResponded := false
	challengeURL := ""
	var challengeParams url.Values

	// Listen for network events
	chromedp.ListenTarget(ctx, func(ev interface{}) {
		switch e := ev.(type) {
		case *network.EventRequestWillBeSent:
			fmt.Printf("Request: %s %s\n", e.Request.Method, e.Request.URL)
			
			if strings.Contains(e.Request.URL, "pass-challenge") {
				passChallengeRequested = true
				challengeURL = e.Request.URL
				challengeRequests = append(challengeRequests,
					fmt.Sprintf("Challenge Request: %s %s", e.Request.Method, e.Request.URL))
				
				// Parse the challenge URL to extract parameters
				if parsedURL, err := url.Parse(e.Request.URL); err == nil {
					challengeParams = parsedURL.Query()
					fmt.Println("\n=== Challenge Parameters ===")
					for k, v := range challengeParams {
						fmt.Printf("%s: %s\n", k, v[0])
					}
				}
				
				fmt.Println("\n=== Pass-Challenge Request Details ===")
				fmt.Printf("URL: %s\n", e.Request.URL)
				fmt.Printf("Method: %s\n", e.Request.Method)
				fmt.Println("Headers:")
				for name, value := range e.Request.Headers {
					fmt.Printf("%s: %v\n", name, value)
				}
			} else if strings.Contains(e.Request.URL, "make-challenge") {
				challengeRequests = append(challengeRequests,
					fmt.Sprintf("Make Challenge Request: %s %s", e.Request.Method, e.Request.URL))
			}
			
		case *network.EventResponseReceived:
			fmt.Printf("Response: %s (Status: %d)\n", e.Response.URL, int(e.Response.Status))
			
			if strings.Contains(e.Response.URL, "pass-challenge") {
				passChallengeResponded = true
				challengeResponses = append(challengeResponses,
					fmt.Sprintf("Pass-Challenge Response: %d %s", int(e.Response.Status), e.Response.URL))
				
				fmt.Println("\n=== Pass-Challenge Response Details ===")
				fmt.Printf("URL: %s\n", e.Response.URL)
				fmt.Printf("Status: %d\n", int(e.Response.Status))
				fmt.Println("Headers:")
				for name, value := range e.Response.Headers {
					headerStr := fmt.Sprintf("%s: %v", name, value)
					responseHeaders = append(responseHeaders, headerStr)
					fmt.Println(headerStr)
				}
			} else if strings.Contains(e.Response.URL, "make-challenge") {
				challengeResponses = append(challengeResponses,
					fmt.Sprintf("Make-Challenge Response: %d %s", int(e.Response.Status), e.Response.URL))
				
				fmt.Println("\n=== Make-Challenge Response Headers ===")
				for name, value := range e.Response.Headers {
					headerStr := fmt.Sprintf("%s: %v", name, value)
					responseHeaders = append(responseHeaders, headerStr)
					fmt.Println(headerStr)
				}
			}
			
		case *network.EventLoadingFinished:
			fmt.Printf("Request completed: %s\n", e.RequestID)
		}
	})

	// Override JavaScript properties that might reveal automation
	err := chromedp.Run(ctx, 
		// Enable network monitoring
		network.Enable(),
		
		// Set custom headers to mimic a real browser
		network.SetExtraHTTPHeaders(network.Headers{
			"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
			"accept-language":           "en-US,en;q=0.5",
			"accept-encoding":           "gzip, deflate, br",
			"connection":                "keep-alive",
			"upgrade-insecure-requests": "1",
			"sec-fetch-dest":            "document",
			"sec-fetch-mode":            "navigate",
			"sec-fetch-site":            "none",
			"sec-fetch-user":            "?1",
			"dnt":                       "1",
		}),
		
		// Override JavaScript properties that might reveal automation
		chromedp.Evaluate(`
			// Override navigator properties
			Object.defineProperty(navigator, 'webdriver', {
				get: () => false,
			});
			
			// Override permissions
			const originalQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) => (
				parameters.name === 'notifications' ?
					Promise.resolve({ state: Notification.permission }) :
					originalQuery(parameters)
			);
			
			// Add plugins
			Object.defineProperty(navigator, 'plugins', {
				get: () => {
					return [
						{
							0: {type: "application/pdf", suffixes: "pdf", description: "Portable Document Format"},
							name: "PDF Viewer",
							description: "Portable Document Format",
							filename: "internal-pdf-viewer",
							length: 1
						},
						{
							0: {type: "application/x-shockwave-flash", suffixes: "swf", description: "Shockwave Flash"},
							name: "Shockwave Flash",
							description: "Shockwave Flash 32.0 r0",
							filename: "flash.ocx",
							length: 1
						}
					];
				},
			});
			
			// Add languages
			Object.defineProperty(navigator, 'languages', {
				get: () => ['en-US', 'en'],
			});
			
			// Override toString methods to hide proxy behavior
			const originalFunction = Function.prototype.toString;
			Function.prototype.toString = function() {
				if (this === Function.prototype.toString) return originalFunction.call(this);
				if (this === navigator.permissions.query) return "function query() { [native code] }";
				return originalFunction.call(this);
			};
		`, nil),
		
		// Set hardware concurrency
		chromedp.Evaluate(`
			Object.defineProperty(navigator, 'hardwareConcurrency', {
				get: () => 4,
			});
		`, nil),
		
		// Set device memory
		chromedp.Evaluate(`
			Object.defineProperty(navigator, 'deviceMemory', {
				get: () => 8,
			});
		`, nil),
		
		// Set platform
		chromedp.Evaluate(`
			Object.defineProperty(navigator, 'platform', {
				get: () => 'Linux x86_64',
			});
		`, nil),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Visit the LKML archive
	url := "https://lore.kernel.org/lkml/"
	fmt.Printf("Navigating to: %s\n", url)
	
	err = chromedp.Run(ctx,
		// Set a realistic viewport
		emulation.SetDeviceMetricsOverride(1366, 768, 1.0, false),
		
		// Navigate to the URL
		chromedp.Navigate(url),
		
		// Wait for initial page load
		chromedp.Sleep(randomDuration(3, 5)),
		
		// Perform random mouse movements to appear human-like
		simulateHumanBehavior(ctx),
	)
	if err != nil {
		log.Fatal(err)
	}

	// If we detected a challenge but didn't get a response, try to handle it directly
	if passChallengeRequested && !passChallengeResponded && challengeURL != "" {
		fmt.Printf("Attempting to handle challenge directly: %s\n", challengeURL)
		
		// Try to handle the challenge with a direct HTTP request
		if len(challengeParams) > 0 {
			// Create a direct HTTP client with similar headers
			client := &http.Client{
				Timeout: 30 * time.Second,
			}
			
			req, err := http.NewRequest("GET", challengeURL, nil)
			if err == nil {
				// Add headers to mimic a real browser
				req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/119.0")
				req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
				req.Header.Set("Accept-Language", "en-US,en;q=0.5")
				req.Header.Set("Connection", "keep-alive")
				req.Header.Set("Upgrade-Insecure-Requests", "1")
				req.Header.Set("Sec-Fetch-Dest", "document")
				req.Header.Set("Sec-Fetch-Mode", "navigate")
				req.Header.Set("Sec-Fetch-Site", "same-origin")
				req.Header.Set("Sec-Fetch-User", "?1")
				req.Header.Set("DNT", "1")
				req.Header.Set("Referer", "https://lore.kernel.org/lkml/")
				
				// Make the request
				resp, err := client.Do(req)
				if err == nil {
					defer resp.Body.Close()
					fmt.Printf("Direct challenge request response: %d\n", resp.StatusCode)
					
					// If successful, navigate to the target URL
					if resp.StatusCode == 200 || resp.StatusCode == 302 {
						err = chromedp.Run(ctx,
							chromedp.Navigate(url),
							chromedp.Sleep(randomDuration(3, 5)),
						)
						if err != nil {
							fmt.Printf("Error navigating after challenge: %v\n", err)
						}
					}
				} else {
					fmt.Printf("Error making direct challenge request: %v\n", err)
				}
			}
		}
		
		// Also try with ChromeDP
		err = chromedp.Run(ctx,
			chromedp.Navigate(challengeURL),
			chromedp.Sleep(randomDuration(5, 8)),
			simulateHumanBehavior(ctx),
		)
		if err != nil {
			fmt.Printf("Error navigating to challenge URL: %v\n", err)
		}
	}

	// Wait for challenge to be processed with periodic checks
	maxWaitTime := 60 * time.Second
	checkInterval := 3 * time.Second
	startTime := time.Now()
	
	for time.Since(startTime) < maxWaitTime {
		if passChallengeResponded {
			fmt.Println("Challenge response received, continuing...")
			break
		}
		
		fmt.Println("Waiting for challenge to complete...")
		
		// Perform some random mouse movements and scrolling while waiting
		err = chromedp.Run(ctx, simulateHumanBehavior(ctx))
		if err != nil {
			fmt.Printf("Error during human simulation: %v\n", err)
		}
		
		time.Sleep(checkInterval)
		
		// Check if we're on the expected page
		var currentURL string
		err = chromedp.Run(ctx, chromedp.Location(&currentURL))
		if err == nil {
			fmt.Printf("Current URL: %s\n", currentURL)
			if currentURL == url && !strings.Contains(currentURL, "challenge") {
				fmt.Println("Successfully reached target URL")
				break
			}
		}
	}

	// Final check to see if we're on the expected page
	var pageContent string
	err = chromedp.Run(ctx,
		chromedp.WaitVisible(`body`, chromedp.ByQuery),
		chromedp.OuterHTML(`html`, &pageContent),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Output challenge-related transactions
	fmt.Println("\n=== Challenge Requests ===")
	for _, req := range challengeRequests {
		fmt.Println(req)
	}

	fmt.Println("\n=== Challenge Responses ===")
	for _, resp := range challengeResponses {
		fmt.Println(resp)
	}

	fmt.Println("\n=== Challenge Status ===")
	fmt.Printf("Pass-Challenge Requested: %v\n", passChallengeRequested)
	fmt.Printf("Pass-Challenge Responded: %v\n", passChallengeResponded)

	// Output a portion of the page content to verify we're on the right page
	if len(pageContent) > 200 {
		fmt.Println("\n=== Page Content Preview ===")
		fmt.Println(pageContent[:200] + "...")
	} else {
		fmt.Println("\n=== Page Content ===")
		fmt.Println(pageContent)
	}
}

func simulateHumanBehavior(ctx context.Context) chromedp.Tasks {
	return chromedp.Tasks{
		// Random mouse movements
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Get viewport dimensions
			var width, height int64
			if err := chromedp.Evaluate(`[window.innerWidth, window.innerHeight]`, &[]int64{width, height}).Do(ctx); err != nil {
				width, height = 1000, 800 // fallback values
			}
			
			// Perform 3-7 random mouse movements
			numMovements := rand.Intn(5) + 3
			for i := 0; i < numMovements; i++ {
				x := rand.Float64() * float64(width)
				y := rand.Float64() * float64(height)
				
				if err := input.DispatchMouseEvent(
					input.MouseMoved,
					float64(x),
					float64(y),
				).Do(ctx); err != nil {
					return err
				}
				
				// Random delay between mouse movements
				time.Sleep(randomDuration(0.1, 0.5))
			}
			return nil
		}),
		
		// Random scrolling
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Scroll down randomly
			scrollAmount := rand.Intn(500) + 100
			if err := chromedp.Evaluate(fmt.Sprintf(`window.scrollBy(0, %d)`, scrollAmount), nil).Do(ctx); err != nil {
				return err
			}
			
			time.Sleep(randomDuration(0.5, 1.5))
			
			// Sometimes scroll back up
			if rand.Float64() > 0.5 {
				upAmount := rand.Intn(scrollAmount)
				if err := chromedp.Evaluate(fmt.Sprintf(`window.scrollBy(0, %d)`, -upAmount), nil).Do(ctx); err != nil {
					return err
				}
			}
			
			return nil
		}),
		
		// Random delay
		chromedp.Sleep(randomDuration(1, 3)),
	}
}

// randomDuration returns a random duration between min and max seconds
func randomDuration(min, max float64) time.Duration {
	return time.Duration((min + rand.Float64()*(max-min)) * float64(time.Second))
}

