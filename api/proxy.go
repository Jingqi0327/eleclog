package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Jingqi0327/eleclog/pb"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc/metadata"
)

func (server *Server) getGrpcContext(ctx *gin.Context) context.Context {
	authHeader := ctx.GetHeader(authorizationHeaderKey)
	return metadata.AppendToOutgoingContext(ctx, authorizationHeaderKey, authHeader)
}

// proxyQueryArea GET /proxy/areas
// 获取校区列表（参数固定，无需前端传入）
func (server *Server) proxyQueryArea(ctx *gin.Context) {
	cacheKey := "proxy:areas"
	var result interface{}

	if server.cache != nil {
		if err := server.cache.Get(ctx, cacheKey, &result); err == nil {
			ctx.JSON(http.StatusOK, result)
			return
		}
	}

	// 发起 gRPC 调用，传入带有 auth header 的 context
	grpcCtx := server.getGrpcContext(ctx)
	resp, err := server.proxyClient.QueryArea(grpcCtx, &pb.QueryAreaRequest{})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "代理微服务请求失败"})
		return
	}

	// 解析返回的 JSON 字符串
	if err := json.Unmarshal([]byte(resp.JsonData), &result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "代理服务返回数据格式错误"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryBuilding GET /proxy/buildings?areaId=xxx
// 获取楼栋列表（参数：校区id）
func (server *Server) proxyQueryBuilding(ctx *gin.Context) {
	areaId := ctx.Query("areaId")
	if areaId == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 areaId"})
		return
	}

	cacheKey := fmt.Sprintf("proxy:buildings:%s", areaId)
	var result interface{}

	if server.cache != nil {
		if err := server.cache.Get(ctx, cacheKey, &result); err == nil {
			ctx.JSON(http.StatusOK, result)
			return
		}
	}

	// 发起 gRPC 调用，传入带有 auth header 的 context
	grpcCtx := server.getGrpcContext(ctx)
	resp, err := server.proxyClient.QueryBuilding(grpcCtx, &pb.QueryBuildingRequest{
		AreaId: areaId,
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "代理微服务请求失败"})
		return
	}

	if err := json.Unmarshal([]byte(resp.JsonData), &result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "代理服务返回数据格式错误"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryFloor GET /proxy/floors?areaId=xxx&buildingCode=xxx
// 获取楼层列表（参数：校区id、楼栋编号）
func (server *Server) proxyQueryFloor(ctx *gin.Context) {
	areaId := ctx.Query("areaId")
	buildingCode := ctx.Query("buildingCode")
	if areaId == "" || buildingCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 areaId 或 buildingCode"})
		return
	}

	cacheKey := fmt.Sprintf("proxy:floors:%s:%s", areaId, buildingCode)
	var result interface{}

	if server.cache != nil {
		if err := server.cache.Get(ctx, cacheKey, &result); err == nil {
			ctx.JSON(http.StatusOK, result)
			return
		}
	}

	// 发起 gRPC 调用，传入带有 auth header 的 context
	grpcCtx := server.getGrpcContext(ctx)
	resp, err := server.proxyClient.QueryFloor(grpcCtx, &pb.QueryFloorRequest{
		AreaId:       areaId,
		BuildingCode: buildingCode,
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "代理微服务请求失败"})
		return
	}

	if err := json.Unmarshal([]byte(resp.JsonData), &result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "代理服务返回数据格式错误"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryRoom GET /proxy/rooms?areaId=xxx&buildingCode=xxx&floorCode=xxx
// 获取寝室列表（参数：校区id、楼栋编号、楼层编号）
func (server *Server) proxyQueryRoom(ctx *gin.Context) {
	areaId := ctx.Query("areaId")
	buildingCode := ctx.Query("buildingCode")
	floorCode := ctx.Query("floorCode")
	if areaId == "" || buildingCode == "" || floorCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数 areaId、buildingCode 或 floorCode"})
		return
	}

	cacheKey := fmt.Sprintf("proxy:rooms:%s:%s:%s", areaId, buildingCode, floorCode)
	var result interface{}

	if server.cache != nil {
		if err := server.cache.Get(ctx, cacheKey, &result); err == nil {
			ctx.JSON(http.StatusOK, result)
			return
		}
	}

	// 发起 gRPC 调用，传入带有 auth header 的 context
	grpcCtx := server.getGrpcContext(ctx)
	resp, err := server.proxyClient.QueryRoom(grpcCtx, &pb.QueryRoomRequest{
		AreaId:       areaId,
		BuildingCode: buildingCode,
		FloorCode:    floorCode,
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "代理微服务请求失败"})
		return
	}

	if err := json.Unmarshal([]byte(resp.JsonData), &result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "代理服务返回数据格式错误"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryRoomSurplus GET /proxy/room-surplus?areaId=xxx&buildingCode=xxx&floorCode=xxx&roomCode=xxx
// 用于获取房间全称 displayRoomName
func (server *Server) proxyQueryRoomSurplus(ctx *gin.Context) {
	areaId := ctx.Query("areaId")
	buildingCode := ctx.Query("buildingCode")
	floorCode := ctx.Query("floorCode")
	roomCode := ctx.Query("roomCode")
	if areaId == "" || buildingCode == "" || floorCode == "" || roomCode == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "缺少参数"})
		return
	}

	var result interface{}

	// 发起 gRPC 调用，传入带有 auth header 的 context
	grpcCtx := server.getGrpcContext(ctx)
	resp, err := server.proxyClient.QueryRoomSurplus(grpcCtx, &pb.QueryRoomSurplusRequest{
		AreaId:       areaId,
		BuildingCode: buildingCode,
		FloorCode:    floorCode,
		RoomCode:     roomCode,
	})
	if err != nil {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "代理微服务请求失败"})
		return
	}

	if err := json.Unmarshal([]byte(resp.JsonData), &result); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "代理服务返回数据格式错误"})
		return
	}

	ctx.JSON(http.StatusOK, result)
}
