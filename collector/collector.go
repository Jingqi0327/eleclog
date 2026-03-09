package collector

import (
	"context"
	"log"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/robfig/cron/v3"
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
func (c *Collector) Start() {
	// 每小时的第 5 分钟执行（错开整点 API 刷新延迟）
	_, err := c.cron.AddFunc("5 * * * *", func() {
		c.RunNow()
	})

	if err != nil {
		log.Fatalf("无法添加定时任务: %v", err)
	}

	c.cron.Start()
	log.Println("Collector 定时任务已启动...")
}

// RunNow 立即执行一次抓取逻辑
func (collector *Collector) RunNow() {
	ctx := context.Background()
	rooms, err := collector.store.ListRoomsAll(ctx)
	if err != nil {
		log.Printf("读取寝室列表失败: %v", err)
		return
	}

	for _, room := range rooms {
		// 调用你写好的 FetchSurplus
		balance, err := collector.FetchSurplus(room.AreaID, room.BuildingCode, room.FloorCode, room.RoomCode)
		if err != nil {
			log.Printf("抓取寝室 [%s] 失败: %v", room.Name, err)
			continue
		}

		// 存入数据库 (使用 sqlc 生成的 CreateElectricityRecord)
		_, err = collector.store.CreateElectricityRecord(ctx, db.CreateElectricityRecordParams{
			RoomID:  room.ID,
			Balance: util.ToCents(balance),// 转换为分存储
		})

		if err != nil {
			log.Printf("记录寝室 [%s] 电费失败: %v", room.Name, err)
		} else {
			log.Printf("成功记录寝室 [%s] 余额: %.2f", room.Name, balance)
		}
	}
}
