package grpc

import (
	"os"

	"github.com/YOVO-LABS/workflow/proto/dense"
	"google.golang.org/grpc"
)

// PredictgRPCConnection connects to the auth gRPC service
func PredictgRPCConnection() (dense.PredictClient, error) {
	conn, err := grpc.Dial(os.Getenv("ML_HOST"), grpc.WithInsecure())
	if err != nil {
		return dense.NewPredictClient(conn), err
	}
	return dense.NewPredictClient(conn), nil
}
