package http

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/game-apps/internal/model"
	"github.com/game-apps/internal/service/game"
	"github.com/game-apps/internal/utils"
)

// GameHandler 游戏处理器
type GameHandler struct {
	roomService    *game.RoomService
	sessionService *game.SessionService
	processService *game.ProcessService
}

// NewGameHandler 创建游戏处理器
func NewGameHandler(
	roomService *game.RoomService,
	sessionService *game.SessionService,
	processService *game.ProcessService,
) *GameHandler {
	return &GameHandler{
		roomService:    roomService,
		sessionService: sessionService,
		processService: processService,
	}
}

// CreateRoom 创建房间
func (h *GameHandler) CreateRoom(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	var req game.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	resp, err := h.roomService.CreateRoom(c.Request.Context(), userID, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// JoinRoom 加入房间
func (h *GameHandler) JoinRoom(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	var req game.JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, err.Error()))
		return
	}

	resp, err := h.roomService.JoinRoom(c.Request.Context(), userID, &req)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, resp)
}

// LeaveRoom 离开房间
func (h *GameHandler) LeaveRoom(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的房间ID"))
		return
	}

	if err := h.roomService.LeaveRoom(c.Request.Context(), userID, uint(roomID)); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetRoom 获取房间信息
func (h *GameHandler) GetRoom(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的房间ID"))
		return
	}

	room, err := h.roomService.GetRoom(c.Request.Context(), uint(roomID))
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, room)
}

// ListRooms 列出房间
func (h *GameHandler) ListRooms(c *gin.Context) {
	var status *model.RoomStatus
	if statusStr := c.Query("status"); statusStr != "" {
		s := model.RoomStatus(0)
		if _, err := fmt.Sscanf(statusStr, "%d", &s); err == nil {
			status = &s
		}
	}

	limit := 20
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	offset := 0
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	rooms, err := h.roomService.ListRooms(c.Request.Context(), status, limit, offset)
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, rooms)
}

// StartGame 开始游戏
func (h *GameHandler) StartGame(c *gin.Context) {
	userID := GetUserID(c)
	if userID == 0 {
		Error(c, utils.NewError(utils.ErrCodeUnauthorized, "未授权"))
		return
	}

	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的房间ID"))
		return
	}

	if err := h.processService.StartGame(c.Request.Context(), uint(roomID)); err != nil {
		Error(c, err)
		return
	}

	Success(c, nil)
}

// GetGameState 获取游戏状态
func (h *GameHandler) GetGameState(c *gin.Context) {
	roomIDStr := c.Param("id")
	roomID, err := strconv.ParseUint(roomIDStr, 10, 32)
	if err != nil {
		Error(c, utils.NewError(utils.ErrCodeInvalidInput, "无效的房间ID"))
		return
	}

	state, err := h.processService.GetGameState(c.Request.Context(), uint(roomID))
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, state)
}

