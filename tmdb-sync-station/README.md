# TMDB 同步站（PostgreSQL 本地版）

Go API 和 Vue 3 页面读取 PostgreSQL。数据库已从本地 TMDB 离线快照导入：1,424,937 条影视、1,931,926 条别名、382,376 季、6,512,105 集。源 SQLite 保持原样。同步器按 TMDB 变更清单更新条目，并在 PostgreSQL 中更新剧集季集；海报只存 TMDB 路径，不下载图片文件。

## 本机启动

需要 Go 1.26+、Node.js 20.19+、npm 和 PostgreSQL 服务。

```sh
createdb tmdb_sync_station
./dev.sh
```

默认连接 `postgres://localhost:5432/tmdb_sync_station?sslmode=disable`，使用本机当前 PostgreSQL 用户。其他连接方式通过 `DATABASE_URL` 配置；连接字符串只在 Go 服务端使用。

页面地址：<http://127.0.0.1:5173/>，API 地址：`127.0.0.1:8788`。按 `Ctrl+C` 同时停止前后端。

重命名应用连接本同步站时，在该应用的设置页填写同步站地址（同一台 NAS 默认 `http://127.0.0.1:8788`）。校准接口 `GET /api/catalog/search?kind=电影&q=片名&year=年份` 只读 PostgreSQL 片库并返回 TMDB ID、片名别名、语言、地区和海报路径；重命名应用的后端调用该接口，不直接连接数据库。同步站必须在 NAS 上启动，开发机的 `127.0.0.1` 仅对开发机本身有效。

## TMDB 增量同步

Go 服务启动 5 秒后会自动同步一次，此后按 24 小时间隔运行；总览页也可以手动启动。默认通过 `https://tmdb.liubei.org/3` 访问 TMDB 代理，QPS 上限为 20；若代理返回 429，同步器会按 `Retry-After` 退避重试。实际请求速度还受接口响应时间限制。若 Worker 已经配置上游凭据，无需再提供 API 密钥。

```sh
TMDB_API_BASE_URL="https://tmdb.liubei.org" \
TMDB_SYNC_QPS="20" \
TMDB_SYNC_INTERVAL="24h" \
./dev.sh
```

设置 `TMDB_SYNC_AUTOSTART=false` 可关闭启动和每日自动任务，仍可从总览页手动同步。

直连 TMDB 时，可在启动前设置 `TMDB_API_BASE_URL=https://api.themoviedb.org`，并通过 `TMDB_READ_ACCESS_TOKEN` 或 `TMDB_API_KEY` 提供服务端凭据。密钥不会下发到浏览器。可以设置 `DATABASE_URL` 覆盖 PostgreSQL 地址。同步配置在设置页只读展示。

同步器使用电影和电视剧变更 ID 清单，再读取变更条目的详情；电视剧还会刷新分季及分集标题。任务会查询到当前 UTC 日期，但只把水位提交到昨天，确保当天后续变更会在下一次同步中再次检查。水位仅在整个任务成功后推进，失败会保留上次成功水位。TMDB 清单最多查询 14 天且每页返回 100 个 ID；首次水位来自离线快照更新时间，若该时间早于 14 天，首次同步只能覆盖最近 14 天，早于窗口的变化无法通过该接口补齐。同步清单页码和待处理影视 ID 会保存到 PostgreSQL；服务重启或某条请求失败后，下次启动会从最后已保存页和未完成条目继续。详情和剧集会逐个请求，待处理条目较多时会持续较久。

## 数据库

- `titles`：电影/电视剧主体与完整快照字段。
- `aliases`：多语言标题别名与检索键。
- `seasons`、`episodes`：季和分集信息。
- `catalog_meta`：来源、源库更新时间和本机导入时间。
- `sync_runs`、`sync_changes`：后续增量同步任务及变更记录。

结构迁移位于 `internal/catalog/migrations/`，服务启动时按版本自动应用。数据来自本机 `IFIM/catalog.sqlite`，不会打包进仓库；如需在其他机器复用，需要把 SQLite 文件复制过去，再运行同结构的目录导入器。

## 部署到 fnOS

`python3 scripts/build-fpk.py --fnpack <fnpack路径>` 会构建 Vue 页面、编译 Linux x86_64 Go 服务并生成 `dist/TmdbSyncStation.fpk`。PostgreSQL 单独作为 Docker 容器运行，数据卷应放在 `/vol1/docker/tmdb-sync-station/postgres`；数据库容器端口只绑定到 NAS 本机 `127.0.0.1:55432`，服务端口为 `8788`。

安装 FPK 后，在应用数据目录写入权限为 `0600` 的 `station.env`，内容至少包含 `DATABASE_URL=postgres://<用户>:<密码>@127.0.0.1:55432/tmdb_sync_station?sslmode=disable`、`TMDB_SYNC_QPS=20`、`TMDB_SYNC_AUTOSTART=false`。先恢复数据库，再启动应用；应用页面地址为 `http://<NAS地址>:8788`，重命名应用填写 `http://127.0.0.1:8788`。同步站 API 暂无用户级鉴权，只应在可信家庭网络中使用，不要将端口暴露到公网。
