package dnb_ydb

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/yandex-cloud/ydb-go-sdk"
	"github.com/yandex-cloud/ydb-go-sdk/auth/iam"
	"github.com/yandex-cloud/ydb-go-sdk/table"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var Database = os.Getenv("YDB_DATABASE")

type User struct {
	UserID   uint64
	Username string
	Lang     string
	Referral string
	IsAdmin  bool
	IsActive bool

	Coins int

	Bonus         bool
	BonusDatetime time.Time
	BonusCoins    int

	FirstVisit     time.Time
	LastVisit      time.Time
	BannedDatetime time.Time

	CountVip  int
	CountFree int

	Referrals int
	Rub       float64

	CountPayments int
	SumSpent      float64
}

type DB struct {
	Ctx         context.Context
	SessionPool *table.SessionPool
	Close       func()
}

func NewDB() *DB {

	db := new(DB)

	ctx, cancel := context.WithCancel(context.Background())

	db.Ctx = ctx

	key, err := getPrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	credentials, err := iam.NewClient(
		iam.WithIssuer(os.Getenv("YDB_ISSUER")),
		iam.WithKeyID(os.Getenv("YDB_KEYID")),
		iam.WithPrivateKey(key),
		iam.WithDefaultEndpoint(),
		iam.WithSystemCertPool(),
	)
	if err != nil {
		log.Fatal(err)
	}

	cert, err := ioutil.ReadFile("CA.pem")
	if err != nil {
		log.Fatal(err)
	}
	rootCerts, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}
	if ok := rootCerts.AppendCertsFromPEM(cert); !ok {
		log.Fatal(err)
	}

	dialer := &ydb.Dialer{
		DriverConfig: &ydb.DriverConfig{
			Database:    Database,
			Credentials: credentials,
		},
		TLSConfig: &tls.Config{
			RootCAs: rootCerts,
		},
		Timeout: time.Second * 2,
	}

	driver, err := dialer.Dial(ctx, os.Getenv("YDB_ADDR"))
	if err != nil {
		log.Fatal(err)
	}

	sp := table.SessionPool{
		IdleThreshold: time.Second,
		Builder: &table.Client{
			Driver: driver,
		},
	}

	db.SessionPool = &sp

	db.Close = func() {
		cancel()
		driver.Close()
		sp.Close(context.Background())
	}
	return db
}

func getPrivateKey() (key *rsa.PrivateKey, err error) {
	pemString := os.Getenv("YDB_PRIVATE_KEY")

	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		return nil, fmt.Errorf("invalid pem encoding")
	}
	key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	if err == nil {
		return
	}
	x, _ := x509.ParsePKCS8PrivateKey(block.Bytes)
	if key, _ = x.(*rsa.PrivateKey); key != nil {
		err = nil
	}
	return
}

