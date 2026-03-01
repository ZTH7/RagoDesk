package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/ZTH7/RagoDesk/apps/server/internal/conf"
	"github.com/ZTH7/RagoDesk/apps/server/internal/data"
	knowledgebiz "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/biz"
	knowledgedata "github.com/ZTH7/RagoDesk/apps/server/internal/knowledge/data"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"

	_ "go.uber.org/automaxprocs"
)

// go build -ldflags "-X main.Name=ragodesk-ingester -X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name = "ragodesk-ingester"
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string

	id, _ = os.Hostname()
)

func init() {
	flag.StringVar(&flagconf, "conf", "../../configs", "config path, eg: -conf config.yaml")
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
	helper := log.NewHelper(logger)

	sources, err := loadConfigSources(flagconf)
	if err != nil {
		panic(err)
	}
	c := config.New(config.WithSource(sources...))
	defer c.Close()

	if err := c.Load(); err != nil {
		panic(err)
	}

	var bc conf.Bootstrap
	if err := c.Scan(&bc); err != nil {
		panic(err)
	}

	if os.Getenv("RAGODESK_INGESTION_ASYNC") == "" {
		_ = os.Setenv("RAGODESK_INGESTION_ASYNC", "1")
	}

	dataData, cleanup, err := data.NewData(bc.Data)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	queue := knowledgedata.NewIngestionQueue(bc.Data, logger)
	if queue == nil {
		helper.Warn("ingestion queue disabled or not configured")
		return
	}
	defer func() {
		if err := queue.Close(); err != nil {
			helper.Warnf("ingestion queue close failed: %v", err)
		}
	}()

	repo := knowledgedata.NewKnowledgeRepo(dataData, bc.Data, logger)
	uc := knowledgebiz.NewKnowledgeUsecase(repo, queue, bc.Data, logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := uc.StartIngestionConsumer(ctx); err != nil {
		helper.Warnf("ingestion consumer start failed: %v", err)
		return
	}

	helper.Info("ingestion worker started")
	<-ctx.Done()
	helper.Info("ingestion worker stopped")
}

func loadConfigSources(confPath string) ([]config.Source, error) {
	info, err := os.Stat(confPath)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []config.Source{file.NewSource(confPath)}, nil
	}
	entries, err := os.ReadDir(confPath)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}
		files = append(files, filepath.Join(confPath, name))
	}
	sort.Slice(files, func(i, j int) bool {
		pi := configPriority(files[i])
		pj := configPriority(files[j])
		if pi == pj {
			return strings.ToLower(filepath.Base(files[i])) < strings.ToLower(filepath.Base(files[j]))
		}
		return pi < pj
	})
	sources := make([]config.Source, 0, len(files))
	for _, f := range files {
		sources = append(sources, file.NewSource(f))
	}
	return sources, nil
}

func configPriority(path string) int {
	base := strings.ToLower(filepath.Base(path))
	switch {
	case strings.Contains(base, ".local."):
		return 2
	case base == "config.yaml" || base == "config.yml":
		return 0
	default:
		return 1
	}
}
