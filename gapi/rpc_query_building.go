package gapi

import (
	"context"

	"github.com/Jingqi0327/eleclog/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryBuilding 实现 gRPC 接口
func (server *Server) QueryBuilding(ctx context.Context, req *pb.QueryBuildingRequest) (*pb.QueryBuildingResponse, error) {
	_, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}
	
	if req.GetAreaId() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "缺少参数 areaId")
	}

	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform": "YUNMA_APP",
			"areaId":   req.GetAreaId(),
		}).
		Get(xiaofubaoBase + "queryBuilding")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "请求校付宝失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, status.Errorf(codes.Internal, "校付宝返回错误状态码: %d", resp.StatusCode())
	}

	return &pb.QueryBuildingResponse{
		JsonData: string(resp.Body()),
	}, nil
}
