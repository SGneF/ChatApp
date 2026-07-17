# LightChat 后端学习笔记

本文档整理当前 `backend` 目录里的后端模块、接口、数据模型和核心业务流程，适合复习 Go + Gin + GORM + WebSocket 即时通讯项目的实现思路。

## 1. 后端整体架构

当前后端是一个基于 Go + Gin 的 REST API + WebSocket 服务。

目录结构重点：

```text
backend/
├── cmd/server/main.go              # 服务启动入口
├── internal/
│   ├── user/                       # 用户模块
│   ├── friend/                     # 好友模块
│   ├── conversation/               # 会话模块
│   ├── message/                    # 消息模块
│   └── websocket/                  # WebSocket 实时通信模块
└── pkg/
    ├── db/                         # MySQL 初始化和 AutoMigrate
    ├── jwt/                        # JWT 生成和解析
    ├── middleware/                 # Gin 鉴权中间件
    ├── redis/                      # Redis 初始化
    └── response/                   # 统一响应结构
```

后端启动流程：

1. `cmd/server/main.go` 调用 `db.InitMySQL()` 初始化 MySQL。
2. `InitMySQL()` 内部执行 `AutoMigrate()`，自动创建/更新数据表。
3. 启动时尝试初始化 Redis：`pkg/redis.InitRedis()`。
4. 注册 Gin CORS 中间件。
5. 注册 `/api` 下的用户、好友、会话、消息 REST 接口。
6. 注册 `/ws` WebSocket 长连接入口。
7. 服务监听 `:8080`。

## 2. 公共基础模块

### 2.1 统一响应 `pkg/response`

所有 REST 接口返回统一结构：

```json
{
  "code": 1,
  "msg": "success",
  "data": {}
}
```

字段含义：

| 字段 | 含义 |
|---|---|
| `code` | `1` 表示成功，`0` 表示失败 |
| `msg` | 提示消息 |
| `data` | 业务数据，失败时通常为空 |

核心方法：

- `response.Success(c, data)`
- `response.Fail(c, httpStatus, msg)`

### 2.2 JWT `pkg/jwt`

JWT 用于登录态。

`Claims` 里保存：

- `user_id`
- `username`
- JWT 标准字段：过期时间、签发时间、生效时间

规则：

- 默认密钥：`light-chat-secret`
- 可通过环境变量 `LIGHTCHAT_JWT_SECRET` 覆盖
- Token 有效期：7 天

核心方法：

- `GenerateToken(userID, username)`：登录成功后生成 token
- `ParseToken(tokenString)`：接口鉴权和 WebSocket 鉴权时解析 token

### 2.3 鉴权中间件 `pkg/middleware`

`AuthMiddleware()` 会从请求中读取 token：

1. 优先读取请求头：`Authorization: Bearer <token>`
2. 也支持 query 参数：`?token=<token>`，主要给 WebSocket 使用

解析成功后，会把用户信息写入 Gin 上下文：

- `user_id`
- `username`

业务 handler 通过下面方法获取当前用户：

- `GetCurrentUserID(c)`
- `GetCurrentUsername(c)`

### 2.4 MySQL `pkg/db`

数据库 DSN：

- 环境变量：`LIGHTCHAT_MYSQL_DSN`
- 默认值：`root:123456@tcp(127.0.0.1:3307)/lightchat?charset=utf8mb4&parseTime=True&loc=Local`

连接池设置：

- 最大空闲连接：10
- 最大打开连接：100
- 连接最大生命周期：1 小时

自动迁移的表：

- `users`
- `friend_requests`
- `friends`
- `conversations`
- `messages`

### 2.5 Redis `pkg/redis`

Redis 当前主要用于 WebSocket 在线状态。

配置项：

