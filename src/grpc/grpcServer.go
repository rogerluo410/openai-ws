package grpc

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"net"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/metadata"

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

// 认证服务器地址
var verfiyUrl = ""

func (s *mathServer) Max(srv pb.Math_MaxServer) error {

	log.Info("start new server")
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
	md, ok := metadata.FromIncomingContext(stream.Context()) // get context from stream
	if !ok {
		log.Error("Get metadata failed")
		return status.Errorf(codes.Unauthenticated, "Get token failed")
	}
	token, ok := md["token"]
  if !ok {
		log.Error("Get token from metadata failed")
		return status.Errorf(codes.Unauthenticated, "Get token failed")
	}
	if ok := verifyToken(token[0]); !ok {
		log.Error("Token is invalid")
		return status.Errorf(codes.Unauthenticated, "Token is invalid")
	}

	ctx := stream.Context()
	done := make(chan bool)
	SendMsg := make(chan *pb.StreamingSpeechRequest)
	ReceiveMsg := make(chan *pb.StreamingSpeechResponse)
	ErrorMsg := make(chan string)

	defer func() {
		close(done)
		if err := recover(); err != nil {
			log.Error("Channel is already closed...")
		}
	}()
	
	go func() {
		for {
			select {
			case <- ctx.Done():
				return
			default:  //需要加default， 否则阻塞在case <- ctx.Done()， 后面的流程不会执行
			}

			r, err := stream.Recv()
			if err == io.EOF {
				log.Error("Openai proxy server stream.Recv received io.EOF")
				return
				// return stream.Send(&pb.StreamingSpeechResponse{Status: &pb.StreamingSpeechStatus{ProcessedTimestamp: time.Now().Unix()}})
			}
			if err != nil {
				log.WithField("Err", err).Error("Openai proxy server stream.Recv received error")
				return
			}
			log.Info("Openai proxy server stream.Recv payload")

			SendMsg <- r
		}
	}()

	go func() {
		for {
			select {
			case <- ctx.Done():
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
		log.WithField("Err", err).Error("Failed to start up Yitu grpc client", err)
		return err
	}

	for {
		select {
		case error := <- ErrorMsg:
      log.WithField("Err", error).Error("Finished Openai grpc proxy with receiving error message from Yitu grpc server")
	    return status.Errorf(codes.InvalidArgument, error)
		case <- done:
			return nil
		default:
		}
	}

	log.Info("Finished Openai grpc proxy.")

	return nil
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func verifyToken(token string) bool {
	if len(token) == 0 {
		return false
	}
  apiUrl := verfiyUrl
	resource := "/api/v1/verfiy_token"
	data := url.Values{}
	data.Set("token", token)

	u, _ := url.ParseRequestURI(apiUrl)
	u.Path = resource
	urlStr := u.String() // "https://xxxx.com/api/v1/verfiy_token"

	client := &http.Client{}
	r, err := http.NewRequest(http.MethodPost, urlStr, strings.NewReader(data.Encode())) // URL-encoded payload

	if err != nil {
		log.Error(err)
		return false
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)

	defer func() {
    r.Close = true
		r.Body.Close();
		if resp != nil {
		  resp.Body.Close();
		}
	}()
	
	if err != nil {
		log.Error(err)
		return false
	}
	log.WithField("status", resp.Status).Info("token验证结果...")
  if "204 No Content" == resp.Status {
		return true
	} else {
		return false
	}
}


func StartGrpcServer(port string, url string) {
	verfiyUrl = url
	intVar, _ := strconv.Atoi(port)
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", intVar))
	if err != nil {
		log.WithField("Err", err).Fatalf("Failed to start up grpc server listening on %d", intVar)
	}
	s := grpc.NewServer()
	pb.RegisterSpeechRecognitionServer(s, &speechRecognitionServer{})
	pb.RegisterGreeterServer(s, &server{})
	pb.RegisterMathServer(s, &mathServer{})
	log.Printf("Grpc server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.WithField("Err", err).Fatal("Failed to serve grpc")
	}
}