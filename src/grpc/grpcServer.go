package grpc

import (
	"context"
	"fmt"
	"net"
	"time"
	"io"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	pb "github.com/rogerluo410/openai-ws/src/grpc/pb"
)

type SpeechRecognition struct {
	pb.UnimplementedSpeechRecognitionServer
}

type server struct {
	pb.UnimplementedGreeterServer
}

func (s *SpeechRecognition) RecognizeStream(stream pb.SpeechRecognition_RecognizeStreamServer) error {
	for {
		r, err := stream.Recv()
		if err == io.EOF {
			return stream.Send(&pb.StreamingSpeechResponse{Status: &pb.StreamingSpeechStatus{ProcessedTimestamp: time.Now().Unix()}})
		}
		if err != nil {
			return err
		}
		log.Printf("stream.Recv payload: %s", r.GetRequestPayload())
	}
	return nil
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}


func StartGrpcServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatal("Failed to start up grpc server listening on: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterSpeechRecognitionServer(s, &SpeechRecognition{})
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("Grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: %v", err)
	}
}