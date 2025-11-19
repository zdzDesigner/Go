## 协议


- `Fixed Header`
```md
 0                   1                   2                   3
 0   1   2    3   4   5   6   7   8901234567890123456789012345678
+---------------+---------------+-------------------------------+
|   控制包类型  | 特定包标志    |         剩余长度              |
|   (4 bits)    | (4 bits)      |       (1-4 bytes)             |
+-------------------------------+-------------------------------+


- 控制包类型
1: CONNECT     5: PUBACK      9:  SUBACK         13: PINGRESP
2: CONNACK     6: PUBREC      10: UNSUBSCRIBE    14: DISCONNECT
3: PUBLISH     7: PUBREL      11: UNSUBACK
4: PUBREC      8: SUBSCRIBE   12: PINGREQ
                              

- 特定包标志
CONNECT:   保留位必须为0
PUBLISH:   [DUP][QoS][QoS][RETAIN]
   DUP:    重复投递标志 (1位)
   QoS:    服务质量等级 (2位: 00=0, 01=1, 10=2)
   RETAIN: 持久化标志 (1位)


- 剩余长度
1. 使用变长编码表示`可变头`和`负载`的**总字节数**
2. 每个字节的低7位表示数值，最高位为延续标志
3. 最大支持 256MB 消息 (4字节时最大为 0xFFFFFFF)
```


- `Variable Header`(可选)
它在固定报头和负载之间。
可变报头的内容根据报文类型的不同而不同。
可变报头的报文标识符（Packet Identifier）字段存在于在多个类型的报文里。




## Packet Demo

- `CONNECT`
```md
Fixed Header: [0001][0000] + Remaining length
Variable Header:
  Protocol Name: [00][04] "MQTT"
  Protocol Level: [0x03] (3.1), [0x04] (3.1.1)
  Connect Flags: 8bits
  Keep Alive: [MSB][LSB] (秒)
Payload:
  ClientID: [length MSB][length LSB] + string
  (可选) Username: [length] + string
  (可选) Password: [length] + string


Connect Flags:
   7        6      5    4  3    2    1     0
+--------+------+----+--------+----+----+------+
| 用户名 | 密码 | WR |  QoS   | WF | CS | 保留 |
|   标志 | 标志 |    | (2位)  |    |    |      |
+--------+------+----+--------+----+----+------+
```
```md
[16 15 0 4 77 81 84 84 4 2 0 60 0 3 97 97 97]
16: 00010000
    控制包类型: 0001
    特定包标志: 0000
15: 剩余长度
    15个字节长度

0 4: Protocol Name([Length] [77 81 84 84: MQTT])
4:   Protocol Level([0x04] (3.1.1))
2:   Connect Flags(CS)
0 60: Keep Alive
0 3: Payload Length
97 97 97: aaa(负载)
    
```


- `CONNACK`
```md
Fixed Header: [0010][0000] + [00000010] (剩余长度=2)
Variable Header:
  Acknowledge Flags: [SP] (Session Present)
  Connect Return Code: 
    0x00: 接受
    0x01: 协议版本不支持
    0x02: ClientID 拒绝
    0x03: 服务不可用
    0x04: 用户名/密码错误
    0x05: 未授权
```

- `PUBLISH`
```md
Fixed Header: [0011][DUP][QoS][QoS][RETAIN] + Remaining Length
Variable Header:
  Topic Name: [length MSB][length LSB] + string
  Packet ID: [MSB][LSB] (仅当 QoS > 0)
Payload:
  Application Message (二进制数据)


Fixed Header:
DUP 重发标志位: 
1. 如果DUP标志被设置为0，表示这是客户端或服务端第一次请求发送这个PUBLISH报文。
2. 如果DUP标志被设置为1，表示这可能是一个早前报文请求的重发。


```
```md
packet: [48 33 0 19 114 101 109 111 116 101 47 114 101 97 100 121 47 110 111 116 105 102 121 99 108 105 101 110 116 32 114 101 97 100 121]

48: [0011][0000]
33: Remaining Length
0 19: Topic Name Length [remote/ready/notify](114 101 109 111 116 101 47 114 101 97 100 121 47 110 111 116 105 102 121)
Payload: [client ready](99 108 105 101 110 116 32 114 101 97 100 121)


```






- `PUBACK`
```md
Fixed Header: [0100][0000] + [00000010] (剩余长度=2)
Variable Header:
  Packet ID: [MSB][LSB]
```

- `SUBSCRIBE`
```md
Fixed Header: [1000][0010] + Remaining Length
Variable Header:
  Packet ID: [MSB][LSB]
Payload:
  Topic Filters: 
    [length MSB][length LSB] + Topic Filter
    [QoS Requested] (0/1/2)
    ... (多组订阅)



```
```md
packet: [130 21 0 1 0 16 97 112 112 47 114 101 97 100 121 47 110 111 116 105 102 121 1]
130: [1000][0010]
21 : Remaining Length
0 1: Packet ID
0 16: Payload Length [app/ready/notify](97 112 112 47 114 101 97 100 121 47 110 111 116 105 102 121)
1: QOS


```


## Packet ID
报文标识符范围`1～65535`
它在确保消息可靠传输中起着核心作用，特别是对于需要确认的服务质量（QoS）等级（QoS 1和QoS 2）。

- 关联请求与响应
*客户端*和*代理服务器（Broker）*使用**相同**的`Packet ID`来匹配请求和相应的确认包
1. 发送: 当客户端发送一个`QoS为1`的`PUBLISH`消息时，会附带一个`Packet ID`。
2. 响应: 代理服务器在接收到该消息后，会返回一个`PUBACK`消息，其中包含**相同**的`Packet ID`，以**确认**接收到该消息












