# -*- coding: UTF-8 -*-
import sys
sys.path.append('./sdk')
import queue
import grpc
import threading

import client_auth_service
import real_time_asr_pb2
import real_time_asr_pb2_grpc

# 实时语音连接地址
asr_host_port = 'localhost:8090'
# 依图语音开放平台分配的DevId
dev_id = ''
# 用来加密的DevKey，与DevId共同生成请求header参数
dev_key = ''


class RealTimeAsrClient:
    def __init__(self):
        self.stub = None
        self.isFinished = False
        self.request_iterator = None
        self.global_stream_id = None
        self.responses = None
        self.request_queue = queue.Queue()
        self.response_queue = queue.Queue()

    # 建立链接
    def build_connection(self, request_iterator):
        print("connect to " + asr_host_port)
        channel = grpc.insecure_channel(asr_host_port)
        self.request_iterator = request_iterator
        self.stub = real_time_asr_pb2_grpc.SpeechRecognitionStub(channel)
        auth = client_auth_service.get_auth_info(dev_id, dev_key)
        print(auth)
        # 获取转写结果
        self.responses = self.stub.RecognizeStream(self.request_iterator, metadata=auth)
        # 开启线程接收转写结果
        threading.Thread(target=self.get_response).start()

    # 把转写结果放入结果队列中
    def get_response(self):
        try:
            for response in self.responses:
                if response.result:
                    self.response_queue.put(response)
                    if not self.global_stream_id:
                        self.global_stream_id = response.globalStreamId
            self.isFinished = True
        except Exception:
            exit(1)

    # 将音频数据放入迭代器的请求队列，等待上传）
    def post_audio(self, audio_data):
        self.request_queue.put(audio_data)

    # 停止传输
    def close(self):
        self.request_queue.put(False)

    # 获取全局唯一ID
    def get_global_stream_id(self):
        return self.global_stream_id

    # 获取转写结果队列
    def get_result_queue(self):
        return self.response_queue

    # 是否转写结束
    def is_finished(self):
        return self.isFinished


class request_iter(object):
    def __init__(self, config, request_queue):
        self.config = config
        self.request_queue = request_queue
        self.isFirst = True

    def __iter__(self):
        return self

    def __next__(self):
        if self.isFirst:
            # 每次建立连接后第一个请求必须为Speech Config
            self.isFirst = False
            print(self.config)
            return real_time_asr_pb2.StreamingSpeechRequest(streamingSpeechConfig=self.config)
        else:
            audio_data = self.request_queue.get()
            if audio_data:
                return real_time_asr_pb2.StreamingSpeechRequest(audioData=audio_data)
            else:
                raise StopIteration
