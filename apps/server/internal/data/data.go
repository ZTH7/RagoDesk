package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"

	_ "github.com/go-sql-driver/mysql"
)

// Data .
type Data struct {
	DB *sql.DB
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	if c == nil || c.Database == nil || c.Database.Driver == "" || c.Database.Source == "" {
		return nil, nil, errors.InternalServer("DB_CONFIG_MISSING", "database config missing")
	}
	db, err := sql.Open(c.Database.Driver, c.Database.Source)
	if err != nil {
		return nil, nil, err
	}
	db.SetConnMaxLifetime(1 * time.Hour)
	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(50)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, nil, err
	}
	cleanup := func() {
		log.Info("closing the data resources")
		if err := db.Close(); err != nil {
			log.Errorf("close database error: %v", err)
		}
	}
	return &Data{DB: db}, cleanup, nil
}
