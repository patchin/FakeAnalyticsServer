package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"net/url"
	"os"
	"strings"
	"time"
)

type MySet map[string]bool

var legend = map[string]string{
	"v":    "Analytics Api Version",
	"tid":  "Analytics Account Id",
	"cid":  "Client Id",
	"ai":   "Anonymize IP",
	"sc":   "Session Control",
	"dr":   "Document Referrer",
	"cn":   "Campaign Name",
	"cs":   "Campaign Source",
	"cm":   "Campaign Medium",
	"ck":   "Campaign Keyword",
	"cc":   "Campaign Content",
	"ci":   "Campaign Id",
	"sr":   "Screen Resolution",
	"vp":   "Viewport Size",
	"de":   "Document Encoding",
	"sd":   "Screen Colors",
	"ul":   "User Language",
	"je":   "Java Enabled",
	"fl":   "Flash Version",
	"t":    "Hit Type",
	"ni":   "Non-Interaction Hit",
	"dl":   "Documentation Location Url",
	"dh":   "Document Host Name",
	"dp":   "Document Path",
	"dt":   "Document Title",
	"an":   "Application Name",
	"av":   "Application Version",
	"ec":   "Event Category",
	"ea":   "Event Action",
	"el":   "Event Label",
	"ev":   "Event Value",
	"sn":   "Social Network",
	"sa":   "Social Action",
	"st":   "Social Action Target",
	"utc":  "User Timing Category",
	"utv":  "User Timing Variable Name",
	"utt":  "User Timing Time",
	"utl":  "User Timing Label",
	"exd":  "Exception Description",
	"exf":  "Is Exception Fatal?",
	"cd1":  "cd1",
	"cd2":  "cd2",
	"cd3":  "cd3",
	"cd4":  "cd4",
	"cd5":  "cd5",
	"cd6":  "cd6",
	"cd7":  "cd7",
	"cd8":  "cd8",
	"cd9":  "cd9",
	"cd10": "cd10",
	"cd11": "cd11",
	"cd12": "cd12",
	"z":    "Cache Buster",
}

var g_googleKeysArray = []string{
	"v",
	"tid",
	"cid",
	"ai",
	"sc",
	"dr",
	"cn",
	"cs",
	"cm",
	"ck",
	"cc",
	"ci",
	"sr",
	"vp",
	"de",
	"dl",
	"sd",
	"ul",
	"je",
	"fl",
	"t",
	"ni",
	"dh",
	"dp",
	"dt",
	"an",
	"av",
	"ec",
	"ea",
	"el",
	"ev",
	"sn",
	"sa",
	"st",
	"utc",
	"utv",
	"utt",
	"utl",
	"exd",
	"exf",
	"z",
}

var g_cdArray = []string{
	"cd1", "cd2", "cd3", "cd4", "cd5", "cd6", "cd7", "cd8", "cd9", "cd10", "cd11", "cd12",
}

var g_allKeys []string = append(g_googleKeysArray, g_cdArray...)

var g_keySet MySet = make(MySet)

var g_requiredList []string = []string{"v", "cid", "t"}

var g_lastEmailTime time.Time
var g_sentEmailForFirstTime bool = false

// Request.RemoteAddress contains port, which we want to remove i.e.: 
// "[::1]:58292" => "[::1]" 
func ipAddrFromRemoteAddr(s string) string {
	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return s
	}
	return s[:idx]
}

func getIpAddress(r *http.Request) string {
	hdr := r.Header
	hdrRealIp := hdr.Get("X-Real-Ip")
	hdrForwardedFor := hdr.Get("X-Forwarded-For")
	if hdrRealIp == "" && hdrForwardedFor == "" {
		return ipAddrFromRemoteAddr(r.RemoteAddr)
	}
	if hdrForwardedFor != "" {
		// X-Forwarded-For is potentially a list of addresses separated with "," 
		parts := strings.Split(hdrForwardedFor, ",")
		for i, p := range parts {
			parts[i] = strings.TrimSpace(p)
		}
		// TODO: should return first non-local address 
		return parts[0]
	}
	return hdrRealIp
}

func splitAndPrintString(buffer *bytes.Buffer, str string) {
	if len(str) < 40 {
		buffer.WriteString(fmt.Sprintf("%s\n", str))
	} else {
		numslices := len(str) / 40

		// print first line right after '='' sign
		low := 0
		high := 40
		buffer.WriteString(fmt.Sprintf("%s\n", str[low:high]))

		// print rest of full lines
		for i := 1; i < numslices; i++ {
			low = i * 40
			high = low + 40
			buffer.WriteString(fmt.Sprintf("%43s%s\n", "", str[low:high]))
		}

		// print remainder if any
		remainder := len(str) % 40
		if remainder > 0 {
			buffer.WriteString(fmt.Sprintf("%43s%s\n", "", str[high:high+remainder]))
		}
	}
}

