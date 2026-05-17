package worker

import (
	"context"
	"time"

	db "github.com/Jingqi0327/eleclog/db/sqlc"
	"github.com/Jingqi0327/eleclog/logger"
	"github.com/Jingqi0327/eleclog/mail"
	"github.com/Jingqi0327/eleclog/util"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

type Mail_Alerter struct {
	store       db.Store
	cron        *cron.Cron
	emailSender mail.EmailSender
}

func NewMailAlerter(store db.Store, emailSender mail.EmailSender) *Mail_Alerter {
	return &Mail_Alerter{
		store:       store,
		emailSender: emailSender,
		cron:        cron.New(),
	}
}

func (alerter *Mail_Alerter) Start() error {
	// 每小时的第 10 分钟执行（错开整点 API 刷新延迟）
	_, err := alerter.cron.AddFunc("10 * * * *", func() {
		alerter.RunNow()
	})

	if err != nil {
		return err
	}

	alerter.cron.Start()
	return nil
}

func (alerter *Mail_Alerter) Stop() context.Context {
	return alerter.cron.Stop()
}

// RunNow 立即执行一次检查通知逻辑
func (alerter *Mail_Alerter) RunNow() {
	logger.Log.Info("Mail_Alerter 开始执行检查通知逻辑...")
	ctx := context.Background()
	// 从Notifications表中查询需要检测的房间，以及对应的用户和阈值
	roomList, err := alerter.store.ListDueUserRoomNotifications(ctx)
	if err != nil {
		logger.Log.Error("查询需检测房间列表失败", zap.Error(err))
		return
	}

	rooms_surplus := make(map[int64]int64) // room_id -> current surplus
	for _, item := range roomList {
		// 获取当前剩余电量
		if _, exists := rooms_surplus[item.RoomID]; !exists {
			record, err := alerter.store.GetLatestBalance(ctx, item.RoomID)
			if err != nil {
				logger.Log.Error("查询房间当前剩余电量失败",
					zap.Int64("room_id", item.RoomID),
					zap.Error(err),
				)
				continue
			}
			rooms_surplus[item.RoomID] = record.Balance
		}
		surplus := rooms_surplus[item.RoomID]

		if surplus <= int64(item.Threshold)*100 { // 阈值单位为元，余额单位为分
			// 查询用户信息
			userinfo, err := alerter.store.GetUser(ctx, item.Username)
			if err != nil {
				logger.Log.Error("查询用户信息失败",
					zap.String("username", item.Username),
					zap.Error(err),
				)
				continue
			}

			// TODO: 发送和修改使用事务

			// 发送邮件通知用户
			// 使用 Zap 打印日志，代替 log.Printf，可以获得更详细的上下文信息
			logger.Log.Info("发送邮件通知",
				zap.String("username", item.Username),
				zap.Int64("room_id", item.RoomID),
				zap.String("surplus", util.FormatCentsToYuan(surplus)),
				zap.String("threshold", util.FormatCentsToYuan(int64(item.Threshold)*100)),
			)
			subject := "寝室电量不足通知"
			content := "您好，您的寝室电量剩余 " + util.FormatCentsToYuan(surplus) + " 元，已低于您设置的阈值 " + util.FormatCentsToYuan(int64(item.Threshold)*100) + " 元，请及时充值。"
			to := []string{userinfo.Email}
			err = alerter.emailSender.SendEmail(subject, content, to, nil, nil, nil)
			if err != nil {
				logger.Log.Error("发送邮件通知失败",
					zap.String("username", item.Username),
					zap.Int64("room_id", item.RoomID),
					zap.Error(err),
				)
				continue
			}

			// 更新 last_notified_at 字段为当前时间
			arg := db.UpdateUserRoomNotificationLastNotifiedAtParams{
				Username:       item.Username,
				RoomID:         item.RoomID,
				LastNotifiedAt: time.Now(),
			}
			_, err = alerter.store.UpdateUserRoomNotificationLastNotifiedAt(ctx, arg)
			if err != nil {
				logger.Log.Error("更新通知时间失败",
					zap.String("username", item.Username),
					zap.Int64("room_id", item.RoomID),
					zap.Error(err),
				)
			}
		}
	}
}
