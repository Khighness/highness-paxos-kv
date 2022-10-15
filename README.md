# highness-paxos-kv
Basic paxos kv store.

## build proto
```shell
protoc --proto_path=. --go_out=plugins=grpc:. api/paxos.proto
```
