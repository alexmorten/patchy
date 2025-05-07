package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/tebeka/selenium"
	"github.com/tebeka/selenium/chrome"
)

// NetworkEvent represents a simplified network request/response
type NetworkEvent struct {
	Type        string            `json:"type"`
	Method      string            `json:"method"`
	URL         string            `json:"url"`
	StatusCode  int               `json:"statusCode,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Timestamp   time.Time         `json:"timestamp"`
	ElapsedTime int64             `json:"elapsedTime,omitempty"`
}

func main() {
	// Set up Selenium WebDriver
	const (
		seleniumPath    = "chromedriver" // Make sure chromedriver is in your PATH
		port            = 4444
		geckoDriverPath = "geckodriver" // Make sure geckodriver is in your PATH if using Firefox
	)

	// Create a directory to store network logs
	logDir := "network_logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("Error creating log directory: %v", err)
	}

	opts := []selenium.ServiceOption{}
	service, err := selenium.NewChromeDriverService(seleniumPath, port, opts...)
	if err != nil {
		fmt.Printf("Error starting Chrome driver: %v\n", err)

		// Try Firefox as a fallback
		service, err = selenium.NewGeckoDriverService(geckoDriverPath, port, opts...)
		if err != nil {
			log.Fatalf("Error starting Firefox driver: %v", err)
		}
	}
	defer service.Stop()

	// Connect to the WebDriver instance
	caps := selenium.Capabilities{
		"browserName": "chrome",
	}

	// Enable Chrome DevTools Protocol (CDP) logging
	loggingPrefs := map[string]string{
		"browser":     "ALL",
		"driver":      "ALL",
		"performance": "ALL",
	}
	caps["loggingPrefs"] = loggingPrefs

	networkEnabled := true
	pageEnabled := true
	timelineEnabled := true

	// Set Chrome-specific options
	chromeCaps := chrome.Capabilities{
		Args: []string{
			"--no-sandbox",
			"--disable-dev-shm-usage",
			"--disable-blink-features=AutomationControlled",
			"--start-maximized",
			"--disable-extensions",
			"--disable-popup-blocking",
			"--disable-infobars",
			"--ignore-certificate-errors",
			"--enable-gpu", // Enable GPU acceleration
			"--window-size=1920,1080",
			"--enable-logging",
			"--v=1",
			"--enable-network-information",
			"--log-level=0",
			"--enable-features=NetworkService,NetworkServiceInProcess",
		},
		ExcludeSwitches: []string{"enable-automation"},
		// Enable performance logging
		PerfLoggingPrefs: &chrome.PerfLoggingPreferences{
			EnableNetwork:   &networkEnabled,
			EnablePage:      &pageEnabled,
			EnableTimeline:  &timelineEnabled,
			TraceCategories: "devtools.timeline,loading",
		},
	}

	caps.AddChrome(chromeCaps)

	// Set user preferences
	prefs := map[string]interface{}{
		"profile.default_content_settings.popups":   0,
		"credentials_enable_service":                false,
		"profile.password_manager_enabled":          false,
		"devtools.preferences.network.log.preserve": true,
	}
	caps.AddChrome(chrome.Capabilities{
		Prefs: prefs,
	})

	// Connect to WebDriver
	wd, err := selenium.NewRemote(caps, fmt.Sprintf("http://localhost:%d/wd/hub", port))
	if err != nil {
		log.Fatalf("Error connecting to WebDriver: %v", err)
	}
	defer wd.Quit()

	// Set window size
	if err := wd.ResizeWindow("", 1920, 1080); err != nil {
		log.Printf("Error resizing window: %v", err)
	}

	// Execute JavaScript to enable network monitoring
	script := `
		// Create a global array to store network events
		window.networkEvents = [];
		
		// Create a PerformanceObserver to monitor network requests
		const observer = new PerformanceObserver((list) => {
			for (const entry of list.getEntries()) {
				if (entry.entryType === 'resource') {
					window.networkEvents.push({
						type: 'request',
						url: entry.name,
						initiatorType: entry.initiatorType,
						startTime: entry.startTime,
						duration: entry.duration,
						timestamp: new Date().toISOString()
					});
				}
			}
		});
		
		// Start observing
		observer.observe({entryTypes: ['resource']});
		
		// Override fetch to monitor network requests
		const originalFetch = window.fetch;
		window.fetch = async function(input, init) {
			const url = typeof input === 'string' ? input : input.url;
			const method = init?.method || 'GET';
			
			// Log the request
			const requestEvent = {
				type: 'fetch-request',
				method: method,
				url: url,
				headers: init?.headers || {},
				timestamp: new Date().toISOString()
			};
			window.networkEvents.push(requestEvent);
			
			try {
				// Make the actual request
				const response = await originalFetch.apply(this, arguments);
				
				// Clone the response to avoid consuming it
				const clonedResponse = response.clone();
				
				// Log the response
				const responseEvent = {
					type: 'fetch-response',
					url: url,
					status: clonedResponse.status,
					statusText: clonedResponse.statusText,
					headers: Object.fromEntries([...clonedResponse.headers.entries()]),
					timestamp: new Date().toISOString()
				};
				window.networkEvents.push(responseEvent);
				
				return response;
			} catch (error) {
				// Log the error
				const errorEvent = {
					type: 'fetch-error',
					url: url,
					error: error.toString(),
					timestamp: new Date().toISOString()
				};
				window.networkEvents.push(errorEvent);
				
				throw error;
			}
		};
		
		// Override XMLHttpRequest to monitor AJAX requests
		const originalXHROpen = XMLHttpRequest.prototype.open;
		const originalXHRSend = XMLHttpRequest.prototype.send;
		
		XMLHttpRequest.prototype.open = function(method, url) {
			this._method = method;
			this._url = url;
			return originalXHROpen.apply(this, arguments);
		};
		
		XMLHttpRequest.prototype.send = function() {
			// Log the request
			const requestEvent = {
				type: 'xhr-request',
				method: this._method,
				url: this._url,
				timestamp: new Date().toISOString()
			};
			window.networkEvents.push(requestEvent);
			
			// Add response listener
			this.addEventListener('load', function() {
				const responseEvent = {
					type: 'xhr-response',
					url: this._url,
					status: this.status,
					statusText: this.statusText,
					headers: this.getAllResponseHeaders().split('\r\n').reduce((acc, header) => {
						const parts = header.split(': ');
						if (parts.length === 2) {
							acc[parts[0]] = parts[1];
						}
						return acc;
					}, {}),
					timestamp: new Date().toISOString()
				};
				window.networkEvents.push(responseEvent);
			});
			
			this.addEventListener('error', function() {
				const errorEvent = {
					type: 'xhr-error',
					url: this._url,
					timestamp: new Date().toISOString()
				};
				window.networkEvents.push(errorEvent);
			});
			
			return originalXHRSend.apply(this, arguments);
		};
		
		// Function to retrieve all network events
		window.getNetworkEvents = function() {
			return window.networkEvents;
		};
	`
	_, err = wd.ExecuteScript(script, nil)
	if err != nil {
		log.Printf("Error executing network monitoring script: %v", err)
	}

	// Execute JavaScript to hide automation
	antiDetectionScript := `
		Object.defineProperty(navigator, 'webdriver', {
			get: () => false,
		});
		
		// Override permissions
		if (window.navigator.permissions) {
			const originalQuery = window.navigator.permissions.query;
			window.navigator.permissions.query = (parameters) => (
				parameters.name === 'notifications' ?
					Promise.resolve({ state: Notification.permission }) :
					originalQuery(parameters)
			);
		}
		
		// Set hardware concurrency
		Object.defineProperty(navigator, 'hardwareConcurrency', {
			get: () => 8,
		});
		
		// Set platform
		Object.defineProperty(navigator, 'platform', {
			get: () => 'Linux x86_64',
		});
	`
	_, err = wd.ExecuteScript(antiDetectionScript, nil)
	if err != nil {
		log.Printf("Error executing anti-detection script: %v", err)
	}

	// Navigate to the target URL
	url := "https://lore.kernel.org/lkml/"
	fmt.Printf("Navigating to: %s\n", url)

	if err := wd.Get(url); err != nil {
		log.Fatalf("Error navigating to %s: %v", url, err)
	}

	// Wait for page to load
	time.Sleep(5 * time.Second)

	// Simulate human-like behavior
	simulateHumanBehavior(wd)

	// Collect network events from JavaScript
	var jsNetworkEvents []map[string]interface{}
	if result, err := wd.ExecuteScript("return window.getNetworkEvents();", nil); err == nil {
		if events, ok := result.([]interface{}); ok {
			for _, event := range events {
				if eventMap, ok := event.(map[string]interface{}); ok {
					jsNetworkEvents = append(jsNetworkEvents, eventMap)
				}
			}
		}
	} else {
		log.Printf("Error getting network events from JavaScript: %v", err)
	}

	// Print JavaScript-captured network events
	fmt.Println("\n=== JavaScript Network Events ===")
	for i, event := range jsNetworkEvents {
		eventType := event["type"]
		urlValue := event["url"]
		timestamp := event["timestamp"]

		fmt.Printf("%d. [%s] %s - %s\n", i+1, timestamp, eventType, urlValue)

		// Print more details for certain events
		if eventType == "fetch-response" || eventType == "xhr-response" {
			status := event["status"]
			fmt.Printf("   Status: %v\n", status)

			if headers, ok := event["headers"].(map[string]interface{}); ok && len(headers) > 0 {
				fmt.Println("   Headers:")
				for k, v := range headers {
					fmt.Printf("     %s: %v\n", k, v)
				}
			}
		}

		// Look for challenge-related events
		if urlStr, ok := urlValue.(string); ok {
			if strings.Contains(urlStr, "challenge") {
				fmt.Printf("\n!!! CHALLENGE EVENT DETECTED !!!\n")
				fmt.Printf("Event Type: %s\n", eventType)
				fmt.Printf("URL: %s\n", urlStr)

				// Print all event details
				eventJSON, _ := json.MarshalIndent(event, "", "  ")
				fmt.Println(string(eventJSON))
				fmt.Println()
			}
		}
	}

	// Get browser logs
	logs, err := wd.Log("performance")
	if err != nil {
		log.Printf("Error getting performance logs: %v", err)
	} else {
		fmt.Printf("\n=== Browser Performance Logs (%d entries) ===\n", len(logs))

		// Process performance logs to extract network events
		for i, entry := range logs {
			var logEntry map[string]interface{}
			if err := json.Unmarshal([]byte(entry.Message), &logEntry); err != nil {
				continue
			}

			if message, ok := logEntry["message"].(map[string]interface{}); ok {
				if method, ok := message["method"].(string); ok {
					// Filter for network events
					if strings.HasPrefix(method, "Network.") {
						fmt.Printf("%d. %s\n", i+1, method)

						// Extract request/response details
						if params, ok := message["params"].(map[string]interface{}); ok {
							// Check for request details
							if request, ok := params["request"].(map[string]interface{}); ok {
								if reqURL, ok := request["url"].(string); ok {
									fmt.Printf("   URL: %s\n", reqURL)

									// Check if this is a challenge-related request
									if strings.Contains(reqURL, "challenge") {
										fmt.Printf("\n!!! CHALLENGE REQUEST DETECTED !!!\n")
										fmt.Printf("Method: %s\n", method)
										fmt.Printf("URL: %s\n", reqURL)

										// Print request details
										if reqMethod, ok := request["method"].(string); ok {
											fmt.Printf("HTTP Method: %s\n", reqMethod)
										}

										// Print headers
										if headers, ok := request["headers"].(map[string]interface{}); ok {
											fmt.Println("Headers:")
											for k, v := range headers {
												fmt.Printf("  %s: %v\n", k, v)
											}
										}

										// Save full details to a file
										detailsJSON, _ := json.MarshalIndent(message, "", "  ")
										filename := filepath.Join(logDir, fmt.Sprintf("challenge_request_%d.json", i))
										if err := ioutil.WriteFile(filename, detailsJSON, 0644); err != nil {
											log.Printf("Error saving challenge details: %v", err)
										} else {
											fmt.Printf("Full details saved to: %s\n", filename)
										}

										fmt.Println()
									}
								}
							}

							// Check for response details
							if response, ok := params["response"].(map[string]interface{}); ok {
								if respURL, ok := response["url"].(string); ok {
									fmt.Printf("   Response URL: %s\n", respURL)

									if statusCode, ok := response["status"].(float64); ok {
										fmt.Printf("   Status: %.0f\n", statusCode)
									}

									// Check if this is a challenge-related response
									if strings.Contains(respURL, "challenge") {
										fmt.Printf("\n!!! CHALLENGE RESPONSE DETECTED !!!\n")
										fmt.Printf("Method: %s\n", method)
										fmt.Printf("URL: %s\n", respURL)
										fmt.Println(response)

										// Print status
										if statusCode, ok := response["status"].(float64); ok {
											fmt.Printf("Status Code: %.0f\n", statusCode)
										}

										// Print headers with special focus on Set-Cookie
										if headers, ok := response["headers"].(map[string]interface{}); ok {
											fmt.Println("Headers:")

											// First check specifically for Set-Cookie header
											if setCookie, ok := headers["Set-Cookie"]; ok {
												fmt.Printf("  >>> Set-Cookie: %v <<<\n", setCookie)
											} else if setCookie, ok := headers["set-cookie"]; ok {
												fmt.Printf("  >>> Set-Cookie: %v <<<\n", setCookie)
											}

											// Then print all headers
											for k, v := range headers {
												fmt.Printf("  %s: %v\n", k, v)
											}
										}

										// Save full details to a file
										detailsJSON, _ := json.MarshalIndent(message, "", "  ")
										filename := filepath.Join(logDir, fmt.Sprintf("challenge_response_%d.json", i))
										if err := ioutil.WriteFile(filename, detailsJSON, 0644); err != nil {
											log.Printf("Error saving challenge details: %v", err)
										} else {
											fmt.Printf("Full details saved to: %s\n", filename)
										}

										fmt.Println()

										// Special handling for pass-challenge responses
										if strings.Contains(respURL, "pass-challenge") {
											fmt.Printf("\n!!! PASS-CHALLENGE RESPONSE DETECTED !!!\n")
											fmt.Printf("URL: %s\n", respURL)

											if headers, ok := response["headers"].(map[string]interface{}); ok {
												fmt.Println("=== COOKIE INFORMATION ===")

												// Check for Set-Cookie with different case variations
												cookieFound := false
												for k, v := range headers {
													if strings.ToLower(k) == "set-cookie" {
														fmt.Printf("SET-COOKIE HEADER: %v\n", v)
														cookieFound = true
													}
												}

												if !cookieFound {
													fmt.Println("No Set-Cookie header found in this response")
												}
											}
										}
									}

								}
							}
						}
					}
				}
			}
		}
	}

	// Check if we're on the expected page
	currentURL, err := wd.CurrentURL()
	if err != nil {
		log.Printf("Error getting current URL: %v", err)
	} else {
		fmt.Printf("\nCurrent URL: %s\n", currentURL)
	}

	// Get page source to verify content
	pageSource, err := wd.PageSource()
	if err != nil {
		log.Printf("Error getting page source: %v", err)
	} else {
		// Save page source to file for inspection
		err = os.WriteFile("lkml_page.html", []byte(pageSource), 0644)
		if err != nil {
			log.Printf("Error saving page source: %v", err)
		}

		// Print a preview
		if len(pageSource) > 200 {
			fmt.Println("\n=== Page Content Preview ===")
			fmt.Println(pageSource[:200] + "...")
		} else {
			fmt.Println("\n=== Page Content ===")
			fmt.Println(pageSource)
		}
	}

	// Try to get the Chrome DevTools Protocol logs using Chrome's logging capabilities
	if logs, err := wd.Log("browser"); err == nil {
		fmt.Printf("\n=== Browser Logs (%d entries) ===\n", len(logs))
		for i, log := range logs {
			fmt.Printf("%d. [%s] %s: %s\n", i+1, log.Timestamp, log.Level, log.Message)
		}
	} else {
		log.Printf("Error getting browser logs: %v", err)
	}

	// Get final cookies before closing
	fmt.Println("\n=== Final Browser Cookies ===")
	finalCookies, err := wd.GetCookies()
	if err != nil {
		log.Printf("Error getting final cookies: %v", err)
	} else {
		for i, cookie := range finalCookies {
			fmt.Printf("%d. %s = %s\n", i+1, cookie.Name, cookie.Value)
			fmt.Printf("   Domain: %s, Path: %s\n", cookie.Domain, cookie.Path)
			fmt.Printf("   Secure: %t, Expiry: %v\n",
				cookie.Secure, time.Unix(int64(cookie.Expiry), 0))
			fmt.Println()
		}
	}

	fmt.Println("\nBrowser automation completed.")
}

func simulateHumanBehavior(wd selenium.WebDriver) {
	// Scroll down slowly
	script := `window.scrollBy(0, 300);`
	_, err := wd.ExecuteScript(script, nil)
	if err != nil {
		log.Printf("Error scrolling: %v", err)
	}
	time.Sleep(2 * time.Second)

	// Scroll up a bit
	script = `window.scrollBy(0, -100);`
	_, err = wd.ExecuteScript(script, nil)
	if err != nil {
		log.Printf("Error scrolling: %v", err)
	}
	time.Sleep(1 * time.Second)

	// Try to find and click on a link
	elements, err := wd.FindElements(selenium.ByCSSSelector, "a")
	if err != nil {
		log.Printf("Error finding links: %v", err)
	} else if len(elements) > 0 {
		// Click on a random link (but not the first few which might be navigation)
		index := 5
		if len(elements) > 10 {
			index = 10
		}
		if index < len(elements) {
			if err := elements[index].Click(); err != nil {
				log.Printf("Error clicking link: %v", err)
			}
			time.Sleep(3 * time.Second)

			// Go back
			if err := wd.Back(); err != nil {
				log.Printf("Error navigating back: %v", err)
			}
			time.Sleep(2 * time.Second)
		}
	}

	// Wait a bit more
	time.Sleep(3 * time.Second)
}
