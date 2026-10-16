# 影视文件整理

飞牛 fnOS 原生应用，按文件名识别并整理电影和剧集。手动搜索或自动校准时会向 TMDB 发送片名；应用不扫描未授权目录。首次启动后请在应用设置中授权源目录与目标目录，再在页面中配置路径。

## 识别和目录规则

- 电影：识别年份（如 `Arrival.2016.1080p.mkv`），整理到 `电影/Arrival (2016)/Arrival (2016).mkv`。没有年份时按清理后的文件名建目录。
- 电视剧：识别 `Show.Name.S01E02.mkv` 或 `Show.Name.1x02.mkv`，整理到 `电视剧/Show Name/Season 01/Show Name - S01E02.mkv`。
- 文件名识别组合 GuessIt 4.4.0 与本项目中文规则；GuessIt 用于补充发行名字段和有目录依据的非标准集数。季号、集号与现有规则冲突时保留现有结果，避免将年份误当成季号。
- 识别器支持常见的 `S01E02`、`1x02`、`Season 1 Episode 2`、`第1季第2集`、`第2集` 等写法、多集范围、季目录内编号集数、常见画质编码标签清理，并可从明确的季目录或电影目录补充季号、年份。非标准专名和发行方自定义格式仍可能需要人工校准。
- 剧名文件夹里只有 `01.mkv`、`02.mkv` 时，如果同目录数字文件至少组成一组连续集号，扫描会把它们识别为第一季；同目录有明确剧集文件名时，也会用剧集目录名补充纯数字文件的剧名。孤立的纯数字文件无法确认是电影还是剧集，会跳过等待人工处理；无季目录时季号按第一季处理。
- 扫描后可在每行修改一级媒体类型、二级分类、片名、电影年份或电视剧季/集号，再点“更新预览”核对完整目标路径。只有预览确认后点击“执行整理”才会移动文件；执行会采用当前修正值。
- 设置页可添加多个互不重叠的待整理文件夹，扫描时合并成一个预览并整理到同一目标片库；旧版单文件夹设置会自动兼容。日志记录操作时间、结果数量、来源/目标目录及跳过或失败原因，可按操作类型筛选。
- 电影二级分类按语言选择：华语电影、英语电影、日语电影、韩语电影、其他语言电影；电视剧二级分类按地区选择：国产剧、美剧、日剧、韩剧、港剧、台剧、英剧、其他地区剧。TMDB 校准或用户手动选择可以指定具体分类；缺少语言或地区资料时默认放入“待确认”，避免误分到其他语言或地区。应用不根据文件名猜测语言或地区。
- 目录示例：`电影/华语电影/片名 (年份)/文件.mkv`、`电视剧/美剧/剧名/Season 01/文件.mkv`。
- 可在目录设置中填写 TMDB v3 API Key 和 TMDB API 反向代理域名（例如 `https://tmdb.liubei.org`）。也可粘贴完整 TMDB API 链接，应用会提取代理域名并使用其 `/3/search/...` 接口。预览结果每行提供“TMDB 校准”，可查看候选名称、年份、原始语言、国家/地区和简介；选中后将 TMDB 名称/年份写入方案，并按原始语言或国家/地区建议分类，仍可手动更改。搜索时会将当前查询词发送给 TMDB；Key 保存在应用私有配置文件中且状态接口不返回 Key。
- 设置页可连接 TMDB 同步站（例如同一台 NAS 上的 `http://127.0.0.1:8788`）。自动校准依次搜索同步站 PostgreSQL 片库、导入的 SQLite 离线片库，再使用已配置的 TMDB API；同步站不可用时会明确显示连接错误。同步站 API 只在整理应用后端调用，不向浏览器暴露数据库连接信息。
- 也可在设置页导入自有 SQLite 离线片库（最大 2 GB）。数据库保存在 NAS 私有应用数据目录；SQLite 数据文件不会进入仓库或发布包。
- TMDB 校准请求按全局 0.8 QPS 限速，相邻请求至少间隔 1.25 秒，降低连续搜索触发限流的概率。文件名扫描本身在本地执行，不请求 TMDB。
- 扫描后可按片名启动自动校准：逐个唯一片名搜索 TMDB，仅采用标题高度匹配且年份不冲突的结果，剧集信息会同步到同剧各集。支持查看进度与停止；自动校准只更新预览，执行整理前仍需核对。
- TMDB 搜索结果显示海报预览；选中条目并更新预览后会保留海报路径。剧集按识别片名分组显示，校准组内任一集会把剧名、分类、TMDB 条目和海报同步到该组所有文件，季号、集号仍分别保留。
- 目前处理常见视频文件（MKV、MP4、AVI、MOV、WMV、M4V、TS、M2TS、WEBM、MPG、MPEG）。无法识别或目标重名的文件会跳过，不覆盖已有文件。
- 整理操作移动文件。页面先显示预览，点击“执行整理”并确认后才会手动移动。定时自动整理默认关闭，开启后按选定的间隔移动识别到的文件。

