package handler

import (
	"net/http"
	"strconv"

	"github.com/edsonmubezi/myapp/api/middleware"
	"github.com/edsonmubezi/myapp/internal/chat"
	"github.com/edsonmubezi/myapp/pkg/securejson"
	"github.com/gorilla/mux"
)

var KnowledgeRepo chat.KnowledgeRepository

func SetKnowledgeRepository(repo chat.KnowledgeRepository) {
	KnowledgeRepo = repo
}

// ListKnowledgeArticlesHandler lists knowledge base articles.
func ListKnowledgeArticlesHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	orgID := auth.OrganizationID

	category := r.URL.Query().Get("category")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	articles, total, err := KnowledgeRepo.ListArticles(r.Context(), &orgID, category, page, pageSize)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	if articles == nil {
		articles = []chat.KnowledgeArticle{}
	}
	securejson.JSON(w, http.StatusOK, map[string]any{
		"status": 200,
		"data":   articles,
		"total":  total,
		"page":   page,
	})
}

// CreateKnowledgeArticleHandler creates a knowledge article with embedding.
func CreateKnowledgeArticleHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	orgID := auth.OrganizationID

	input, validationErrors, err := middleware.ParseAndValidateBody[chat.CreateKnowledgeInput](r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid input", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	// Get raw API key for embedding generation
	apiKey, err := ChatUC.GetRawAPIKey(r.Context(), orgID)
	if err != nil || apiKey == "" {
		SendJSONResponse(w, http.StatusBadRequest, "OpenAI API key not configured. Set it in AI Assistant settings first.", nil)
		return
	}

	// Generate embedding
	embeddingSvc := chat.NewEmbeddingService(apiKey)
	embedding, err := embeddingSvc.Embed(r.Context(), input.Content)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate embedding: "+err.Error(), nil)
		return
	}

	article := &chat.KnowledgeArticle{
		OrganizationID: &orgID,
		Category:       input.Category,
		Title:          input.Title,
		Content:        input.Content,
	}

	created, err := KnowledgeRepo.CreateArticle(r.Context(), article, embedding)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	securejson.JSON(w, http.StatusCreated, map[string]any{"status": 201, "data": created})
}

// UpdateKnowledgeArticleHandler updates a knowledge article.
func UpdateKnowledgeArticleHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	orgID := auth.OrganizationID

	id, err := parseKnowledgeID(r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	input, validationErrors, parseErr := middleware.ParseAndValidateBody[chat.UpdateKnowledgeInput](r)
	if parseErr != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid input", nil)
		return
	}
	if validationErrors != nil {
		SendValidationErrors(w, validationErrors)
		return
	}

	// Re-generate embedding if content changed
	var embedding []float32
	if input.Content != nil {
		apiKey, err := ChatUC.GetRawAPIKey(r.Context(), orgID)
		if err != nil || apiKey == "" {
			SendJSONResponse(w, http.StatusBadRequest, "OpenAI API key not configured", nil)
			return
		}
		embeddingSvc := chat.NewEmbeddingService(apiKey)
		embedding, err = embeddingSvc.Embed(r.Context(), *input.Content)
		if err != nil {
			SendJSONResponse(w, http.StatusInternalServerError, "Failed to generate embedding: "+err.Error(), nil)
			return
		}
	}

	updated, err := KnowledgeRepo.UpdateArticle(r.Context(), id, &orgID, input, embedding)
	if err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	securejson.JSON(w, http.StatusOK, map[string]any{"status": 200, "data": updated})
}

// DeleteKnowledgeArticleHandler soft-deletes a knowledge article.
func DeleteKnowledgeArticleHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	orgID := auth.OrganizationID

	id, err := parseKnowledgeID(r)
	if err != nil {
		SendJSONResponse(w, http.StatusBadRequest, "Invalid ID", nil)
		return
	}

	if err := KnowledgeRepo.DeleteArticle(r.Context(), id, &orgID); err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Article deleted", nil)
}

// SeedKnowledgeHandler seeds global knowledge articles.
func SeedKnowledgeHandler(w http.ResponseWriter, r *http.Request) {
	auth := middleware.GetAuthContext(r)
	orgID := auth.OrganizationID

	apiKey, err := ChatUC.GetRawAPIKey(r.Context(), orgID)
	if err != nil || apiKey == "" {
		SendJSONResponse(w, http.StatusBadRequest, "OpenAI API key not configured. Set it in AI Assistant settings first.", nil)
		return
	}

	embeddingSvc := chat.NewEmbeddingService(apiKey)
	if err := chat.SeedGlobalKnowledge(r.Context(), KnowledgeRepo, embeddingSvc); err != nil {
		SendJSONResponse(w, http.StatusInternalServerError, "Seeding failed: "+err.Error(), nil)
		return
	}
	SendJSONResponse(w, http.StatusOK, "Knowledge base seeded successfully", nil)
}

func parseKnowledgeID(r *http.Request) (int64, error) {
	raw := mux.Vars(r)["id"]
	return strconv.ParseInt(raw, 10, 64)
}