| 环境变量 | 作用 | 默认值 |
|---|---|---|
| `LIGHTCHAT_REDIS_ADDR` | Redis 地址 | `127.0.0.1:6379` |
| `LIGHTCHAT_REDIS_PASSWORD` | Redis 密码 | 空 |
| `LIGHTCHAT_REDIS_DB` | Redis DB 下标 | `0` |

如果 Redis 初始化失败，服务不会退出，而是降级为本机内存连接管理。

## 3. 用户模块 `internal/user`

### 3.1 数据表 `users`

模型：`user.User`

| 字段 | 说明 |
|---|---|
| `id` | 用户 ID，主键 |
| `username` | 用户名，唯一 |
| `password` | bcrypt 加密后的密码 |
| `nickname` | 昵称 |
| `avatar` | 头像地址 |
| `signature` | 个性签名 |
| `create_time` | 创建时间 |
| `update_time` | 更新时间 |

### 3.2 接口

| 方法 | 路径 | 是否鉴权 | 功能 |
|---|---|---|---|
| `POST` | `/api/user/register` | 否 | 注册用户 |
| `POST` | `/api/user/login` | 否 | 登录并返回 token |
| `GET` | `/api/user/info` | 是 | 获取当前用户信息 |
| `POST` | `/api/user/profile` | 是 | 修改资料 |
| `GET` | `/api/user/search?keyword=` | 是 | 搜索用户 |

### 3.3 核心业务规则

注册：

1. 检查用户名是否已经存在。
2. 使用 bcrypt 对密码加密。
3. 如果没有传昵称，默认昵称等于用户名。
4. 写入 `users` 表。
5. 返回不包含密码的用户信息。

登录：

1. 根据用户名查询用户。
2. 使用 bcrypt 校验密码。
3. 校验通过后生成 JWT。
4. 返回 token 和用户信息。

搜索用户：

- 不返回自己。
- 支持按 `username`、`nickname` 模糊搜索。
- 如果关键字是数字，也支持匹配 `id`。
- 最多返回 20 条。

## 4. 好友模块 `internal/friend`

### 4.1 数据表

#### `friend_requests`

好友申请表。

| 字段 | 说明 |
|---|---|
| `id` | 申请 ID |
| `from_user_id` | 申请人 ID |
| `to_user_id` | 接收人 ID |
| `remark` | 申请备注 |
| `status` | `pending` / `accepted` / `rejected` |
| `create_time` | 创建时间 |
| `update_time` | 更新时间 |

#### `friends`

好友关系表。

| 字段 | 说明 |
|---|---|
| `id` | 主键 |
| `user_id` | 当前用户 ID |
| `friend_id` | 好友 ID |
| `remark` | 好友备注 |
| `create_time` | 创建时间 |

注意：好友关系是双向写入的。例如 A 和 B 成为好友，会写入两条记录：

```text
A -> B
B -> A
```

### 4.2 接口

| 方法 | 路径 | 功能 |
|---|---|---|
| `POST` | `/api/friend/apply` | 发送好友申请 |
| `GET` | `/api/friend/requests` | 获取收到的待处理好友申请 |
| `POST` | `/api/friend/accept` | 同意好友申请 |
| `POST` | `/api/friend/reject` | 拒绝好友申请 |
| `GET` | `/api/friend/list` | 获取好友列表 |
| `DELETE` | `/api/friend/:friend_id` | 删除好友 |

以上接口都需要登录。

### 4.3 核心业务规则

发送好友申请：

1. 不能添加自己。
2. 目标用户必须存在。
3. 如果已经是好友，拒绝申请。
4. 如果已经存在待处理申请，拒绝重复申请。
5. 创建 `friend_requests` 记录，状态为 `pending`。

同意好友申请：

1. 申请必须存在。
2. 只能由 `to_user_id` 本人处理。
3. 申请状态必须是 `pending`。
4. 将申请状态改成 `accepted`。
5. 在 `friends` 表中写入双向好友关系。
6. 使用事务保证状态更新和好友关系写入一致。

