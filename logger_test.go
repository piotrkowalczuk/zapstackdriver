package zapstackdriver_test

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"testing"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"net/http"
	"time"

	"github.com/piotrkowalczuk/zapstackdriver"
	"github.com/piotrkowalczuk/zapstackdriver/internal/testutil"
	"github.com/piotrkowalczuk/zapstackdriver/operation"
	"github.com/piotrkowalczuk/zapstackdriver/zapstackdrivergrpc"
)

func Example() {
	logger, err := zapstackdriver.NewStackdriverConfig().Build()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	logger = logger.With(zap.Object("serviceContext", &zapstackdriver.ServiceContext{
		Service: "example-service",
		Version: "v0.1.0",
	}))

	ctx := operation.WithContext(context.Background())

	logger.Debug("http request received",
		operation.FromContextFirst(ctx),
		zap.Object("httpRequest", zapstackdriver.HTTPRequest{
			Method:    "GET",
			URL:       "example.com",
			UserAgent: "curl",
			Referrer:  "test",
			RemoteIP:  "127.0.0.1",
		}),
	)
	logger.Debug("something important happened", operation.FromContext(ctx))

	logger.Debug("http response send",
		operation.FromContextLast(ctx),
		zap.Object("httpRequest", zapstackdriver.HTTPRequest{
			Method:    "GET",
			URL:       "example.com",
			UserAgent: "curl",
			Referrer:  "test",
			Status:    http.StatusTeapot,
			RemoteIP:  "127.0.0.1",
			Latency:   10 * time.Second,
		}),
	)
}

func BenchmarkEncoder(b *testing.B) {
	for _, env := range []string{"production", "development", "stackdriver"} {
		b.Run(env, func(b *testing.B) {
			b.ReportAllocs()

			l, err := testutil.Init(testutil.Opts{
				Environment: env,
				Level:       "debug",
			})
			if err != nil {
				b.Fatalf("unexpected error: %s", err.Error())
			}
			opr := operation.FromContextFirst(context.Background())

			b.ResetTimer()

			for n := 0; n < b.N; n++ {
				l.Debug("debug message", opr,
					zap.Object("httpRequest", zapstackdriver.HTTPRequest{
						Method:    "GET",
						URL:       "example.com",
						UserAgent: "curl",
						Referrer:  "test",
						Status:    http.StatusTeapot,
						RemoteIP:  "127.0.0.1",
						Latency:   10 * time.Second,
					}),
					zap.Object("serviceContext", zapstackdriver.ServiceContext{
						Service: "example-service",
						Version: "v0.1.0",
					}),
				)
			}
		})
	}
}

func TestStreamServerInterceptor(t *testing.T) {
	l, err := testutil.Init(testutil.Opts{Environment: "stackdriver"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	testStreamServerInterceptor(t, l, status.Errorf(codes.OK, "ok"))
	testStreamServerInterceptor(t, l, status.Errorf(codes.InvalidArgument, "example error"))
	testStreamServerInterceptor(t, l, status.Errorf(codes.Internal, "internal"))
	testStreamServerInterceptor(t, l, errors.New("unhandled error"))
}

func testStreamServerInterceptor(t *testing.T, l *zap.Logger, exp error) {
	t.Helper()

	itr := zapstackdrivergrpc.StreamServerInterceptor(l)
	err := itr(testContext(), &stream{}, &grpc.StreamServerInfo{
		FullMethod: "service/method",
	}, func(srv interface{}, stream grpc.ServerStream) error {
		return exp
	})
	if err != exp {
		t.Fatalf("wrong error: %s", err.Error())
	}
}

func TestUnaryServerInterceptor(t *testing.T) {
	l, err := testutil.Init(testutil.Opts{Environment: "stackdriver"})
	if err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}

	testUnaryServerInterceptor(t, l, status.Errorf(codes.OK, "ok"))
	testUnaryServerInterceptor(t, l, status.Errorf(codes.InvalidArgument, "example error"))
	testUnaryServerInterceptor(t, l, status.Errorf(codes.Internal, "internal"))
	testUnaryServerInterceptor(t, l, errors.New("unhandled error"))
}

func testUnaryServerInterceptor(t *testing.T, l *zap.Logger, exp error) {
	t.Helper()

	itr := zapstackdrivergrpc.UnaryServerInterceptor(l)
	_, err := itr(testContext(), nil, &grpc.UnaryServerInfo{
		FullMethod: "service/method",
	}, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, exp
	})
	if err != exp {
		t.Fatalf("wrong error: %s", err.Error())
	}
}

func testContext() context.Context {
	ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
		"user-agent": []string{"fake agent"},
	})
	return peer.NewContext(ctx, &peer.Peer{
		Addr: &net.IPAddr{
			IP: net.IP{127, 0, 0, 1},
		},
	})
}

type stream struct {
	grpc.ServerStream
}

func (s *stream) Context() context.Context {
	return context.Background()
}
