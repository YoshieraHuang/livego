<p align='center'>
    <img src='./logo.png' width='200px' height='80px'/>
</p>

简单高效的直播服务器：
- 安装和使用非常简单；
- 纯 Golang 编写，性能高，跨平台；
- 支持常用的传输协议、文件格式、编码格式；

#### 支持的传输协议
- RTMP
- HTTP-FLV

#### 支持的容器格式
- FLV
- TS

#### 支持的编码格式
- H264
- AAC
- MP3

## 安装
直接下载编译好的[二进制文件](https://github.com/Yoshiera/livego/releases)后，在命令行中执行。

#### 从 Docker 启动
执行`docker run -p 1935:1935 -p 7001:7001 -p 7002:7002 -p 8090:8090 -d Yoshiera/livego`启动

#### 从源码编译
1. 下载源码 `git clone https://github.com/Yoshiera/livego.git`
2. 去 livego 目录中 执行 `go build`

## 使用
1. 启动服务：执行 `livego` 二进制文件启动 livego 服务；
3. 推流: 通过`RTMP`协议推送视频流到地址 `rtmp://localhost:1935/{serverName}/{serverkey}` (appname默认是`live`), 例如： 使用 `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/{serverName}/{serverKey}` 推流([下载demo flv](https://s3plus.meituan.net/v1/mss_7e425c4d9dcb4bb4918bbfa2779e6de1/mpack/default/demo.flv));
4. 播放: 支持多种播放协议，播放地址如下:
    - `FLV`:`http://127.0.0.1:7001/{serverName}/{serverChannel}.flv`

命令行配置项: 
```bash
./livego  -h
Usage of ./livego:
  -config string
        config file (default "livego.yaml")
```

文件配置项:
| 配置项 | 默认值 | 备注 |
| :---------: | :--------: | :---- |
| `level` | `info` | logging level. Legal levels see [here](https://github.com/sirupsen/logrus) |
| `flv_dir` | `./tmp/app` | path to store flv cache |
| `rtmp_addr` | `:1935` | address for rtmp stream server to push in |
| `httpflv_addr` | `:7001` | address of http-flv stream out |
| `read_timeout`| 10 | reading timeout for stream in. unit is `s` |
| `write_timeout`| 10 | writeing timeout for stream in. unit is `s` |
| `gop_num` | 1 | number of gop |
| `server.name` | `live` | name of server |
| `server.channel` | `movie` | channel name |
| `server.key` | `123456` | key of server |

### [和 flv.js 搭配使用](https://github.com/Yoshiera/blog/issues/3)

对Golang感兴趣？请看[Golang 中文学习资料汇总](http://go.wuhaolin.cn/)

