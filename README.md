# sekai-game-api

Project Sekai 游戏 API 网关，聚合 Haruki / Exmeaning 等公开上游，hedged 抢答返回结果。

## 接口

所有 `/api/*` 路由需要 `Authorization: Bearer {bearer_token}`。`/healthz` 不需鉴权。

错误统一为 `{"detail":{"msg":"..."}}`。

### `GET /healthz`
健康检查，返回 `ok`。

### `GET /api/status`
```json
{"status":"ok"}
```

### `GET /api/{region}/event/{event_id}/ranking`
合并上游 `ranking-top100` 和 `ranking-border` 两个端点。5 秒内存缓存。
```json
{
  "top100": { "rankings": [...], "userWorldBloomChapterRankings": [...] },
  "border": { "borderRankings": [...], "userWorldBloomChapterRankingBorders": [...] }
}
```
响应头：`X-Cache: hit|miss`、`X-Source: haruki|exmeaning`。

### `GET /api/{region}/user/{uid}/profile`
透传 query string（例如 `?use_cache=true`）。返回上游 profile JSON 原文。

### `POST /api/{region}/mysekai/photo`
请求体为一条 `userMysekaiPhotos` 元素（JSON 对象），从中提取 `uid`/`seq` 后拼出上游 image URL，返回 PNG bytes（`Content-Type: image/png`）。

候选字段名（按优先级，命中即用）：
- `uid`：`userId` / `user_id` / `uid`
- `seq`：`seq` / `sequenceId` / `sequence_id` / `photoIndex` / `userMysekaiPhotoId` / `mysekaiPhotoId` / `photoId` / `id`

如果未命中，响应 400，detail 会带回 `photo_keys` 实际值，照着调整 `photo.go` 的候选列表即可。

### 占位接口 (501)
以下路由已注册但尚未实现，命中返回 `501 Not Implemented`：

| 路径 | 方法 |
|---|---|
| `/api/{region}/user/{uid}/send_boost` | POST |
| `/api/{region}/create_account` | POST |
| `/api/{region}/user/{uid}/ad_result` | GET |
| `/api/{region}/ad_result/update_time` | GET |

## 配置

`config.yaml`：

```yaml
server:
  port: 11000
  auth_enabled: true           # false 则跳过 Authorization 校验
  bearer_token: "xxxxxxxx"     # 可被 env GAMEAPI_BEARER_TOKEN 覆盖

cache:
  ttl_seconds: 5

http:
  timeout_seconds: 10

upstreams:
  - name: haruki
    base: "https://public-api.haruki.seiunx.com/sekai-api/v5/api"
    image_base: "https://public-api.haruki.seiunx.com/sekai-api/v5"
    auth_header: "X-Haruki-Sekai-Token"
    token_env: "HARUKI_TOKEN"
  - name: exmeaning
    base: "https://seka-api.exmeaning.com/api"
    image_base: "https://seka-api.exmeaning.com"
    auth_header: "x-moe-sekai-token"
    token_env: "MOE_TOKEN"
```

上游 token 通过环境变量注入：`HARUKI_TOKEN` / `MOE_TOKEN`。

## 运行

本地：

```bash
export HARUKI_TOKEN=...
export MOE_TOKEN=...
go build -o sekai-game-api .
./sekai-game-api -config config.yaml
```

Docker：

```bash
docker build -t sekai-game-api .
docker run -d --name game-api \
  -p 11000:11000 \
  -e HARUKI_TOKEN=... \
  -e MOE_TOKEN=... \
  -e GAMEAPI_BEARER_TOKEN=... \
  sekai-game-api
```

挂载自定义 `config.yaml`：

```bash
docker run -d --name game-api \
  -p 11000:11000 \
  -e HARUKI_TOKEN=... -e MOE_TOKEN=... \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  sekai-game-api
```

## 行为

- **hedged 抢答**：所有上游并发请求，首个成功立即返回并取消其它。
- **ranking 缓存**：TTL 5 秒，避免高频采样压上游。
- **photo 字段探测**：第一次调用如果字段名不命中，detail 里会列出 photo 实际包含的 keys。
