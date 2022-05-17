# -*- coding: UTF-8 -*-
import threading
import time
import os

import sdk.real_time_asr_client as real_time_asr_client
import sdk.real_time_asr_pb2 as real_time_asr_pb2


audio_path = os.path.join(os.getcwd(), 'gametest.wav')
time_per_chunk = 0.2    # 多长时间发送一次音频数据，单位：s
num_channel = 1         # 声道数
num_quantify = 16       # 量化位数
sample_rate = 16000     # 采样频率
# 字节数 ＝ 采样频率 × 量化位数 × 时间(秒) × 声道数(时长为时间秒的音频大小为数据量大小) /8
bytes_per_chunk = int(sample_rate * num_quantify * time_per_chunk * num_channel / 8)


# 音频转写相关设置
def set_audio_config():
    # 音频相关设置。
    audioConfig = real_time_asr_pb2.AudioConfig(
        # 音频的编码。对应aue为PCM
        aue=real_time_asr_pb2.AudioConfig.AudioEncoding.PCM,  # 1,
        # 采样率（范围为8000-48000）。
        sampleRate=sample_rate)

    # 指定规则替换文本。
    wordsReplace = real_time_asr_pb2.WordsReplace(
        # 待替换的文本。最多支持100个词
        keywords=u"回忆",
        # 替换后的字符。不指定时替换为空。最多支持100个符号，和待替换文本一一对应。
        replace=u"什么"
    )

    # 识别相关设置。
    speechConfig = real_time_asr_pb2.SpeechConfig(
        # 转写的语言。标示lang为 MANDARIN
        lang=real_time_asr_pb2.SpeechConfig.Language.MANDARIN,  # 1,
        # 情景模式，针对不同的应用场景可定制模型，例如医疗。scene为GENERAL_SCENE
        scene=real_time_asr_pb2.SpeechConfig.Scene.GENERALSCENE,  # 0,
        # 识别类型（全部，仅逐句, 仅逐字）
        #   real_time_asr_pb2.SpeechConfig.RecognizeType.UTTERANCE: 逐句模式
        #   real_time_asr_pb2.SpeechConfig.RecognizeType.STREAMING: 逐字模式
        #   real_time_asr_pb2.SpeechConfig.RecognizeType.ALL: 全部
        recognizeType=real_time_asr_pb2.SpeechConfig.RecognizeType.ALL,  #
        # 统一数字的转换方式。默认false，开启阿拉伯数字能力。true为汉字一二三四五六七八九十。
        disableConvertNumber=False,  # 1,
        # 加标点。默认false，开启添加标点。
        disablePunctuation=False,  # 1,
        # 指定规则替换文本。
        wordsReplace=wordsReplace,
        # 待替换的文本。最多支持100个词
        keyWords=[u"英雄联盟"],
        # 自定义词语（支持中文2-4个字，中英混合4-8个字符）。
        customWord=[u"依图"],
    )

    # 音频流请求的相关设置。
    requestConfig = real_time_asr_pb2.StreamingSpeechConfig(
        # 音频设置。
        audioConfig=audioConfig,
        # 识别设置。
        speechConfig=speechConfig
    )
    return requestConfig


# 线程1：将待转写语音发送出去
class Thread1(threading.Thread):
    def __init__(self, asr_client):
        super(Thread1, self).__init__()
        self.asr_client = asr_client

    def run(self):
        fd = open(audio_path, 'rb')
        audio_data = fd.read(bytes_per_chunk)
        sleep_time = time.time()
        while len(audio_data) > 0:
            self.asr_client.post_audio(audio_data)
            audio_data = fd.read(bytes_per_chunk)
            # 延时时间补偿处理
            if (time.time() - sleep_time) < time_per_chunk:
                time.sleep(time_per_chunk - (time.time() - sleep_time))
            sleep_time = time.time()
        # 关闭连接
        self.asr_client.close()


# 线程2：读取转写结果
class Thread2(threading.Thread):
    def __init__(self, asr_client):
        super(Thread2, self).__init__()
        self.asr_client = asr_client

    def run(self):
        response_queue = self.asr_client.get_result_queue()
        # 判断是否转写结束且都已将全部转写结果读出
        while not self.asr_client.is_finished():
            while not response_queue.empty():
                response = response_queue.get()
                # print(response)
                print(response.result.bestTranscription.transcribedText)
            # 等待轮询获取转写结果
            time.sleep(time_per_chunk)
        print("转写结束")


def start():
    client = real_time_asr_client.RealTimeAsrClient()
    request_iter = real_time_asr_client.request_iter(set_audio_config(), client.request_queue)
    client.build_connection(request_iter)

    thread1 = Thread1(client)
    thread2 = Thread2(client)

    thread1.start()
    thread2.start()

    thread1.join()
    thread2.join()


if __name__ == '__main__':
    start()
