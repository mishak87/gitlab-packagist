package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	help := flag.Bool("help", false, "Show usage")
	url := flag.String("url", "", "GitLab API url including version string ie. https://gitlab.com/api/v3/")
	token := flag.String("token", "", "GitLab API token")
	verify := flag.Bool("verify-ssl", true, "GitLab API verify SSL")
	httpAddr := flag.String("addr", ":7070", "HTTP server address")
	interval := flag.Uint("interval", 5, "Interval between updates")
	flag.Parse()

	if *help == true || *url == "" || *token == "" || *httpAddr == "" {
		flag.Usage()
		return
	}

	g := &GitLab{
		Url:   *url,
		Token: *token,
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !*verify,
				},
			},
		},
	}

	c := NewComposerRepository(g)

	go func() {
		for {
			log.Println("Fetching data...")
			if err := c.Update(); err != nil {
				log.Println(err)
			}
			time.Sleep(time.Duration(*interval) * time.Minute)
		}
	}()

	http.HandleFunc("/packages.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		x, err := c.Content()
		if err != nil {
			http.Error(w, fmt.Sprintf("%v", err), 500)
		}
		http.ServeContent(w, r, "packages.json", c.ModifiedTime(), x)
	})

	log.Printf("Server Listening on %v", *httpAddr)
	if err := http.ListenAndServe(*httpAddr, nil); err != nil {
		log.Fatal(err)
	}
}
