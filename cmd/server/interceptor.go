package main

import (
	"context"
	"errors"
	"io"
	"log"

	"google.golang.org/grpc"
)

// インターセプタ関数の引数は決められている
// ミドルウェアのことをGoの世界ではインターセプターと呼ぶ

func myUnaryServerInterceptor1(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("[pre unary] unary server interceptor 1: ", info.FullMethod) // ハンドラの前に割り込ませる前処理
	res, err := handler(ctx, req)                                            // 本来の処理
	log.Println("[post unary] unary server interceptor 1: ", res)            // ハンドラの後に割り込ませる後処理
	return res, err
}

func myUnaryServerInterceptor2(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	log.Println("[pre unary] unary server interceptor 2: ", info.FullMethod, req)
	res, err := handler(ctx, req)
	log.Println("[post unary] unary server interceptor 2: ", res)
	return res, err
}

type myServerStreamWrapper1 struct {
	grpc.ServerStream
}

type myServerStreamWrapper2 struct {
	grpc.ServerStream
}

func myStreamServerInterceptor1(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// ストリームがopenされたときに行われる前処理
	log.Println("[pre stream] stream server interceptor 1: ", info.FullMethod)

	err := handler(srv, &myServerStreamWrapper1{ss}) // 本来の処理

	// ストリームがcloseされたときに行われる後処理
	log.Println("[post stream] stream server interceptor 1: ")
	return err
}

func myStreamServerInterceptor2(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	log.Println("[pre stream] stream server interceptor 2: ", info.FullMethod)
	err := handler(srv, &myServerStreamWrapper2{ss})
	log.Println("[post stream] stream server interceptor 2: ")
	return err
}

func (s *myServerStreamWrapper1) RecvMsg(m interface{}) error {
	// ストリームからリクエストを受信
	err := s.ServerStream.RecvMsg(m)

	// 受信したリクエストを、ハンドラで処理する前に差し込む前処理
	if !errors.Is(err, io.EOF) {
		log.Println("[pre message] RecvMsg: ", m)
	}
	return err
}

func (s *myServerStreamWrapper1) SendMsg(m interface{}) error {
	// ハンドラで作成したレスポンスを、ストリームから返信する直前に差し込む後処理
	log.Println("[post message] RecvMsg:", m)
	return s.ServerStream.SendMsg(m)
}

func (s *myServerStreamWrapper2) RecvMsg(m interface{}) error {
	err := s.ServerStream.RecvMsg(m)
	if !errors.Is(err, io.EOF) {
		log.Println("[pre message] my stream server interceptor 2: ", m)
	}
	return err
}

func (s *myServerStreamWrapper2) SendMsg(m interface{}) error {
	log.Println("[post message] my stream server interceptor 2: ", m)
	return s.ServerStream.SendMsg(m)
}
