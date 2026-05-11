package handlers

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Nawaf-Almansour/diwan-server/internal/models"
	"github.com/Nawaf-Almansour/diwan-server/internal/ws"
)

type MeetingHandler struct {
	db         *pgxpool.Pool
	hub        *ws.Hub
	pnmHost    string
	pnmKey     string
	pnmSecret  string
	clientHost string
}

func NewMeetingHandler(db *pgxpool.Pool, hub *ws.Hub, pnmHost, pnmKey, pnmSecret, clientHost string) *MeetingHandler {
	return &MeetingHandler{
		db:         db,
		hub:        hub,
		pnmHost:    pnmHost,
		pnmKey:     pnmKey,
		pnmSecret:  pnmSecret,
		clientHost: clientHost,
	}
}

type StartMeetingRequest struct {
	Title string `json:"title"`
}

type pnmCreateRoomReq struct {
	RoomID   string      `json:"room_id"`
	Metadata pnmMetadata `json:"metadata"`
}

type pnmMetadata struct {
	RoomTitle      string          `json:"room_title"`
	WelcomeMessage string          `json:"welcome_message"`
	RoomFeatures   pnmRoomFeatures `json:"room_features"`
}

type pnmRoomFeatures struct {
	AllowWebcams          bool                     `json:"allow_webcams"`
	MuteOnStart           bool                     `json:"mute_on_start"`
	AllowScreenShare      bool                     `json:"allow_screen_share"`
	AllowRecording        bool                     `json:"allow_recording"`
	AllowRTMP             bool                     `json:"allow_rtmp"`
	AdminOnlyWebcams      bool                     `json:"admin_only_webcams"`
	AllowViewOtherWebcams bool                     `json:"allow_view_other_webcams"`
	AllowViewOthersList   bool                     `json:"allow_view_other_users_list"`
	AllowPolls            bool                     `json:"allow_polls"`
	RoomDuration          int                      `json:"room_duration"`
	ChatFeatures          pnmChatFeatures          `json:"chat_features"`
	SharedNotePadFeatures pnmSharedNotePadFeatures `json:"shared_note_pad_features"`
	WhiteboardFeatures    pnmWhiteboardFeatures    `json:"whiteboard_features"`
	ExternalMediaPlayer   pnmExternalMediaPlayer   `json:"external_media_player_features"`
	WaitingRoomFeatures   pnmWaitingRoomFeatures   `json:"waiting_room_features"`
	BreakoutRoomFeatures  pnmBreakoutRoomFeatures  `json:"breakout_room_features"`
	DisplayExternalLink   pnmDisplayExternalLink   `json:"display_external_link_features"`
}

type pnmChatFeatures struct {
	IsAllow         bool `json:"is_allow"`
	AllowFileUpload bool `json:"allow_file_upload"`
}

type pnmSharedNotePadFeatures struct {
	IsAllow bool `json:"is_allow"`
}

type pnmWhiteboardFeatures struct {
	IsAllow bool `json:"is_allow"`
}

type pnmExternalMediaPlayer struct {
	IsAllow bool `json:"is_allow"`
}

type pnmWaitingRoomFeatures struct {
	IsAllow bool `json:"is_allow"`
}

type pnmBreakoutRoomFeatures struct {
	IsAllow bool `json:"is_allow"`
}

type pnmDisplayExternalLink struct {
	IsAllow bool `json:"is_allow"`
}

type pnmJoinTokenReq struct {
	RoomID   string      `json:"room_id"`
	UserInfo pnmUserInfo `json:"user_info"`
}

type pnmUserInfo struct {
	Name    string `json:"name"`
	UserID  string `json:"user_id"`
	IsAdmin bool   `json:"is_admin"`
}

