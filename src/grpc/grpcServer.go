package grpc

import (
	"context"
	"fmt"
	"net"
	_ "time"
	"io"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	log.Println("Start RecognizeStream server")
	// Get token from context
	// md, ok := metadata.FromIncomingContext(stream.Context()) // get context from stream
  
	ctx := stream.Context()
	done := make(chan bool)
	SendMsg := make(chan *pb.StreamingSpeechRequest)
	ReceiveMsg := make(chan *pb.StreamingSpeechResponse)
	ErrorMsg := make(chan string)
	
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Channel is already closed...")
			}
		}()
		for {
			select {
			case <- ctx.Done():
				close(done)
				return
			default:  //需要加default， 否则阻塞在case <- ctx.Done()， 后面的流程不会执行
			}

			r, err := stream.Recv()
			if err == io.EOF {
				close(done)
				log.Error("Openai proxy server stream.Recv received io.EOF")
				return
				// return stream.Send(&pb.StreamingSpeechResponse{Status: &pb.StreamingSpeechStatus{ProcessedTimestamp: time.Now().Unix()}})
			}
			if err != nil {
				close(done)
				log.WithField("Err", err).Error("Openai proxy server stream.Recv received error")
				return
			}
			log.WithField("Response", r).Info("Openai proxy server stream.Recv payload")

			SendMsg <- r
		}
	}()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Error("Channel is already closed...")
			}
		}()
		for {
			select {
			case <- ctx.Done():
				close(done)
				return
			case response := <- ReceiveMsg:
				if err := stream.Send(response); err != nil {
					log.WithField("Err", err).Error("Openai proxy server stream.Send error")
				}
			default:
			}
		}
	}()

	// 创建依图grpc客户端
	if err := YituAsrClient(SendMsg, ReceiveMsg, ErrorMsg); err != nil {
		log.Error("Failed to start up Yitu grpc client: %v", err)
		return err
	}

	// go func() {
	// 	<- ctx.Done()
	// 	if err := ctx.Err(); err != nil {
	// 		log.Println(err)
	// 	}
	// 	close(done)
	// }()

	// time.Sleep(2 * time.Second)

	for {
		select {
		case error := <- ErrorMsg:
      log.WithField("Err message", error).Error("Received error message from Yitu grpc server")
	    return status.Errorf(codes.InvalidArgument, error)
		case <- done:
			return nil
		default:
		}
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
	pb.RegisterSpeechRecognitionServer(s, &speechRecognitionServer{})
	pb.RegisterGreeterServer(s, &server{})
	pb.RegisterMathServer(s, &mathServer{})
	log.Printf("Grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatal("Failed to serve: %v", err)
	}
}