## 项目结构

- `manifest`：应用元数据及服务端口。
- `app/server/media_parser.py`：组合 GuessIt 与本项目规则，解析文件名和目录名中的影视信息。
- `app/server/guessit_adapter.py`：规范化 GuessIt 结果，并在有明确季集或目录上下文时补充识别。
- `app/server/requirements.txt`：固定服务端解析依赖版本；FPK 构建时将依赖装入独立目录。
- `app/server/THIRD_PARTY_NOTICES.md`：列出 GuessIt 及其依赖的许可证和来源。
- `app/server/organizer.py`：整理方案、安全移动逻辑及人工修正应用。
- `app/server/tmdb.py`：TMDB 搜索 API 客户端。
- `app/server/server.py`：本地 HTTP 服务、目录授权检查、设置保存和定时扫描。
- `frontend/src`：Vue 3 页面、组件、API 调用和样式。
- `frontend/vite.config.js`：Vite 开发代理与 fnOS 静态目录构建配置。
- `app/ui/www`：Vite 构建后由 Python 服务提供的静态页面资源。
- `app/ui/config`、`app/ui/images`：桌面入口及图标。
- `cmd/main`：服务启停和状态检查。
- `config/privilege`、`config/resource`：运行用户和应用共享目录声明。

## 构建与安装

### 在线安装与升级

本项目提供 FnDepot V2 应用源。首次可在 fnOS 的 FnDepot 或兼容的 New Store 中添加仓库地址 `https://github.com/CNLiuBei/MediaFileRenaming`，然后从应用源安装应用。之后每次发布新版本，GitHub Actions 会自动构建 FPK、创建 GitHub Release 并更新应用源索引；刷新应用源后即可通过商店升级，无需再次手动上传 FPK。GitHub Release 本身不会让 fnOS 自动检查更新。

发布新版时，先同步更新 `manifest` 中的版本号，再提交并推送对应标签（例如 `v1.0.19`）。工作流会验证标签与清单版本一致。

运行 `python3 tests/generate_filename_corpus.py` 可生成 1000 个零字节视频占位文件、预期识别清单和 ZIP，供扫描预览使用。当前整理流程在该合成语料的 1000 条预期结果上全部匹配；这不代表对真实片库的准确率。占位文件不含可播放内容；请只核对扫描结果，不要执行整理。

开发机需安装 Node.js 20.19+（或 22.12+）与 npm，并准备与开发设备架构匹配的 `fnpack`。应用通过 `manifest` 声明飞牛 `python312` 运行时依赖。先构建 Vue 前端，再从项目目录构建 FPK：

```sh
npm --prefix frontend ci
npm --prefix frontend run build
python3 scripts/build-fpk.py
```

构建后在项目目录生成 `.fpk` 文件。打包脚本仅把 fnOS 运行需要的 `app`、`cmd`、`config`、`wizard` 和 `manifest` 放入临时目录，前端源码与 `node_modules` 不会进入 FPK。将其安装到 fnOS 测试设备，在应用设置中授权目录，再配置源目录、目标目录并预览结果。服务端口为 `8787`。

## 安全与限制

页面可通过“选择文件夹”浏览 fnOS 应用设置中已授权的目录。服务只接受位于 `TRIM_DATA_ACCESSIBLE_PATHS` 或 `TRIM_DATA_SHARE_PATHS` 中的绝对目录；目录选择 API 只列出授权路径中的真实文件夹，并跳过符号链接。请在 fnOS 应用设置中授权需要整理的源和目标目录。目标目录不能与源目录相同或位于源目录内部。运行用户为应用专属用户。平台声明为 `all`；服务端使用 Python 3.12 和随包提供的纯 Python GuessIt 依赖。

TMDB 搜索按官方流程查询电影或电视剧候选条目，由用户选择具体结果；不自动采纳首个候选。参见 [TMDB 搜索说明](https://developer.themoviedb.org/docs/search-and-query-for-details) 与 [API 认证说明](https://developer.themoviedb.org/docs/authentication-application)。本产品使用 TMDB API，但未获 TMDB 认可或认证。未配置 TMDB Key 时仍可按文件名预览和手动修正。
