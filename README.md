# Simple-Go-GRPC 
Simple implementation of gRPC in Golang.

## Service and RPC List 
```sh
+---------------+--------------+---------------------+----------------------+
|    SERVICE    |     RPC      |    REQUEST TYPE     |    RESPONSE TYPE     |
+---------------+--------------+---------------------+----------------------+
| AuthService   | Login        | LoginRequest        | LoginResponse        |
| LaptopService | CreateLaptop | CreateLaptopRequest | CreateLaptopResponse |
| LaptopService | SearchLaptop | SearchLaptopRequest | SearchLaptopResponse |
| LaptopService | UploadImage  | UploadImageRequest  | UploadImageResponse  |
| LaptopService | RateLaptop   | RateLaptopRequest   | RateLaptopResponse   |
+---------------+--------------+---------------------+----------------------+
```

## Quickstart 
1. Install Module Requirement
```sh
go mod tidy
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




