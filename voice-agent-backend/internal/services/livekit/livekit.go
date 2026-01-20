package livekit

import (
	"context"
	"fmt"
	"time"

	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/voice-agent/backend/internal/config"
)

// Service handles LiveKit operations
type Service struct {
	url        string
	apiKey     string
	apiSecret  string
	roomClient *lksdk.RoomServiceClient
}

// NewService creates a new LiveKit service
func NewService(cfg *config.Config) *Service {
	roomClient := lksdk.NewRoomServiceClient(cfg.LiveKitURL, cfg.LiveKitAPIKey, cfg.LiveKitAPISecret)

	return &Service{
		url:        cfg.LiveKitURL,
		apiKey:     cfg.LiveKitAPIKey,
		apiSecret:  cfg.LiveKitAPISecret,
		roomClient: roomClient,
	}
}

// CreateRoom creates a new LiveKit room
func (s *Service) CreateRoom(ctx context.Context, roomName string) (*livekit.Room, error) {
	room, err := s.roomClient.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    300, // 5 minutes
		MaxParticipants: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create room: %w", err)
	}
	return room, nil
}

// DeleteRoom deletes a LiveKit room
func (s *Service) DeleteRoom(ctx context.Context, roomName string) error {
	_, err := s.roomClient.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomName,
	})
	return err
}

// ListRooms lists all active rooms
func (s *Service) ListRooms(ctx context.Context) ([]*livekit.Room, error) {
	resp, err := s.roomClient.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, fmt.Errorf("failed to list rooms: %w", err)
	}
	return resp.Rooms, nil
}

// GenerateToken generates an access token for a participant
func (s *Service) GenerateToken(roomName, participantName string, isAgent bool) (string, error) {
	at := auth.NewAccessToken(s.apiKey, s.apiSecret)

	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     roomName,
	}

	// Agents get additional permissions
	if isAgent {
		grant.RoomAdmin = true
		grant.CanPublish = boolPtr(true)
		grant.CanSubscribe = boolPtr(true)
		grant.CanPublishData = boolPtr(true)
	} else {
		grant.CanPublish = boolPtr(true)
		grant.CanSubscribe = boolPtr(true)
		grant.CanPublishData = boolPtr(true)
	}

	at.AddGrant(grant).
		SetIdentity(participantName).
		SetValidFor(24 * time.Hour)

	return at.ToJWT()
}

// GetParticipants gets participants in a room
func (s *Service) GetParticipants(ctx context.Context, roomName string) ([]*livekit.ParticipantInfo, error) {
	resp, err := s.roomClient.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: roomName,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list participants: %w", err)
	}
	return resp.Participants, nil
}

// RemoveParticipant removes a participant from a room
func (s *Service) RemoveParticipant(ctx context.Context, roomName, participantID string) error {
	_, err := s.roomClient.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
		Room:     roomName,
		Identity: participantID,
	})
	return err
}

// SendData sends data to participants in a room
func (s *Service) SendData(ctx context.Context, roomName string, data []byte, destinationIdentities []string) error {
	_, err := s.roomClient.SendData(ctx, &livekit.SendDataRequest{
		Room:                  roomName,
		Data:                  data,
		Kind:                  livekit.DataPacket_RELIABLE,
		DestinationIdentities: destinationIdentities,
	})
	return err
}

// GetURL returns the LiveKit server URL
func (s *Service) GetURL() string {
	return s.url
}

func boolPtr(b bool) *bool {
	return &b
}
