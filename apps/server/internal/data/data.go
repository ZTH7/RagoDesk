package data

import (
	"github.com/ZTH7/RAGDesk/apps/server/internal/conf"
	analyticsdata "github.com/ZTH7/RAGDesk/apps/server/internal/analytics/data"
	apimgmtdata "github.com/ZTH7/RAGDesk/apps/server/internal/apimgmt/data"
	conversationdata "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/data"
	iamdata "github.com/ZTH7/RAGDesk/apps/server/internal/iam/data"
	knowledgedata "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/data"
	platformdata "github.com/ZTH7/RAGDesk/apps/server/internal/platform/data"
	ragdata "github.com/ZTH7/RAGDesk/apps/server/internal/rag/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	analyticsdata.ProviderSet,
	apimgmtdata.ProviderSet,
	conversationdata.ProviderSet,
	iamdata.ProviderSet,
	knowledgedata.ProviderSet,
	platformdata.ProviderSet,
	ragdata.ProviderSet,
)

// Data .
type Data struct {
	// TODO wrapped database client
}

// NewData .
func NewData(c *conf.Data) (*Data, func(), error) {
	cleanup := func() {
		log.Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}
