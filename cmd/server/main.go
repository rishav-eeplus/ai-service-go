package main

import (
	"ai-service-go/internals/chats_db"
	"ai-service-go/internals/config"
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/handlers"
	"ai-service-go/internals/logger"
	"ai-service-go/internals/orchestrator"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/vector_db"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	// initialize logger
	logger.InitLogger()
	// load config
	config.LoadConfig()
	// initialise aiManager
	controllers.InitializeAiManager()
	// initialise vector store
	vector_db.NewVectorStore()
	// intialize chat db
	chatDBManager, err := chats_db.NewDB()
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing chat database")
	}
	// initialise tool registry and register tools
	toolRegistory := tools.NewToolRegistry()
	toolRegistory.RegisterTool(&tools.GetUserGuideInformation{
		VectorManager: &vector_db.VectorStoreManager,
	})
	toolRegistory.RegisterTool(&tools.GetLayerInformation{})
	toolRegistory.RegisterTool(&tools.GetUpdateInformation{})
	toolRegistory.RegisterTool(&tools.GetAllAvailableLayers{})
	toolRegistory.RegisterTool(&tools.LocateALayer{})
	toolRegistory.RegisterTool(&tools.GetHelpSupport{})
	// intialise orchestrator
	orch := orchestrator.NewOrchestrator(controllers.AiManager, toolRegistory, &vector_db.VectorStoreManager, chatDBManager)
	// Create a new router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight OPTIONS requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
	h := handlers.NewHandler(&vector_db.VectorStoreManager, &controllers.AiManager, toolRegistory, orch, chatDBManager)
	// Serve static files
	r.Handle("/ui/*", http.StripPrefix("/ui/", http.FileServer(http.Dir("./public"))))
	// Routes for embedding and vector management
	r.Get("/load-embeddings", h.HandleLoadEmbeddings)
	r.Get("/vectors", h.HandleGetAllVectors)
	// Routes for chat management
	r.Get("/status", h.HandleStatus)
	r.Post("/handle-query-v1", h.HandleQueryV1)
	r.Post("/handle-query-v2", h.HandleSSEQuery)
	r.Get("/temp", h.AllAvailableLayersHandler)
	r.Get("/temp2", h.UpdateCycleHandler)
	// Routes for chat database management
	r.Get("/chats", h.HandleGetAllChats)
	r.Get("/chats/user", h.HandleGetChatForAUser)
	r.Post("/chats/feedback", h.HandleChatFeedBack)

	// 404 handler
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 not found", http.StatusNotFound)
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: r,
	}
	log.Info().Msgf("Starting server on port %d", 8080)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("Error starting server")
	}

}
