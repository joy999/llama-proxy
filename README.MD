# Llama Proxy

Llama Proxy 是一个基于 GoFrame 框架实现的 OpenAI 协议兼容代理服务，用于将 OpenAI API 请求转发到本地运行的 llama.cpp 服务。

## 功能特性

- **OpenAI 协议兼容**：支持 `/v1/models`、`/v1/chat/completions`、`/v1/embeddings` 等标准接口
- **动态模型加载**：根据请求中的 `model` 参数自动加载对应的模型
- **模型切换**：支持在多个模型之间动态切换
- **进程监控**：自动检测 llama-server 进程状态，意外退出时自动重置
- **闲置超时**：支持配置闲置超时自动卸载模型，释放系统资源
- **请求队列**：使用单 worker 确保请求串行处理，避免并发问题

## 技术栈

- Go 1.21+
- GoFrame v2
- llama.cpp (llama-server)

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置模型

编辑 `manifest/config/config.yaml` 配置文件：

```yaml
openai:
  llama_server_path: "llama-server"
  idle_timeout: 300

  models:
    - id: "llama-2-7b-chat"
      name: "Llama 2 7B Chat"
      model_path: "/models/llama-2-7b-chat.gguf"
      ctx_size: 4096
      threads: 8
      gpu_layers: 35
```

### 3. 构建运行

```bash
go build -o llama-proxy.exe .
./llama-proxy.exe
```

## API 接口

### 获取模型列表

```bash
GET /v1/models
```

### 获取模型详情

```bash
GET /v1/models/{model_id}
```

### 聊天补全

```bash
POST /v1/chat/completions
Content-Type: application/json

{
  "model": "llama-2-7b-chat",
  "messages": [
    {"role": "user", "content": "Hello!"}
  ]
}
```

### 嵌入生成

```bash
POST /v1/embeddings
Content-Type: application/json

{
  "model": "bge-m3",
  "input": "Hello world"
}
```

### 其他接口

除上述接口外，其他 `/v1/*` 请求会被反向代理到当前加载的 llama-server 后端服务。这意味着你可以使用 llama-server 支持的任何其他接口。

## 配置说明

配置文件位于 `manifest/config/config.yaml`，详细配置请参考 `manifest/config/config.tpl.yaml`。

### 主配置项

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `llama_server_path` | string | `llama-server` | llama-server 可执行文件路径 |
| `backend_url` | string | `http://127.0.0.1:8123` | 备用后端代理地址 |
| `idle_timeout` | int | 300 | 闲置超时秒数，0 表示禁用自动卸载 |

### 默认模型参数 (default_params)

这些参数会被所有模型继承，可以在每个模型中覆盖：

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `ctx_size` | int | 4096 | 上下文窗口大小 |
| `threads` | int | 8 | 线程数 |
| `parallel` | int | 1 | 并行处理数 |
| `gpu_layers` | int | - | GPU 加速层数，不设置表示不传递参数，0 表示纯 CPU 计算，>0 表示使用 GPU 层数 |
| `cache_type_k` | string | q4_0 | K 缓存量化类型 |
| `cache_type_v` | string | q4_0 | V 缓存量化类型 |
| `flash_attn` | string | on | 闪注意力优化 |
| `no_mmap` | bool | false | 是否禁用 mmap |
| `mlock` | bool | false | 是否锁定内存到物理内存 |
| `temp` | float | 0.6 | 温度参数 |
| `top_p` | float | 0.95 | top-p 采样 |
| `top_k` | int | 20 | top-k 采样 |
| `min_p` | float | 0.0 | min-p 采样 |
| `presence_penalty` | float | 0.0 | 存在惩罚 |
| `repeat_penalty` | float | 1.0 | 重复惩罚 |

### 模型配置 (models)

| 参数 | 类型 | 说明 |
|------|------|------|
| `id` | string | 模型唯一标识符，用于 API 请求 |
| `name` | string | 模型名称 |
| `owned_by` | string | 模型拥有者 |
| `model_path` | string | 模型文件路径 |
| `vision_model_path` | string | 视觉模型路径（多模态模型） |
| 其他参数 | - | 参考 default_params，可以覆盖默认值 |

### 配置示例

```yaml
openai:
  llama_server_path: "/path/to/llama-server"
  idle_timeout: 300

  default_params:
    ctx_size: 4096
    threads: 8
    gpu_layers: 35

  models:
    - id: "llama-2-7b-chat"
      name: "Llama 2 7B Chat"
      model_path: "/models/llama-2-7b-chat.gguf"
      ctx_size: 4096
      threads: 8
      gpu_layers: 35
```

### 参数传递规则

参数传递遵循以下优先级和规则：

1. **优先级**：模型配置 > 默认配置 > 不传递（使用 llama-server 默认值）

2. **不传递参数的方法**：
   - 对于**数字类型**（如 `ctx_size`, `threads`, `gpu_layers`）：注释掉该配置项
   - 对于**字符串类型**（如 `cache_type_k`, `flash_attn`）：设置为空字符串 `""` 或注释掉
   - 对于**布尔类型**（如 `no_mmap`, `mlock`）：设置为 `false` 或注释掉

3. **特殊说明**：
   - `gpu_layers` 使用指针类型，可以正确区分"未设置"和"设置为 0"
   - 未设置时：不传递 `--n-gpu-layers` 参数（使用 llama-server 默认行为）
   - 设置为 0：传递 `--n-gpu-layers 0`（纯 CPU 计算）
   - 设置为 >0：传递对应的 GPU 层数

3. **示例**：
   ```yaml
   models:
     - id: "my-model"
       name: "My Model"
       model_path: "/models/my-model.gguf"
       ctx_size: 4096        # 传递 --ctx-size 4096
       threads: 8            # 传递 --threads 8
       gpu_layers: 0         # 传递 --n-gpu-layers 0（纯 CPU 计算）
       # gpu_layers: 35      # 传递 --n-gpu-layers 35（使用 35 层 GPU）
       # cache_type_k:       # 不传递 --cache-type-k（使用 llama-server 默认）
       flash_attn: ""        # 不传递 --flash-attn
   ```

4. **布尔类型**（`no_mmap`, `mlock`）：设置为 `true` 时传递对应参数，否则不传递

## 调试日志

启动模型时，会打印完整的 llama-server 启动命令参数，方便调试和确认配置：

```
[INFO] Starting llama-server with command: /path/to/llama-server --model /models/xxx.gguf --port 8125 --ctx-size 4096 --parallel 1 ...
```

## 项目结构

```
llama-proxy/
├── api/                    # API 定义
│   └── openai/v1/          # OpenAI 协议接口定义
├── internal/
│   ├── cmd/               # 命令行入口
│   ├── controller/        # 控制器
│   ├── logic/             # 业务逻辑
│   ├── middleware/        # 中间件
│   ├── model/             # 数据模型
│   └── service/           # 服务接口
├── manifest/
│   └── config/            # 配置文件
└── README.md
```

## License

MIT