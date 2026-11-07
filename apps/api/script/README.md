# API 数据库脚本

以下命令在仓库根目录执行，需要 Bash、Docker 和 Go：

```bash
./apps/api/script/test_integration.sh -v
# 兼容入口，复用同一套容器生命周期逻辑：
./apps/api/tests/run.sh -v
./apps/api/script/build_jet.sh
```

脚本通过自身路径定位仓库，因此也可以从其他工作目录使用绝对路径调用。

## 临时数据库

集成测试和 Jet 生成脚本均启动独立的 `postgres:17.7-alpine3.23` 容器，使用
`tmpfs` 数据目录，执行根目录 `sql/init.sql`，最多进行 60 次健康状态检查
（每次间隔 1 秒）。脚本退出、失败或收到 INT / TERM 信号时，会强制清理临时容器。

数据库端口仅绑定 `127.0.0.1`，默认由 Docker 动态分配，避免与部署数据库或其他
工作区冲突。需要固定端口时，分别设置 `TEST_DB_PORT=7003` 或 `JET_DB_PORT=7004`；
设置为 `0` 也表示动态分配。脚本不访问持久化数据库，不读取应用配置。
Docker 命令本身遵循 Docker CLI / daemon 的超时设置。

## 集成测试

包装脚本显式传入 `TEST_DB_HOST`、`TEST_DB_PORT`、`TEST_DB_USER`、
`TEST_DB_PASSWORD`、`TEST_DB_NAME` 和 `TEST_DB_SSLMODE`。
测试使用 `integration` 构建标签，禁用 Go 测试结果缓存，默认超时为 60 秒；
额外参数会转发给 `go test`。

当前烟雾测试通过 `infra.OpenDB`、`SELECT 1` 和 `GET /ready` 验证真实连接，
不创建业务表、不插入数据，也不要求后续 schema 永远为空。

## Jet 代码生成

仅使用固定版本命令 `go run github.com/go-jet/jet/v2/cmd/jet@v2.13.0`，
不依赖全局安装的 `jet`。首次实际生成需要网络或模块缓存；工具版本不加入 API 的
`go.mod`。

当 `sql/init.sql` 尚未定义 public 业务表时，脚本明确跳过生成，不下载 Jet，
也不创建占位业务表。有实际业务表后，生成目标为 `apps/api/.gen`。
