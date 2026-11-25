# ChargingHub API 契约（v1）

基础路径 `/api`。除登录外所有请求带 `Authorization: Bearer <token>`。
金额一律用**分**（整数），时间用 ISO 8601 字符串（本地时区 Asia/Shanghai）。

响应统一格式：

```json
{ "code": 0, "message": "ok", "data": {} }
```

失败时 `code != 0`，HTTP 状态为 400/401/403/404/409/500。

## 角色

| role | 说明 |
| --- | --- |
| `super` | 平台超管，管理租户 |
| `tenant_admin` | 租户管理员，全量业务权限 |
| `operator` | 租户操作员，除退款/SMTP 配置外同 tenant_admin |
| `user` | 充电用户，只能看自己的数据 |

## 认证

- `POST /api/login` body `{username, password, deviceId, tenantId?}` → `{token, role, tenantId, user:{id,phone,email,name}}`
  - 同 deviceId 重复登录复用原会话；不同 deviceId 登录踢掉该用户全部旧会话。
- `GET /api/me` → 同上（无 token 字段）
- `POST /api/logout`

## 超管

- `GET /api/tenants`，`PUT /api/tenants/{id}` body `{name,status}`
- `POST /api/tenants` body `{name,status,adminName,adminPhone,adminEmail,adminPassword}`：一步创建租户+管理员账号，租户ID自增，管理员自动关联该租户（role=tenant_admin），用 adminPhone/adminPassword 登录即自动归属该租户。`adminEmail` 用于接收通知。返回 `{id,name,adminPhone,adminUserId}`
- `POST /api/users` body `{phone,email,password,role,tenantId,name}`（超管/租户管理员创建用户；超管需传 tenantId，列表对超管显示全部租户用户并带 tenantName）
- 超管（role=super）拥有全部接口权限；各列表接口（场站/桩/枪/用户/钱包/流水/预约/订单/通知）对超管返回全部租户数据并带 tenantName/tenantId。超管写操作需显式传 tenantId（场站/钱包），桩按所属场站反推租户。
- `GET|PUT /api/platform/config`（仅超管）：全局对外域名，用于生成二维码。body `{baseUrl}`，如 `https://ff.dadaf.com:9010`。必须先配置才能下载二维码。

## 租户管理（tenant_admin / operator）

用户由租户创建；**钱包在首次充值时自动创建**，建用户时不建：

- `POST /api/users` body `{phone,email,password,name,confirmPhone?}` → `{id,phone,email,name}`
  - 若该手机号已属于其他租户，须传 `confirmPhone` 且等于完整手机号才能关联到本租户（防止误把他人加进来）。未传时返回 409 提示。
- `GET /api/users?keyword=` → 用户列表（含钱包余额）
- `PUT /api/users/{id}` body `{status,email,name}`

场站 / 桩 / 枪：

- `GET|POST /api/stations`；POST body：
  `{name,address,areaCode:"320106",areaName:"安徽省 / 合肥市 / 蜀山区",defaultStartTime:"22:00",startThresholdCents:5000,stopThresholdCents:2000,warnThresholdCents:3000,settlementTimeoutHours:24,priceCentsPerKwh:150}`
  - `areaCode` 为 6 位行政区划码（PC 端按地址级联选择自动带出），是桩编号前缀。
- `PUT /api/stations/{id}`（同上字段）
- `GET|POST /api/piles`；POST body `{stationId,gunCount}` → 服务端自动生成桩和枪，**不接受 deviceNo**
  - 桩编号 16 位 = 场站 `areaCode`(6) + 全局流水(10，补 0)；枪编号 18 位 = 桩编号 + 枪序号(2，每桩从 01 起)。
  - 每把枪生成唯一 `qr_token`。返回 `{id,deviceNo,secret,gunCount,gunCodes[],qrTokens[]}`。
- `PUT /api/piles/{id}` body `{status}`
- `GET /api/guns?stationId=` → 枪列表含 `gunCode`(18位)、`qrToken`
- `GET /api/guns/{id}/qrcode` → 该枪二维码 PNG（内容 = 全局域名 + `/m/#/charge?token=<qrToken>`，单个下载；需先配域名）
- `POST /api/piles/{id}/regenerate-secret`（重置设备密钥）