func printFormValues(buffer *bytes.Buffer, keys []string, form_values url.Values) bool {
	var found bool = false

	for i := range keys {
		key := keys[i]
		var arr []string = form_values[key]
		if len(arr) > 0 {
			if key == "dl" {
				display, _ := legend["dl"]
				buffer.WriteString(fmt.Sprintf("%40s = ", fmt.Sprintf("%s (dl)", display)))
				var str = arr[0]
				splitAndPrintString(buffer, str)
			} else if key == "dr" {
				display, _ := legend["dr"]
				buffer.WriteString(fmt.Sprintf("%40s = ", fmt.Sprintf("%s (dr)", display)))
				var str = arr[0]
				splitAndPrintString(buffer, str)
			} else {
				found = true

				// Display a human-readable caption if available.
				display, ok := legend[key]
				if len(arr) == 1 {
					if ok {
						buffer.WriteString(fmt.Sprintf("%40s = %s\n", fmt.Sprintf("%s (%s)", display, key), arr[0]))
					} else {
						buffer.WriteString(fmt.Sprintf("%40s = %s\n", key, arr[0]))
					}
				} else {
					if ok {
						buffer.WriteString(fmt.Sprintf("%40s = ", fmt.Sprintf("%s (%s)", display, key)))
					} else {
						buffer.WriteString(fmt.Sprintf("%40s = ", key))
					}

					for i, x := range arr {
						if i < len(arr)-1 {
							buffer.WriteString(fmt.Sprintf("%s, ", x))
						} else {
							buffer.WriteString(x)
						}
					}
				}
			}
		}
	}

	return found
}

func handler(w http.ResponseWriter, r *http.Request) {
	// A goroutine is created for each incoming request so there's no need to put
	// logAnalytics in its own goroutine. When I did put logAnalytics() in a goroutine it causes a bug where
	// posted form values disappear, probably because the main handler is already done before the goroutine
	// finishes. Some race condition?
	logAnalytics(r)

	// Return a fake image to calm the browser's nerves.
	w.Header().Set("Content-Type", "image/gif")
	w.Write([]byte(`GIF89a...`))
}

func logAnalytics(r *http.Request) {
	var bSendEmail bool = false

	r.ParseForm()

	var buffer bytes.Buffer

	buffer.WriteString("\n****************************************************************************************\n")

	buffer.WriteString("HTTP Vars:\n")
	buffer.WriteString(fmt.Sprintf("%40s = %s\n", "source", getIpAddress(r)))
	buffer.WriteString(fmt.Sprintf("%40s = %s\n", "host", r.Host))
	buffer.WriteString(fmt.Sprintf("%40s = %s\n", "method", r.Method))
	buffer.WriteString(fmt.Sprintf("%40s = %s\n", "path", r.URL.Path))

	// print header
	buffer.WriteString("HTTP Headers:\n")
	for k, v := range r.Header {
		/*
			if k == "Cookie" {
				var cookies []string = strings.Split(v[0], ";")
				for k2, v2 := range cookies {
					buffer.WriteString(fmt.Sprintf("%40s = %v\n", k2, v2))
				}
			} */
		if k == "User-Agent" || k == "Referer" {
			buffer.WriteString(fmt.Sprintf("%40s = ", k))
			if len(v) > 0 {
				splitAndPrintString(&buffer, v[0])
			}
		} else {
			buffer.WriteString(fmt.Sprintf("%40s = %s\n", k, strings.Join(v, ",")))
		}
	}

	// Standard Google Analytics
	buffer.WriteString("Standard Google Analytics Vars:\n")
	printFormValues(&buffer, g_googleKeysArray, r.Form)

	// Print out dl + dp for testing purposes.
	buffer.WriteString("For test: ")
	buffer.WriteString("cid=")
	if value, ok := r.Form["cid"]; ok {
		if len(value) > 0 {
			buffer.WriteString(value[0])
		} else {
			buffer.WriteString("<no-value>")
		}
	}
	buffer.WriteString(" dp=")
	if value, ok := r.Form["dp"]; ok {
		if len(value) > 0 {
			buffer.WriteString(value[0])
		} else {
			buffer.WriteString("<no-value>")
		}
	}
	buffer.WriteString("\n")

	// Print out Custom Definitions
	buffer.WriteString("Custom Definitions:\n")
	if !printFormValues(&buffer, g_cdArray, r.Form) {
		buffer.WriteString(fmt.Sprintf("%40s", "None\n"))
	}

	// Print out Unused
	buffer.WriteString(fmt.Sprintf("Unused: \n"))
	if !printUnusedKeys(&buffer, g_allKeys, r.Form) {
		buffer.WriteString(fmt.Sprintf("%40s", "None\n"))
	}

	// Print out Unknown
	buffer.WriteString(fmt.Sprintf("Unknown: \n"))
	if !printUnknownKeys(&buffer, g_keySet, r.Form) {
		buffer.WriteString(fmt.Sprintf("%40s", "None\n"))
	}

	// Determine which required measurements are missing.
	buffer.WriteString(fmt.Sprintf("Required but Missing:\n"))
	if !printMissingRequired(&buffer, g_requiredList, r.Form) {
		buffer.WriteString(fmt.Sprintf("%40s", "None\n"))
	} else {
		bSendEmail = true
	}

	// TODO: Move this into a goroutine
	// Don't send more than one an hour.	
	if bSendEmail {
		if !g_sentEmailForFirstTime {
			g_sentEmailForFirstTime = true
			g_lastEmailTime = time.Now()
			sendEmail(buffer.String())
		} else {
			if time.Since(g_lastEmailTime) > time.Hour {
				sendEmail(buffer.String())
				g_lastEmailTime = time.Now()
			}
		}
		bSendEmail = false
	}

	// Finally print out buffer to stdout.
	fmt.Printf("%s", buffer.String())
}

