# 第三方组件

本应用在服务端使用以下开源 Python 软件包解析影视文件名。完整许可证文本随发布包放在 `app/server/vendor/<package>.dist-info/` 下；部分包将文件放在 `licenses/` 子目录。

| 软件包 | 版本 | 许可证 | 项目地址 |
| --- | --- | --- | --- |
| GuessIt | 4.4.0 | LGPL-3.0-or-later | https://github.com/guessit-io/guessit |
| Rebulk | 6.0.1 | MIT | https://github.com/Toilal/rebulk |
| Babelfish | 0.6.1 | BSD-3-Clause | https://github.com/Diaoul/babelfish |
| python-dateutil | 2.9.0.post0 | BSD-3-Clause / Apache-2.0 | https://github.com/dateutil/dateutil |
| six | 1.17.0 | MIT | https://github.com/benjaminp/six |

GuessIt 及其依赖以独立 Python 软件包形式随 FPK 分发，没有修改其上游源码。应用自身仍按仓库中的许可证和源码发布；GuessIt 的 LGPL 许可证文本随包提供。构建时从 PyPI 下载上表所列固定版本。
