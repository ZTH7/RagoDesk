package service

import (
	"context"
	"time"

	v1 "github.com/ZTH7/RAGDesk/apps/server/api/knowledge/v1"
	iambiz "github.com/ZTH7/RAGDesk/apps/server/internal/iam/biz"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/knowledge/biz"
	"github.com/ZTH7/RAGDesk/apps/server/internal/tenant"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// KnowledgeService handles knowledge service layer.
type KnowledgeService struct {
	v1.UnimplementedKnowledgeAdminServer

	uc    *biz.KnowledgeUsecase
	iamUC *iambiz.IAMUsecase
	log   *log.Helper
}

// NewKnowledgeService creates a new KnowledgeService
func NewKnowledgeService(uc *biz.KnowledgeUsecase, iamUC *iambiz.IAMUsecase, logger log.Logger) *KnowledgeService {
	return &KnowledgeService{uc: uc, iamUC: iamUC, log: log.NewHelper(logger)}
}

func (s *KnowledgeService) CreateKnowledgeBase(ctx context.Context, req *v1.CreateKnowledgeBaseRequest) (*v1.KnowledgeBaseResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionKnowledgeBaseWrite); err != nil {
		return nil, err
	}
	created, err := s.uc.CreateKnowledgeBase(ctx, biz.KnowledgeBase{
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.KnowledgeBaseResponse{KnowledgeBase: toKnowledgeBase(created)}, nil
}

func (s *KnowledgeService) GetKnowledgeBase(ctx context.Context, req *v1.GetKnowledgeBaseRequest) (*v1.KnowledgeBaseResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionKnowledgeBaseRead); err != nil {
		return nil, err
	}
	kb, err := s.uc.GetKnowledgeBase(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	return &v1.KnowledgeBaseResponse{KnowledgeBase: toKnowledgeBase(kb)}, nil
}

func (s *KnowledgeService) UpdateKnowledgeBase(ctx context.Context, req *v1.UpdateKnowledgeBaseRequest) (*v1.KnowledgeBaseResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionKnowledgeBaseWrite); err != nil {
		return nil, err
	}
	updated, err := s.uc.UpdateKnowledgeBase(ctx, biz.KnowledgeBase{
		ID:          req.GetId(),
		Name:        req.GetName(),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, err
	}
	return &v1.KnowledgeBaseResponse{KnowledgeBase: toKnowledgeBase(updated)}, nil
}

func (s *KnowledgeService) DeleteKnowledgeBase(ctx context.Context, req *v1.DeleteKnowledgeBaseRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionKnowledgeBaseWrite); err != nil {
		return nil, err
	}
	if err := s.uc.DeleteKnowledgeBase(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *KnowledgeService) ListKnowledgeBases(ctx context.Context, req *v1.ListKnowledgeBasesRequest) (*v1.ListKnowledgeBasesResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionKnowledgeBaseRead); err != nil {
		return nil, err
	}
	items, err := s.uc.ListKnowledgeBases(ctx)
	if err != nil {
		return nil, err
	}
	resp := &v1.ListKnowledgeBasesResponse{Items: make([]*v1.KnowledgeBase, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toKnowledgeBase(item))
	}
	return resp, nil
}

func (s *KnowledgeService) ListDocuments(ctx context.Context, req *v1.ListDocumentsRequest) (*v1.ListDocumentsResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionDocumentRead); err != nil {
		return nil, err
	}
	items, err := s.uc.ListDocuments(ctx, req.GetKbId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		return nil, err
	}
	resp := &v1.ListDocumentsResponse{Items: make([]*v1.Document, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toDocument(item))
	}
	return resp, nil
}

func (s *KnowledgeService) ListBotKnowledgeBases(ctx context.Context, req *v1.ListBotKnowledgeBasesRequest) (*v1.ListBotKnowledgeBasesResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionBotRead); err != nil {
		return nil, err
	}
	items, err := s.uc.ListBotKnowledgeBases(ctx, req.GetBotId())
	if err != nil {
		return nil, err
	}
	resp := &v1.ListBotKnowledgeBasesResponse{Items: make([]*v1.BotKnowledgeBase, 0, len(items))}
	for _, item := range items {
		resp.Items = append(resp.Items, toBotKnowledgeBase(item))
	}
	return resp, nil
}

func (s *KnowledgeService) BindBotKnowledgeBase(ctx context.Context, req *v1.BindBotKnowledgeBaseRequest) (*v1.BotKnowledgeBaseResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionBotKBBind); err != nil {
		return nil, err
	}
	linked, err := s.uc.BindBotKnowledgeBase(ctx, req.GetBotId(), req.GetKbId(), req.GetPriority(), req.GetWeight())
	if err != nil {
		return nil, err
	}
	return &v1.BotKnowledgeBaseResponse{BotKb: toBotKnowledgeBase(linked)}, nil
}

