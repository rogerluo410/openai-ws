package main

import (
	"fmt"
	"context"
	"flag"
	"time"
	"io"
	"os"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pb "github.com/rogerluo410/openai-ws/src/grpc/pb"
)

const (
	STATUS_FIRST_FRAME    = 0
	STATUS_CONTINUE_FRAME = 1
	STATUS_LAST_FRAME     = 2
)

var (
	addr = flag.String("addr", "localhost:8090", "the address to connect to")
	timePerChunk = 0.2    // 多长时间发送一次音频数据，单位：s
	numChannel = 1         // 声道数
	numQuantify = 16       // 量化位数
	sampleRate = 16000     // 采样频率
	// 字节数 ＝ 采样频率 × 量化位数 × 时间(秒) × 声道数(时长为时间秒的音频大小为数据量大小) /8
	bytesPerChunk = int(int(float64(sampleRate) * float64(numQuantify) * float64(timePerChunk) * float64(numChannel)) / 8)
	file      = "./gametest.wav" //请填写您的音频文件路径
)

// 音频转写相关设置
func setAudioConfig() *pb.StreamingSpeechConfig {

	return &pb.StreamingSpeechConfig {
		// 音频的编码。对应aue为PCM
		// 采样率（范围为8000-48000）。
    AudioConfig: &pb.AudioConfig{SampleRate: int32(sampleRate), Aue: pb.AudioConfig_PCM},
		SpeechConfig: &pb.SpeechConfig{
      Lang: pb.SpeechConfig_MANDARIN,
			Scene: pb.SpeechConfig_GENERALSCENE,
			CustomWord: []string{"依图"},
			RecognizeType: pb.SpeechConfig_ALL,
			DisableConvertNumber: false,
			DisablePunctuation: false,
			WordsReplace: &pb.WordsReplace{Keywords: []string{"回忆"}, Replace: []string{"什么"}},
			KeyWords: []string{"英雄联盟"},
		},
	} 
}

func main() {
	flag.Parse()
	// Set up a connection to the server.
	conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	// 添加metadata元数据
	ctx := context.Background()
	ctx = metadata.AppendToOutgoingContext(ctx, "token", "1234")

	c := pb.NewSpeechRecognitionClient(conn)
	stream, err := c.RecognizeStream(ctx)
	streamCtx := stream.Context()
	done := make(chan bool)

	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.WithFields(log.Fields{"Err": err}).Error("Openai grpc proxy server receives error")
				return
			}
			log.WithField("Message", resp).Info("Received from Openai grpc proxy server")
		}
	}()

	go func() {
		var status = STATUS_FIRST_FRAME      //音频的状态信息，标识音频是第一帧，还是中间帧、最后一帧
		var intervel = 40 * time.Millisecond //发送音频间隔
		var buffer = make([]byte, bytesPerChunk)
		audioFile, err := os.Open(file)
		if err != nil {
			panic(err)
		}
		fmt.Println("Opened audio file")
		for {
			select {
			case <- streamCtx.Done():
				log.WithField("StreamCtx Err", streamCtx.Err()).Info("Received from stream context err message")
				return
			default:
			}

			switch status {
			case STATUS_FIRST_FRAME: //发送第一帧
			  request := &pb.StreamingSpeechRequest{RequestPayload: &pb.StreamingSpeechRequest_StreamingSpeechConfig{StreamingSpeechConfig: setAudioConfig()} }
				log.WithField("Msg", request).Info("send first")
        if err := stream.Send(request); err != nil {
					log.WithField("Err", err).Error("Yitu client send first error")
				}
				status = STATUS_CONTINUE_FRAME
			case STATUS_CONTINUE_FRAME:
				fmt.Println("send binary data")
				len, err := audioFile.Read(buffer)
				if err != nil {
					if err == io.EOF { //文件读取完了，改变status = STATUS_LAST_FRAME
						status = STATUS_LAST_FRAME
						continue
					} else {
						panic(err)
					}
				}
        if err := stream.Send(&pb.StreamingSpeechRequest{RequestPayload: &pb.StreamingSpeechRequest_AudioData{AudioData: buffer[:len]} }); err != nil {
					log.WithField("Err", err).Error("Yitu client send audio data error")
				}
			case STATUS_LAST_FRAME:
				fmt.Println("send last")
				return
			}

			//模拟音频采样间隔
			time.Sleep(intervel)
		}

	}()

	// third goroutine closes done channel
	// if context is done
	go func() {
		<- streamCtx.Done()
		if err := streamCtx.Err(); err != nil {
			log.WithField("streamCtx Err", err).Info("Received from stream context err message")
		}
		close(done)
	}()

	<- done
	log.Printf("Finished Yitu grpc client.")
	os.Exit(0)
}