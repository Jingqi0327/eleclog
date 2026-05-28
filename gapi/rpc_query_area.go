package gapi

import (
	"context"

	"github.com/Jingqi0327/eleclog/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const xiaofubaoBase = "https://application.xiaofubao.com/app/electric/"

// QueryArea 实现 gRPC 接口
func (server *Server) QueryArea(ctx context.Context, req *pb.QueryAreaRequest) (*pb.QueryAreaResponse, error) {
	_, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	// 直接获取原始字符串作为结果
	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform": "YUNMA_APP",
			"type":     "1",
		}).
		Get(xiaofubaoBase + "queryArea")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "请求校付宝失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, status.Errorf(codes.Internal, "校付宝返回错误状态码: %d", resp.StatusCode())
	}

	return &pb.QueryAreaResponse{
		JsonData: string(resp.Body()), // 直接将获取到的 JSON 返回
	}, nil
}
