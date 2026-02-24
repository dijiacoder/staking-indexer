# Staking Indexer

一个用于跟踪以太坊兼容网络上质押事件的稳健区块链索引器。该服务扫描区块链上的质押相关交易，并在本地数据库中维护同步状态，支持链重组（reorgs）处理。

**项目说明**: 本项目是从实际生产环境中抽离出来的技术沉淀项目，基于真实业务场景的技术落地实践，旨在为个人技术积累和经验总结提供参考。这是一个"手搓"项目，展示了完整的区块链数据索引解决方案。

## 特性

- **实时索引**: 持续扫描区块链上的质押事件
- **重组处理**: 稳健处理区块链重组
- **多链支持**: 兼容以太坊兼容网络
- **事件跟踪**: 监控存款、提款和申领
- **数据库同步**: 维护一致的本地状态

## 架构

- `cmd/scanner`: 应用程序入口点
- `internal/config`: 配置管理
- `internal/repository`: 数据访问层
- `internal/service/scanner`: 核心业务逻辑（区块处理、事件解码、重组处理）
- `internal/gen`: 自动生成的 GORM 模型和查询

## 系统要求

- Go 1.25+
- MySQL 8.0+

## 快速开始

### 设置数据库

```bash
# 创建数据库架构
mysql -u root -p < sql/ddl.sql
```

### 配置

使用你的设置更新 `config/config.toml`：

```toml
[database]
# 数据库连接配置
# 格式: username:password@tcp(host:port)/database_name
dsn = "username:password@tcp(localhost:3306)/stake_db?charset=utf8mb4&parseTime=True&loc=Local"
debug = false

[ethereum]
# 以太坊网络配置
rpc_url = "https://your-rpc-endpoint.com"  # 替换为实际的 RPC 节点地址
chain_id = 1                               # 主网: 1, Sepolia测试网: 11155111
contract_address = "0xYourContractAddress" # 替换为实际的合约地址
confirmations = 12                         # 区块确认数

[scanner]
# 扫描器配置
batch_size = 10        # 每次扫描的区块数量
scan_interval = 1      # 扫描间隔(秒)
scan_timeout = 30      # 扫描超时时间(秒)

[prometheus]
# 监控配置
enabled = true         # 是否启用 Prometheus 监控
port = 9090           # 监控端口
```

### 构建与运行

```bash
# 安装依赖
go mod tidy

# 构建
go build -o staking-scanner ./cmd/scanner

# 运行
go run ./cmd/scanner --config=config/config.toml
```

## 数据库表

- `chain_scan_cursor`: 跟踪扫描进度并处理重组
- `chain_blocks`: 存储区块头用于重组检测
- `staking_pools`: 定义质押池
- `staking_user_positions`: 用户质押位置（聚合状态）
- `staking_events`: 来自区块链的原始质押事件