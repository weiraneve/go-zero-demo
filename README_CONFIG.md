# OpenAI 配置化客户端使用说明

## 概述

这个项目已经升级为支持配置文件的OpenAI客户端，可以轻松管理不同的API配置，包括自定义代理、模型选择等。

## 配置文件结构

### config.yaml

```yaml
# OpenAI 配置文件
chat_models:
  platform: "openai"
  configs:
    api_key: "your-api-key-here"
    proxy_url: "https://apis.linkin.love/v1"
    default_model: "lilchat-13b"
    summary_model: "lilfunc"
    
# 其他配置
settings:
  max_tokens: 100
  stream: true
  timeout: 30s
```

## 配置说明

### ChatModels 配置

- **platform**: 平台类型，目前支持 "openai"
- **api_key**: OpenAI API密钥
- **proxy_url**: 自定义API代理地址
- **default_model**: 默认使用的模型名称
- **summary_model**: 摘要模型名称（预留）

### Settings 配置

- **max_tokens**: 最大生成token数量
- **stream**: 是否启用流式响应
- **timeout**: 请求超时时间

## 环境变量支持

为了安全起见，建议使用环境变量覆盖敏感配置：

```bash
# 设置API密钥
export OPENAI_API_KEY="your-real-api-key"

# 设置代理URL
export OPENAI_PROXY_URL="https://your-proxy.com/v1"

# 设置默认模型
export OPENAI_DEFAULT_MODEL="gpt-4"
```

## 使用方法

1. **修改配置文件**：
   ```bash
   # 编辑 config.yaml 文件
   vim config.yaml
   ```

2. **设置环境变量**（推荐）：
   ```bash
   export OPENAI_API_KEY="sk-your-real-api-key"
   ```

3. **运行程序**：
   ```bash
   go run .
   ```

## 安全建议

1. **不要将真实的API密钥提交到版本控制系统**
2. **使用环境变量管理敏感信息**
3. **在生产环境中使用配置管理工具**
4. **定期轮换API密钥**

## 自定义模型支持

该配置支持使用自定义模型，如：
- `lilchat-13b`
- `lilfunc`
- 或其他兼容OpenAI API的模型

## 代理配置

支持通过 `proxy_url` 配置自定义API端点，适用于：
- 企业内部代理
- 第三方API网关
- 自建OpenAI兼容服务

## 故障排除

### 常见错误

1. **401 Unauthorized**：检查API密钥是否正确
2. **连接超时**：检查代理URL是否可访问
3. **模型不存在**：确认模型名称是否正确

### 调试模式

程序启动时会显示当前使用的配置信息，便于调试：

```
使用配置:
- 平台: openai
- 代理URL: https://apis.linkin.love/v1
- 默认模型: lilchat-13b
- 最大Token: 100
```

## 扩展功能

### 添加新的配置项

1. 在 `config.go` 中添加新的结构体字段
2. 在 `config.yaml` 中添加对应配置
3. 在主程序中使用新配置

### 支持多个API提供商

可以扩展配置结构以支持多个API提供商，如Azure OpenAI、Anthropic等。