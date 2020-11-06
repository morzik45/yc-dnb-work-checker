package dnb_ydb

import (
	"context"
	"fmt"
	"github.com/yandex-cloud/ydb-go-sdk"
	"github.com/yandex-cloud/ydb-go-sdk/table"
)

func (db DB) InsertWork(userID uint64, status uint8) error {
	readTx := table.TxControl(
		table.BeginTx(
			table.WithSerializableReadWrite(),
		),
		table.CommitTx(),
	)

	query := fmt.Sprintf(`
		PRAGMA TablePathPrefix("%s");
		DECLARE $user_id AS Uint64;
		DECLARE $status AS Uint8;
		UPSERT INTO works
		(
			user_id,
			time,
			status
		)
		VALUES
		(
			$user_id,
			CurrentUtcTimestamp(),
			$status
		);`, Database)

	err := table.Retry(db.Ctx, db.SessionPool,
		table.OperationFunc(func(ctx context.Context, s *table.Session) (err error) {
			_, _, err = s.Execute(ctx,
				readTx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$user_id", ydb.Uint64Value(userID)),
					table.ValueParam("$status", ydb.Uint8Value(status)),
				),
			)
			return
		}))
	return err
}

func (db DB) CountFreeWorkFromCurrentDay(userID uint64) error {
	readTx := table.TxControl(
		table.BeginTx(
			table.WithSerializableReadWrite(),
		),
		table.CommitTx(),
	)

	query := fmt.Sprintf(`
		PRAGMA TablePathPrefix("%s");
		DECLARE $user_id AS Uint64;
		SELECT COUNT(*) FROM works WHERE user_id = $user_id AND time > `, Database)

	err := table.Retry(db.Ctx, db.SessionPool,
		table.OperationFunc(func(ctx context.Context, s *table.Session) (err error) {
			_, _, err = s.Execute(ctx,
				readTx,
				query,
				table.NewQueryParameters(
					table.ValueParam("$user_id", ydb.Uint64Value(userID)),
				),
			)
			return
		}))
	return err
}