func (h *MeetingHandler) StartMeeting(c *fiber.Ctx) error {
	channelID := c.Params("id")
	userID := c.Locals("user_id").(string)

	var req StartMeetingRequest
	if err := c.BodyParser(&req); err != nil {
		req.Title = ""
	}

	// Get user info
	var userName string
	_ = h.db.QueryRow(context.Background(),
		`SELECT display_name FROM users WHERE id = $1`, userID,
	).Scan(&userName)

	// Get channel info for room title
	var channelName string
	_ = h.db.QueryRow(context.Background(),
		`SELECT name FROM channels WHERE id = $1`, channelID,
	).Scan(&channelName)

	if req.Title == "" {
		req.Title = fmt.Sprintf("اجتماع - %s", channelName)
	}

	roomID := fmt.Sprintf("diwan-%s", channelID)

	// Create room in PlugNmeet with all features enabled
	createBody := pnmCreateRoomReq{
		RoomID: roomID,
		Metadata: pnmMetadata{
			RoomTitle:      req.Title,
			WelcomeMessage: fmt.Sprintf("مرحباً في اجتماع قناة #%s", channelName),
			RoomFeatures: pnmRoomFeatures{
				AllowWebcams:          true,
				MuteOnStart:           false,
				AllowScreenShare:      true,
				AllowRecording:        true,
				AllowRTMP:             true,
				AdminOnlyWebcams:      false,
				AllowViewOtherWebcams: true,
				AllowViewOthersList:   true,
				AllowPolls:            true,
				RoomDuration:          0,
				ChatFeatures:          pnmChatFeatures{IsAllow: true, AllowFileUpload: true},
				SharedNotePadFeatures: pnmSharedNotePadFeatures{IsAllow: true},
				WhiteboardFeatures:    pnmWhiteboardFeatures{IsAllow: true},
				ExternalMediaPlayer:   pnmExternalMediaPlayer{IsAllow: true},
				WaitingRoomFeatures:   pnmWaitingRoomFeatures{IsAllow: false},
				BreakoutRoomFeatures:  pnmBreakoutRoomFeatures{IsAllow: true},
				DisplayExternalLink:   pnmDisplayExternalLink{IsAllow: true},
			},
		},
	}

	_, err := h.callPlugNmeet("room/create", createBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("failed to create meeting: %v", err)})
	}

	// Generate join token for the user
	tokenBody := pnmJoinTokenReq{
		RoomID: roomID,
		UserInfo: pnmUserInfo{
			Name:    userName,
			UserID:  userID,
			IsAdmin: true,
		},
	}

	tokenRes, err := h.callPlugNmeet("room/getJoinToken", tokenBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("failed to get join token: %v", err)})
	}

	token, _ := tokenRes["token"].(string)
	joinURL := fmt.Sprintf("%s/?access_token=%s", h.clientHost, token)

	// Send meeting message to channel
	var msg models.Message
	meetingMeta, _ := json.Marshal(map[string]string{
		"meeting_room_id": roomID,
		"join_url":        joinURL,
		"title":           req.Title,
	})

	err = h.db.QueryRow(context.Background(),
		`INSERT INTO messages (channel_id, sender_id, body, message_type, metadata)
		 VALUES ($1, $2, $3, 'meeting', $4)
		 RETURNING id, channel_id, sender_id, body, message_type, created_at, updated_at`,
		channelID, userID,
		fmt.Sprintf("🎥 بدأ %s اجتماعاً: %s", userName, req.Title),
		string(meetingMeta),
	).Scan(&msg.ID, &msg.ChannelID, &msg.SenderID, &msg.Body, &msg.MessageType, &msg.CreatedAt, &msg.UpdatedAt)
	if err != nil {
		// Meeting created but message failed — still return success
		return c.JSON(fiber.Map{
			"room_id":  roomID,
			"join_url": joinURL,
			"title":    req.Title,
		})
	}

	msg.SenderName = userName

	// Broadcast meeting started to channel
	h.hub.BroadcastToChannel(channelID, ws.Event{
		Type:    "message.new",
		Payload: msg,
	})

	return c.JSON(fiber.Map{
		"room_id":  roomID,
		"join_url": joinURL,
		"title":    req.Title,
	})
}

func (h *MeetingHandler) JoinMeeting(c *fiber.Ctx) error {
	channelID := c.Params("id")
	userID := c.Locals("user_id").(string)

	// Get user info
	var userName string
	_ = h.db.QueryRow(context.Background(),
		`SELECT display_name FROM users WHERE id = $1`, userID,
	).Scan(&userName)

	roomID := fmt.Sprintf("diwan-%s", channelID)

	tokenBody := pnmJoinTokenReq{
		RoomID: roomID,
		UserInfo: pnmUserInfo{
			Name:    userName,
			UserID:  userID,
			IsAdmin: false,
		},
	}

	tokenRes, err := h.callPlugNmeet("room/getJoinToken", tokenBody)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "failed to join meeting — meeting may not be active"})
	}

	token, _ := tokenRes["token"].(string)
	joinURL := fmt.Sprintf("%s/?access_token=%s", h.clientHost, token)

	return c.JSON(fiber.Map{
		"room_id":  roomID,
		"join_url": joinURL,
	})
}

func (h *MeetingHandler) callPlugNmeet(path string, body interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	signature := h.hmacSign(string(jsonBody))
	url := fmt.Sprintf("%s/auth/%s", h.pnmHost, path)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("API-KEY", h.pnmKey)
	req.Header.Set("HASH-SIGNATURE", signature)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if status, ok := result["status"].(bool); ok && !status {
		msg, _ := result["msg"].(string)
		return result, fmt.Errorf("plugnmeet error: %s", msg)
	}

	return result, nil
}

func (h *MeetingHandler) hmacSign(body string) string {
	mac := hmac.New(sha256.New, []byte(h.pnmSecret))
	mac.Write([]byte(body))
	return hex.EncodeToString(mac.Sum(nil))
}
