package orgs

import (
	"net/http"

	"github.com/edsonmubezi/myapp/api/handler"
	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/gorilla/mux"
)

func registerChatRoutes(r *mux.Router) {
	// Chat routes require only JWT auth (no extra permissions).
	// The frontend already gates the widget to org admins only.

	// List user's chat threads
	r.Handle("/chat/threads",
		http.HandlerFunc(handler.ListChatThreadsHandler),
	).Methods("GET", "OPTIONS")

	// Create new thread + stream first response (SSE)
	r.Handle("/chat/threads/new/stream",
		http.HandlerFunc(handler.CreateThreadAndStreamHandler),
	).Methods("POST", "OPTIONS")

	// Get messages for a thread
	r.Handle("/chat/threads/{thread_id}/messages",
		middleware.ChainMiddleware(
			http.HandlerFunc(handler.GetChatMessagesHandler),
			middleware.DecryptMiddleware("thread_id"),
		),
	).Methods("GET", "OPTIONS")

	// Send message to existing thread + stream response (SSE)
	r.Handle("/chat/threads/{thread_id}/stream",
		middleware.ChainMiddleware(
			http.HandlerFunc(handler.SendMessageStreamHandler),
			middleware.DecryptMiddleware("thread_id"),
		),
	).Methods("POST", "OPTIONS")

	// Update thread title
	r.Handle("/chat/threads/{thread_id}",
		middleware.ChainMiddleware(
			http.HandlerFunc(handler.UpdateChatThreadHandler),
			middleware.DecryptMiddleware("thread_id"),
		),
	).Methods("PUT", "OPTIONS")

	// Delete thread
	r.Handle("/chat/threads/{thread_id}",
		middleware.ChainMiddleware(
			http.HandlerFunc(handler.DeleteChatThreadHandler),
			middleware.DecryptMiddleware("thread_id"),
		),
	).Methods("DELETE", "OPTIONS")

	// API Key management
	r.Handle("/chat/api-key",
		http.HandlerFunc(handler.GetChatAPIKeyHandler),
	).Methods("GET", "OPTIONS")

	r.Handle("/chat/api-key",
		http.HandlerFunc(handler.UpdateChatAPIKeyHandler),
	).Methods("PUT", "OPTIONS")

	// Knowledge base management
	r.Handle("/chat/knowledge",
		http.HandlerFunc(handler.ListKnowledgeArticlesHandler),
	).Methods("GET", "OPTIONS")

	r.Handle("/chat/knowledge",
		http.HandlerFunc(handler.CreateKnowledgeArticleHandler),
	).Methods("POST", "OPTIONS")

	r.Handle("/chat/knowledge/seed",
		http.HandlerFunc(handler.SeedKnowledgeHandler),
	).Methods("POST", "OPTIONS")

	r.Handle("/chat/knowledge/{id}",
		middleware.ChainMiddleware(
			http.HandlerFunc(handler.UpdateKnowledgeArticleHandler),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("PUT", "OPTIONS")

	r.Handle("/chat/knowledge/{id}",
		middleware.ChainMiddleware(
			http.HandlerFunc(handler.DeleteKnowledgeArticleHandler),
			middleware.DecryptMiddleware("id"),
		),
	).Methods("DELETE", "OPTIONS")
}
