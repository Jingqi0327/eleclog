package db

import (
	"context"
	"time"
)

type UpdateUserRoomNotificationLastNotifiedAtTxParams struct {
	UpdateUserRoomNotificationLastNotifiedAtParams
	AfterUpdate func(notification UserRoomNotification) error
}

func (store *SQLStore) UpdateRoomNotificationLastNotifiedAtTx(ctx context.Context, arg UpdateUserRoomNotificationLastNotifiedAtTxParams) (UserRoomNotification, error) {
	var userRoomNotification UserRoomNotification
	err := store.execTx(ctx, func(q *Queries) error {
		arg.LastNotifiedAt = time.Now()
		var updateErr error
		userRoomNotification, updateErr = q.UpdateUserRoomNotificationLastNotifiedAt(ctx, arg.UpdateUserRoomNotificationLastNotifiedAtParams)
		if updateErr != nil {
			return updateErr
		}

		if cbErr := arg.AfterUpdate(userRoomNotification); cbErr != nil {
			return cbErr
		}

		return nil
	})
	return userRoomNotification, err
}
