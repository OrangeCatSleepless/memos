package mysql

import (
	"context"
	"database/sql"

	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/usememos/memos/internal/profile"
	"github.com/usememos/memos/store"
)

type DB struct {
	db      *sql.DB
	profile *profile.Profile
	config  *mysql.Config
}

func NewDB(profile *profile.Profile) (store.Driver, error) {
	// Open MySQL connection with parameter.
	// multiStatements=true is required for migration.
	// See more in: https://github.com/go-sql-driver/mysql#multistatements
	dsn, err := mergeDSN(profile.DSN)
	if err != nil {
		return nil, err
	}

	driver := DB{profile: profile}
	driver.config, err = mysql.ParseDSN(dsn)
	if err != nil {
		return nil, errors.New("Parse DSN eroor")
	}

	driver.db, err = sql.Open("mysql", dsn)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open db: %s", profile.DSN)
	}

	return &driver, nil
}

func (d *DB) GetDB() *sql.DB {
	return d.db
}

func (d *DB) Close() error {
	return d.db.Close()
}

func (d *DB) IsInitialized(ctx context.Context) (bool, error) {
	var exists bool
	err := d.db.QueryRowContext(ctx, "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE TABLE_NAME = 'memo' AND TABLE_TYPE = 'BASE TABLE')").Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "failed to check if database is initialized")
	}
	return exists, nil
}

func mergeDSN(baseDSN string) (string, error) {
	config, err := mysql.ParseDSN(baseDSN)
	if err != nil {
		return "", errors.Wrapf(err, "failed to parse DSN: %s", baseDSN)
	}

	config.MultiStatements = true
	return config.FormatDSN(), nil
}
