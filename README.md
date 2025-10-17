# openai-server

## 服务实现内容，
1、代理其他 openai 服务的 api 接口
2、提供独立的 access_key，只有合法的access_key才能请求openai接口
3、通过 userId 置换 access_key，access_key由是根据userId + 服务端配置 进行加密的
4、当校验access_key是否合法时，需要根据服务端配置 解密access_key 并获取 userId，同时校验userId是否存在
5、支持根据 userId配置黑名单，黑名单内的禁止访问
6、支持基于令牌桶的限流策略，根据userId限流，为了保持逻辑简单，所有userId采用同一个限流策略


## 服务提供接口如下

1、获取access_key接口
{base_url}/api/v1/get_access_key
入参： userId
出参： access_key


2、open ai接口:
示例如下：
curl -X POST {base_url}/api/v1/chat/completions \
-H "Authorization: {access_key}" \
-H "Content-Type: application/json" \
-d '{
    "model": "{model_name}",
    "messages": [
        {
            "role": "system",
            "content": "You are a helpful assistant."
        },
        {
            "role": "user", 
            "content": "你是谁？"
        }
    ]
}'



## 被代理的接口示例：
curl {target_url} \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer {access_key}" \
  -d '{
    "model": "{model_name}",
    "messages": "{messages}""
  }'

其中
target_url、access_key 本地配置
model_name、messages 从参数读取