func printMissingRequired(buffer *bytes.Buffer, keys []string, values url.Values) bool {
	var tempStr bytes.Buffer

	var found bool = false // whether we found missing vars
	for _, key := range keys {
		_, ok := values[key]
		if !ok {
			found = true
			tempStr.WriteString(fmt.Sprintf("%s ", key))
		}
	}

	if found {
		tempStr.WriteString("\n")
	}

	if tempStr.Len() > 0 {
		// Add padding if the string is small.
		if tempStr.Len() < 40 {
			var numSpaces int = (80 - tempStr.Len()) / 2
			buffer.WriteString(strings.Repeat(" ", numSpaces))
		}
		buffer.WriteString(tempStr.String())
	}

	return found
}

func printUnusedKeys(buffer *bytes.Buffer, keys []string, values url.Values) bool {
	var found bool = false
	var newline bool = true
	count := 0
	for i := range keys {
		key := keys[i]
		arr, ok := values[key]
		if !ok || len(arr) <= 0 {
			found = true
			if newline {
				buffer.WriteString("          ")
				newline = false
			}

			buffer.WriteString(fmt.Sprintf("%s ", key))
			if count++; count%20 == 0 {
				buffer.WriteString("\n")
				newline = true
			}
		}
	}
	if found {
		buffer.WriteString("\n")
	}

	return found
}

func printUnknownKeys(buffer *bytes.Buffer, key_set MySet, values url.Values) bool {
	var found bool = false
	var unknownKeys []string

	// Determine which form values are new.
	for k, _ := range values {
		if _, ok := key_set[k]; !ok {
			found = true
			unknownKeys = append(unknownKeys, k)
		}
	}

	for _, key := range unknownKeys {
		arr := values[key]
		if len(arr) > 0 {
			buffer.WriteString(fmt.Sprintf("%40s = %s\n", key, arr[0]))
		} else {
			buffer.WriteString(fmt.Sprintf("%40s = <no value>\n", key))
		}
	}
	return found
}

func makeSet(strings []string) MySet {
	var newSet MySet = make(MySet)

	for _, v := range strings {
		newSet[v] = true
	}

	return newSet
}

func sendEmail(msg string) {
	auth := smtp.PlainAuth("", "email@domain.com", "password", "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:25", auth, "recipient@destination.com", []string{"sender@source.com"}, []byte(msg))
	if err != nil {
		fmt.Printf("Failed to send alert email: %s", msg)
		//log.Fatal(err)
	}
}

func main() {
	fmt.Printf("Starting fake analytics server.\n")

	// create set to find if there are any new params
	g_keySet = makeSet(g_allKeys)
	g_keySet["dl"] = true // tack this on.

	http.HandleFunc("/", handler)
	//http.ListenAndServe(":8080", nil)
	err := http.ListenAndServe(":"+os.Getenv("PORT"), nil)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
