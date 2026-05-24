package db

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
)

type BindRoomToUserTxParams struct {
	Username     string `json:"username"`
	Name         string `json:"name"`
	AreaID       string `json:"area_id"`
	BuildingCode string `json:"building_code"`
	FloorCode    string `json:"floor_code"`
	RoomCode     string `json:"room_code"`
	Threshold    int32  `json:"threshold"`
}

type BindRoomToUserTxResult struct {
	Room     Room     `json:"room"`
	UserRoom UserRoom `json:"user_room"`
}

func (store *SQLStore) BindRoomToUserTx(ctx context.Context, arg BindRoomToUserTxParams) (BindRoomToUserTxResult, error) {
	var result BindRoomToUserTxResult

	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		// 1. Try to get the room by unique codes
		room, err := q.GetRoomByCodes(ctx, GetRoomByCodesParams{
			AreaID:       arg.AreaID,
			BuildingCode: arg.BuildingCode,
			FloorCode:    arg.FloorCode,
			RoomCode:     arg.RoomCode,
		})

		if err != nil {
			if errors.Is(err, ErrRecordNotFound) {
				// 2. If room does not exist, create it
				room, err = q.CreateRoom(ctx, CreateRoomParams{
					Name:         arg.Name,
					AreaID:       arg.AreaID,
					BuildingCode: arg.BuildingCode,
					FloorCode:    arg.FloorCode,
					RoomCode:     arg.RoomCode,
				})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		} else if room.Name != arg.Name {
			// Update the room name if it's different (e.g. fixing previously incorrect short names)
			room, err = q.UpdateRoom(ctx, UpdateRoomParams{
				ID:   room.ID,
				Name: pgtype.Text{String: arg.Name, Valid: true},
			})
			if err != nil {
				return err
			}
		}

		result.Room = room

		// 3. Check if user room already exists
		userRoom, err := q.GetUserRoom(ctx, GetUserRoomParams{
			Username: arg.Username,
			RoomID:   room.ID,
		})

		if err != nil {
			if errors.Is(err, ErrRecordNotFound) {
				// 4. Bind the room to the user
				userRoom, err = q.CreateUserRoom(ctx, CreateUserRoomParams{
					Username:  arg.Username,
					RoomID:    room.ID,
					Threshold: arg.Threshold,
					IsEnabled: true,
				})
				if err != nil {
					return err
				}
			} else {
				return err
			}
		}

		result.UserRoom = userRoom
		return nil
	})

	return result, err
}
