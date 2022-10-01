package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

func LoadApp(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	appID := params["appid"]

	fmt.Printf("URL called with Param: %s", appID)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "<div><b>Hello</b> World <button style='color: green;'>Click me!</button></div>: App = %s\n", appID)
}

func VersionHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Zuri Chat API - Version 0.0255\n")
}

func RequestDurationMiddleware(h http.Handler) http.Handler {
	const durationLimit = 10

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, r)
		duration := time.Since(start)

		postToSlack := func() {
			m := make(map[string]interface{})
			m["timeTaken"] = duration.Seconds()

			if duration.Seconds() < durationLimit {
				return
			}

			scheme := "http"

			if r.TLS != nil {
				scheme += "s"
			}

			m["endpoint"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.URL.Path)
			m["timeTaken"] = duration.Seconds()

			b, _ := json.Marshal(m)
			resp, err := http.Post("https://companyfiles.zuri.chat/api/v1/slack/message", "application/json", strings.NewReader(string(b)))

			if err != nil {
				return
			}

			if resp.StatusCode != http.StatusOK {
				fmt.Printf("got error %d", resp.StatusCode)
			}

			defer resp.Body.Close()
		}

		if strings.Contains(r.Host, "api.zuri.chat") {
			go postToSlack()
		}
	})
}
