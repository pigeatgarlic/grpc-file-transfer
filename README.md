# File transfer via GRPC Proof of Concept

## Build

### protocol buffers

You can build protocol buffer files using the command 

```shell
build\proto-gen.bat
```

## Run

### Server

```shell
go run .\cmd\server\main.go start -path="C:\Desktop"
```

### Client

```shell
go run .\cmd\client\main.go upload -path="C:\Desktop\file.txt"
```

## Dependencies

### Protoc 
Download the latest protoc here https://github.com/protocolbuffers/protobuf/releases/
Copy the binary file into your $GOPATH/bin. You can find your $GOPATH using `go env`