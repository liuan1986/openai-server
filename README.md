# openai-server

该项目实现了一个简单的 OpenAI 代理服务，核心能力包括：

* 为合法的 `userId` 签发访问凭证 `access_key`。
* 使用令牌桶对用户进行限流，所有用户共享同一套限流配置。
* 拦截并校验调用方提供的 `access_key`，支持黑名单控制。
* 代理调用下游 OpenAI Chat Completions 接口并返回结果。

## 快速开始

1. 准备配置文件，可以直接拷贝 `config.sample.json` 并按需修改：
   ```bash
   cp config.sample.json config.json
   ```
   需要重点关注以下字段：
   * `service.secret_key`：用于生成 access_key 的服务端密钥。
   * `service.users`：允许访问的用户列表。
   * `service.blacklist`：被拒绝访问的用户。
   * `openai.target_url` / `openai.api_key`：被代理的 OpenAI 接口与凭证。
   * `rate_limit`：限流配置。

2. 运行服务：
   ```bash
   CONFIG_PATH=./config.json go run ./cmd/server
   ```

3. 获取 access_key：
   ```bash
   curl -X POST http://localhost:8080/api/v1/get_access_key \
     -H "Content-Type: application/json" \
     -d '{"userId":"alice"}'
   ```

4. 代理调用 OpenAI 接口：
   ```bash
   curl -X POST http://localhost:8080/api/v1/chat/completions \
     -H "Authorization: <access_key>" \
     -H "Content-Type: application/json" \
     -d '{
       "model": "gpt-4o-mini",
       "messages": [
         {"role": "system", "content": "You are a helpful assistant."},
         {"role": "user", "content": "你好"}
       ]
     }'
   ```

## Docker 镜像

使用下面的命令即可构建镜像：
```bash
docker build -t openai-server:latest .
```

运行镜像时可以通过挂载配置文件覆盖默认配置：
```bash
docker run -p 8080:8080 \
  -v $(pwd)/config.json:/app/config.json \
  -e CONFIG_PATH=/app/config.json \
  openai-server:latest
```

## Kubernetes 部署

`k8s/deployment.yaml` 提供了一个基础示例，包含 ConfigMap、Deployment 与 Service：
```bash
kubectl apply -f k8s/deployment.yaml
```

请根据实际环境修改镜像地址 (`image`) 以及 ConfigMap 中的配置内容。

## 项目结构

```
cmd/server/          # HTTP 服务入口
internal/auth/       # access_key 生成与校验逻辑
internal/config/     # 配置加载逻辑
internal/proxy/      # 下游 OpenAI 接口代理
internal/ratelimit/  # 简单的令牌桶限流实现
```

## 开发

* 格式化代码：`go fmt ./...`
* 依赖管理：`go mod tidy`

欢迎根据业务需求继续扩展。
