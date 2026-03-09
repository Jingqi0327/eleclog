package collector

import (
	"fmt"

	"github.com/go-resty/resty/v2"
)

// 根据 API 实际返回的 JSON 结构定义
type SurplusResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Amount          float64 `json:"amount"`          // 剩余金额 (元)
		DisplayRoomName string  `json:"displayRoomName"` // 房间全称
	} `json:"data"`
}

func (collector *Collector) FetchSurplus(areaID, buildingCode, floorCode, roomCode string) (float64, error) {
	client := resty.New()

	apiURL := "https://application.xiaofubao.com/app/electric/queryRoomSurplus"

	var result SurplusResponse
	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/148.0.0.0 Safari/537.36").
		SetHeader("Cookie", fmt.Sprintf("shiroJID=%s",collector.config.ShiroJID)).
		SetQueryParams(map[string]string{
			"platform":     "YUNMA_APP",
			"areaId":       areaID,
			"buildingCode": buildingCode,
			"floorCode":    floorCode,
			"roomCode":     roomCode,
		}).
		SetResult(&result). // 自动将 JSON 解析到 struct
		Post(apiURL)

	if err != nil {
		return 0, fmt.Errorf("请求失败: %v", err)
	}

	if !resp.IsSuccess() {
		return 0, fmt.Errorf("接口返回错误代码: %d", resp.StatusCode())
	}

	if !result.Success {
		return 0, fmt.Errorf("接口返回失败: %v", result)
	}

	return result.Data.Amount, nil
}
