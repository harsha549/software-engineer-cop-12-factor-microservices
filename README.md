你好！
很冒昧用这样的方式来和你沟通，如有打扰请忽略我的提交哈。我是光年实验室（gnlab.com）的HR，在招Golang开发工程师，我们是一个技术型团队，技术氛围非常好。全职和兼职都可以，不过最好是全职，工作地点杭州。
我们公司是做流量增长的，Golang负责开发SAAS平台的应用，我们做的很多应用是全新的，工作非常有挑战也很有意思，是国内很多大厂的顾问。
如果有兴趣的话加我微信：13515810775  ，也可以访问 https://gnlab.com/，联系客服转发给HR。
12-factor microservices
=======================

### Initial setup

```
# start a postgres container
docker run --name postgres -p 5432:5432 -e POSTGRES_PASSWORD=password -e POSTGRES_DB=babynames -d postgres

# start a kafka container
docker run -p 2181:2181 -p 9092:9092 --env ADVERTISED_HOST=127.0.0.1 --env ADVERTISED_PORT=9092 -d spotify/kafka

# import the raw data
go run baby-names-import/main.go
```

### 12-factor microservices

```
# start the baby names API
go run 12-factor-microservices/baby-names-api/main.go

# query the baby names API
curl localhost:8080/top10 | jq

# start the baby names input API
go run 12-factor-microservices/baby-names-input-api/main.go

# register a new baby
curl -v -d '{"name": "OLIVER", "sex": "male"}' localhost:8081/baby

# query the baby names API
curl localhost:8080/top10 | jq
```

### Event driven architecture

```
# start the baby names API
go run event-driven-architecture/baby-names-api/main.go

# query the baby names API
curl localhost:8080/top10 | jq

# start the baby names input API
go run event-driven-architecture/baby-names-input-api/main.go

# register a new baby
curl -v -d '{"name": "OLIVER", "sex": "male"}' localhost:8081/baby

# start the baby names processor
go run event-driven-architecture/baby-names-processor/main.go

# start the baby names streaming API
go run event-driven-architecture/baby-names-streaming-api/main.go

# follow the stream
curl -v localhost:8082/stream

# register a new baby
curl -v -d '{"name": "OLIVER", "sex": "male"}' localhost:8081/baby
```