func (db *DB) SetWorkStatus(userID uint64, token string) (workStatus uint8, count int, err error) {

	readTx := table.TxControl(
		table.BeginTx(
			table.WithSerializableReadWrite(),
		),
	)

	err = table.Retry(db.Ctx, db.SessionPool,
		table.OperationFunc(func(ctx context.Context, s *table.Session) (err error) {

			user := new(User)

			getUserQuery := fmt.Sprintf(`
				PRAGMA TablePathPrefix("%s");
				DECLARE $user_id AS Uint64;
				SELECT 
					user_id,
					is_admin,
					lang,
					coins,
					bonus,
					bonus_datetime,
					bonus_coins,
					banned_datetime,
				FROM
					users
				WHERE
					user_id=$user_id;`, Database)

			t, res, err := s.Execute(ctx,
				readTx,
				getUserQuery,
				table.NewQueryParameters(
					table.ValueParam("$user_id", ydb.Uint64Value(userID)),
				),
			)
			if err != nil || res.Err() != nil {
				return err
			}
			for res.NextSet() {
				for res.NextRow() {
					res.NextItem()
					user.UserID = res.OUint64()
					if user.UserID != userID {
						return errors.New("user.UserID != userID")
					}
					res.NextItem()
					user.IsAdmin = res.OBool()
					res.NextItem()
					user.Lang = string(res.OString())
					res.NextItem()
					user.Coins = int(res.OUint64())
					res.NextItem()
					user.Bonus = res.OBool()
					res.NextItem()
					user.BonusDatetime = time.Unix(0, int64(res.OTimestamp())*int64(time.Microsecond))
					res.NextItem()
					user.BonusCoins = int(res.OUint64())
					res.NextItem()
					user.BannedDatetime = time.Unix(0, int64(res.OTimestamp())*int64(time.Microsecond))
				}
			}

			var updateUserQuery string

			switch {
			case user.BannedDatetime.After(time.Now().UTC()):

			case user.IsAdmin:
				workStatus = 3
				updateUserQuery = fmt.Sprintf(`
					PRAGMA TablePathPrefix("%s");
					DECLARE $user_id AS Uint64;
					DECLARE $token AS String;
					DECLARE $status AS Uint8;
					UPDATE users SET token = $token WHERE user_id=$user_id;
					UPSERT INTO works (user_id, time, status) VALUES ($user_id, CurrentUtcTimestamp(), $status);`, Database)

			case user.Bonus && user.BonusCoins > 0 && user.BonusDatetime.After(time.Now().UTC()):
				count = user.BonusCoins - 1
				workStatus = 2
				updateUserQuery = fmt.Sprintf(`
					PRAGMA TablePathPrefix("%s");
					DECLARE $user_id AS Uint64;
					DECLARE $token AS String;
					DECLARE $status AS Uint8;
					UPDATE users SET bonus_coins = COALESCE( bonus_coins, 0 ) - 1, count_vip = COALESCE( count_vip, 0 ) + 1, token = $token WHERE user_id=$user_id;
					UPSERT INTO works (user_id, time, status) VALUES ($user_id, CurrentUtcTimestamp(), $status);`, Database)

			case user.Coins > 0:
				count = user.Coins - 1
				workStatus = 1
				updateUserQuery = fmt.Sprintf(`
					PRAGMA TablePathPrefix("%s");
					DECLARE $user_id AS Uint64;
					DECLARE $token AS String;
					DECLARE $status AS Uint8;
					UPDATE users SET coins = COALESCE( coins, 0 ) - 1, count_vip = COALESCE( count_vip, 0 ) + 1, token = $token WHERE user_id=$user_id;
					UPSERT INTO works (user_id, time, status) VALUES ($user_id, CurrentUtcTimestamp(), $status);`, Database)

			default:
				query := fmt.Sprintf(`
					PRAGMA TablePathPrefix("%s");
					DECLARE $user_id AS Uint64;
					$time = DateTime::MakeTimestamp(CurrentUtcDate());
					SELECT
					COUNT_IF(user_id = $user_id AND status = 0 AND time > $time) AS count FROM works;`, Database)

				var workCount uint64
				resC, err := t.Execute(ctx,
					query,
					table.NewQueryParameters(
						table.ValueParam("$user_id", ydb.Uint64Value(userID)),
					),
				)
				if err != nil || resC.Err() != nil {
					return err
				}
				if resC.NextSet() && resC.NextRow() && resC.NextItem() {
					workCount = res.Uint64()
				}
				count = 3 - int(workCount) - 1

				if workCount < 3 {
					updateUserQuery = fmt.Sprintf(`
					PRAGMA TablePathPrefix("%s");
					DECLARE $user_id AS Uint64;
					DECLARE $token AS String;
					DECLARE $status AS Uint8;
					UPDATE users SET count_free = COALESCE( count_free, 0 ) + 1, token = $token WHERE user_id=$user_id;
					UPSERT INTO works (user_id, time, status) VALUES ($user_id, CurrentUtcTimestamp(), $status);`, Database)
				}
			}
			_, err = t.Execute(
				ctx,
				updateUserQuery,
				table.NewQueryParameters(
					table.ValueParam("$user_id", ydb.Uint64Value(userID)),
					table.ValueParam("$token", ydb.StringValue([]byte(token))),
					table.ValueParam("$status", ydb.Uint8Value(workStatus)),
				))
			if err != nil {
				return err
			}
			err = t.Commit(ctx)
			return err
		}))

	return workStatus, count, err
}
