# ChargingHub

个人用的多租户充电桩管理系统。后端 Go，PC 管理端 Vue 3 + TypeScript + Pinia + Element Plus，手机端 Vue 3 + TypeScript + Pinia + Vant（H5，手机浏览器打开）。

## 已实现的核心规则

1. **用户互踢**：同一设备重复登录复用会话；不同设备登录踢掉该用户全部旧会话。
2. **多租户**：用户身份全局共用（手机号唯一），每个租户各自创建关联用户与独立钱包；所有业务表带 `tenant_id`，后端强制过滤。
3. **预约**：用户可选；场站级固定启动时刻，用户不可改；当天过点自动顺延次日；有预约单时禁止退款。
4. **钱包**：线下充值/退款全部走 append-only 流水；退款只退当前余额；存在充电中/待结算订单或未关闭预约单时拒绝退款，已结算历史订单不拦截。
5. **余额控制**：启动门槛默认 50 元，低于 20 元自动断电，低于 30 元邮件预警；三个阈值均可在场站配置。
6. **结算**：设备上报充电结束后超过 24 小时（场站可配）自动按最后上报数据结算，也支持手动结算。
7. **SMTP 通知**：租户级配置，异步发送，记录每次发送状态。事件包括预约、启动、失败、完成、结算、余额预警、充值/退款到账。
8. **设备接入**：平台作为 TCP 服务端监听 8768，实现 0x68 帧、CRC16、登录、心跳、实时数据、远程启停、指令补发；一期不做离线卡、并充、OTA。

## 本地运行

```bash
# 后端
go mod tidy
go run ./cmd/server

# 前端开发（两个终端）
cd web/admin && npm install && npm run dev
cd web/mobile && npm install && npm run dev
```

后端默认监听 8080（HTTP API + PC 管理端）与 8768（设备网关）。PC 管理端 `http://localhost:8080/`，手机端 `http://<服务器>:8080/m/`。

首次启动会自动创建超级管理员：`admin / admin123`，登录后请先创建租户和租户管理员。

## 编译与部署

代码推送到 GitHub 后：

1. 仓库 → Actions → **Build ChargingHub** → **Run workflow**
2. 填写 `image_tag`（例如 `v1`）→ 点绿色按钮
3. 编译成功后镜像推送到 `ghcr.io/jshmcz/charginghub:v1` 和 `:latest`

在服务器/NAS 上：

```bash
mkdir -p charginghub && cd charginghub
# 修改 docker-compose.yml 中的 SECRET_KEY 和 image tag
docker compose up -d
docker compose logs -f
```

升级只需改 `docker-compose.yml` 的 `image` 标签，然后 `docker compose up -d`。数据库文件在 `./data/charginghub.db`，备份直接复制该文件（停止容器后复制最安全）。

> 私有仓库的 GHCR 镜像需要先 `echo <PAT> | docker login ghcr.io -u jshmcz --password-stdin`，PAT 需 `read:packages` 权限。

## 目录说明

- `cmd/server`：程序入口
- `internal/app`：业务 API、钱包、预约、订单、结算、SMTP、认证
- `internal/device`：TCP 设备网关与协议解析
- `web/admin`：PC 管理端
- `web/mobile`：手机 H5 端
- `docs/API.md`：HTTP 接口契约
- `docs/DEVICE_PROTOCOL.md`：设备协议说明

## 环境变量

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `PORT` | `8080` | HTTP 端口 |
| `DEVICE_PORT` | `8768` | 设备 TCP 端口 |
| `DB_PATH` | `data/charginghub.db` | SQLite 数据库路径 |
| `SECRET_KEY` | `dev-secret-change-me` | 用于加密 SMTP 密码，部署前务必修改 |
| `STATIC_DIR` | `web/admin/dist` | PC 前端静态目录 |
| `MOBILE_DIR` | `web/mobile/dist` | 手机 H5 静态目录 |
