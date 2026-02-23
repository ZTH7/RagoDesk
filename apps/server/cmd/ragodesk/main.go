package main

import (
	"context"
	"flag"
	"os"

	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	knowledgebiz "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/biz"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Name=ragodesk -X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "ragodesk"
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
}

func newApp(logger log.Logger, gs *grpc.Server, hs *http.Server, knowledgeUC *knowledgebiz.KnowledgeUsecase) *kratos.App {
	options := []kratos.Option{
		kratos.ID(id),
		kratos.Name(Name),
		kratos.Version(Version),
		kratos.Metadata(map[string]string{}),
		kratos.Logger(logger),
		kratos.Server(
			gs,
			hs,
		),
	}
	if knowledgeUC != nil && knowledgeUC.AsyncEnabled() {
		helper := log.NewHelper(logger)
		options = append(options,
			kratos.AfterStart(func(ctx context.Context) error {
				if err := knowledgeUC.StartIngestionConsumer(ctx); err != nil {
					helper.Warnf("ingestion consumer start failed: %v", err)
				} else {
					helper.Info("ingestion consumer started")
				}
				return nil
			}),
			kratos.BeforeStop(func(ctx context.Context) error {
				knowledgeUC.CloseIngestionQueue()
				helper.Info("ingestion consumer stopped")
				return nil
			}),
		)
	}
	return kratos.New(options...)
}

func main() {
	flag.Parse()
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", id,
		"service.name", Name,
		"service.version", Version,
		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)
	c := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	app, cleanup, err := wireApp(bc.Server, bc.Data, logger)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		panic(err)
	}
}
