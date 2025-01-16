package service

import (
	"catbox-scanner-master/internal/database"
	"context"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/panjf2000/ants/v2"
	"github.com/rs/dnscache"
)

type LinkChecker struct {
	db        *database.Database
	client    *http.Client
	isRunning bool
	wg        *sync.WaitGroup
	pool      ants.Pool
}

func NewLinkChecker(db *database.Database) *LinkChecker {
	r := &dnscache.Resolver{}
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false,
			MaxIdleConnsPerHost: 0,
			MaxConnsPerHost:     0,
			ForceAttemptHTTP2:   true,
			DialContext: func(ctx context.Context, network string, addr string) (conn net.Conn, err error) {
				host, port, err := net.SplitHostPort(addr)
				if err != nil {
					return nil, err
				}
				ips, err := r.LookupHost(ctx, host)
				if err != nil {
					return nil, err
				}
				for _, ip := range ips {
					var dialer net.Dialer
					conn, err = dialer.DialContext(ctx, network, net.JoinHostPort(ip, port))
					if err == nil {
						break
					}
				}
				return
			},
		},
	}

	pool, err := ants.NewPool(10)
	if err != nil {
		log.Fatalln(err.Error())
	}

	return &LinkChecker{
		db:     db,
		client: client,
		wg:     &sync.WaitGroup{},
		pool:   *pool,
	}
}

func (lc *LinkChecker) Start() {
	for lc.isRunning {
		lc.pool.Submit(lc.checkLinks)
	}
}

func (lc *LinkChecker) Stop() {
	lc.isRunning = false
	lc.wg.Wait()
}

func (lc *LinkChecker) checkLinks() {
	lc.wg.Add(1)
	defer lc.wg.Done()
	entries, err := lc.db.GetRandomEntries(10)
	if err != nil {
		log.Printf("Error fetching entries for link check: %v", err)
		return
	}

	for _, entry := range entries {
		if !lc.isRunning {
			return
		}
		url := "https://files.catbox.moe/" + entry.ID + "." + entry.Ext
		if !lc.checkLink(url) {
			log.Printf("Invalid link: %s. Removing from database.", url)
			if err := lc.db.RemoveEntry(entry.ID, entry.Ext); err != nil {
				log.Printf("Error removing invalid entry: %v", err)
			}
		}
	}
}

func (lc *LinkChecker) checkLink(url string) bool {
	resp, err := lc.client.Head(url)
	if err != nil {
		log.Printf("Error checking link: %s, %v", url, err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
