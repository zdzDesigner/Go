# @Author: zdzDesigner
# @Date:   2019-02-11 10:18:09
# @Last Modified by:   zdzDesigner
# @Last Modified time: 2019-04-26 11:48:06
#!/bin/bash

echo "mod download ..."
export GOPROXY=https://goproxy.io
export GO111MODULE=on
# export GOPROXY=http://www.alfredzhong.xyz:3000

# 本的环境变量
# export ENV=dev
# export ENV=beta
export ENV=LOCAL_TEST


## 日志地址
export SYSLOG_HOST=
## 内容服务地址
# export KONG_CONTENT_SERVER_INTERNAL=http://10.12.6.76:24846
# export KONG_CONTENT_SERVER_INTERNAL=http://10.12.6.76:8865
export KONG_CONTENT_SERVER_INTERNAL=http://kong.dui.ai

# export KONG_CONTENT_SERVER_INTERNAL=http://127.0.0.1:4001

## cinfo服务地址 http://172.16.20.194:20000
# export CINFOSERVER_INTERNAL=http://dds.dev.dui.ai
export CINFOSERVER_INTERNAL=http://dds.dui.ai

# es
export ES_INTERNAL=http://10.12.6.35:9200
export ES_USERNAME=username
export ES_PASSWORD=password
## ba接口 
export KGES_INTERNAL=http://dev.ba.internal.dui.ai/kges


## redis
export SKILL_CONTEXT_REDIS_HOST=localhost
export SKILL_CONTEXT_REDIS_PORT=6379
export SKILL_CONTEXT_REDIS_PASSWORD=
export WEBHOOK_GOLANG_REDIS_DBNO=

## etcd
export ETCD_SERVER_INTERNAL=http://127.0.0.1:2379
# export ETCD_SERVER_INTERNAL=http://10.12.6.35:2379
# export ETCD_SERVER_INTERNAL='["http://10.12.6.35:2379"]'
# export ETCD_SERVER_INTERNAL='["http://10.12.6.35:2379","http://127.0.0.1:2379"]'



go mod download

go run main.go