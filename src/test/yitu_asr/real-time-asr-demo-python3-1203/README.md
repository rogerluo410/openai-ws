# 环境安装

1. 使用环境为python3，第一次使用需要安装依赖，执行 real-time-demo 下的requirements.txt
```
sudo apt-get install python3-pip
sudo pip3 install --upgrade pip3
cd real-time-demo
sudo pip3 install -r requirements.txt
```

1.1 如果安装依赖失败，先安装wheel，进入whl文件夹，执行whl/下的requirements.txt
```
sudo pip3 install wheel
cd ../whl
sudo pip3 install -r requirements.txt
```

1.1.b 如果1.1失败，或进入source文件夹，执行source/下的requirements.txt
```
cd ../source
sudo pip3 install -r requirements.txt
```

2.运行demo
```
cd ../real-time-demo
sudo python3 real_time_asr_example.py
```


