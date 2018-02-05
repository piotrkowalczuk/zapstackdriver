package zapstackdrivergrpc

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	"github.com/piotrkowalczuk/zapstackdriver"
	"github.com/piotrkowalczuk/zapstackdriver/operation"
)

func UnaryServerInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		now := time.Now()
		httpReq, ctx := contextBackground(ctx, info.FullMethod)

		log.Debug("grpc unary request received by the server", zap.Object("httpRequest", httpReq), operation.FromContextFirst(ctx))

		res, err := handler(ctx, req)

		httpReq.Latency = time.Since(now)
		logRequest(ctx, log, err, httpReq)

		return res, err
	}
}

func StreamServerInterceptor(log *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()
		now := time.Now()
		httpReq, ctx := contextBackground(ctx, info.FullMethod)

		log.Debug("grpc stream request received by the server", zap.Object("httpRequest", httpReq), operation.FromContextFirst(ctx))

		err := handler(srv, ss)

		httpReq.Latency = time.Since(now)
		logRequest(ctx, log, err, httpReq)

		return err
	}
}

func contextBackground(ctx context.Context, fullMethod string) (zapstackdriver.HTTPRequest, context.Context) {
	var httpReq zapstackdriver.HTTPRequest

	httpReq.URL, httpReq.Method = split(fullMethod)
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if userAgent, ok := md["user-agent"]; ok {
			httpReq.UserAgent = userAgent[0]
		}
	}
	if p, ok := peer.FromContext(ctx); ok {
		httpReq.RemoteIP = p.Addr.String()
	}
	return httpReq, operation.WithContext(ctx)
}

func logRequest(ctx context.Context, log *zap.Logger, err error, req zapstackdriver.HTTPRequest) {
	if st, ok := status.FromError(err); ok {
		req.Status = int(st.Code())

		switch st.Code() {
		case codes.OK:
			log.Debug("grpc request processed by the server successfully",
				zap.Object("httpRequest", req),
				operation.FromContextLast(ctx),
			)
		case codes.Internal:
			log.Error("grpc unary request processed by the server with error",
				zap.Object("httpRequest", req),
				operation.FromContextLast(ctx),
				zap.String("error", st.Message()),
			)
		default:
			log.Warn("grpc unary request processed by the server but something went wrong",
				zap.Object("httpRequest", req),
				operation.FromContextLast(ctx),
			)
		}
		return
	}

	log.Error("grpc request processed by the server with unhandled error",
		zap.Object("httpRequest", req),
		operation.FromContextLast(ctx),
		zap.Error(err),
	)
}

func split(name string) (string, string) {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		return name[1:i], name[i+1:]
	}
	return "unknown", "unknown"
}
