package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
)

const xiaofubaoBase = "https://application.xiaofubao.com/app/electric/"

func (server *Server) newRestyClient() *resty.Client {
	return resty.New().
		SetHeader("Cookie", fmt.Sprintf("shiroJID=%s", server.config.ShiroJID)).
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36")
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

	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform": "YUNMA_APP",
			"type":     "1",
		}).
		SetResult(&result).
		Get(xiaofubaoBase + "queryArea")

	if err != nil || !resp.IsSuccess() {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "上游请求失败"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryBuilding GET /proxy/buildings?areaId=xxx
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

	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform": "YUNMA_APP",
			"areaId":   areaId,
		}).
		SetResult(&result).
		Get(xiaofubaoBase + "queryBuilding")

	if err != nil || !resp.IsSuccess() {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "上游请求失败"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryFloor GET /proxy/floors?areaId=xxx&buildingCode=xxx
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

	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform":     "YUNMA_APP",
			"areaId":       areaId,
			"buildingCode": buildingCode,
		}).
		SetResult(&result).
		Get(xiaofubaoBase + "queryFloor")

	if err != nil || !resp.IsSuccess() {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "上游请求失败"})
		return
	}

	if server.cache != nil {
		_ = server.cache.Set(ctx, cacheKey, result, 24*time.Hour)
	}

	ctx.JSON(http.StatusOK, result)
}

// proxyQueryRoom GET /proxy/rooms?areaId=xxx&buildingCode=xxx&floorCode=xxx
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

	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform":     "YUNMA_APP",
			"areaId":       areaId,
			"buildingCode": buildingCode,
			"floorCode":    floorCode,
		}).
		SetResult(&result).
		Get(xiaofubaoBase + "queryRoom")

	if err != nil || !resp.IsSuccess() {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "上游请求失败"})
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
	resp, err := server.newRestyClient().R().
		SetQueryParams(map[string]string{
			"platform":     "YUNMA_APP",
			"areaId":       areaId,
			"buildingCode": buildingCode,
			"floorCode":    floorCode,
			"roomCode":     roomCode,
		}).
		SetResult(&result).
		Post(xiaofubaoBase + "queryRoomSurplus")

	if err != nil || !resp.IsSuccess() {
		ctx.JSON(http.StatusBadGateway, gin.H{"error": "上游请求失败"})
		return
	}
	ctx.JSON(http.StatusOK, result)
}