拒绝好友申请：

1. 申请必须存在。
2. 只能由接收方处理。
3. 状态必须是 `pending`。
4. 将状态改成 `rejected`。

删除好友：

- 删除双方好友关系。
- 如果两边都没有删除到记录，返回好友关系不存在。

## 5. 会话模块 `internal/conversation`

### 5.1 数据表 `conversations`

LightChat 的会话表设计是：**每个用户各自拥有一条会话记录**。

例如 A 和 B 单聊：

```text
A 的会话：user_id=A, target_id=B
B 的会话：user_id=B, target_id=A
```

字段：

| 字段 | 说明 |
|---|---|
| `id` | 会话 ID |
| `user_id` | 当前会话所属用户 |
| `target_id` | 单聊时为对方用户 ID；群聊时预留为群 ID |
| `type` | `single` / `group` |
| `last_message_id` | 最后一条消息 ID |
| `last_message` | 最后一条消息摘要 |
| `unread_count` | 当前用户未读数 |
| `is_top` | 是否置顶 |
| `create_time` | 创建时间 |
| `update_time` | 更新时间 |

唯一索引：

```text
user_id + target_id + type
```

这保证同一个用户对同一个目标只会有一条同类型会话。

### 5.2 接口

| 方法 | 路径 | 功能 |
|---|---|---|
| `POST` | `/api/conversation/single` | 创建或获取单聊会话 |
| `GET` | `/api/conversation/list` | 获取会话列表 |
| `GET` | `/api/conversation/:conversation_id` | 获取会话详情 |
| `DELETE` | `/api/conversation/:conversation_id` | 删除自己的会话 |
| `POST` | `/api/conversation/:conversation_id/read` | 清空该会话未读数 |
| `POST` | `/api/conversation/:conversation_id/top` | 设置/取消置顶 |

### 5.3 核心业务规则

创建或获取单聊会话：

1. 不能和自己创建会话。
2. 目标用户必须存在。
3. 必须已经是好友，才允许创建单聊会话。
4. 使用 `FirstOrCreate`，已存在则直接返回。

获取会话列表：

- 只查询当前用户自己的会话。
- 排序规则：置顶优先，然后按更新时间倒序。
- 单聊会话会组装 `target_user` 信息。

清空未读数：

- 只允许清空自己的会话。
- 将 `unread_count` 更新为 `0`。

删除会话：

- 只删除当前用户自己的会话记录。
- 不删除双方真实消息。

## 6. 消息模块 `internal/message`

### 6.1 数据表 `messages`

字段：

| 字段 | 说明 |
|---|---|
| `id` | 消息 ID |
| `conversation_id` | 发送方自己的会话 ID |
| `sender_id` | 发送者 ID |
| `receiver_id` | 接收者 ID |
| `type` | `text` / `image` / `file` / `voice` |
| `content` | 消息内容 |
| `status` | `sent` / `read` / `revoked` |
| `create_time` | 创建时间 |
| `update_time` | 更新时间 |

注意：`conversation_id` 是发送方自己的会话 ID，因此历史消息查询不能只按 `conversation_id` 查。实际查询逻辑是按双方用户 ID 查询：

```sql
(sender_id = 当前用户 AND receiver_id = 对方)
OR
(sender_id = 对方 AND receiver_id = 当前用户)
```

### 6.2 REST 接口

| 方法 | 路径 | 功能 |
|---|---|---|
| `POST` | `/api/message/send` | 发送消息 |
| `GET` | `/api/message/history?conversation_id=&page=&page_size=` | 获取历史消息 |
| `POST` | `/api/message/:message_id/revoke` | 撤回消息 |

实际桌面前端主要通过 WebSocket 发送消息、已读和撤回；REST 接口仍然保留。

### 6.3 消息发送流程

`message.Service.Send()` 的流程：

