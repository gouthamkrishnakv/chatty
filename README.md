# Chatty - gRPC-based Chat Application

Chatty is a gRPC-based chat application that allows users to send messages to each other.

For specification, check the [gRPC specification file](./proto/chatty.proto).

### Run

```bash
make run
```

### Generate gRPC code

> Note: Make sure `protoc` protocol buffer compiler is installed on your system.

```bash
make generate
```

### Dev mode build

> Note: Install go's air reloader before running the following command
> ```bash
> go install github.com/air-verse/air@latest
> ```

```bash
make dev
```
