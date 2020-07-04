<p align='center'>
    <img src='./logo.png' width='200px' height='80px'/>
</p>

[中文](./README_cn.md)

<!-- [![Test](https://github.com/Yoshiera/livego/workflows/Test/badge.svg)](https://github.com/Yoshiera/livego/actions?query=workflow%3ATest)
[![Release](https://github.com/Yoshiera/livego/workflows/Release/badge.svg)](https://github.com/Yoshiera/livego/actions?query=workflow%3ARelease) -->

Simple and efficient live broadcast server:
- Very simple to install and use;
- Pure Golang, high performance, and cross-platform;
- Supports commonly used transmission protocols, file formats, and encoding formats;

#### Supported transport protocols
- RTMP
- HTTP-FLV

#### Supported container formats
- FLV
- TS

#### Supported encoding formats
- H264
- AAC
- MP3

## Installation
After directly downloading the compiled [binary file](https://github.com/Yoshiera/livego/releases), execute it on the command line.

#### Boot from Docker
Run `docker run -p 1935:1935 -p 7001:7001 -p 7002:7002 -p 8090:8090 -d Yoshiera/livego` to start

#### Compile from source
1. Download the source code `git clone https://github.com/Yoshiera/livego.git`
2. Go to the livego directory and execute `go build` or `make build`

## Use
1. Start the service: execute the livego binary file or `make run` to start the livego service;
3. Upstream push: Push the video stream to `rtmp://localhost:1935/{serverName}/{serverKey}` through the` RTMP` protocol(default appname is `live`), for example, use `ffmpeg -re -i demo.flv -c copy -f flv rtmp://localhost:1935/{serverName}/{serverKey}` push([download demo flv](https://s3plus.meituan.net/v1/mss_7e425c4d9dcb4bb4918bbfa2779e6de1/mpack/default/demo.flv));
4. Downstream playback: The following three playback protocols are supported, and the playback address is as follows:
    - `FLV`:`http://127.0.0.1:7001/{serverName}/{serverChannel}.flv`
   
commandline options: 
```bash
./livego  -h
Usage of ./livego:
  -config string
        config file (default "livego.yaml")
```

file configurations:
| key | default value | remark |
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

### [Use with flv.js](https://github.com/gwuhaolin/blog/issues/3)

Interested in Golang? Please see [Golang Chinese Learning Materials Summary](http://go.wuhaolin.cn/)
