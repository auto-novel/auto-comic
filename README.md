# Auto Comic 漫画服务

[![GPL-3.0](https://img.shields.io/github/license/auto-novel/auto-comic)](https://github.com/auto-novel/auto-comic#license)
[![cd-web](https://github.com/auto-novel/auto-comic/actions/workflows/cd-web.yml/badge.svg)](https://github.com/auto-novel/auto-comic/actions/workflows/cd-web.yml)
[![cd-api](https://github.com/auto-novel/auto-comic/actions/workflows/cd-api.yml/badge.svg)](https://github.com/auto-novel/auto-comic/actions/workflows/cd-api.yml)

漫画服务项目框架，包含用户端、管理端、API 和 PostgreSQL 数据库，暂不包含具体业务功能和示例数据。

## 贡献

请务必在编写代码前阅读 [贡献指南](CONTRIBUTING.md)，感谢所有为本项目做出贡献的人们！

## 部署

> [!WARNING]
> 注意：本项目并不是为了个人部署设计的，不保证所有功能可用和前向兼容。

```bash
# 1. 克隆仓库
git clone https://github.com/auto-novel/auto-comic.git
cd auto-comic

# 2. 生成环境变量配置
cat > .env << EOF
POSTGRES_PASSWORD=$(openssl rand -base64 48)
EOF

# 3. 启动服务
docker compose up -d --build
```

启动后，可通过以下地址访问：

- 用户端：http://localhost:7000/
- 管理端：http://localhost:7000/admin/
