# ichat-go

基于GO和WebRTC的音视频聊天后端。

[ichat-app前端项目地址](https://github.com/laishere/ichat-app)

[在线体验地址](https://chat.laishere.cn/)

## 架构

- `gin`框架
- `gorm` + mysql + redis
- 中间件 + API + 业务逻辑 + 数据库访问的分层架构
- 无状态服务设计，服务间通信主要依靠消息队列

### 目录结构

| 目录         | 说明                    |
|------------|-----------------------|
| api        | API接口定义               |
| config     | 配置                    |
| db         | 数据库初始化                |
| di         | 简单的单例依赖注入实现           |
| errs       | 业务错误定义                |
| jwt        | 简单的JWT token签发、验证封装   |
| logging    | 日志封装                  |
| logic      | 业务逻辑层                 |
| middleware | 中间件(JWT鉴权、业务错误统一响应)   |
| model      | 数据结构定义、DAO实现          |
| sched      | 简单的redis锁、消息队列、延迟队列实现 |
| security   | 密码加密、API白名单           |
| sql        | 数据库表结构定义              |
| tests      | 一些单元测试                |
| utils      | 一些工具函数                |
| validate   | 参数校验配置和初始化            |
| ws         | websocket协议升级函数       |

### 业务逻辑层

| -                        | 说明                  |
|--------------------------|---------------------|
| call/                    | 通话核心逻辑              |
| call/manager.go          | 通话管理器实现             |
| call/manager_api.go      | 通话管理器API            |
| call/manager_delegate.go | 通话管理器的数据访问部分        |
| call/monitor.go          | 通话监视器(异常通话检测和清理)    |
| call/types.go            | 一些数据结构定义            |
| call/ws.go               | 用户通话信令websocket会话实现 |
| call/ws_api.go           | ws会话API             |
| notification/            | 实时通知会话逻辑            |
| notification/send.go     | 实时通知发送接口            |
| notification/session.go  | 会话抽象、查询、管理；会话API    |
| notification/types.go    | 一些数据结构定义            |
| notification/ws.go       | 实时通知的websocket会话实现  |
| 以下是API服务的业务逻辑            |                     |
| call.go                  | 通话业务逻辑              |
| chat.go                  | 聊天业务逻辑              |
| common.go                | 事务通用函数              |
| contact.go               | 联系人业务逻辑             |
| file.go                  | 文件上传下载              |
| group.go                 | 群组业务逻辑              |
| login.go                 | 登录业务逻辑              |
| register.go              | 注册业务逻辑              |
| user.go                  | 用户业务逻辑              |

## 核心逻辑

### 实时通知

- 基于会话id进行消息推送
- 一个会话id对应一个消息队列状态
- 一个用户可以有多个会话id（多客户端登录）
- 会话状态有时效，连接断开超过设定时间自动过期

**实时通知会话建立流程**

- 用户: 建立ws连接
- 用户: 发送token (1)
- 服务端: 验证token，不通过则发送错误并关闭连接
- 用户：发送sessionId (可以紧接(1))
- 服务端: 验证会话id对应的旧会话是否存在，存在则延用，否则新建，最终都返回将要使用的会话id
- 用户: 保存会话id，根据情况决定是否需要进行**增量同步**
- 后续：服务端发送实时通知，用户定时发送心跳（NGINX等服务器一般设定有空窗期超时断连，所以需要心跳）

**实时通知类型**

- 新消息
- 消息更新（通话状态更改、消息撤回）
- 新的联系人请求
- 新的联系人
- 通话已处理通知

**增量同步**

实时通知会话断开期间可能存在新的消息通知，而如果在会话存续期内无法恢复连接，
那么后端的会话状态被清除，用户根据后端返回的新会话id判断旧会话不可用， 因此需要同步断连期间的通知。

主要需要处理的是消息同步，采用增量同步的方式；

- 服务端在投递消息给用户时，会同时插入投递记录。
- 建立会话时，返回最后一条投递记录id。发送消息通知也会附带投递记录id。
- 用户根据最近已同步的投递记录id，通过API查询可能丢失的新消息投递记录并同步。

### 通话

后端负责的通话逻辑主要是通话管理和为WebRTC提供信令(Signaling)服务。

- 通话主要涉及两种服务：通话管理器和用户信令会话。
- 通话管理器工作在一个协程，负责管理通话状态、成员状态、成员间通信。
- 通话管理器通过消息队列API服务、用户信令会话通信。
- 用户信令会话和实时通知会话类似，但会话管理更简单。负责通话信令交流、通话状态通知、心跳。

## 开发

1. 运行docker/dev/compose.yml部署mysql和redis中间件环境
2. 注意执行init.sql语句
3. 启动后端

## docker打包

```sh
docker build . -t ichat-go:version
```

## 部署

参考前端项目
