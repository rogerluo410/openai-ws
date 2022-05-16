package grpc

import (
	"io"
	"fmt"
	"context"
	"time"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/rogerluo410/openai-ws/src/grpc/pb"
)

var (
	asrHostPort = "stream-asr-prod.yitutech.com:50051"
	devId = "22642"
  devKey = "ZDVkZDEwZjg5ZDk4NDVlYjg2NjBmZTE2YTM2MDM2MWU="
)

// 返回依图grpc连接串
func yituAsrConn() (*grpc.ClientConn, error) {
	conn, err := grpc.Dial(asrHostPort, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Did not connect: %v", err)
		return nil, err
	}

	return conn, err
}

// 返回依图认证签名
func yituAsrAuth() string {
  ts := strconv.FormatInt(time.Now().Unix(), 10)
	idTs := devId + ts
  sigHash := hmac.New(sha256.New, []byte(devKey))
	sigHash.Write([]byte(idTs))
	signature := hex.EncodeToString(sigHash.Sum(nil))

	return fmt.Sprintf("%s,%s,%s", devId, ts, signature)
}

func YituAsrClient(sendMsg chan pb.StreamingSpeechRequest, receiveMsg chan pb.StreamingSpeechResponse, errorMsg chan string) error {
	conn, err := yituAsrConn()
	if err != nil {
		log.Error("Connect to Yitu grpc server failed %v", err)
		return err
	}
	defer conn.Close()

	c := pb.NewSpeechRecognitionClient(conn)	
	apiKey := yituAsrAuth()
  log.WithField("apiKey", apiKey).Info("依图签名")

	// 添加metadata元数据给依图服务端
	// ctx := context.Background()
	// ctx = metadata.AppendToOutgoingContext(ctx, "x-api-key", apiKey)
	md := metadata.Pairs("x-api-key", apiKey)
  ctx := metadata.NewOutgoingContext(context.Background(), md)

	stream, err := c.RecognizeStream(ctx)
	streamCtx := stream.Context()
	done := make(chan bool)

	if err := stream.Send(&pb.StreamingSpeechRequest{RequestPayload: &pb.StreamingSpeechRequest_AudioData{AudioData: []byte{'1','2','3'}}}); err != nil {
		log.Error("Yitu client send error111 %v", err)
	}

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.WithFields(log.Fields{"Err": err, "Resp": resp}).Error("Yitu grpc server receives error")
				errorMsg <- err.Error()
				return
			}
			receiveMsg <- *resp
			log.Printf("Received from Yitu grpc server %v", *resp)
		}
	}()

	go func() error {
		for {
			select {
			case <- streamCtx.Done():
				log.WithField("StreamCtx Errr", streamCtx.Err()).Info("Received from stream context err message")
				return streamCtx.Err()
			case request := <- sendMsg:
				log.WithField("Sending message", request).Info("Will send message to Yitu server")
				if err := stream.Send(&request); err != nil {
					log.Error("Yitu client send error %v", err)
				}
			}
		}
	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<- streamCtx.Done()
		if err := streamCtx.Err(); err != nil {
			log.WithField("streamCtx Errr", err).Info("Received from stream context err message")
		}
		close(done)
	}()

	<- done
	log.Printf("Finished Yitu grpc client.")
	return nil
}