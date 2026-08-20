# frostcell

冷链温区越限监测服务：持续接收探头温度，用滑动窗口均值/峰值判定是否越限，经滞回逻辑升级/解除告警，并向值班通道发出告警事件。

## 功能

- **滑动窗口统计**：按温区维护时间有序缓冲，统计 `count`、`overCount`、`maxTemp`、`meanTemp`
- **阈值与滞回**：Active 判定用 `setpoint+delta`；Clearing 判定用 `setpoint+delta-hysteresis`
- **告警状态机**：`Normal → Pending → Active → Clearing → Normal`
- **HMAC 接入**：`POST /v1/probes/sample` 需 `X-Frostcell-Signature` 头
- **告警外发**：`AlarmRaised` / `AlarmCleared` Webhook（可选）
- **运维页面**：`GET /` 嵌入 HTML 状态页

## 快速开始

```bash
cd frostcell
GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go run ./cmd/frostcell
```

默认监听 `:8080`，演示温区 `cell-01`（设定 -18°C，偏差 2°C，滞回 0.5°C）。

## API

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/v1/probes/sample` | 探头上报（HMAC + JSON） |
| GET | `/v1/cells/{id}/window` | 温区窗口与告警状态 |
| GET | `/v1/alarms` | 全部告警列表 |
| GET | `/v1/health` | 健康检查 |
| GET | `/` | 运维状态页 |

### 探头上报示例

```bash
BODY='{"cellID":"cell-01","probeID":"p1","tempC":-15.5,"ts":"2026-08-20T10:00:00Z"}'
SIG=$(printf '%s' "$BODY" | openssl dgst -sha256 -hmac "change-me-in-production" | awk '{print $2}')
curl -s -X POST http://localhost:8080/v1/probes/sample \
  -H "Content-Type: application/json" \
  -H "X-Frostcell-Signature: $SIG" \
  -d "$BODY"
```

## 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `FROSTCELL_LISTEN` | `:8080` | 监听地址 |
| `FROSTCELL_HMAC_SECRET` | `change-me-in-production` | HMAC 密钥 |
| `FROSTCELL_NOTIFY_URL` | （空） | 告警 Webhook URL |
| `FROSTCELL_WINDOW` | `5m` | 滑动窗口时长 |
| `FROSTCELL_EXCURSION_RATIO` | `0.40` | 越限样本占比阈值 |
| `FROSTCELL_CLEAR_RATIO` | `0.20` | 解除占比阈值 |
| `FROSTCELL_PENDING_WINDOWS` | `2` | 连续越限窗口数进入 Active |
| `FROSTCELL_CLEARING_WINDOWS` | `1` | Clearing 确认窗口数 |

## 业务规则

1. 窗口内超限样本占比 ≥ 40% 且窗口闭合 → `Pending`
2. 连续 2 个闭合窗口仍越限 → `Active`，发送 `AlarmRaised`
3. 温度回到滞回线以下且占比 < 20% → `Clearing`
4. 再 1 个确认窗口 → `Normal`，发送 `AlarmCleared`

## 模块结构

```
cmd/frostcell/          入口
internal/config/        配置
internal/model/         领域模型
internal/window/        滑动窗口
internal/threshold/     阈值与滞回
internal/alarmfsm/      告警状态机
internal/ingest/        HMAC 接入
internal/notify/        告警外发
internal/store/         内存快照
internal/app/           应用编排
internal/web/           嵌入运维 UI
```

## 许可证

MIT
