package rtmp

import (
	"fmt"
	"time"

	"github.com/Yoshiera/livego/av"
	"github.com/Yoshiera/livego/protocol/rtmp/cache"

	cmap "github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"
)

var (
	emptyID = ""
)

// Streams is the streams of rtmp
type Streams struct {
	streams cmap.ConcurrentMap //key
}

// NewStreams returns RtmpStream
func NewStreams() *Streams {
	ret := &Streams{
		streams: cmap.New(),
	}
	go ret.CheckAlive()
	return ret
}

// HandleReader handles reader
func (rs *Streams) HandleReader(r av.ReadCloser) {
	info := r.Info()
	log.Debugf("HandleReader: info[%v]", info)

	var stream *Stream
	i, ok := rs.streams.Get(info.Key)
	if stream, ok = i.(*Stream); ok {
		stream.TransStop()
		id := stream.ID()
		if id != emptyID && id != info.UID {
			ns := NewStream()
			stream.Copy(ns)
			stream = ns
			rs.streams.Set(info.Key, ns)
		}
	} else {
		stream = NewStream()
		rs.streams.Set(info.Key, stream)
		stream.info = info
	}

	stream.AddReader(r)
}

// HandleWriter handles writer
func (rs *Streams) HandleWriter(w av.WriteCloser) {
	info := w.Info()
	log.Debugf("HandleWriter: info[%v]", info)

	var s *Stream
	ok := rs.streams.Has(info.Key)
	if !ok {
		s = NewStream()
		rs.streams.Set(info.Key, s)
		s.info = info
	} else {
		item, ok := rs.streams.Get(info.Key)
		if ok {
			s = item.(*Stream)
			s.AddWriter(w)
		}
	}
}

// GetStreams get streams
func (rs *Streams) GetStreams() cmap.ConcurrentMap {
	return rs.streams
}

// CheckAlive check if this stream is alive
func (rs *Streams) CheckAlive() {
	for {
		<-time.After(5 * time.Second)
		for item := range rs.streams.IterBuffered() {
			v := item.Val.(*Stream)
			if v.CheckAlive() == 0 {
				rs.streams.Remove(item.Key)
			}
		}
	}
}

// Stream is one rtmp stream
type Stream struct {
	isStart bool
	cache   *cache.Cache
	r       av.ReadCloser
	ws      cmap.ConcurrentMap
	info    av.Info
}

// PackWriterCloser is a WriteCloser for packet
type PackWriterCloser struct {
	init bool
	w    av.WriteCloser
}

// Writer gets WriteCloser
func (p *PackWriterCloser) Writer() av.WriteCloser {
	return p.w
}

// NewStream returns a Stream
func NewStream() *Stream {
	return &Stream{
		cache: cache.NewCache(),
		ws:    cmap.New(),
	}
}

// ID returns ID
func (s *Stream) ID() string {
	if s.r != nil {
		return s.r.Info().UID
	}
	return emptyID
}

// Reader returns a ReadCloser
func (s *Stream) Reader() av.ReadCloser {
	return s.r
}

// Ws returns a ws
func (s *Stream) Ws() cmap.ConcurrentMap {
	return s.ws
}

// Copy copy this stream to dst stream
func (s *Stream) Copy(dst *Stream) {
	for item := range s.ws.IterBuffered() {
		v := item.Val.(*PackWriterCloser)
		s.ws.Remove(item.Key)
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
	}
}

// AddReader add a reader
func (s *Stream) AddReader(r av.ReadCloser) {
	s.r = r
	go s.TransStart()
}

// AddWriter add a writer
func (s *Stream) AddWriter(w av.WriteCloser) {
	info := w.Info()
	pw := &PackWriterCloser{w: w}
	s.ws.Set(info.UID, pw)
}

// TransStart start the transport
func (s *Stream) TransStart() {
	s.isStart = true
	var p av.Packet

	log.Debugf("TransStart: %v", s.info)

	for {
		if !s.isStart {
			s.closeInter()
			return
		}
		err := s.r.Read(&p)
		if err != nil {
			s.closeInter()
			s.isStart = false
			return
		}

		s.cache.Write(p)

		for item := range s.ws.IterBuffered() {
			v := item.Val.(*PackWriterCloser)
			if !v.init {
				//log.Debugf("cache.send: %v", v.w.Info())
				if err = s.cache.Send(v.w); err != nil {
					log.Debugf("[%s] send cache packet error: %v, remove", v.w.Info(), err)
					s.ws.Remove(item.Key)
					continue
				}
				v.init = true
			} else {
				newPacket := p
				//writeType := reflect.TypeOf(v.w)
				//log.Debugf("w.Write: type=%v, %v", writeType, v.w.Info())
				if err = v.w.Write(&newPacket); err != nil {
					log.Debugf("[%s] write packet error: %v, remove", v.w.Info(), err)
					s.ws.Remove(item.Key)
				}
			}
		}
	}
}

// TransStop stops the transport
func (s *Stream) TransStop() {
	log.Debugf("TransStop: %s", s.info.Key)

	if s.isStart && s.r != nil {
		s.r.Close(fmt.Errorf("stop old"))
	}

	s.isStart = false
}

// CheckAlive checks if this stream is alive or not
func (s *Stream) CheckAlive() (n int) {
	if s.r != nil && s.isStart {
		if s.r.Alive() {
			n++
		} else {
			s.r.Close(fmt.Errorf("read timeout"))
		}
	}
	for item := range s.ws.IterBuffered() {
		v := item.Val.(*PackWriterCloser)
		if v.w != nil {
			if !v.w.Alive() && s.isStart {
				s.ws.Remove(item.Key)
				v.w.Close(fmt.Errorf("write timeout"))
				continue
			}
			n++
		}

	}
	return
}

func (s *Stream) closeInter() {
	for item := range s.ws.IterBuffered() {
		v := item.Val.(*PackWriterCloser)
		if v.w != nil {
			if v.w.Info().IsInterval() {
				v.w.Close(fmt.Errorf("closed"))
				s.ws.Remove(item.Key)
				log.Debugf("[%v] player closed and remove\n", v.w.Info())
			}
		}
	}
}
