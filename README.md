# Simple-Go-GRPC 
Simple implementation of gRPC in Golang.

## Service and RPC List 
```sh
+---------------+--------------+---------------------+----------------------+-----------------------------+
|    SERVICE    |     RPC      |    REQUEST TYPE     |    RESPONSE TYPE     |           RPC TYPE          |
+---------------+--------------+---------------------+----------------------+-----------------------------+
| AuthService   | Login        | LoginRequest        | LoginResponse        | Simple RPC                  |
| LaptopService | CreateLaptop | CreateLaptopRequest | CreateLaptopResponse | Simple RPC                  |
| LaptopService | SearchLaptop | SearchLaptopRequest | SearchLaptopResponse | Server-Side Streaming RPC   |
| LaptopService | UploadImage  | UploadImageRequest  | UploadImageResponse  | Client-Side Streaming RPC   |
| LaptopService | RateLaptop   | RateLaptopRequest   | RateLaptopResponse   | Bidirectional Streaming RPC |
+---------------+--------------+---------------------+----------------------+-----------------------------+
```

## Quickstart 
1. Install Module Requirement
```sh
go mod download
```

2. Generate gRPC Code 
```sh
make gen
```

3. Run gRPC Server
```sh
make server
```

4. Run grPC Client 
    - via Code
    ```sh
    make client
    ```
    - via CLI </br>
    Make sure you have installed [evans-cli](https://github.com/ktr0731/evans) first.
    
    ```sh
    make evans-cli
    ```




