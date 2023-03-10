package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/ilhamfzri/simple-go-grpc/pb"
	"github.com/ilhamfzri/simple-go-grpc/service"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

func seedUsers(userStore service.UserStore) error {
	err := createUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}
	return createUser(userStore, "user1", "secret", "user")
}

func createUser(userStore service.UserStore, username, password, role string) error {
	user, err := service.NewUser(username, password, role)
	if err != nil {
		return err
	}
	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {
	const laptopServicePath = "/simple.pcbook.LaptopService/"
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()

	log.Printf("start server on port %d", *port)

	userStore := service.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatalf("seed users failed: %v", err)
	}

	jwtManager := service.NewJWTmanger(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	laptopServer := service.NewLaptopServer(
		service.NewInMemoryLaptopStore(),
		service.NewDiskImageStore("img"),
		service.NewInMomoryRatingStore())

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)
	reflection.Register(grpcServer)

	address := fmt.Sprintf("0.0.0.0:%d", *port)
	listerner, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}

	err = grpcServer.Serve(listerner)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
