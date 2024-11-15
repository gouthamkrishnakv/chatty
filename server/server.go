package server

import (
	"context"
	"io"
	"log"
	"sync"
	"time"

	"github.com/gouthamkrishnakv/chatty/database"
	"github.com/gouthamkrishnakv/chatty/database/models"
	pb "github.com/gouthamkrishnakv/chatty/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type User struct {
	UserID   uint32
	Nickname string
}

type Message struct {
	MessageID uint32
	Author    User
	Message   string
	CreatedAt time.Time
}

type Server struct {
	pb.UnimplementedChatServiceServer
	connMap sync.Map
}

// Message limits
const PreviousMessageLimits = 100

func NewServer() *Server {
	return &Server{
		connMap: sync.Map{},
	}
}

func (s *Server) Join(ctx context.Context, joinReq *pb.JoinRequest) (*pb.JoinResponse, error) {
	db := database.GetDB()
	userEntity := &models.User{Nickname: joinReq.Nickname}
	if dbErr := db.WithContext(ctx).Create(userEntity); dbErr.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", dbErr.Error)
	}
	return &pb.JoinResponse{
		User: &pb.User{
			UserID:   userEntity.ID,
			Nickname: userEntity.Nickname,
		},
	}, nil
}

func (s *Server) StreamMessage(user *pb.User, stream pb.ChatService_StreamMessageServer) error {
	// make sure user exists
	userEntity := &models.User{ID: user.UserID}
	if txn := database.GetDB().WithContext(stream.Context()).First(userEntity, user.UserID); txn.Error != nil {
		return status.Errorf(codes.NotFound, "user not found: %v", txn.Error)
	}
	if userEntity.Nickname != user.Nickname {
		return status.Errorf(codes.InvalidArgument, "nickname mismatch")
	}
	if _, ok := s.connMap.Load(user.UserID); ok {
		return status.Errorf(codes.AlreadyExists, "user already connected")
	}
	// send out previous messages in DB for persistence
	var previousMessages []models.Message
	if txn := database.GetDB().
		WithContext(stream.Context()).
		InnerJoins("Author").
		Where("author_id != ?", user.UserID).
		Find(&previousMessages); txn.Error != nil {
		return status.Errorf(codes.Internal, "failed to fetch previous messages: %v", txn.Error)
	}
	for _, msg := range previousMessages {
		stream.Send(&pb.MessageObj{
			MessageID: msg.ID,
			Author: &pb.User{
				UserID:   msg.Author.ID,
				Nickname: msg.Author.Nickname,
			},
			Message: msg.Message,
		})
	}
	log.Printf("client connected: %v", user)
	msgChannel := make(chan *Message, 100)
	s.connMap.Store(user.UserID, msgChannel)
	for {
		select {
		case <-stream.Context().Done():
			log.Printf("client disconnected: %v", user)
			s.connMap.Delete(userEntity.ID)
			close(msgChannel)
			return nil
		case msg := <-msgChannel:
			if msg == nil {
				log.Printf("close requested: %v", user)
				s.connMap.Delete(userEntity.ID)
				return nil
			} else if sendErr := stream.Send(&pb.MessageObj{
				MessageID: msg.MessageID,
				Message:   msg.Message,
				Author: &pb.User{
					UserID:   msg.Author.UserID,
					Nickname: msg.Author.Nickname,
				}}); sendErr != nil {
				// if client disconnects, sendErr will be io.EOF
				if sendErr == io.EOF {
					s.connMap.Delete(user.UserID)
					close(msgChannel)
					return nil
				}
				return status.Errorf(codes.Internal, "failed to send message to user %v:  %v", user, sendErr)
			}
		case <-time.After(5 * time.Second):
			continue
		}
	}
}

func (s *Server) Send(ctx context.Context, message *pb.MessageObj) (*pb.MessageObj, error) {
	var userEntity models.User
	if txn := database.GetDB().WithContext(ctx).First(&userEntity, message.Author.UserID); txn.Error != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", txn.Error)
	} else if userEntity.Nickname != message.Author.Nickname {
		return nil, status.Errorf(codes.InvalidArgument, "nickname mismatch")
	}
	messageEntity := &models.Message{
		Message:  message.Message,
		AuthorID: message.Author.UserID,
	}
	if txn := database.GetDB().WithContext(ctx).Create(messageEntity); txn.Error != nil {
		return nil, status.Errorf(codes.Internal, "failed to create message: %v", txn.Error)
	}
	s.connMap.Range(func(key, value any) bool {
		userID, _ := key.(uint32)
		msgChannel, ok := value.(chan *Message)
		if !ok {
			return false
		}
		if userID == message.Author.UserID {
			return true
		}
		msgChannel <- &Message{
			MessageID: messageEntity.ID,
			Author: User{
				UserID:   message.Author.UserID,
				Nickname: message.Author.Nickname,
			},
			Message:   message.Message,
			CreatedAt: messageEntity.CreatedAt,
		}
		return true
	})
	message.MessageID = messageEntity.ID
	return message, nil
}

func (s *Server) Close(_ context.Context, closeReq *pb.CloseRequest) (*pb.CloseResponse, error) {
	rawMsgChan, exists := s.connMap.LoadAndDelete(closeReq.UserID)
	if !exists {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", closeReq.UserID)
	}
	msgChan, ok := rawMsgChan.(chan *Message)
	if !ok {
		return nil, status.Errorf(codes.Internal, "invalid message channel")
	}
	msgChan <- nil
	close(msgChan)
	return &pb.CloseResponse{Status: pb.StatusCode_Success}, nil
}
