# 单机百万级别QPS 网络传输



**TLDR**:

- **UDP**
  
  - 3万 qps的时候偶尔会有一两个丢包。
  
  - 5万 qps的时候丢包严重且频繁，不适合做可信赖的网络探测
  
  - 非可信赖的通讯情况下，可以达到接近百万的QPS

- **TCP**
  
  - 3万 qps的时候丢包就比较严重了。TCP重传厉害
  
  - 非可信赖的通讯情况下，可以达到接近280万的QPS （理论还可以增加，但是client CPU已经100%了）


**注意**： "非可信赖" 指的是可能会发生丢包情况，UDP程序要么忽略，要么通过一种机制进行重发。 TCP本身的重传机制可以保证数据不丢。



## 1、背景

测试两台机器之间UDP、TCP通讯的benchmark。

主要探测两个场景

- 满足可信赖的通讯的qps

- 支持的最大qps



## 2、环境设置

client和server连接到同一台Tor，尽量减少网络的干扰。



**客户端机器**: xxx.xxx.xxx.xxx

Intel(R) Xeon(R) Gold 5118 CPU @ 2.30GHz 两颗

12物理核 48个逻辑核

187GB内存



**服务端机器**: xxx.xxx.xxx.xxx

Intel(R) Xeon(R) Gold 5118 CPU @ 2.30GHz 两颗

12物理核 48个逻辑核

187GB内存



设置了网卡多队列。



## 3、UDP 百万 QPS 测试

- 服务端端口 48 个

- 客户端端口 48 个

- 客户端和服务端端口按顺序建立"连接"



**TLDR**: 

- 3万 qps的时候偶尔会有一两个丢包。

- 5万 qps的时候丢包严重且频繁，不适合做可信赖的网络探测

- 非可信赖的通讯情况下，可以达到接近百万的QPS



### 3.1 3万 QPS 压测

<img src="images/2022-12-14-14-08-40-image.png" title="" alt="" width="621">

client:

<img src="images/2022-12-14-14-07-34-image.png" title="" alt="" width="617">

server:

<img src="images/2022-12-14-14-08-20-image.png" title="" alt="" width="617">



### 3.2 10万 QPS 压测



<img src="images/2022-12-14-14-14-38-image.png" title="" alt="" width="700">



client:

<img src="images/2022-12-14-14-13-08-image.png" title="" alt="" width="707">



server:

<img src="images/2022-12-14-14-14-10-image.png" title="" alt="" width="729">







### 3.3 100万 QPS 压测



<img src="images/2022-12-14-14-19-20-image.png" title="" alt="" width="625">

client:

<img src="images/2022-12-14-14-18-34-image.png" title="" alt="" width="615">



server:

<img src="images/2022-12-14-14-19-02-image.png" title="" alt="" width="619">



## 4、TCP 百万 QPS 测试



- 服务端端口 48 个

- 客户端端口 48 个

- 客户端和服务端建立48条流



**TLDR**:

- 3万 qps的时候丢包就比较严重了。TCP重传厉害

- 非可信赖的通讯情况下，可以达到接近280万的QPS （理论还可以增加，但是client CPU已经100%了）
  
  

### 4.1  3万 QPS 压测


##### 1. dstat

client:

<img src="images/2022-12-14-12-35-24-image.png" title="" alt="" width="558">



server:

<img src="images/2022-12-14-12-35-37-image.png" title="" alt="" width="592">



##### 2. client log

<img src="images/2022-12-14-12-28-05-image.png" title="" alt="" width="590">





### 4.2 5万 QPS 压测


##### 1. dstat

client:

<img src="images/2022-12-14-12-17-26-image.png" title="" alt="" width="502">



server:

<img src="images/2022-12-14-12-17-05-image.png" title="" alt="" width="499">



##### 2. client log

<img src="images/2022-12-14-12-16-29-image.png" title="" alt="" width="422">



### 4.3 100万 QPS 压测


##### 1. dstat

client:

<img src="images/2022-12-14-11-53-52-image.png" title="" alt="" width="749">

server:

<img src="images/2022-12-14-11-54-19-image.png" title="" alt="" width="747">



##### 2. client log

<img src="images/2022-12-14-11-54-50-image.png" title="" alt="" width="610">



### 4.4 200万 QPS 压测

##### 1. dstat

client:

<img src="images/2022-12-14-11-40-21-image.png" title="" alt="" width="580">

server:

<img src="images/2022-12-14-11-41-00-image.png" title="" alt="" width="582">



##### 2. client log

<img src="images/2022-12-14-11-41-19-image.png" title="" alt="" width="646">



### 4.5 300万 QPS 压测

##### 1. dstat

client: 

<img src="images/2022-12-14-11-27-18-image.png" title="" alt="" width="598">

server:

<img src="images/2022-12-14-11-27-47-image.png" title="" alt="" width="601">


##### 2. client log

<img src="images/2022-12-14-11-26-40-image.png" title="" alt="" width="609">
