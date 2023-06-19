package config

import "github.com/jackc/pgx"

var PostgresConf = pgx.ConnConfig{
	Host:                 "localhost",
	Port:                 5432,
	Database:             "forumservice",
	User:                 "postgres",
	Password:             "12345",
	TLSConfig:            nil,
	UseFallbackTLS:       false,
	FallbackTLSConfig:    nil,
	Logger:               nil,
	LogLevel:             0,
	Dial:                 nil,
	RuntimeParams:        nil,
	OnNotice:             nil,
	CustomConnInfo:       nil,
	CustomCancel:         nil,
	PreferSimpleProtocol: false,
	TargetSessionAttrs:   "",
}

var PostgresPoolCong = pgx.ConnPoolConfig{
	ConnConfig:     PostgresConf,
	MaxConnections: 10,
	AfterConnect:   nil,
	AcquireTimeout: 0,
}
