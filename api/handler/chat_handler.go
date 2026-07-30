package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/chat"
	secureid "github.com/edsonmubezi/myapp/pkg/encrypt"
	"github.com/edsonmubezi/myapp/pkg/securejson"
	"github.com/gorilla/mux"
)

var ChatUC chat.UseCase

func SetChatUseCase(uc chat.UseCase) {
	ChatUC = uc
}

// ListChatThreadsHandler lists all threads for the current user.
func ListChatThreadsHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	threads, err := ChatUC.ListThreads(r.Context(), auth.OrganizationID, auth.UserID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if threads == nil {
		threads = []chat.Thread{}
	}
	securejson.JSON(w, http.StatusOK, map[string]any{"status": 200, "data": threads})
}

// GetChatMessagesHandler returns messages for a thread.
func GetChatMessagesHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	threadID, err := parseThreadID(r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid thread ID", nil)
		return
	}
	messages, err := ChatUC.GetMessages(r.Context(), threadID, auth.OrganizationID, auth.UserID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if messages == nil {
		messages = []chat.Message{}
	}
	securejson.JSON(w, http.StatusOK, map[string]any{"status": 200, "data": messages})
}

// CreateThreadAndStreamHandler creates a new thread, sends first message, streams response via SSE.
func CreateThreadAndStreamHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	input, validationErrors, err := middleware.ParseAndValidateBody[chat.CreateThreadInput](r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid input", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	flusher := getFlusher(w)
	if flusher == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Streaming not supported", nil)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	thread, _, err := ChatUC.CreateThreadAndSend(
		r.Context(),
		auth.OrganizationID,
		auth.UserID,
		input,
		func(chunk string) error {
			return writeSSEChunk(w, flusher, chunk)
		},
	)
	if err != nil {
		writeSSEEvent(w, flusher, "[ERROR] "+err.Error())
		writeSSEEvent(w, flusher, "[DONE]")
		return
	}

	// Send encrypted thread ID
	encryptedID, _ := secureid.EncryptID(strconv.FormatInt(thread.ID, 10))
	writeSSEEvent(w, flusher, "[THREAD_ID] "+encryptedID)
	writeSSEEvent(w, flusher, "[DONE]")
}

// SendMessageStreamHandler sends a message to existing thread, streams response via SSE.
func SendMessageStreamHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	threadID, err := parseThreadID(r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid thread ID", nil)
		return
	}

	input, validationErrors, parseErr := middleware.ParseAndValidateBody[chat.SendMessageInput](r)
	if parseErr != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid input", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	flusher := getFlusher(w)
	if flusher == nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Streaming not supported", nil)
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	_, err = ChatUC.SendMessage(
		r.Context(),
		threadID,
		auth.OrganizationID,
		auth.UserID,
		input,
		func(chunk string) error {
			return writeSSEChunk(w, flusher, chunk)
		},
	)
	if err != nil {
		writeSSEEvent(w, flusher, "[ERROR] "+err.Error())
	}
	writeSSEEvent(w, flusher, "[DONE]")
}

// UpdateChatThreadHandler updates a thread title.
func UpdateChatThreadHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	threadID, err := parseThreadID(r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid thread ID", nil)
		return
	}
	input, validationErrors, parseErr := middleware.ParseAndValidateBody[chat.UpdateThreadInput](r)
	if parseErr != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid input", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}
	err = ChatUC.UpdateThread(r.Context(), threadID, auth.OrganizationID, auth.UserID, input)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Thread updated", nil)
}

// DeleteChatThreadHandler soft-deletes a thread.
func DeleteChatThreadHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	threadID, err := parseThreadID(r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid thread ID", nil)
		return
	}
	err = ChatUC.DeleteThread(r.Context(), threadID, auth.OrganizationID, auth.UserID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Thread deleted", nil)
}

// GetChatAPIKeyHandler returns the API key config (masked).
func GetChatAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	config, err := ChatUC.GetAPIKeyConfig(r.Context(), auth.OrganizationID)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	securejson.JSON(w, http.StatusOK, map[string]any{"status": 200, "data": config})
}

// UpdateChatAPIKeyHandler saves/updates the OpenAI API key.
func UpdateChatAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	input, validationErrors, err := middleware.ParseAndValidateBody[chat.APIKeyUpdateInput](r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid input", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}
	err = ChatUC.UpdateAPIKey(r.Context(), auth.OrganizationID, input)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "API key updated", nil)
}

// --- helpers ---

func parseThreadID(r *http.Request) (int64, error) {
	raw := mux.Vars(r)["thread_id"]
	return strconv.ParseInt(raw, 10, 64)
}

func getFlusher(w http.ResponseWriter) http.Flusher {
	if f, ok := w.(http.Flusher); ok {
		return f
	}
	if uw, ok := w.(interface{ Unwrap() http.ResponseWriter }); ok {
		if f, ok := uw.Unwrap().(http.Flusher); ok {
			return f
		}
	}
	return nil
}

func writeSSEChunk(w http.ResponseWriter, flusher http.Flusher, text string) error {
	// Escape for SSE — encode as JSON string to handle newlines/special chars
	escaped, _ := json.Marshal(text)
	// Remove surrounding quotes from JSON string
	inner := string(escaped[1 : len(escaped)-1])
	_, err := fmt.Fprintf(w, "data: %s\n\n", inner)
	if err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func writeSSEEvent(w http.ResponseWriter, flusher http.Flusher, event string) {
	fmt.Fprintf(w, "data: %s\n\n", event)
	flusher.Flush()
}

