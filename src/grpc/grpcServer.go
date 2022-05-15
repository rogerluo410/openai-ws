package grpc

import (
	"context"
	"fmt"
	"net"
	"time"
	"io"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/metadata"

	pb "github.com/rogerluo410/openai-ws/src/grpc/pb"
)

type speechRecognitionServer struct {
	pb.UnimplementedSpeechRecognitionServer
}

type server struct {
	pb.UnimplementedGreeterServer
}

type mathServer struct {
	pb.UnimplementedMathServer
}

func (s *mathServer) Max(srv pb.Math_MaxServer) error {

	log.Println("start new server")
	var max int32
	ctx := srv.Context()

	for {

		// exit if context is done
		// or continue
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// receive data from stream
		req, err := srv.Recv()
		if err == io.EOF {
			// return will close stream from server side
			log.Println("exit")
			return nil
		}
		if err != nil {
			log.Printf("receive error %v", err)
			continue
		}

		// continue if number reveived from stream
		// less than max
		if req.Num <= max {
			continue
		}

		// update max and send it to stream
		max = req.Num
		resp := pb.Response{Result: max}
		if err := srv.Send(&resp); err != nil {
			log.Printf("send error %v", err)
		}
		log.Printf("send new max=%d", max)
	}
}


func (s *speechRecognitionServer) RecognizeStream(stream pb.SpeechRecognition_RecognizeStreamServer) error {
	log.Println("start RecognizeStream server")
	// Get token from context
	// md, ok := metadata.FromIncomingContext(stream.Context()) // get context from stream
  
	ctx := stream.Context()
	done := make(chan bool)
	SendMsg := make(chan pb.StreamingSpeechRequest)
	ReceiveMsg := make(chan pb.StreamingSpeechResponse)

	// 创建依图grpc客户端
	if err := YituAsrClient(SendMsg, ReceiveMsg); err != nil {
		log.Fatal("Failed to start up Yitu grpc client: %v", err)
		return err
	}
	
	go func() error {
		for {
			select {
			case <- ctx.Done():
				return ctx.Err()
			default:
			}

			r, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return stream.Send(&pb.StreamingSpeechResponse{Status: &pb.StreamingSpeechStatus{ProcessedTimestamp: time.Now().Unix()}})
			}
			if err != nil {
				return err
			}
			log.Printf("stream.Recv payload: %s", r)

			SendMsg <- *r
		}
	}()

	go func() error {
		for {
			select {
			case <- ctx.Done():
				return ctx.Err()
			case response := <- ReceiveMsg:
				if err := stream.Send(&response); err != nil {
					log.Printf("send error %v", err)
				}
			}
		}
	}()

	go func() {
		<- ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<- done

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
	pb.RegisterSpeechRecognitionServer(s, &speechRecognitionServer{})
	pb.RegisterGreeterServer(s, &server{})
	pb.RegisterMathServer(s, &mathServer{})
	log.Printf("Grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: %v", err)
	}
}