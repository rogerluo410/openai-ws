package grpc

import (
	_ "context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/rogerluo410/openai-ws/src/grpc/pb"
)

type server struct {
	pb.UnimplementedSpeechRecognitionServer
}

func (s *server) RecognizeStream(stream pb.SpeechRecognition_RecognizeStreamServer) error {
	return nil
}

func StartGrpcServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Failed to start up grpc server listening on: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSpeechRecognitionServer(s, &server{})
	log.Printf("Grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: %v", err)
	}
}