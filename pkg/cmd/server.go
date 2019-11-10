package cmd

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	_ "github.com/go-sql-driver/mysql"

	"github.com/josenpai/go-grpc-http-rest-microservice-tutorial/pkg/logger"
	"github.com/josenpai/go-grpc-http-rest-microservice-tutorial/pkg/protocol/grpc"
	"github.com/josenpai/go-grpc-http-rest-microservice-tutorial/pkg/protocol/rest"
	v1 "github.com/josenpai/go-grpc-http-rest-microservice-tutorial/pkg/service/v1"
)

type Config struct {
	// gRPC server start parameters section
	// gRPC is TCP
	GRPCPort string

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort string

	// DB Datastore parameters section
	// DatastoreDBHost is host of database
	DatastoreDBHost string

	// DatastoreDBUser string
	DatastoreDBUser string

	//DatastoreDBPassword password to connect to database
	DatastoreDBPassword string

	DatstoreDBSchema string

	// Log parameters section
	// LogLevel is a global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel int
	// LogTimeFormat is print time format for logger e.g. 2006-01-02T15:04:05Z07:00
	LogTimeFormat string
}

// RunServer runs gRPC server and HTTP gateway
func RunServer() error {
	ctx := context.Background()

	//get Configuration
	var cfg Config
	flag.StringVar(&cfg.GRPCPort, "grpc-port", "", "gRPC port to bind")
	flag.StringVar(&cfg.HTTPPort, "http-port", "", "HTTP port to bind")
	flag.StringVar(&cfg.DatastoreDBHost, "db-host", "", "Database host")
	flag.StringVar(&cfg.DatastoreDBUser, "db-user", "", "Database user")
	flag.StringVar(&cfg.DatastoreDBPassword, "db-password", "", "Database password")
	flag.StringVar(&cfg.DatstoreDBSchema, "db-schema", "", "Database schema")
	flag.IntVar(&cfg.LogLevel, "log-level", 0, "Global log level")
	flag.StringVar(&cfg.LogTimeFormat, "log-time-format", "", "Print time format for logger e.g. 2006-01-02T15:04:05Z07:00")
	flag.Parse()

	if len(cfg.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC server: '%s'", cfg.GRPCPort)
	}

	if len(cfg.HTTPPort) == 0 {
		return fmt.Errorf("invalid TCP port for HTTP gateway: '%s'", cfg.HTTPPort)
	}

	// initialize logger
	if err := logger.Init(cfg.LogLevel, cfg.LogTimeFormat); err != nil {

	}

	// add MySQL driver specific parameter to parse data/time
	// Drop it for another database
	param := "parseTime=true"

	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?%s",
		cfg.DatastoreDBUser,
		cfg.DatastoreDBPassword,
		cfg.DatastoreDBHost,
		cfg.DatstoreDBSchema,
		param)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	v1API := v1.NewToDoServiceServer(db)

	// run the HTTP gateway
	go func() {
		_ = rest.RunServer(ctx, cfg.GRPCPort, cfg.HTTPPort)
	}()

	return grpc.RunServer(ctx, v1API, cfg.GRPCPort)
}
