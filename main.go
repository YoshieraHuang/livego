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
		serverName := configure.Config.GetString("server.name")
		serverChannel := configure.Config.GetString("server.channel")
		log.Infof("Address to pull stream: http://localhost%s/%s/%s.flv", httpflvAddr, serverName, serverChannel)
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

	rtmpServer = rtmp.NewServer(stream, hlsServer)

	defer func() {
		if r := recover(); r != nil {
			log.Error("RTMP server panic: ", r)
		}
	}()
	log.Info("RTMP Listen On ", rtmpAddr)
	serverName := configure.Config.GetString("server.name")
	serverKey := configure.Config.GetString("server.key")
	log.Infof("Addr to push stream: rtmp://localhost%s/%s/%s", rtmpAddr, serverName, serverKey)
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
	hlsServer := hls.NewServer()
	// startAPI(stream)
	startHTTPFlv(stream)
	startRtmp(stream, hlsServer)
}
