package gapi

import (
	"context"
	"fmt"
	"time"

	"github.com/Jingqi0327/eleclog/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GrpcLogger 是一个 gRPC 的一元拦截器（类似 HTTP 中间件），用于记录每个请求的日志
func GrpcLogger(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (resp any, err error) {
	startTime := time.Now()

	// 先执行实际的 RPC 方法
	result, err := handler(ctx, req)

	duration := time.Since(startTime)
	statusCode := codes.Unknown

	// 提取 gRPC 状态码
	if st, ok := status.FromError(err); ok {
		statusCode = st.Code()
	}

	// 为不同状态码配置颜色
	var color string
	switch statusCode {
	case codes.OK:
		color = "\033[32m" // Green
	case codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.Unauthenticated, codes.PermissionDenied:
		color = "\033[33m" // Yellow
	case codes.Internal, codes.Unknown, codes.DeadlineExceeded, codes.Unimplemented, codes.Unavailable, codes.DataLoss:
		color = "\033[31m" // Red
	default:
		color = "\033[34m" // Blue
	}
	reset := "\033[0m"

	coloredStatus := fmt.Sprintf("%s%s%s", color, statusCode.String(), reset)

	// 准备日志字段
	fields := []zap.Field{
		zap.String("protocol", "grpc"),
		zap.String("method", info.FullMethod),
		zap.Int("status_code", int(statusCode)),
		zap.String("status_text", statusCode.String()), // 原始文本，用于检索
		zap.Duration("duration", duration),
	}

	// 自定义类似 Gin 的带颜色的日志消息头部
	logMessage := fmt.Sprintf(" %s | %s", coloredStatus, info.FullMethod)

	// 如果有错误，附加上去
	if err != nil {
		fields = append(fields, zap.Error(err))
		logger.Log.Error("[gRPC]"+logMessage, fields...)
	} else {
		logger.Log.Info("[gRPC]"+logMessage, fields...)
	}

	return result, err
}
