package dnb_ydb

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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
