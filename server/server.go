package main

import (
    "context"
    "log"
    "net"
    "encoding/json"

    "github.com/nasparria/FirstGo/GRPC/proto"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "google.golang.org/grpc"
	"google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type server struct {
	proto.UnimplementedMyServiceServer // Add this line
	db *mongo.Database
}

func (s *server) GetData(ctx context.Context, in *proto.DataRequest) (*proto.DataResponse, error) {
    collection := s.db.Collection("orders")
    filter := bson.M{"ticker": in.GetQuery()}
    var result bson.M
    err := collection.FindOne(ctx, filter).Decode(&result)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, status.Errorf(codes.NotFound, "No document found with ticker: %s", in.GetQuery())
        }
        return nil, err
    }
    resultJSON, err := json.Marshal(result)
    if err != nil {
        return nil, err
    }
    return &proto.DataResponse{Result: string(resultJSON)}, nil
}


func main() {
    client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    db := client.Database("portfolio")

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("Failed to listen: %v", err)
    }
    s := grpc.NewServer()
    proto.RegisterMyServiceServer(s, &server{db: db})
	log.Println("Server is running on :50051") // <-- Add this line here

    if err := s.Serve(lis); err != nil {
        log.Fatalf("Failed to serve: %v", err)
    }
}
