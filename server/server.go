package main
import (
  "context"
  "log"
  "net"
  "strconv"
  myapi "github.com/nasparria/FirstGo/proto"
  "go.mongodb.org/mongo-driver/bson"
  "go.mongodb.org/mongo-driver/mongo"
  "go.mongodb.org/mongo-driver/mongo/options"
  "google.golang.org/grpc"
  "google.golang.org/grpc/codes"
  "google.golang.org/grpc/status"
  "go.mongodb.org/mongo-driver/bson/primitive"
  "google.golang.org/protobuf/encoding/protojson"
)
type server struct {
  myapi.UnimplementedMyServiceServer
  db *mongo.Database
}

type OrderMapping struct {
  Account        string  `bson:"account"`
  Action         string  `bson:"action"`
  Average_price  string  `bson:"average_price"`
  Created_at     primitive.DateTime `bson:"created_at"`
  Fee            string  `bson:"fee"`
  Is_prime       bool    `bson:"is_prime"`
  Limit_price    string  `bson:"limit_price"`
  Market_time    string  `bson:"market_time"`
  Notional       string  `bson:"notional"`
  Order_id       string  `bson:"order_id"`
  Order_type     string  `bson:"order_type"`
  Quantity       string  `bson:"quantity"`
  Quantity_shares string  `bson:"quantity_shares"`
  Quantity_type  string  `bson:"quantity_type"`
  Status         string  `bson:"status"`
  Ticker         string  `bson:"ticker"`
  Trading_type   string  `bson:"trading_type"`
  Updated_at     primitive.DateTime `bson:"updated_at"`
  User_id        string  `bson:"user_id"`
}

func (s *server) GetData(ctx context.Context, in *myapi.DataRequest) (*myapi.JsonResponse, error) {
  collection := s.db.Collection("orders")
  filter := bson.M{"ticker": in.GetQuery()}

  var resultMapping OrderMapping

  err := collection.FindOne(ctx, filter).Decode(&resultMapping)
  if err != nil {
      if err == mongo.ErrNoDocuments {
          return nil, status.Errorf(codes.NotFound, "No document found with ticker: %s", in.GetQuery())
      }
      return nil, err
  }

  createdAtStr := strconv.FormatInt(int64(resultMapping.Created_at), 10)
  updatedAtStr := strconv.FormatInt(int64(resultMapping.Updated_at), 10)

  result := myapi.Order{
      Account:        resultMapping.Account,
      Action:         resultMapping.Action,
      AveragePrice:   resultMapping.Average_price,
      CreatedAt:      createdAtStr,
      Fee:            resultMapping.Fee,
      IsPrime:        resultMapping.Is_prime,
      LimitPrice:     resultMapping.Limit_price,
      MarketTime:     resultMapping.Market_time,
      Notional:       resultMapping.Notional,
      OrderId:        resultMapping.Order_id,
      OrderType:      resultMapping.Order_type,
      Quantity:       resultMapping.Quantity,
      QuantityShares: resultMapping.Quantity_shares,
      QuantityType:   resultMapping.Quantity_type,
      Status:         resultMapping.Status,
      Ticker:         resultMapping.Ticker,
      TradingType:    resultMapping.Trading_type,
      UpdatedAt:      updatedAtStr,
      UserId:         resultMapping.User_id,
}

jsonResponse, err := protojson.Marshal(&result)
if err != nil {
    return nil, status.Errorf(codes.Internal, "failed to convert to JSON: %v", err)
}
return &myapi.JsonResponse{Json: string(jsonResponse)}, nil
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
  myapi.RegisterMyServiceServer(s, &server{db: db})
log.Println("Server is running on :50051")

  if err := s.Serve(lis); err != nil {
      log.Fatalf("Failed to serve: %v", err)
}
}