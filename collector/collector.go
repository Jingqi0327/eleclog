package collector

import (
	"context"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Collector struct {
	config util.Config
	store  db.Store
	cron   *cron.Cron
}

func NewCollector(config util.Config, store db.Store) *Collector {
	return &Collector{
		config: config,
		store:  store,
		cron:   cron.New(),
	}
}

// Start 启动后台定时任务
func (c *Collector) Start() error {
	// 每小时的第 5 分钟执行（错开整点 API 刷新延迟）
	_, err := c.cron.AddFunc("5 * * * *", func() {
		c.RunNow()
	})

	if err != nil {
		return err
	}

	c.cron.Start()
	return nil
}

func (c *Collector) Stop() context.Context {
	return c.cron.Stop()
}

// RunNow 立即执行一次抓取逻辑
func (collector *Collector) RunNow() {
	ctx := context.Background()
	rooms, err := collector.store.ListRoomsAll(ctx)
	if err != nil {
		logger.Log.Error("读取寝室列表失败", zap.Error(err))
		return
	}

	for _, room := range rooms {
		// 调用你写好的 FetchSurplus
		balance, err := collector.FetchSurplus(room.AreaID, room.BuildingCode, room.FloorCode, room.RoomCode)
		if err != nil {
			logger.Log.Error("抓取寝室电量失败",
				zap.String("room_name", room.Name),
				zap.Error(err),
			)
			continue
		}

		// 存入数据库 (使用 sqlc 生成的 CreateElectricityRecord)
		_, err = collector.store.CreateElectricityRecord(ctx, db.CreateElectricityRecordParams{
			RoomID:  room.ID,
			Balance: util.ToCents(balance), // 转换为分存储
		})

		if err != nil {
			logger.Log.Error("记录寝室电费失败",
				zap.String("room_name", room.Name),
				zap.Error(err),
			)
		} else {
			logger.Log.Info("成功记录寝室电量",
				zap.String("room_name", room.Name),
				zap.String("balance", util.FormatCentsToYuan(util.ToCents(balance))),
			)
		}
	}
}
