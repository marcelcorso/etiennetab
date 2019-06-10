package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/onrik/logrus/filename"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

func main() {

	port := os.Getenv("PORT")
	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	loglvl := os.Getenv("LOG_LEVEL")

	logLevel, err := log.ParseLevel(loglvl)
	if err != nil {
		log.Fatal(err)
	}
	log.SetLevel(logLevel)
	log.AddHook(filename.NewHook())

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	http.HandleFunc("/gifs.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err := os.Stat("/tmp/gifs.json"); err == nil {
			f, err := os.Open("/tmp/gifs.json")
			if err != nil {
				errors.Wrap(err, "err opening cache")
				log.Errorf("%v", err)
				return
			}
			_, err = io.Copy(w, f)
			if err != nil {
				errors.Wrap(err, "err copying from cache to response writer")
				log.Errorf("%v", err)
				return
			}
			return
		}

		tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
			ScreenName: "etiennejcb",
			Count:      500,
		})

		if err != nil {
			errors.Wrap(err, "err requesting timeline")
			log.Errorf("%v", err)
			return
		}

		var gifs []string
		for _, t := range tweets {
			if t.ExtendedEntities != nil {
				for _, me := range t.ExtendedEntities.Media {

					if (me.Type == "animated_gif") &&
						(len(me.VideoInfo.Variants) > 0) {

						gifs = append(gifs, me.VideoInfo.Variants[0].URL)
					}
				}
			}
		}

		json.NewEncoder(w).Encode(gifs)

		// cache
		f, err := os.Create("/tmp/gifs.json")
		if err != nil {
			errors.Wrap(err, "err opening cache")
			log.Errorf("%v", err)
			return
		}
		json.NewEncoder(f).Encode(gifs)
	})

	// clear cache
	ticker := time.NewTicker(6 * time.Hour)
	go func() {
		for range ticker.C {
			os.Remove("/tmp/gifs.json")
		}
	}()

	log.Debugf("listening for http on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
