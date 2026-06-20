package main

import (
	"compress/gzip"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
	"meow.net/controllers/hub"
	"meow.net/controllers/player"
	"meow.net/db"
	"meow.net/routes"
	"meow.net/utils"
)

func getHost() string {
	if h := os.Getenv("HOST"); h != "" {
		return h
	}
	port := getPort()
	return "http://localhost" + port
}

func getCDNHost() string {
	if h := os.Getenv("CDN_HOST"); h != "" {
		return h
	}
	return getHost()
}

func getPort() string {
	if p := os.Getenv("PORT"); p != "" {
		if !strings.HasPrefix(p, ":") {
			return ":" + p
		}
		return p
	}
	return ":8080"
}

func getNSPort() string {
	p := os.Getenv("NS_PORT")
	if p == "" {
		return ""
	}
	if !strings.HasPrefix(p, ":") {
		return ":" + p
	}
	return p
}

const gzipMinSize = 1024

var gzipWriterPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz       *gzip.Writer
	buf      []byte
	status   int
	wrote    bool
	gzActive bool
}

func (g *gzipResponseWriter) WriteHeader(status int) {
	g.status = status
}

func (g *gzipResponseWriter) writeHeaderOnce() {
	if g.wrote {
		return
	}
	if g.status == 0 {
		g.status = http.StatusOK
	}
	g.ResponseWriter.WriteHeader(g.status)
	g.wrote = true
}

func (g *gzipResponseWriter) startGzip() {
	g.Header().Set("Content-Encoding", "gzip")
	g.Header().Del("Content-Length")
	g.writeHeaderOnce()
	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(g.ResponseWriter)
	g.gz = gz
	g.gzActive = true
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.gzActive {
		return g.gz.Write(b)
	}
	g.buf = append(g.buf, b...)
	if len(g.buf) >= gzipMinSize {
		g.startGzip()
		buffered := g.buf
		g.buf = nil
		if _, err := g.gz.Write(buffered); err != nil {
			return 0, err
		}
	}
	return len(b), nil
}

func (g *gzipResponseWriter) finish() {
	if g.gzActive {
		g.gz.Close()
		gzipWriterPool.Put(g.gz)
		g.gz = nil
		return
	}
	g.writeHeaderOnce()
	if len(g.buf) > 0 {
		g.ResponseWriter.Write(g.buf)
		g.buf = nil
	}
}

func noCacheMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
		}
		next.ServeHTTP(w, r)
	})
}

func getMaxConcurrentRequests() int {
	if v := os.Getenv("MAX_CONCURRENT_REQUESTS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return 100
}

func concurrencyLimitMiddleware(next http.Handler) http.Handler {
	limit := getMaxConcurrentRequests()
	if limit <= 0 {
		log.Printf("limiter disabled")
		return next
	}
	log.Printf("limiting concurrent requests to %d", limit)

	const acquireTimeout = 15 * time.Second
	sem := make(chan struct{}, limit)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}

		timer := time.NewTimer(acquireTimeout)
		defer timer.Stop()

		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next.ServeHTTP(w, r)
		case <-r.Context().Done():
		case <-timer.C:
			log.Printf("503 saturated")
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"error":"server busy"}`)
		}
	})
}

func rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !utils.RateLimitAllow(r) {
			log.Printf("[RATELIMIT] 429 ip=%s remote=%s %s %s", utils.ClientIP(r), r.RemoteAddr, r.Method, r.URL.Path)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"rate limit exceeded"}`)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func gzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") ||
			strings.EqualFold(r.Header.Get("Upgrade"), "websocket") {
			next.ServeHTTP(w, r)
			return
		}
		gw := &gzipResponseWriter{ResponseWriter: w}
		defer gw.finish()
		next.ServeHTTP(gw, r)
	})
}

func main() {
	seedOnly := flag.Bool("seed", false, "seed the database and exit")
	flag.Parse()

	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL must be set in .env file")
	}

	db.Connect(dsn)
	db.Migrate()

	if *seedOnly {
		db.Seed()
		return
	}

	utils.InitR2()
	utils.InitRateLimiter()

	player.BuildAvatarItemsCache()

	routes.RegisterRoutes()

	hub.StartInstanceReaper()

	host := getHost()
	cdnHost := getCDNHost()
	port := getPort()
	nsPort := getNSPort()

	patcherPubKey := ""
	if seed, err := base64.StdEncoding.DecodeString(os.Getenv("PATCHER_PRIVATE_KEY")); err == nil && len(seed) == ed25519.SeedSize {
		privateKey := ed25519.NewKeyFromSeed(seed)
		patcherPubKey = hex.EncodeToString(privateKey.Public().(ed25519.PublicKey))
	}

	writeNSResponse := func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"Accounts":              host,
			"API":                   host,
			"Auth":                  host,
			"BugReporting":          host,
			"CDN":                   cdnHost,
			"Chat":                  host,
			"Clubs":                 host,
			"CMS":                   host,
			"Commerce":              host,
			"DataCollection":        host,
			"Discovery":             host,
			"Econ":                  host,
			"GameLogs":              host,
			"Images":                cdnHost,
			"Leaderboard":           host,
			"Link":                  host,
			"Lists":                 host,
			"Matchmaking":           host,
			"Moderation":            host,
			"Notifications":         host,
			"PlatformNotifications": host,
			"PlayerSettings":        host,
			"RoomComments":          host,
			"Rooms":                 host,
			"Storage":               host,
			"Strings":               host,
			"StringsCDN":            cdnHost,
			"Studio":                host,
			"WWW":                   host,
		})
	}

	if nsPort != "" {
		nsMux := http.NewServeMux()
		nsMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			log.Printf("[NS] [%s] %s%s", r.Method, r.URL.Path, utils.QueryString(r))
			writeNSResponse(w)
		})
		go func() {
			log.Printf("Name server running on %s", nsPort)
			log.Fatal(http.ListenAndServe(nsPort, nsMux))
		}()
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		log.Printf("[%s] %s%s  body=%q", r.Method, r.URL.Path, utils.QueryString(r), body)
		if r.URL.Path == "/" && nsPort == "" {
			if r.Header.Get("key") != patcherPubKey {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, `{"error":"unauthorized"}`)
				return
			}
			writeNSResponse(w)
			return
		}

		log.Printf("[UNHANDLED] %s %s%s  body=%q", r.Method, r.URL.Path, utils.QueryString(r), body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprintf(w, `{"error":"unhandled route","path":%q}`, r.URL.Path)
	})

	log.Printf("Private server running on %s", port)
	log.Fatal(http.ListenAndServe(port, rateLimitMiddleware(concurrencyLimitMiddleware(gzipMiddleware(noCacheMiddleware(http.DefaultServeMux))))))
}
