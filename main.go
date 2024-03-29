package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/onrik/logrus/filename"
	"github.com/pkg/errors"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
)

func worker(handle string, client *twitter.Client, wg *sync.WaitGroup, gifs chan [3]string) {
	defer wg.Done()

	tweets, _, err := client.Timelines.UserTimeline(&twitter.UserTimelineParams{
		ScreenName: handle,
		Count:      500,
	})

	if err != nil {
		errors.Wrap(err, "err requesting timeline")
		log.Errorf("%v", err)
		return
	}

	for _, t := range tweets {
		if t.ExtendedEntities != nil {
			for _, me := range t.ExtendedEntities.Media {

				if ((me.Type == "animated_gif") || (me.Type == "video")) &&
					(len(me.VideoInfo.Variants) > 0) {

					if strings.Contains(me.VideoInfo.Variants[0].URL, ".m3u8") {
						// TODO dunno how to play this yet. find out
						// https://stackoverflow.com/questions/19782389/playing-m3u8-files-with-html-video-tag
						// https://stackoverflow.com/questions/23388135/how-to-play-html5-video-play-m3u8-on-mobile-and-desktop/23388308#23388308
						continue
					}
					gifs <- [3]string{
						handle,
						strconv.FormatInt(t.ID, 10),
						me.VideoInfo.Variants[0].URL,
					}
				}
			}
		}
	}
}

func main() {

	port := os.Getenv("PORT")
	consumerKey := os.Getenv("CONSUMER_KEY")
	consumerSecret := os.Getenv("CONSUMER_SECRET")
	accessToken := os.Getenv("ACCESS_TOKEN")
	accessTokenSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	loglvl := os.Getenv("LOG_LEVEL")

	logLevel, err := log.ParseLevel(loglvl)
	if err != nil {
		log.SetLevel(log.DebugLevel)
		log.WithFields(log.Fields{"error": err}).Error("cannot parse log level. Setting to DEBUG")
	} else {
		log.SetLevel(logLevel)
	}

	log.SetLevel(logLevel)
	log.AddHook(filename.NewHook())

	config := oauth1.NewConfig(consumerKey, consumerSecret)
	token := oauth1.NewToken(accessToken, accessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	mux := http.NewServeMux()

	var expires string

	mux.HandleFunc("/gifs.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if _, err := os.Stat("/tmp/gifs.json"); err == nil {
			f, err := os.Open("/tmp/gifs.json")
			if err != nil {
				errors.Wrap(err, "err opening cache")
				log.Errorf("%v", err)
				return
			}

			w.Header().Set("Cache-Control", "public")
			w.Header().Set("Expires", expires)

			_, err = io.Copy(w, f)
			if err != nil {
				errors.Wrap(err, "err copying from cache to response writer")
				log.Errorf("%v", err)
				return
			}
			return
		}

		handles := []string{"etiennejcb", "KangarooPhysics", "jn3008", "satoshi_aizawa", "RavenKwok", "nicolasdnl", "kyndinfo", "DirkKoy", "connrbell", "HedronApp", "DirkKoy", "yuruyurau", "Yugemaku", "MAKIO135", "zozuar", "JosefBsharah", "incre_ment", "rustysniper1", "MIRAI_MIZUE", "MacTuitui", "tdhooper", "dansmath", "jagarikin", "cs_kaplan", "any_user", "Mark_Lindner", "tompop99", "cacheflowe", "KeishiroUeki", "Bbbn192", "kimasendorf", "pickover", "TaterGFX", "marioecg", "canvas_51", "lejeunerenard", "ShanzhaiB", "dstern_motion", "AkiyoshiKitaoka", "TatsuyaBot", "BilndArt", "kamoshika_vrc", "davestrickgifs", "voorbeeld", "amandaghassaei", "smangiat", "sjpalmer1994", "beesandbombs", "Yann_LeGall", "kusakarism", "QuentinHocde", "echophons", "hexeosis", "kurtruslfanclub", "richprjcts", "p1xelfool", "p4stoboy", "wanoco4D", "andreasgysin", "smjtyazdi", "LorenBednar", "kineticgraphics", "skybase", "Nicolas_Sassoon", "tasty_plots", "newrafael", "ziyangwen", "shiffman", "Gakutadar", "derek_morash", "Poppel20", "RiversHaveWings", "HughesMichi", "bluecocoa", "midjourney"}

		gifchan := make(chan [3]string)
		var wg sync.WaitGroup

		for _, handle := range handles {
			wg.Add(1)
			go worker(handle, client, &wg, gifchan)
		}

		var gifs [][3]string
		go func() {
			for g := range gifchan {
				gifs = append(gifs, g)
			}
		}()

		wg.Wait()
		close(gifchan)

		w.Header().Set("Cache-Control", "public")
		w.Header().Set("Expires", expires)
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

	s := rand.NewSource(time.Now().Unix())
	rn := rand.New(s) // initialize local pseudorandom generator
	mux.HandleFunc("/one.mp4", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "video/mp4")

		if _, err := os.Stat("/tmp/gifs.json"); err == nil {
			f, err := os.Open("/tmp/gifs.json")
			if err != nil {
				errors.Wrap(err, "err opening cache")
				log.Errorf("%v", err)
				return
			}

			var gifs [][3]string
			err = json.NewDecoder(f).Decode(gifs)

			// pick a vid
			o := gifs[rn.Intn(len(gifs))]
			videoURL := o[2]

			resp, err := http.Get(videoURL)
			if err != nil {
				errors.Wrap(err, "error fetching vid from the twitz")
				log.Errorf("%v", err)
				return
			}

			// read from twitter, write to response
			_, err = io.Copy(w, resp.Body)
			if err != nil {
				errors.Wrap(err, "err copying from the twittz to response writer")
				log.Errorf("%v", err)
				return
			}
			return
		}
	})

	burstCacheEvery := 6 * time.Hour
	expires = time.Now().Add(burstCacheEvery).Format(http.TimeFormat)
	// clear cache
	ticker := time.NewTicker(burstCacheEvery)
	go func() {

		for range ticker.C {
			os.Remove("/tmp/gifs.json")
			expires = time.Now().Add(burstCacheEvery).Format(http.TimeFormat)
		}
	}()

	handler := cors.Default().Handler(mux)

	log.Debugf("listening for http on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
