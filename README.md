PushServer设计
==========
![github](https://raw.githubusercontent.com/shawnfeng/imagsbed/master/github_push_service.png "puser")

<br/>

INSTALL
----------
### 1. 配置自己的go运行环境
### 2. clone 代码
> git clone https://github.com/shawnfeng/PushServer.git

### 3. 安装依赖的库
> go get github.com/fzzy/radix/redis
>
> go get github.com/sdming/gosnow
>
> go get code.google.com/p/goprotobuf/proto
>
> go get code.google.com/p/go-uuid/uuid

### 4. build
> cd PushServer/server/linker && go build
>
> cd PushServer/server/router && go build
>
> 根据自己的测试需求，跳转到测试目录(PushServer/test)build






一、连接状态变迁图
----------

### 1. 服务器状态变迁
![github](https://raw.githubusercontent.com/shawnfeng/imagsbed/master/server_state.jpg "server_state")

当服务需要下线时候，通过人工干预，设置下线状态
* 会触发服务给主动客户端下发reroute协议，来完成客户端的重新路由
* 客户端每有协议发送给服务器时候，服务器除了完成该协议功能之外，会额外给客户端回复一条reroute协议
* 为了实现服务器的平滑下线，下线的状态并不会影响客户端的正常协议逻辑，reroute的过程是完全依赖客户端主动完成的，服务器端只起到提示，通知的作用

### 2. 客户端状态变迁
![github](https://raw.githubusercontent.com/shawnfeng/imagsbed/master/client_state.jpg "client_state")



二、协议说明
----------

推送系统为了简化设计，协议采用基于数据报的方式进行发送，后期可以考虑连续性顺序性扩展。需要回执的消息如果没有回执，需要触发重传，或者激发其他策略，例如重新连接

接入层分包协议，完成TCP流到数据报的转化
![github](https://raw.githubusercontent.com/shawnfeng/imagsbed/master/push_proto.split.png "push_proto.split.png")



三个部分
1. 协议长度，使用varint 128编码<=4Byte，长度不包含pad
2. 协议体，google protobuf编码
3. 协议垫衬，跟在每个数据报之后，用于拆包校验，如果发现不是0，则数据通道损坏，重连clientid，不易变的



三、重传消息设计策略




