package hls

import (
	"fmt"
	"net"
	"time"

	"github.com/Yoshiera/livego/av"

	cmap "github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"
)

const (
	duration = 3000
)

var (
	// // ErrNoPublisher means no publisher
	// ErrNoPublisher = fmt.Errorf("no publisher")

	// // ErrInvalidReq means invalid req url path
	// ErrInvalidReq = fmt.Errorf("invalid req url path")

	// ErrUnsupportedVideoCodec means unsupported video codec
	ErrUnsupportedVideoCodec = fmt.Errorf("unsupported video codec")

	// ErrUnsupportedAudioCodec means unsupported audio codec
	ErrUnsupportedAudioCodec = fmt.Errorf("unsupported audio codec")
)

// var crossdomainxml = []byte(
// 	`<?xml version="1.0" ?>
// <cross-domain-policy>
// 	<allow-access-from domain="*" />
// 	<allow-http-request-headers-from domain="*" headers="*"/>
// </cross-domain-policy>`)

// Server is a HLS server
type Server struct {
	listener net.Listener
	conns    cmap.ConcurrentMap
}

// NewServer returns a Server
func NewServer() *Server {
	ret := &Server{
		conns: cmap.New(),
	}
	go ret.checkStop()
	return ret
}

// // Serve serves http requests
// func (server *Server) Serve(listener net.Listener) error {
// 	mux := http.NewServeMux()
// 	mux.HandleFunc("/", server.handle)
// 	server.listener = listener
// 	http.Serve(listener, mux)
// 	return nil
// }

// Writer get writer
func (server *Server) Writer(info av.Info) av.WriteCloser {
	var s *Source
	ok := server.conns.Has(info.Key)
	if !ok {
		log.Debug("new hls source")
		s = NewSource(info)
		server.conns.Set(info.Key, s)
	} else {
		v, _ := server.conns.Get(info.Key)
		s = v.(*Source)
	}
	return s
}

// func (server *Server) getConn(key string) *Source {
// 	v, ok := server.conns.Get(key)
// 	if !ok {
// 		return nil
// 	}
// 	return v.(*Source)
// }

func (server *Server) checkStop() {
	for {
		<-time.After(5 * time.Second)
		for item := range server.conns.IterBuffered() {
			v := item.Val.(*Source)
			if !v.Alive() {
				log.Debug("check stop and remove: ", v.Info())
				server.conns.Remove(item.Key)
			}
		}
	}
}

// func (server *Server) handle(w http.ResponseWriter, r *http.Request) {
// 	if path.Base(r.URL.Path) == "crossdomain.xml" {
// 		w.Header().Set("Content-Type", "application/xml")
// 		w.Write(crossdomainxml)
// 		return
// 	}
// 	switch path.Ext(r.URL.Path) {
// 	case ".m3u8":
// 		key, _ := server.parseM3u8(r.URL.Path)
// 		conn := server.getConn(key)
// 		if conn == nil {
// 			http.Error(w, ErrNoPublisher.Error(), http.StatusForbidden)
// 			return
// 		}
// 		tsCache := conn.GetCacheInc()
// 		if tsCache == nil {
// 			http.Error(w, ErrNoPublisher.Error(), http.StatusForbidden)
// 			return
// 		}
// 		body, err := tsCache.GenM3U8PlayList()
// 		if err != nil {
// 			log.Debug("GenM3U8PlayList error: ", err)
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}

// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Cache-Control", "no-cache")
// 		w.Header().Set("Content-Type", "application/x-mpegURL")
// 		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
// 		w.Write(body)
// 	case ".ts":
// 		key, _ := server.parseTs(r.URL.Path)
// 		conn := server.getConn(key)
// 		if conn == nil {
// 			http.Error(w, ErrNoPublisher.Error(), http.StatusForbidden)
// 			return
// 		}
// 		tsCache := conn.GetCacheInc()
// 		item, err := tsCache.GetItem(r.URL.Path)
// 		if err != nil {
// 			log.Debug("GetItem error: ", err)
// 			http.Error(w, err.Error(), http.StatusBadRequest)
// 			return
// 		}
// 		w.Header().Set("Access-Control-Allow-Origin", "*")
// 		w.Header().Set("Content-Type", "video/mp2ts")
// 		w.Header().Set("Content-Length", strconv.Itoa(len(item.Data)))
// 		w.Write(item.Data)
// 	}
// }

// func (server *Server) parseM3u8(pathstr string) (key string, err error) {
// 	pathstr = strings.TrimLeft(pathstr, "/")
// 	key = strings.Split(pathstr, path.Ext(pathstr))[0]
// 	return
// }

// func (server *Server) parseTs(pathstr string) (key string, err error) {
// 	pathstr = strings.TrimLeft(pathstr, "/")
// 	paths := strings.SplitN(pathstr, "/", 3)
// 	if len(paths) != 3 {
// 		err = fmt.Errorf("invalid path=%s", pathstr)
// 		return
// 	}
// 	key = paths[0] + "/" + paths[1]

// 	return
// }
