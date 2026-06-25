# queue-service

Для запуска:

В первом терминале
```
go run main.go --port 80
```

Во 2 терминале
```
curl http://127.0.0.1/pet -v
```

В 3 терминале
```
curl -X PUT http://127.0.0.1:80/pet\?v\=cat
```