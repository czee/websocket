// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build ignore
// +build ignore

package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
	"crypto/tls"
	"net/http"

	"github.com/gorilla/websocket"
)

//var addr = flag.String("addr", "bla.finza.tech:443", "http service address")
//var addr = flag.String("addr", "localhost:443", "http service address")
//var addr = flag.String("addr", "127.0.0.1:80", "http service address")
//var addr = flag.String("addr", "127.0.0.1:8081", "http service address")
//var addr = flag.String("addr", "localhost:8081", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	//u := url.URL{Scheme: "ws", Host: "127.0.0.1:8001", Path: "/echo"}
	u := url.URL{Scheme: "wss", Host: "127.0.0.1:8002", Path: "/echo"}
	//u := url.URL{Scheme: "wss", Host: *addr, Path: "/echo"}
	//u := url.URL{Scheme: "ws", Host: *addr, Path: "/~bl/wsecho"}
	log.Printf("connecting to %s", u.String())

	header := make(http.Header)
	//header.Add("Host", "127.0.0.1:5000")
	header.Add("Host", "127.0.0.1:5001")

	dialer := *websocket.DefaultDialer
    	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	c, _, err := dialer.Dial(u.String(), header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			//err = c.WriteMessage(websocket.TextMessage, []byte("test"))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
