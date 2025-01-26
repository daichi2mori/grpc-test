package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"time"

	hellopb "grpc-test/pkg/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

type myServer struct {
	hellopb.UnimplementedGreetingServiceServer
}

// Unary RPCがレスポンスを返す
func (s *myServer) Hello(ctx context.Context, req *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		log.Println(md)
	}

	headerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "header"})
	if err := grpc.SetHeader(ctx, headerMD); err != nil {
		return nil, err
	}

	trailerMD := metadata.New(map[string]string{"type": "unary", "from": "server", "in": "trailer"})
	if err := grpc.SetTrailer(ctx, trailerMD); err != nil {
		return nil, err
	}

	return &hellopb.HelloResponse{
		Message: fmt.Sprintf("Hello, %s!", req.GetName()),
	}, nil

	// 何か処理をしてエラーが発生した
	// stat := status.New(codes.Unknown, "unknown error occurred")
	// stat, _ = stat.WithDetails(&errdetails.DebugInfo{
	// 	Detail: "detail reason of err",
	// })
	// err := stat.Err()

	// return &hellopb.HelloResponse{
	// 	Message: "Hello, " + req.GetName(),
	// }, err
}

// Server Stream RPCがレスポンスを受け取る
func (s *myServer) HelloServerStream(req *hellopb.HelloRequest, stream grpc.ServerStreamingServer[hellopb.HelloResponse]) error {
	resCount := 5
	for i := 0; i < resCount; i++ {
		if err := stream.Send(&hellopb.HelloResponse{
			Message: fmt.Sprintf("[%d] Hello, %s!", i, req.GetName()),
		}); err != nil {
			return err
		}
		time.Sleep(time.Second * 1)
	}
	return nil
}

// Client Stream RPCがレスポンスを受け取る
func (s *myServer) HelloClientStream(stream grpc.ClientStreamingServer[hellopb.HelloRequest, hellopb.HelloResponse]) error {
	nameList := make([]string, 0)
	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			message := fmt.Sprintf("Hello, %v!", nameList)
			return stream.SendAndClose(&hellopb.HelloResponse{
				Message: message,
			})
		}
		if err != nil {
			return err
		}
		nameList = append(nameList, req.GetName())
	}
}

// 双方向ストリーミング
func (s *myServer) HelloBiStreams(stream grpc.BidiStreamingServer[hellopb.HelloRequest, hellopb.HelloResponse]) error {
	if md, ok := metadata.FromIncomingContext(stream.Context()); ok {
		log.Println(md)
	}

	headerMD := metadata.New(map[string]string{"type": "stream", "from": "server", "in": "header"})

	//（パターン1）すぐにヘッダーを送信する場合
	// if err := stream.SendHeader(headerMD); err != nil {
	// 	return err
	// }

	//（パターン2）本来のタイミングでヘッダーを送信する場合
	if err := stream.SetHeader(headerMD); err != nil {
		return err
	}

	trailerMD := metadata.New(map[string]string{"type": "stream", "from": "server", "in": "trailer"})
	stream.SetTrailer(trailerMD)

	for {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		message := fmt.Sprintf("Hello, %v!", req.GetName())
		if err := stream.Send(&hellopb.HelloResponse{
			Message: message,
		}); err != nil {
			return err
		}
	}
}

func NewMyServer() *myServer {
	return &myServer{}
}

func main() {
	// 1. 8080番ポートで起動
	port := 8080
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		panic(err)
	}

	// 2. gRPCサーバーを作成
	s := grpc.NewServer(
		// interceptorファイルで実装したミドルウェアを適用

		// grpc.UnaryInterceptor(myUnaryServerInterceptor1),
		grpc.ChainUnaryInterceptor(
			myUnaryServerInterceptor1,
			myUnaryServerInterceptor2,
		),
		// grpc.StreamInterceptor(myStreamServerInterceptor1),
		grpc.ChainStreamInterceptor(
			myStreamServerInterceptor1,
			myStreamServerInterceptor2,
		),
	)

	// 3. gRPCサーバーにGreetingServiceを登録
	hellopb.RegisterGreetingServiceServer(s, NewMyServer())

	// 4. サーバーリフレクションの設定
	reflection.Register(s)

	// 5. 作成したgRPCサーバーを、8080番ポートで稼働させる
	go func() {
		log.Printf("start gRPC server port: %v", port)
		s.Serve(listener)
	}()

	// 6. Ctrl+Cが入力されたらGraceful shutdownされるようにする
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("stopping gRPC server...")
	s.GracefulStop()
}
