package grpc

import (
	_ "context"
	_ "time"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	_ "github.com/rogerluo410/openai-ws/src/grpc/pb"
)

var (
	asrHostPort = "stream-asr-prod.yitutech.com:50051"
	devId = "22642"
  devKey = "ZDVkZDEwZjg5ZDk4NDVlYjg2NjBmZTE2YTM2MDM2MWU="
)

func YituAsrConn() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(asrHostPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Did not connect: %v", err)
		return nil, err
	}

	return conn, err
}

func GetYituAsrClient(conn *grpc.ClientConn) {
	c := pb.NewSpeechRecognitionClient(conn)	
}

// func main() {
// 	c := pb.NewGreeterClient(conn)

// 	// Contact the server and print out its response.
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
// 	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
// 	if err != nil {
// 		log.Fatalf("could not greet: %v", err)
// 	}
// 	log.Printf("Greeting: %s", r.GetMessage())
// }