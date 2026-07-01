# xiaozhi-esp32-server-golang 源码部署与运行教程

本文档完整记录从 Git 克隆到服务正常运行的全过程，包含各平台依赖安装、常见坑点及解决方案。

> 适用读者：有基础 Go 经验的开发者。如为小白，建议先阅读 [知识点拓展](#8-知识点拓展) 了解基础概念。

---

## 目录

1. [环境准备](#1-环境准备)
2. [获取源码](#2-获取源码)
3. [安装系统依赖（macOS）](#3-安装系统依赖macOS)
4. [安装 OnnxRuntime](#4-安装-onnxruntime)
5. [编译与运行](#5-编译与运行)
6. [配置详解](#6-配置详解)
7. [常见问题](#7-常见问题)

---

## 1. 环境准备

### 1.1 硬件与系统要求

- macOS 11+ / Ubuntu 20.04+ / Windows 10+
- 4GB+ 内存（推荐 8GB）
- 10GB+ 可用磁盘空间

### 1.2 必需软件

| 软件 | 版本要求 | 检查命令 |
|------|----------|----------|
| Go | 1.24+ | `go version` |
| Git | 任意 | `git --version` |
| Homebrew（macOS） | 最新 | `brew --version` |

### 1.3 可选但推荐

| 软件 | 用途 |
|------|------|
| Redis | 设备配置存储 & 聊天历史 |
| Node.js 20+ | 构建管理后台前端 |
| Docker | 基础设施（Redis、FunASR等） |

---

## 2. 获取源码

### 2.1 克隆仓库

```bash
git clone https://github.com/hackers365/xiaozhi-esp32-server-golang.git
cd xiaozhi-esp32-server-golang
```

> 注意：`asr_server/` 是 git submodule，如需要该功能，使用 `git clone --recursive`。

### 2.2 下载 Go 依赖

如果使用代理：

```bash
HTTP_PROXY=http://127.0.0.1:10808 \
HTTPS_PROXY=http://127.0.0.1:10808 \
go mod tidy
```

如不使用代理，直接执行：

```bash
go mod tidy
```

---

## 3. 安装系统依赖（macOS）

本项目大量使用 CGo，需要系统级 C 依赖库才能编译。

### 3.1 安装 Opus 音频库

```bash
brew install opus opusfile pkg-config
```

### 3.2 验证安装

```bash
pkg-config --cflags --libs opusfile
```

预期输出包含 `-lopusfile`（**不是**仅 `-lopus`）：

```
-I/usr/local/Cellar/opusfile/0.12_1/include/opus -L/usr/local/Cellar/opusfile/0.12_1/lib -lopusfile
```

> ⚠️ **坑点说明**：macOS 下 `brew install opusfile` 后，`/usr/local/lib/pkgconfig/opusfile.pc` 可能是一个**假的 stub 文件**，内容仅有 `-lopus` 而缺少 `-lopusfile`，导致链接失败。修复方法：
>
> ```bash
> sudo cp /usr/local/Cellar/opusfile/*/lib/pkgconfig/opusfile.pc \
>         /usr/local/lib/pkgconfig/opusfile.pc
> ```
>
> 验证修复后输出应包含 `-lopusfile`。

### 3.3 复制运行时库到系统目录

```bash
# libopusfile
sudo cp /usr/local/Cellar/opusfile/*/lib/libopusfile*.dylib /usr/local/lib/

# opusfile.h 头文件（某些版本需要）
sudo ln -sf /usr/local/Cellar/opusfile/*/include/opus/opusfile.h \
            /usr/local/include/opus/opusfile.h
```

---

## 4. 安装 OnnxRuntime

Silero VAD 和某些 ASR 模块需要 OnnxRuntime 1.21.0。

### 方式一：Homebrew 安装（推荐但耗时）

需要编译 21 个依赖，耗时较长（10-30 分钟）：

```bash
yes | brew install onnxruntime
```

### 方式二：从 GitHub 下载预编译包（更快速）

根据系统架构选择合适的包：

```bash
# 查看本机架构
uname -m   # x86_64 或 arm64
```

**Intel Mac (x86_64)：**

```bash
export ALL_PROXY=http://127.0.0.1:10808   # 如需要代理

cd /tmp
curl -L -o onnxruntime-osx-x86_64-1.21.0.tgz \
  "https://github.com/microsoft/onnxruntime/releases/download/v1.21.0/onnxruntime-osx-x86_64-1.21.0.tgz"
tar -xzf onnxruntime-osx-x86_64-1.21.0.tgz

# 复制头文件
sudo cp onnxruntime-osx-x86_64-1.21.0/include/onnxruntime_c_api.h \
        /usr/local/include/
sudo cp onnxruntime-osx-x86_64-1.21.0/include/onnxruntime_cxx_api.h \
        /usr/local/include/
sudo cp onnxruntime-osx-x86_64-1.21.0/include/onnxruntime_cxx_inline.h \
        /usr/local/include/

# 复制动态库
sudo cp onnxruntime-osx-x86_64-1.21.0/lib/libonnxruntime*.dylib \
        /usr/local/lib/
```

**Apple Silicon (arm64)：**

URL 改为 `https://github.com/microsoft/onnxruntime/releases/download/v1.21.0/onnxruntime-osx-arm64-1.21.0.tgz`

**Linux (x64)：**

```bash
cd /tmp
wget https://github.com/microsoft/onnxruntime/releases/download/v1.21.0/onnxruntime-linux-x64-1.21.0.tgz
tar -xzf onnxruntime-linux-x64-1.21.0.tgz
sudo cp -r onnxruntime-linux-x64-1.21.0/include/* /usr/local/include/
sudo cp -r onnxruntime-linux-x64-1.21.0/lib/* /usr/local/lib/
sudo ldconfig
```

> ⚠️ **坑点说明**：下载地址的文件名是 `onnxruntime-osx-x86_64-1.21.0.tgz`，**不是** `onnxruntime-osx-x64-1.21.0.tgz`。后者返回 404。

### 4.1 验证安装

```bash
ls /usr/local/include/onnxruntime_c_api.h   # 头文件
ls /usr/local/lib/libonnxruntime.dylib      # 动态库（Linux 为 .so）
```

---

## 5. 编译与运行

### 5.1 基本编译（无管理后台）

```bash
cd /Users/xinbao/space/xiaozhi-esp32-server-golang

CGO_ENABLED=1 go build -o xiaozhi-server ./cmd/server

./xiaozhi-server -c config/config.yaml
```

### 5.2 带管理后台编译

#### 第一步：构建前端

```bash
cd manager/frontend
npm install
npm run build
cd ../..

# 将构建产物复制到 Go embed 目录
cp -r manager/frontend/dist manager/backend/static/dist
```

#### 第二步：编译并运行

```bash
CGO_ENABLED=1 go run -tags "manager embed_ui" ./cmd/server \
  -c config/config.yaml \
  --manager-enable \
  --manager-config build/common/manager.json
```

> **参数说明：**
>
> | 参数 | 说明 | 默认值 |
> |------|------|--------|
> | `-c` | 主配置文件路径 | `config/config.yaml` |
> | `--manager-enable` | 启用内嵌管理后台 | 关闭 |
> | `--manager-config` | 管理后台配置文件 | 需要手动指定 |
> | `--asr-enable` | 启用内嵌 ASR 服务 | 关闭 |

#### 第三步：访问管理后台

浏览器打开 **http://localhost:8080**

首次访问进入初始化页面，创建管理员账号后登录。

---

## 6. 配置详解

### 6.1 配置文件结构

主配置文件 `config/config.yaml` 包含所有服务模块的配置。

### 6.2 关键配置项

```yaml
# 配置提供者（决定设备配置来源）
config_provider:
  type: "memory"    # memory / manager / redis

# 管理后台地址（manager 模式时使用）
manager:
  backend_url: "http://127.0.0.1:8080"
```

#### 三种配置提供者的区别

| 类型 | 说明 | 适用场景 |
|------|------|----------|
| `memory` | 从本地 config.yaml 读取，无需外部依赖 | 本地开发测试 |
| `manager` | 从管理后台 API 获取设备配置 | 正式部署 |
| `redis` | 从 Redis 读取 | 已使用 Redis 的场景 |

### 6.3 默认配置提供者（memory）的配置项

当使用 `memory` 模式时，设备将使用 `config.yaml` 中的以下配置：

```yaml
# ASR（任选其一配置）
asr:
  provider: "funasr"
  funasr:
    host: "127.0.0.1"
    port: "10096"
    mode: "offline"

# TTS（任选其一配置，需要填入 API Key）
tts:
  provider: "doubao_ws"
  doubao_ws:
    appid: "你的appid"
    access_token: "你的access_token"
  # 或使用 Edge TTS（无需 API Key，但中文效果一般）
  # provider: "edge"
  # edge:
  #   voice: "zh-CN-XiaoxiaoNeural"

# LLM（任选其一配置，需要填入 API Key）
llm:
  provider: "qwen_72b"
  qwen_72b:
    api_key: "你的api_key"
    base_url: "https://api.siliconflow.cn/v1"
    model_name: "Qwen/Qwen2.5-72B-Instruct"

# VAD（推荐 ten_vad，无需外部依赖）
vad:
  provider: "ten_vad"

# MCP（如未启动 MCP 服务，先禁用）
mcp:
  global:
    enabled: false
```

### 6.4 激活设备

在管理后台完成配置后，设备通过 WebSocket 连接：

```
ws://服务器IP:8989/xiaozhi/v1/
```

设备需要发送合法 MAC 地址作为 Device ID，如 `34:cd:b0:a6:fb:8c`。

---

## 7. 常见问题

### 7.1 编译错误：onnxruntime_c_api.h not found

**原因：** 未安装 OnnxRuntime 头文件（见第 4 章）。

### 7.2 编译错误：opusfile.h not found

**原因：** opusfile 头文件未在系统头文件目录。

**解决：**
```bash
sudo ln -sf /usr/local/Cellar/opusfile/*/include/opus/opusfile.h \
            /usr/local/include/opus/opusfile.h
```

### 7.3 链接错误：undefined symbol "_op_free" / "_op_read"

**原因：** `pkg-config` 返回的链接标志缺少 `-lopusfile`。

**解决：** 修复 `opusfile.pc`（见 3.2 节）。

### 7.4 运行错误：不支持的ASR引擎类型: 

**原因：** 设备配置中的 ASR Provider 为空字符串。

**解决一（memory 模式）：** 确保 `config.yaml` 中 `asr.provider` 已正确填写。

**解决二（manager 模式）：** 在管理后台 → 配置管理 → ASR 配置 → 添加 ASR 配置，并关联到智能体。

### 7.5 MCP 连接被拒绝

```
failed to connect to SSE stream: Get "http://localhost:3001/sse": connection refused
```

**原因：** 配置了全局 MCP 服务但未启动。

**解决：** 在 `config.yaml` 中设置 `mcp.global.enabled: false`，或启动对应的 MCP 服务器。

### 7.6 Redis 连接被拒绝

```
init redis error: failed to connect to redis: dial tcp 127.0.0.1:6379: connect: connection refused
```

**原因：** 配置了 Redis 但服务未运行。

**解决：**
```bash
brew install redis
brew services start redis
```

或修改 `config.yaml` 中 Redis 配置为空（取决于配置提供者是否需要 Redis）。

### 7.7 MQTT 走了代理导致连接失败

```yaml
socks connect tcp 127.0.0.1:10808->127.0.0.1:2883: connection refused
```

**原因：** 终端设置了 `ALL_PROXY` 环境变量，MQTT 连接被路由到了代理。

**解决：** 运行前清除代理环境变量：

```bash
env -u ALL_PROXY -u HTTP_PROXY -u HTTPS_PROXY CGO_ENABLED=1 go run ./cmd/server ...
```

### 7.8 可执行文件找不到 main 函数

```
undefined: defaultConfigFilePath
undefined: StartManagerHTTP
```

**原因：** 使用 `go run cmd/server/main.go`（文件模式）而非 `go run ./cmd/server`（包模式）。

**解决：** 改为包模式运行：

```bash
go run ./cmd/server -c config/config.yaml
```

### 7.9 macOS 安全提示

首次运行二进制文件时，macOS 可能提示"已损坏"或安全拦截：

```bash
xattr -cr xiaozhi-server
```

---

## 附录 A：完整启动命令

### 基础启动（无管理后台）

```bash
cd /Users/xinbao/space/xiaozhi-esp32-server-golang

CGO_ENABLED=1 go run ./cmd/server -c config/config.yaml
```

### 完整启动（带管理后台 + 前端）

```bash
cd /Users/xinbao/space/xiaozhi-esp32-server-golang

# 确保前端已构建
cp -r manager/frontend/dist manager/backend/static/ 2>/dev/null

env -u ALL_PROXY -u HTTP_PROXY -u HTTPS_PROXY \
CGO_ENABLED=1 go run -tags "manager embed_ui" ./cmd/server \
  -c config/config.yaml \
  --manager-enable \
  --manager-config build/common/manager.json
```

### 启动后端口

| 端口 | 服务 | 说明 |
|------|------|------|
| 8080 | 管理后台 | HTTP Web 控制台 |
| 8989 | WebSocket | 设备/客户端连接入口 |
| 2883 | MQTT | 设备 MQTT 通信（内置 MQTT 服务启用时） |
| 8990 | UDP | 设备 UDP 通信 |

---

## 附录 B：依赖安装速查表（Ubuntu）

```bash
# Opus 编解码
sudo apt-get install -y pkg-config libopus0 libopusfile-dev

# ONNX Runtime 1.21.0
wget https://github.com/microsoft/onnxruntime/releases/download/v1.21.0/onnxruntime-linux-x64-1.21.0.tgz
tar -xzf onnxruntime-linux-x64-1.21.0.tgz
sudo cp -r onnxruntime-linux-x64-1.21.0/include/* /usr/local/include/
sudo cp -r onnxruntime-linux-x64-1.21.0/lib/* /usr/local/lib/
sudo ldconfig

# ten_vad 运行时依赖
sudo apt install -y libc++1 libc++abi1
```

---

## 附录 C：依赖安装速查表（Windows）

> 参考 [doc/compile_deploy.md](compile_deploy.md) 完整教程。

环境变量设置：

```powershell
$env:CGO_ENABLED = "1"
$env:PATH = "C:\msys64\mingw64\bin;$env:PATH"
$env:C_INCLUDE_PATH = "E:\onnxruntime-win-x64-1.21.0\include"
$env:LIBRARY_PATH = "E:\onnxruntime-win-x64-1.21.0\lib"
```

---

## 8. 知识点拓展

### 8.1 CGo 是什么？为什么需要它？

Go 本身是一门自带 GC 的编译型语言，能直接调用 C 语言代码。这种能力称为 **CGo**。

本项目中使用 CGo 的原因：

```
Go 标准库 → 不支持 Opus 音频编解码
            不支持 OnnxRuntime 推理框架
            调用系统原生库性能最优

CGo 方案 → 直接链接 opus/libopusfile 共享库
           直接调用 onnxruntime C API
           调用 ten_vad 原生框架库
```

CGo 的工作方式并非"Go 执行 C 代码"，而是：
1. Go 编译器在链接阶段调用系统 C 编译器（clang/gcc）
2. C 代码被编译为 `.o` 目标文件
3. 链接器（ld）将 Go 目标文件和 C 目标文件链接为单一可执行文件
4. 运行时，Go 通过 cgo 机制调用 C 共享库中的函数

这就是为什么需要 `CGO_ENABLED=1` 环境变量——Go 默认**不启用** CGo。

### 8.2 pkg-config 的作用

`pkg-config` 是一个元数据查询工具，解决 C/C++ 编译时的依赖问题：

```bash
# 查看 opusfile 的编译标志
$ pkg-config --cflags --libs opusfile
-I/usr/local/Cellar/opusfile/0.12_1/include/opus
-L/usr/local/Cellar/opusfile/0.12_1/lib
-lopusfile -lopus
```

Go 的 CGo 通过 `#cgo pkg-config: opusfile` 指令自动调用 `pkg-config` 获取链接参数。如果 `.pc` 文件缺失或内容错误，就会导致链接时找不到符号。

`.pc` 文件示例（`/usr/local/lib/pkgconfig/opusfile.pc`）：

```
prefix=/usr/local/Cellar/opusfile/0.12_1
exec_prefix=${prefix}
libdir=${exec_prefix}/lib
includedir=${prefix}/include

Name: opusfile
Description: High-level Opus decoding library
Version: 0.12
Requires: opus
Libs: -L${libdir} -lopusfile
Cflags: -I${includedir}
```

> **核心要点**：`Libs:` 字段决定了链接器的 `-l` 参数。缺失 `-lopusfile` 就会导致 `undefined symbol: _op_free` 等链接错误。

### 8.3 Go Build Tags（编译标签）

Go 使用**编译标签**（Build Tags）实现条件编译。本项目中：

```go
//go:build manager && embed_ui
```

| 标签 | 作用 |
|------|------|
| `manager` | 启用管理后台 Go 后端代码 |
| `embed_ui` | 将 Vue 前端构建产物嵌入二进制文件 |
| `asr_enabled` | 启用 ASR 相关的额外代码 |

使用 `-tags` 参数指定：

```bash
go run -tags "manager embed_ui" ./cmd/server
```

不加 `manager` 标签时，管理后台相关代码完全不被编译。嵌入 Go 二进制中的是一个完整的 Web 应用（后端 API + Vue SPA）。

### 8.4 `//go:embed` 静态资源嵌入

Go 1.16 引入的 `//go:embed` 指令，允许将静态文件直接编译进二进制：

```go
//go:embed dist/*
var staticFS embed.FS
```

这意味着：
- 部署时只需要一个二进制文件，无需附带 HTML/CSS/JS 文件
- 前端 `dist/` 目录必须在编译前存在，否则编译报错 `pattern dist/*: no matching files found`
- 热更新时需重新编译整个二进制

### 8.5 VAD、ASR、LLM、TTS 全链路详解

本项目实现了一条完整的 AI 语音交互流水线：

```mermaid
graph LR
    A[麦克风] --> B[VAD]
    B --> C[ASR]
    C --> D[LLM]
    D --> E[TTS]
    E --> F[喇叭]
```

**各模块职责：**

| 模块 | 中文 | 任务 | 典型技术 |
|------|------|------|----------|
| VAD | 语音活动检测 | 判断人是否在说话，拆分句子 | Silero VAD / WebRTC VAD / ten_vad |
| ASR | 自动语音识别 | 将语音转为文本 | FunASR / Doubao ASR / 讯飞 |
| LLM | 大语言模型 | 理解对话并生成回复 | Qwen / DeepSeek / 豆包 |
| TTS | 文本转语音 | 将文字转为语音 | Doubao TTS / Edge TTS / CosyVoice |

**全流式架构**：与"录音→发送→等待→播放"的传统方式不同，本项目采用**流式处理**——ASR 边识别边输出中间结果，LLM 边生成边输出，TTS 边合成边播放。这种架构将端到端延迟从秒级降至毫秒级。

### 8.6 资源池（Resource Pool）设计

项目中大量使用**资源池**模式管理外部连接：

```go
// 注册资源类型
manager.Register("vad", constructor)
manager.Register("asr", constructor)
manager.Register("llm", constructor)
manager.Register("tts", constructor)
```

资源池的核心配置项：

| 配置项 | 含义 | 默认值 |
|--------|------|--------|
| `max_size` | 最大资源数量 | 1000 |
| `min_size` | 最小资源数量（预创建） | 1 |
| `acquire_timeout` | 获取资源超时时间 | 5s |
| `idle_timeout` | 空闲资源回收时间 | 10m |

当每个设备连接时，会从池中获取一个 ASR 资源、一个 LLM 资源等。池化设计避免了每次对话都创建新连接的开销。

### 8.7 配置提供者（Config Provider）模式

本项目引入了一套**配置优先级**系统：

```
设备发送的 UDP hello 包
    ↓
管理后台设备配置（manager 模式）
    ↓
本地 config.yaml（memory 模式）
    ↓
Redis 配置（redis 模式）
```

`memory` 提供者最初是为测试设计的——它只返回空配置。本教程中将其增强为"回退到本地配置"，使其更适合开发环境使用。

### 8.8 Go run vs Go build：使用包路径而非文件路径

```bash
# ❌ 文件模式 - 只编译指定文件
go run cmd/server/main.go

# ✅ 包模式 - 编译整个包
go run ./cmd/server
```

**区别：**

| 模式 | 命令 | 编译范围 | 适用场景 |
|------|------|----------|----------|
| 文件模式 | `go run file.go` | 仅特定文件 | 单文件程序 |
| 包模式 | `go run ./pkg` | 整个包（含所有 `.go` 文件） | 多文件项目 |

`cmd/server/` 目录下包含多个文件（`main.go`、`config.go`、`defaults_config_standard.go` 等），使用文件模式会导致"undefined"编译错误。

### 8.9 macOS 动态库路径查找机制

macOS 使用 `dyld` 动态链接器查找共享库，查找顺序为：

1. `DYLD_LIBRARY_PATH` 环境变量
2. dyld 共享缓存（系统库）
3. `/usr/lib/`
4. `/usr/local/lib/`
5. Binary 中嵌入的 `@rpath`

Go 的 CGo 在链接时指定的 `-L` 和 `-rpath` 会影响运行时查找：

```bash
# 编译时写入二进制
-Wl,-rpath,/Users/xxx/lib/ten-vad/lib/macOS

# 运行时 dyld 会搜索该路径
```

如果运行时提示 `dyld: Library not loaded`，可以通过 `DYLD_LIBRARY_PATH` 临时补救：

```bash
DYLD_LIBRARY_PATH=/usr/local/lib go run ./cmd/server
```

### 8.10 管理后台的双模块架构

管理后台由两个独立的 Go 模块构成：

```
xiaozhi-esp32-server-golang/        ← go.mod（主模块）
  ├── cmd/server/                   ← 主程序入口
  └── manager/backend/              ← go.mod（管理后台模块）

manager/backend/  ──── HTTP API（Gin）
                     ├── controllers/   ← 路由处理器
                     ├── services/      ← 业务逻辑
                     ├── models/        ← 数据模型
                     └── static/        ← 前端 dist/（Go embed）
```

主模块通过 `replace` 指令引用管理后台模块，编译时通过 build tags 决定是否包含管理后台功能。

---

## 附录 D：Go 环境搭建（如未安装 Go）

```bash
# macOS
brew install go

# 验证安装
go version

# 设置 GOPROXY（国内加速）
go env -w GOPROXY=https://goproxy.cn,direct

# 设置环境变量（建议写入 ~/.zshrc）
echo 'export GOPATH=$HOME/go' >> ~/.zshrc
echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.zshrc
source ~/.zshrc
```

---

> 本文档覆盖了 macOS 平台的完整部署流程。Linux 和 Windows 的更多细节请参考官方文档：
> - [编译与部署指南](compile_deploy.md)
> - [Docker 部署](docker.md)
> - [配置详解](config.md)