func (s *KnowledgeService) UnbindBotKnowledgeBase(ctx context.Context, req *v1.UnbindBotKnowledgeBaseRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionBotKBUnbind); err != nil {
		return nil, err
	}
	if err := s.uc.UnbindBotKnowledgeBase(ctx, req.GetBotId(), req.GetKbId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *KnowledgeService) UploadDocument(ctx context.Context, req *v1.UploadDocumentRequest) (*v1.UploadDocumentResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionDocumentUpload); err != nil {
		return nil, err
	}
	doc, ver, err := s.uc.UploadDocument(ctx, req.GetKbId(), req.GetTitle(), req.GetSourceType(), req.GetRawUri())
	if err != nil {
		return nil, err
	}
	return &v1.UploadDocumentResponse{
		Document: toDocument(doc),
		Version:  toDocumentVersion(ver),
	}, nil
}

func (s *KnowledgeService) GetDocument(ctx context.Context, req *v1.GetDocumentRequest) (*v1.GetDocumentResponse, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionDocumentRead); err != nil {
		return nil, err
	}
	doc, versions, err := s.uc.GetDocument(ctx, req.GetId())
	if err != nil {
		return nil, err
	}
	resp := &v1.GetDocumentResponse{
		Document: toDocument(doc),
		Versions: make([]*v1.DocumentVersion, 0, len(versions)),
	}
	for _, item := range versions {
		resp.Versions = append(resp.Versions, toDocumentVersion(item))
	}
	return resp, nil
}

func (s *KnowledgeService) DeleteDocument(ctx context.Context, req *v1.DeleteDocumentRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionDocumentDelete); err != nil {
		return nil, err
	}
	if err := s.uc.DeleteDocument(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *KnowledgeService) ReindexDocument(ctx context.Context, req *v1.ReindexDocumentRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionDocumentReindex); err != nil {
		return nil, err
	}
	if _, err := s.uc.ReindexDocument(ctx, req.GetId()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *KnowledgeService) RollbackDocument(ctx context.Context, req *v1.RollbackDocumentRequest) (*emptypb.Empty, error) {
	if err := requireTenantContext(ctx); err != nil {
		return nil, err
	}
	if err := s.iamUC.RequirePermission(ctx, biz.PermissionDocumentRollback); err != nil {
		return nil, err
	}
	if err := s.uc.RollbackDocument(ctx, req.GetId(), req.GetVersion()); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

// ProviderSet is knowledge service providers.
var ProviderSet = wire.NewSet(NewKnowledgeService)

func requireTenantContext(ctx context.Context) error {
	if _, err := tenant.RequireTenantID(ctx); err != nil {
		return errors.Forbidden("TENANT_MISSING", "tenant missing")
	}
	return nil
}

func toTimestamp(value time.Time) *timestamppb.Timestamp {
	if value.IsZero() {
		return nil
	}
	return timestamppb.New(value)
}

func toKnowledgeBase(kb biz.KnowledgeBase) *v1.KnowledgeBase {
	if kb.ID == "" && kb.Name == "" {
		return nil
	}
	return &v1.KnowledgeBase{
		Id:          kb.ID,
		TenantId:    kb.TenantID,
		Name:        kb.Name,
		Description: kb.Description,
		CreatedAt:   toTimestamp(kb.CreatedAt),
		UpdatedAt:   toTimestamp(kb.UpdatedAt),
	}
}

func toBotKnowledgeBase(link biz.BotKnowledgeBase) *v1.BotKnowledgeBase {
	if link.ID == "" && link.BotID == "" && link.KBID == "" {
		return nil
	}
	return &v1.BotKnowledgeBase{
		Id:        link.ID,
		TenantId:  link.TenantID,
		BotId:     link.BotID,
		KbId:      link.KBID,
		Priority:  link.Priority,
		Weight:    link.Weight,
		CreatedAt: toTimestamp(link.CreatedAt),
	}
}

func toDocument(doc biz.Document) *v1.Document {
	if doc.ID == "" && doc.Title == "" {
		return nil
	}
	return &v1.Document{
		Id:             doc.ID,
		TenantId:       doc.TenantID,
		KbId:           doc.KBID,
		Title:          doc.Title,
		SourceType:     doc.SourceType,
		Status:         doc.Status,
		CurrentVersion: doc.CurrentVersion,
		CreatedAt:      toTimestamp(doc.CreatedAt),
		UpdatedAt:      toTimestamp(doc.UpdatedAt),
	}
}

func toDocumentVersion(v biz.DocumentVersion) *v1.DocumentVersion {
	if v.ID == "" && v.DocumentID == "" {
		return nil
	}
	return &v1.DocumentVersion{
		Id:         v.ID,
		TenantId:   v.TenantID,
		DocumentId: v.DocumentID,
		Version:    v.Version,
		Status:     v.Status,
		CreatedAt:  toTimestamp(v.CreatedAt),
	}
}
