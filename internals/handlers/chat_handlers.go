package handlers

import (
	"ai-service-go/internals/logger"
	"net/http"
	"os"
	"strconv"
)



func (h *Handler) HandleGetChatForAUser(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	if secret != os.Getenv("SECRET") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	userName := r.URL.Query().Get("user_name")
	chat, err := h.ChatDBManager.GetChatsForAUser(r.Context(), userName)
	if err != nil {
		logger.Logger.Errorf("Error while fetching chat for user %s: %v", userName, err)
		http.Error(w, "Error while fetching chat", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Chat fetched successfully", chat)
}

func (h *Handler) HandleChatFeedBack(w http.ResponseWriter, r *http.Request) {
	chatID := r.URL.Query().Get("chat_id")
	feedback := r.URL.Query().Get("feedback")
	chatIdInt, ok := strconv.Atoi(chatID)
	if ok != nil {
		logger.Logger.Errorf("Invalid chat ID %s: %v", chatID, ok)
		http.Error(w, "Invalid chat ID", http.StatusBadRequest)
		return
	}
	feedbackInt, ok := strconv.Atoi(feedback)
	if ok != nil {
		logger.Logger.Errorf("Invalid feedback %s: %v", feedback, ok)
		http.Error(w, "Invalid feedback", http.StatusBadRequest)
		return
	}
	err := h.ChatDBManager.UpdateChatFeedback(r.Context(), chatIdInt, feedbackInt)
	if err != nil {
		logger.Logger.Errorf("Error while updating chat feedback for chat ID %s: %v", chatID, err)
		http.Error(w, "Error while updating chat feedback", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Chat feedback updated successfully", nil)
}



func (h *Handler) HandleGetAllChats(w http.ResponseWriter, r *http.Request) {
	secret := r.URL.Query().Get("secret")
	if secret != os.Getenv("SECRET") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	chats, err := h.ChatDBManager.GetAllChats(r.Context())
	if err != nil {
		logger.Logger.Errorf("Error while fetching chats: %v", err)
		http.Error(w, "Error while fetching chats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	SendSuccessResponse(w, 200, "Chats fetched successfully", map[string]interface{}{
		"count": len(chats),
		"chats": chats,
	})
}