1. 根据 `conversation_id + current_user_id` 查询当前用户自己的会话。
2. 得到接收方 ID：`receiverID = currentConversation.TargetID`。
3. 创建一条 `messages` 记录，状态为 `sent`。
4. 更新发送方会话：
   - `last_message_id`
   - `last_message`
   - `update_time`
5. 确保接收方也存在一条反向会话：
   - `user_id = receiverID`
   - `target_id = currentUserID`
   - `type = single`
6. 更新接收方会话：
   - `last_message_id`
   - `last_message`
   - `unread_count + 1`
   - `update_time`
7. 整个过程在事务中完成。

### 6.4 历史消息流程

`History()` 的流程：

1. 校验当前用户拥有该会话。
2. 通过会话找到 `target_id`。
3. 查询双方之间的消息。
4. 先按时间倒序查询最近消息，再反转为正序返回。
5. 支持分页，`page_size` 最大 100。

### 6.5 撤回消息流程

`Revoke()` 的规则：

1. 消息必须存在。
2. 只能发送者本人撤回。
3. 已撤回的消息不能重复撤回。
4. 只能撤回 2 分钟内的消息。
5. 将消息状态改成 `revoked`，内容清空。
6. 如果这条消息正好是双方会话的最后一条消息，则会话摘要改成“撤回了一条消息”。
7. 返回 `message_id / sender_id / receiver_id / status`。

### 6.6 已读流程

`MarkConversationRead()` 的规则：

1. 当前用户必须拥有该会话。
2. 清空当前用户该会话的 `unread_count`。
3. 把“对方发给我”的 `sent` 消息改成 `read`。
4. 返回：
   - `conversation_id`
   - `reader_id`
   - `target_id`
   - `read_count`

## 7. WebSocket 模块 `internal/websocket`

### 7.1 WebSocket 入口

路径：

```text
GET /ws?token=<jwt>
```

连接流程：

1. 从 query 或 header 中读取 token。
2. 解析 JWT。
3. 将 HTTP 升级为 WebSocket。
4. 创建一个 `Client` 对象。
5. 注册到 `Hub`。
6. 写入 Redis 在线状态。
7. 发送 `connected` 事件。
8. 启动三个 goroutine：
   - `WritePump()`：统一写 WebSocket 消息和心跳 Ping。
   - `ReadPump()`：读取客户端发来的消息。
   - `sendOfflineSync()`：上线后同步离线消息。

### 7.2 Hub 连接管理

`Hub` 的核心结构：

```go
map[userID]map[*Client]bool
```

含义：

- 一个用户可能有多个连接，比如多个桌面端、移动端。
- `Hub.SendToUser(userID, msg)` 会把消息推送给该用户的所有本机连接。
- `Hub` 使用 `sync.RWMutex` 保护 map，避免并发读写 panic。

### 7.3 Client 读写协程

`ReadPump()`：

- 读取客户端消息。
- 按事件类型分发：
  - `chat_message`
  - `message_read`
  - `message_revoke`
- 连接断开时从 Hub 移除。
- 如果该用户在本机没有其他连接，则调用 Redis `SetOffline()`。

`WritePump()`：

- 从 `Client.Send` channel 中取消息并写入 WebSocket。
- 定时发送 Ping。
- 刷新 Redis 在线状态 TTL。

### 7.4 WebSocket 事件协议

服务端支持的事件：

| 事件 | 方向 | 说明 |
|---|---|---|
| `connected` | 服务端 -> 客户端 | WebSocket 连接成功 |
| `chat_message` | 双向 | 客户端发消息；服务端推送给接收者 |
| `chat_ack` | 服务端 -> 发送者 | 消息发送成功确认 |
| `chat_error` | 服务端 -> 客户端 | 消息操作失败 |
| `offline_sync` | 服务端 -> 客户端 | 上线后同步未读/离线消息 |
| `message_read` | 双向 | 客户端发已读；服务端通知对方 |
| `message_read_ack` | 服务端 -> 已读方 | 已读操作确认 |
| `message_revoke` | 双向 | 客户端发撤回；服务端通知接收者 |
| `message_revoke_ack` | 服务端 -> 撤回方 | 撤回操作确认 |

