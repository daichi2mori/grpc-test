package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"

	"google.golang.org/grpc"
)

func myUnaryClientInterceptor1(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	fmt.Println("[pre unary] unary client interceptor 1: ", method, req)
	err := invoker(ctx, method, req, res, cc, opts...)
	fmt.Println("[post unary] unary client interceptor 1: ", res)
	return err
}

func myUnaryClientInterceptor2(ctx context.Context, method string, req, res interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	fmt.Println("[pre unary] unary client interceptor 1: ", method, req)
	err := invoker(ctx, method, req, res, cc, opts...)
	fmt.Println("[post unary] unary client interceptor 1: ", res)
	return err
}

type myClientStreamWrapper1 struct {
	grpc.ClientStream
}

type myClientStreamWrapper2 struct {
	grpc.ClientStream
}

func myStreamClientInterceptor1(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("[pre] stream client interceptor 1: ", method)
	stream, err := streamer(ctx, desc, cc, method, opts...)
	return &myClientStreamWrapper1{stream}, err
}

func myStreamClientInterceptor2(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	log.Println("[pre] stream client interceptor 1: ", method)
	stream, err := streamer(ctx, desc, cc, method, opts...)
	return &myClientStreamWrapper2{stream}, err
}

func (s *myClientStreamWrapper1) SendMsg(m interface{}) error {
	log.Println("[pre message] stream client interceptor 1: ", m)
	return s.ClientStream.SendMsg(m)
}

func (s *myClientStreamWrapper1) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if !errors.Is(err, io.EOF) {
		log.Println("[post message] stream client interceptor 1: ", m)
	}
	return err
}

func (s *myClientStreamWrapper1) CloseSend() error {
	err := s.ClientStream.CloseSend()
	log.Println("[post] stream client interceptor 1")
	return err
}

func (s *myClientStreamWrapper2) SendMsg(m interface{}) error {
	log.Println("[pre message] stream client interceptor 1: ", m)
	return s.ClientStream.SendMsg(m)
}

func (s *myClientStreamWrapper2) RecvMsg(m interface{}) error {
	err := s.ClientStream.RecvMsg(m)
	if !errors.Is(err, io.EOF) {
		log.Println("[post message] stream client interceptor 1: ", m)
	}
	return err
}

func (s *myClientStreamWrapper2) CloseSend() error {
	err := s.ClientStream.CloseSend()
	log.Println("[post] stream client interceptor 1")
	return err
}