钱包（现金流水，`operator` 无退款/SMTP 权限）：

- `GET /api/wallets?userId=` → 钱包列表（含 balanceCents）
- `GET /api/ledgers?userId=` → 流水列表
- `POST /api/wallet/recharge` body `{userId,amountCents,method,voucherNo,remark}`
- `POST /api/wallet/refund` body `{userId,amountCents,method,voucherNo,remark}`
  - 校验：余额足够；该用户在该租户下无 `charging`/`pending_settle` 订单；无 `pending`/`active` 预约单。已结算历史订单不拦截。
- `POST /api/wallet/{userId}/freeze` body `{frozen:true|false}`（可选）

预约与订单：

- `GET /api/reservations?userId=`
- `POST /api/reservations` body `{stationId,gunId}`
  - 启动时间 = 场站 `defaultStartTime`；当天已过则顺延次日；用户不可修改。
  - 创建时校验余额 ≥ `startThresholdCents`，枪不可被其他预约/充电占用。
- `POST /api/reservations/{id}/cancel`
- `GET /api/orders?userId=`
- `POST /api/orders/{id}/settle`（手动结算，`pending_settle` → `settled`）

通知配置：

- `GET|PUT /api/smtp` body `{host,port,username,password,from,encryption:"none|ssl|starttls",enabled}`
- `GET /api/notifications?userId=` → 发送记录

## 用户端

- `GET /api/stations` → 本租户场站（含枪状态）
- `GET /api/my/wallet`，`GET /api/my/ledgers`
- `GET /api/my/orders`，`GET /api/my/reservations`（与 `/api/orders`、`/api/reservations` 等价，自动按当前用户过滤）
- `GET /api/gun/by-token?token=` → 扫码后按枪令牌定位枪，返回枪编号/场站/状态/当前钱包余额（自动锁定本租户）
- `POST /api/reservations` body `{stationId,gunId}`（同管理员接口，自动绑定当前用户）→ `{id,startAt,status}`
- `POST /api/reservations/{id}/cancel`
- `POST /api/charge/start` body `{gunId}` → `{id,orderId,status}`，立即启动，校验余额 ≥ 场站启动门槛
- `POST /api/charge/stop` body `{gunId}` → 停止 charging 中的订单

## 业务规则（服务端强制）

1. **互踢**：同设备免踢；不同设备登录踢掉该用户所有旧会话。
2. **预约**：场站级固定时刻，用户不可改；当天过点顺延次日；预约单未关闭时禁止退款。
3. **钱包**：tenant_id + user_id 唯一；建用户时不建钱包，首次充值时按当前租户自动创建；所有变动写 append-only 流水；余额以流水为准。
4. **启动**：余额 ≥ 场站 `startThresholdCents`（默认 50 元）。
5. **跨租户同手机号**：用户主档按手机号全局唯一，可在多个租户各有一份关联和独立钱包；其他租户首次关联该手机号时须完整输入手机号二次确认。
6. **编号**：桩号 = 6 位区划码 + 10 位流水 = 16 位；枪号 = 桩号 + 2 位枪序号 = 18 位；设备通信与展示统一用 18 位枪号。
7. **二维码**：内容 = 全局域名 + `/m/#/charge?token=<随机令牌>`；扫码定位按令牌查库；域名由超管前置配置，改域名后旧码失效。
5. **断电**：充电中余额 < 场站 `stopThresholdCents`（默认 20 元）自动断电并生成待结算订单；< `warnThresholdCents`（默认 30 元）发提醒。
6. **退款**：仅退当前余额；存在 `charging` / `pending_settle` 订单或 `pending` / `active` 预约时拒绝；已结算订单不拦截。
7. **结算**：设备上报结束后超过 `settlementTimeoutHours`（默认 24h）自动按最后上报数据结算。
8. **通知**：SMTP 租户级配置，异步发送，事件包括预约提醒、启动成功/失败、充电完成、余额预警、充值/退款到账。

## 设备侧（非 HTTP）

TCP 长连接，平台为服务端。帧头 `0x68`，CRC16（低字节在前）。一期实现：

- 设备登录（deviceNo + secret）、心跳、实时数据上报、远程启停指令、交易记录上报、参数下发。
- 一期不实现：离线卡同步、并充、OTA 远程升级。