### 7.5 实时发送消息流程

客户端发送：

```json
{
  "type": "chat_message",
  "data": {
    "conversation_id": 1,
    "type": "text",
    "content": "你好"
  }
}
```

服务端流程：

1. `ReadPump()` 收到 `chat_message`。
2. 调用 `message.Service.Send()` 写入 MySQL 并更新双方会话。
3. 给发送者返回 `chat_ack`。
4. 通过 `Hub.SendToUser(receiverID, ...)` 给接收者推送 `chat_message`。

### 7.6 实时已读流程

客户端发送：

```json
{
  "type": "message_read",
  "data": {
    "conversation_id": 1
  }
}
```

服务端流程：

1. 调用 `message.Service.MarkConversationRead()`。
2. 给当前用户返回 `message_read_ack`。
3. 给对方推送 `message_read`，告诉对方消息已读。

### 7.7 实时撤回流程

客户端发送：

```json
{
  "type": "message_revoke",
  "data": {
    "message_id": 10
  }
}
```

服务端流程：

1. 调用 `message.Service.Revoke()`。
2. 给撤回方返回 `message_revoke_ack`。
3. 给接收方推送 `message_revoke`。

### 7.8 Redis 在线状态

在线状态 key：

```text
im:online:user:<userID>
```

特点：

- 用户连接成功时：`SetOnline()`。
- 心跳时：`RefreshOnline()` 刷新 TTL。
- 最后一个本机连接断开时：`SetOffline()` 删除 key。
- TTL 当前为 120 秒，用来防止服务崩溃后用户永远显示在线。

注意：当前 Redis 只保存在线状态，不负责消息持久化。

### 7.9 离线消息同步

当前离线同步不是 Redis 队列，而是基于 MySQL：

1. 用户离线期间，别人发来的消息已经写入 `messages` 表。
2. 接收方会话的 `unread_count` 会增加。
3. 用户重新连接 WebSocket 后，服务端执行 `sendOfflineSync()`。
4. 查询 `conversations` 中 `unread_count > 0` 的会话。
5. 从 `messages` 表中查询未读消息。
6. 推送 `offline_sync` 给前端。

限制：

- 默认最多同步 200 条离线消息。
- 当前没有 Redis Stream / PubSub。
- 如果未来做多后端实例，实时跨实例推送需要 Redis Pub/Sub 或消息队列。

## 8. 模块之间的调用关系

### 8.1 好友和会话的关系

创建单聊会话时，会话模块会查询好友表：

```text
conversation.CreateOrGetSingle()
  -> 检查 target 用户存在
  -> 查询 friend.Friend 是否存在
  -> FirstOrCreate conversation
```

所以：不是好友不能创建单聊会话。

### 8.2 会话和消息的关系

发送消息时，消息模块依赖会话模块的数据：

```text
message.Send()
  -> 根据 conversation_id + current_user_id 找发送方会话
  -> 得到 receiver_id
  -> 写 messages
  -> 更新发送方 conversation
  -> 创建/更新接收方 conversation
```

### 8.3 WebSocket 和消息模块的关系

WebSocket 不直接写 SQL，而是调用消息 service：

```text
websocket.Client.handleChatMessage()
  -> message.NewService(db).Send()
```

已读和撤回也是同样思路：

```text
handleMessageRead()   -> message.Service.MarkConversationRead()
handleMessageRevoke() -> message.Service.Revoke()
```

这种设计的好处是：REST 接口和 WebSocket 可以复用同一套业务逻辑。

## 9. 常见复习重点

### 9.1 为什么好友关系要写两条？

因为查询“我的好友列表”时只需要：

