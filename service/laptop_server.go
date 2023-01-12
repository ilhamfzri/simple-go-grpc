package service

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"

	"github.com/google/uuid"
	"github.com/ilhamfzri/simple-go-grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MaxImageSize = 1MB
const MaxImageSize = 1 << 20

type LaptopServer struct {
	pb.UnimplementedLaptopServiceServer
	LaptopStore LaptopStore
	ImageStore  ImageStore
	RatingStore RatingStore
}

func NewLaptopServer(laptopStore LaptopStore, imageStore ImageStore, ratingStore RatingStore) *LaptopServer {
	return &LaptopServer{
		LaptopStore: laptopStore,
		ImageStore:  imageStore,
		RatingStore: ratingStore,
	}
}

func (server *LaptopServer) CreateLaptop(
	ctx context.Context,
	req *pb.CreateLaptopRequest,
) (*pb.CreateLaptopResponse, error) {
	laptop := req.GetLaptop()
	log.Printf("receive a create-laptop request with id: %s", laptop.Id)

	if err := contextError(ctx); err != nil {
		return nil, err
	}

	if len(laptop.Id) > 0 {
		_, err := uuid.Parse(laptop.Id)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "laptop ID is not a valid UUID: %v", err)
		}
	} else {
		id, err := uuid.NewRandom()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "cannot generate a new laptop ID: %v", err)
		}
		laptop.Id = id.String()
	}

	err := server.LaptopStore.Save(laptop)
	if err != nil {
		code := codes.Internal
		if errors.Is(err, ErrAlreadyExists) {
			code = codes.AlreadyExists
		}
		return nil, status.Errorf(code, "cannot save laptop to the store: %v", err)
	}

	log.Printf("saved laptop with id: %s", laptop.Id)

	res := &pb.CreateLaptopResponse{
		Id: laptop.Id,
	}
	return res, nil
}

func (server *LaptopServer) SearchLaptop(
	req *pb.SearchLaptopRequest,
	stream pb.LaptopService_SearchLaptopServer,
) error {
	filter := req.GetFilter()
	log.Printf("receive a search-laptop request with filter: %v", filter)

	err := server.LaptopStore.Search(
		stream.Context(),
		filter,
		func(laptop *pb.Laptop) error {
			res := &pb.SearchLaptopResponse{Laptop: laptop}

			err := stream.Send(res)
			if err != nil {
				return err
			}

			log.Printf("sent laptop with id: %s", laptop.GetId())
			return nil
		},
	)

	if err != nil {
		return status.Errorf(codes.Internal, "unexpected error: %v", err)
	}

	return nil
}

func (server *LaptopServer) UploadImage(stream pb.LaptopService_UploadImageServer) error {
	req, err := stream.Recv()
	if err != nil {
		log.Print("cannot receive image info", err)
		return status.Errorf(codes.Internal, "cannot receive image info:")
	}

	laptopID := req.GetInfo().GetLaptopId()
	imageType := req.GetInfo().GetImageType()
	log.Printf("receive an upload-image request for laptop %s with image type %s", laptopID, imageType)

	laptop, err := server.LaptopStore.Find(laptopID)
	if err != nil {
		log.Printf("cannot find laptop : %v", err)
		return status.Errorf(codes.Internal, "cannot find laptop")
	}

	if laptop == nil {
		log.Printf("laptop doesnt exist : %s", laptopID)
		return status.Errorf(codes.InvalidArgument, "laptop doesnt exist: %s", laptopID)
	}

	imageData := bytes.Buffer{}
	imageSize := 0

	for {
		if err := contextError(stream.Context()); err != nil {
			return err
		}
		log.Print("waiting to receive more data")

		req, err := stream.Recv()
		if err == io.EOF {
			log.Print("no more data")
			break
		}

		if err != nil {
			log.Printf("cannot receive chunk data : %v", err)
			return status.Errorf(codes.Unknown, "cannot receive chunk data: %v", err)
		}

		chunk := req.GetChunkData()
		size := len(chunk)

		imageSize += size
		if imageSize > MaxImageSize {
			log.Printf("image size to large: %d > %d", imageSize, MaxImageSize)
			return status.Errorf(codes.InvalidArgument, "image size to large: %d > %d", imageSize, MaxImageSize)
		}

		_, err = imageData.Write(chunk)
		if err != nil {
			log.Printf("error writing chunk data : %v", err)
			return status.Errorf(codes.Internal, "error writing chunk data : %v", err)
		}
	}

	imageID, err := server.ImageStore.Save(laptopID, imageType, imageData)
	if err != nil {
		log.Printf("cannot save image to the store: %v", err)
		return status.Errorf(codes.Internal, "cannot save image to the store: %v", err)
	}

	res := &pb.UploadImageResponse{
		Id:   imageID,
		Size: uint32(imageSize),
	}

	err = stream.SendAndClose(res)
	if err != nil {
		log.Printf("cannot send response %v", err)
		return status.Errorf(codes.Internal, "cannot send response")
	}

	log.Printf("saved image with id: %s, size: %d", imageID, imageSize)

	return nil
}

func (server *LaptopServer) RateLaptop(stream pb.LaptopService_RateLaptopServer) error {
	for {
		err := contextError(stream.Context())
		if err != nil {
			return err
		}

		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("no more else")
			break
		}

		if err != nil {
			log.Printf("cannot receive stream request: %v", err)
			return status.Errorf(codes.Unknown, "cannot receive stream request: %v", err)
		}

		laptopID := req.GetLaptopId()
		score := req.GetScore()

		log.Printf("received a rate-laptop request: id = %s, score = %f", laptopID, score)

		found, err := server.LaptopStore.Find(laptopID)
		if err != nil {
			log.Printf("cannot find laptop: %v", err)
			return status.Errorf(codes.Internal, "cannot find laptop: %v", err)
		}

		if found == nil {
			log.Printf("laptopID %s not found", laptopID)
			return status.Errorf(codes.NotFound, "laptopID %s not found", laptopID)
		}

		rating, err := server.RatingStore.Add(laptopID, score)
		if err != nil {
			log.Printf("cannot add rating to the store: %v", err)
			return status.Errorf(codes.Internal, "cannot add rating to the store: %v", err)
		}

		res := &pb.RateLaptopResponse{
			LaptopId:     laptopID,
			RatedCount:   rating.Count,
			AverageScore: rating.Sum / float64(rating.Count),
		}

		err = stream.Send(res)
		if err != nil {
			log.Printf("cannot send stream response: %v", err)
			return status.Errorf(codes.Internal, "cannot send stream response: %v", err)
		}
	}
	return nil
}

func contextError(ctx context.Context) error {
	switch ctx.Err() {
	case context.Canceled:
		log.Print("request is canceled")
		return status.Error(codes.Canceled, "request is canceled")
	case context.DeadlineExceeded:
		log.Print("deadline is exceeded")
		return status.Error(codes.DeadlineExceeded, "deadline is exceeded")
	default:
		return nil
	}
}
