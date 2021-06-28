# File transfer via GRPC Proof of Concept

## Build

### protocol buffers

You can build protocol buffer files using the command 

```shell
build\proto-gen.bat
```

## Dependencies

### Protoc 
Download the latest protoc here https://github.com/protocolbuffers/protobuf/releases/
Copy the binary file into your $GOPATH/bin. You can find your $GOPATH using `go env`