```sql
SELECT * FROM friends WHERE user_id = 当前用户
```

如果只写一条关系，每次查好友都要判断 `user_id` 或 `friend_id`，查询和权限判断会复杂。

### 9.2 为什么单聊会话双方各一条？

因为每个人对同一个聊天的状态不同：

- 未读数不同
- 是否置顶不同
- 是否删除会话不同

所以会话记录应该归属于具体用户。

### 9.3 为什么历史消息不能只按 `conversation_id` 查询？

因为 `messages.conversation_id` 保存的是发送方自己的会话 ID。

A 发给 B 的消息，`conversation_id` 是 A 的会话 ID；B 发给 A 的消息，`conversation_id` 是 B 的会话 ID。

所以查询双方历史时，必须按 `sender_id / receiver_id` 双向查询。

### 9.4 为什么 WebSocket 要有 `Send` channel？

同一个 WebSocket 连接不应该被多个 goroutine 同时写。

当前设计是：

- 业务代码只把消息放进 `Client.Send`。
- `WritePump()` 作为唯一写协程，负责真正写入 WebSocket。

这样可以避免并发写 WebSocket 的问题。

### 9.5 为什么 Redis 在线状态要设置 TTL？

如果服务崩溃，`SetOffline()` 可能来不及执行。

没有 TTL 的话，Redis 里会残留“在线”状态，导致用户永远显示在线。

TTL + 心跳刷新可以解决这个问题。

### 9.6 已读和撤回为什么也走 WebSocket？

因为它们属于实时状态同步：

- A 读了 B 的消息，B 应该马上看到“已读”。
- A 撤回消息，B 的聊天窗口应该马上变成撤回提示。

REST 接口可以完成数据修改，但 WebSocket 更适合通知对方实时更新 UI。

## 10. 当前能力清单

已实现：

- 用户注册、登录、资料修改、搜索用户
- JWT 鉴权
- 好友申请、同意、拒绝、删除、好友列表
- 单聊会话创建/获取
- 会话列表、详情、删除、置顶、清空未读
- 文本/图片/文件/语音类型的消息模型预留
- 消息发送、历史消息、撤回
- 消息状态：`sent`、`read`、`revoked`
- WebSocket 实时消息
- WebSocket 已读同步
- WebSocket 撤回同步
- Redis 在线状态
- 基于 MySQL unread_count 的离线消息同步

预留或后续可完善：

- 群聊业务目前只预留了 `group` 类型，还没有完整群组模块
- 图片、文件、语音消息的上传和存储还没有完整实现
- 多后端实例之间的实时推送还没有 Redis Pub/Sub / MQ
- 在线状态目前写入 Redis，但接口层还没有独立的“查询用户在线状态”API
- 离线同步当前基于 MySQL，不是 Redis Stream 或消息队列
- 撤回后如果需要展示“谁撤回了消息”，前端已能根据 sender_id 判断，后端目前只返回 sender_id/receiver_id/status

## 11. 推荐复习顺序

1. 先看 `cmd/server/main.go`，理解服务启动和路由注册。
2. 再看 `pkg/middleware/auth.go` 和 `pkg/jwt/jwt.go`，理解登录态如何流转。
3. 看 `internal/user`，理解最基础的注册登录闭环。
4. 看 `internal/friend`，理解好友申请和双向好友关系。
5. 看 `internal/conversation`，理解“每个用户一条会话记录”的设计。
6. 看 `internal/message`，理解消息发送如何同时更新双方会话。
7. 最后看 `internal/websocket`，理解实时消息、已读、撤回、在线状态和离线同步。

## 12. 一句话总结

LightChat 后端的核心设计是：

```text
用户通过 JWT 登录；好友关系决定能否创建单聊；
每个用户拥有自己的会话记录；消息按 sender/receiver 保存；
WebSocket 负责实时消息、已读和撤回；MySQL 负责最终持久化；Redis 负责在线状态。
```