package db

import (
	"context"
	"time"
)

type UpdateUserRoomLastNotifiedAtTxParams struct {
	UpdateUserRoomLastNotifiedAtParams
	AfterUpdate func(notification UserRoom) error
}

func (store *SQLStore) UpdateRoomLastNotifiedAtTx(ctx context.Context, arg UpdateUserRoomLastNotifiedAtTxParams) (UserRoom, error) {
	var userRoom UserRoom
	err := store.execTx(ctx, func(q *Queries) error {
		arg.LastNotifiedAt = time.Now()
		var updateErr error
		userRoom, updateErr = q.UpdateUserRoomLastNotifiedAt(ctx, arg.UpdateUserRoomLastNotifiedAtParams)
		if updateErr != nil {
			return updateErr
		}

		if cbErr := arg.AfterUpdate(userRoom); cbErr != nil {
			return cbErr
		}

		return nil
	})
	return userRoom, err
}
