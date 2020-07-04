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
	// ErrUnsupportedVideoCodec means unsupported video codec
	ErrUnsupportedVideoCodec = fmt.Errorf("unsupported video codec")

	// ErrUnsupportedAudioCodec means unsupported audio codec
	ErrUnsupportedAudioCodec = fmt.Errorf("unsupported audio codec")
)

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
