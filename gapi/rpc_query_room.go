package gapi

import (
	"context"

	"github.com/Jingqi0327/eleclog/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// QueryRoom 实现 gRPC 接口
func (server *Server) QueryRoom(ctx context.Context, req *pb.QueryRoomRequest) (*pb.QueryRoomResponse, error) {
	_, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if req.GetAreaId() == "" || req.GetBuildingCode() == "" || req.GetFloorCode() == "" {
		return nil, status.Errorf(codes.InvalidArgument, "缺少参数 areaId、buildingCode 或 floorCode")
	}

	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform":     "YUNMA_APP",
			"areaId":       req.GetAreaId(),
			"buildingCode": req.GetBuildingCode(),
			"floorCode":    req.GetFloorCode(),
		}).
		Get(xiaofubaoBase + "queryRoom")

	if err != nil {
		return nil, status.Errorf(codes.Internal, "请求校付宝失败: %v", err)
	}

	if !resp.IsSuccess() {
		return nil, status.Errorf(codes.Internal, "校付宝返回错误状态码: %d", resp.StatusCode())
	}

	return &pb.QueryRoomResponse{
		JsonData: string(resp.Body()),
	}, nil
}
