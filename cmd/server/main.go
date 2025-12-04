package main

import (
	"ai-service-go/internals/config"
	"ai-service-go/internals/controllers"
	"ai-service-go/internals/handlers"
	"ai-service-go/internals/orchestrator"
	"ai-service-go/internals/tools"
	"ai-service-go/internals/utils"
	"ai-service-go/internals/vector_db"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

func main() {
	// initialize logger
	utils.InitLogger()
	// load config
	config.LoadConfig()
	// initialise aiManager
	controllers.InitializeAiManager()
	// initialise vector store
	vector_db.NewVectorStore()
	// initialise tool registry and register tools
	toolRegistory := tools.NewToolRegistry()
	toolRegistory.RegisterTool(&tools.GetUserGuideInformation{
		VectorManager: &vector_db.VectorStoreManager,
	})
	toolRegistory.RegisterTool(&tools.GetLayerInformation{})
	toolRegistory.RegisterTool(&tools.GetUpdateInformation{})
	toolRegistory.RegisterTool(&tools.GetAllAvailableLayers{})
	toolRegistory.RegisterTool(&tools.LocateALayer{}) 
	// intialise orchestrator
	orch := orchestrator.NewOrchestrator(controllers.AiManager, toolRegistory, &vector_db.VectorStoreManager)
	// Create a new router
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	h := handlers.NewHandler(&vector_db.VectorStoreManager, &controllers.AiManager, toolRegistory, orch)
	// Serve static files
	r.Handle("/ui/*", http.StripPrefix("/ui/", http.FileServer(http.Dir("./public"))))
	r.Get("/load-embeddings", h.HandleLoadEmbeddings)
	r.Get("/load-embeddings-v2", h.HandleLoadEmbeddingsV2)
	r.Get("/vectors", h.HandleGetAllVectors)
	r.Get("/status", h.HandleStatus)
	r.Post("/handle-query-v1", h.HandleQueryV1)
	r.Post("/handle-query-v1.5", h.HandleQueryAndTools)
	r.Post("/handle-query-v2", h.HandleQueryV2)
	r.Get("/ws/query", h.HandleWebSocketQuery)
	r.Get("/temp", h.TempHandler)

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
