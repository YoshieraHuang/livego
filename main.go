package main

import (
	"fmt"
	"net"
	"path"
	"runtime"
	"time"

	"github.com/Yoshiera/livego/configure"
	"github.com/Yoshiera/livego/protocol/hls"
	"github.com/Yoshiera/livego/protocol/httpflv"
	"github.com/Yoshiera/livego/protocol/rtmp"

	log "github.com/sirupsen/logrus"
)

const version = "master"

func startHls() *hls.Server {
	hlsAddr := configure.Config.GetString("hls_addr")
	hlsListen, err := net.Listen("tcp", hlsAddr)
	if err != nil {
		log.Fatal(err)
	}

	hlsServer := hls.NewServer()
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Error("HLS server panic: ", r)
			}
		}()
		log.Info("HLS listen On ", hlsAddr)
		hlsServer.Serve(hlsListen)
	}()
	return hlsServer
}

var rtmpAddr string

func startHTTPFlv(stream *rtmp.Streams) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("HTTP-FLV server panic: ", r)
		}
	}()

	httpflvAddr := configure.Config.GetString("httpflv_addr")

	flvListen, err := net.Listen("tcp", httpflvAddr)
	if err != nil {
		log.Fatal(err)
	}

	hdlServer := httpflv.NewServer(stream)
	go func() {
		log.Info("HTTP-FLV listen On ", httpflvAddr)
		hdlServer.Serve(flvListen)
	}()
}

func startRtmp(stream *rtmp.Streams, hlsServer *hls.Server) {
	rtmpAddr = configure.Config.GetString("rtmp_addr")

	rtmpListen, err := net.Listen("tcp", rtmpAddr)
	if err != nil {
		log.Fatal(err)
	}

	var rtmpServer *rtmp.Server

	if hlsServer == nil {
		rtmpServer = rtmp.NewServer(stream, nil)
		log.Info("HLS server disable....")
	} else {
		rtmpServer = rtmp.NewServer(stream, hlsServer)
		log.Info("HLS server enable....")
	}

	defer func() {
		if r := recover(); r != nil {
			log.Error("RTMP server panic: ", r)
		}
	}()
	log.Info("RTMP Listen On ", rtmpAddr)
	rtmpServer.Serve(rtmpListen)
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})
}

func main() {
	defer func() {
		if r := recover(); r != nil {
			log.Error("livego panic: ", r)
			time.Sleep(1 * time.Second)
		}
	}()

	log.Infof(`
     _     _            ____       
    | |   (_)_   _____ / ___| ___  
    | |   | \ \ / / _ \ |  _ / _ \ 
    | |___| |\ V /  __/ |_| | (_) |
    |_____|_| \_/ \___|\____|\___/ 
        version: %s
	`, version)

	stream := rtmp.NewStreams()
	hlsServer := startHls()
	// startAPI(stream)
	startHTTPFlv(stream)
	startRtmp(stream, hlsServer)
}
