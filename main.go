package main

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/admission/v1"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	responseBody = []byte("OK")
	responseType = "text/plain; charset=utf-8"
)

const (
	certFile = "/autoops-data/tls/tls.crt"
	keyFile  = "/autoops-data/tls/tls.key"
)

func exit(err *error) {
	if *err != nil {
		log.Println("exited with error:", (*err).Error())
		os.Exit(1)
	} else {
		log.Println("exited")
	}
}

func main() {
	var err error
	defer exit(&err)

	log.SetFlags(0)
	log.SetOutput(os.Stdout)

	requestID := uint64(0)
	requestLocker := &sync.Mutex{}

	s := &http.Server{
		Addr: ":443",
		Handler: http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			requestLocker.Lock()
			defer requestLocker.Unlock()
			requestID++
			// line start / line end
			lineStart := fmt.Sprintf("================== %s ==== #%d ==================", time.Now().Format(time.RFC3339), requestID)
			log.Println(lineStart)
			lineEnd := &strings.Builder{}
			for range lineStart {
				lineEnd.WriteRune('=')
			}
			defer log.Println(lineEnd.String())
			// proto / method / url
			log.Printf("\n%s %s %s", req.Proto, req.Method, req.URL.String())
			// headers
			log.Printf("\nHost: %s", req.Host)
			for k, vs := range req.Header {
				for _, v := range vs {
					log.Printf("%s: %s", k, v)
				}
			}

			// response
			var review v1.AdmissionReview
			if err := json.NewDecoder(req.Body).Decode(&review); err != nil {
				log.Println("Failed to decode a AdmissionReview:", err.Error())
				http.Error(rw, err.Error(), http.StatusBadRequest)
				return
			}

			reviewPretty, _ := json.MarshalIndent(&review, "", "  ")
			log.Printf("%s", reviewPretty)

			// response
			review.Response = &v1.AdmissionResponse{
				UID:     review.Request.UID,
				Allowed: true,
			}
			review.Request = nil

			reviewJSON, _ := json.Marshal(review)
			rw.Header().Set("Content-Type", "application/json")
			rw.Header().Set("Content-Length", strconv.Itoa(len(reviewJSON)))
			_, _ = rw.Write(reviewJSON)
		}),
	}

	// channels
	chErr := make(chan error, 1)
	chSig := make(chan os.Signal, 1)
	signal.Notify(chSig, syscall.SIGTERM, syscall.SIGINT)

	// start server
	go func() {
		log.Println("listening at :443")
		chErr <- s.ListenAndServeTLS(certFile, keyFile)
	}()

	// wait signal or failed start
	select {
	case err = <-chErr:
	case sig := <-chSig:
		log.Println("signal caught:", sig.String())
		_ = s.Shutdown(context.Background())
	}
}
