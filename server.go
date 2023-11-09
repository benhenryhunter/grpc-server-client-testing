package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/benhenryhunter/grpc-server-client-testing/proto"
)

type Server struct {
	proto.StreamerServer
	Lock        sync.RWMutex
	StreamChans map[string]chan *proto.StreamStuffResponse
}

func (s *Server) StreamStuff(req *proto.StreamStuffRequest, srv proto.Streamer_StreamStuffServer) error {
	id := fmt.Sprintf("%v", req.Id)
	s.Lock.Lock()
	s.StreamChans[id] = make(chan *proto.StreamStuffResponse, 100)
	s.Lock.Unlock()

	for {
		select {
		case <-srv.Context().Done():
			delete(s.StreamChans, id)
			return nil
		case response := <-s.StreamChans[id]:
			if err := srv.Send(response); err != nil {
				return err
			}
		}
	}

}

func (s *Server) StartStreaming() {
	for {
		s.Lock.RLock()
		if len(s.StreamChans) <= 1 {
			s.Lock.RUnlock()
			time.Sleep(1 * time.Second)
			continue
		}
		num := len(s.StreamChans)
		chanToSend := rand.Intn(num) + 1
		c, ok := s.StreamChans[fmt.Sprintf("%v", chanToSend)]
		if ok {
			c <- &proto.StreamStuffResponse{Id: int32(chanToSend)}
		} else {
			println("channel not found")
		}
		s.Lock.RUnlock()

		time.Sleep(1 * time.Second)
	}
}

func NewServer() *Server {
	s := &Server{
		StreamChans: make(map[string]chan *proto.StreamStuffResponse),
	}

	go s.StartStreaming()
	return s
}
