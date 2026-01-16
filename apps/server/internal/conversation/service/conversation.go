package service

import (
	"context"
	"time"

	v1 "github.com/ZTH7/RAGDesk/apps/server/api/conversation/v1"
	biz "github.com/ZTH7/RAGDesk/apps/server/internal/conversation/biz"
	"github.com/go-kratos/kratos/v2/errors"
	"github.com/google/wire"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ConversationService handles conversation service layer.
type ConversationService struct {
	v1.UnimplementedConversationServer
	v1.UnimplementedConsoleConversationServer

	uc *biz.ConversationUsecase
}

// NewConversationService creates a new ConversationService
func NewConversationService(uc *biz.ConversationUsecase) *ConversationService {
	return &ConversationService{uc: uc}
}

func (s *ConversationService) CreateSession(ctx context.Context, req *v1.CreateSessionRequest) (*v1.CreateSessionResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	meta := structToMap(req.Metadata)
	session, err := s.uc.CreateSession(ctx, req.BotId, req.UserExternalId, meta)
	if err != nil {
		return nil, err
	}
	return &v1.CreateSessionResponse{Session: toAPISession(session)}, nil
}

func (s *ConversationService) GetSession(ctx context.Context, req *v1.GetSessionRequest) (*v1.GetSessionResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	session, messages, err := s.uc.GetSession(ctx, req.SessionId, req.IncludeMessages, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	resp := &v1.GetSessionResponse{
		Session:  toAPISession(session),
		Messages: toAPIMessages(messages),
	}
	return resp, nil
}

func (s *ConversationService) CloseSession(ctx context.Context, req *v1.CloseSessionRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := s.uc.CloseSession(ctx, req.SessionId, req.CloseReason); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ConversationService) CreateFeedback(ctx context.Context, req *v1.CreateFeedbackRequest) (*emptypb.Empty, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	if err := s.uc.CreateFeedback(ctx, req.SessionId, req.MessageId, req.Rating, req.Comment, req.Correction); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

func (s *ConversationService) ListSessions(ctx context.Context, req *v1.ListSessionsRequest) (*v1.ListSessionsResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	sessions, err := s.uc.ListSessions(ctx, req.BotId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	out := make([]*v1.Session, 0, len(sessions))
	for _, item := range sessions {
		out = append(out, toAPISession(item))
	}
	return &v1.ListSessionsResponse{Sessions: out}, nil
}

func (s *ConversationService) ListMessages(ctx context.Context, req *v1.ListMessagesRequest) (*v1.ListMessagesResponse, error) {
	if req == nil {
		return nil, errors.BadRequest("REQUEST_EMPTY", "request empty")
	}
	messages, err := s.uc.ListMessages(ctx, req.SessionId, int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, err
	}
	return &v1.ListMessagesResponse{Messages: toAPIMessages(messages)}, nil
}

// ProviderSet is conversation service providers.
var ProviderSet = wire.NewSet(NewConversationService)

func toAPISession(session biz.Session) *v1.Session {
	if session.ID == "" {
		return nil
	}
	return &v1.Session{
		Id:            session.ID,
		TenantId:      session.TenantID,
		BotId:         session.BotID,
		Status:        session.Status,
		CloseReason:   session.CloseReason,
		UserExternalId: session.UserExternal,
		Metadata:      session.Metadata,
		CreatedAt:     timestamppb.New(session.CreatedAt),
		UpdatedAt:     timestamppb.New(session.UpdatedAt),
		ClosedAt:      timeOrNil(session.ClosedAt),
	}
}

func toAPIMessages(messages []biz.Message) []*v1.Message {
	if len(messages) == 0 {
		return nil
	}
	out := make([]*v1.Message, 0, len(messages))
	for _, item := range messages {
		out = append(out, &v1.Message{
			Id:         item.ID,
			SessionId:  item.SessionID,
			Role:       item.Role,
			Content:    item.Content,
			Confidence: item.Confidence,
			References: toAPIReferences(item.References),
			CreatedAt:  timestamppb.New(item.CreatedAt),
		})
	}
	return out
}

func toAPIReferences(raw string) []*v1.Reference {
	refs := biz.DecodeReferences(raw)
	if len(refs) == 0 {
		return nil
	}
	out := make([]*v1.Reference, 0, len(refs))
	for _, ref := range refs {
		out = append(out, &v1.Reference{
			DocumentId:        ref.DocumentID,
			DocumentVersionId: ref.DocumentVersionID,
			ChunkId:           ref.ChunkID,
			Score:             ref.Score,
			Rank:              ref.Rank,
			Snippet:           ref.Snippet,
		})
	}
	return out
}

func structToMap(value *structpb.Struct) map[string]any {
	if value == nil {
		return nil
	}
	return value.AsMap()
}

func timeOrNil(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}
