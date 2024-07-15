package main

import (
	"context"
	"fmt"
	"github.com/project/server/handler"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/project/config"
	"github.com/project/internal/model"
	"github.com/project/internal/repository"
	"github.com/project/internal/service"
	"github.com/project/pkg/github"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("loading env error: %v", err)
	}

	db := config.GetDB()
	if err := db.AutoMigrate(&model.Repository{}, &model.Commit{}); err != nil {
		log.Fatalf("Failed to run production migrations: %v", err)
	}
	gitRepo := repository.NewGitDBRepo(db.DB)
	gitService := service.NewGitInfo(gitRepo, github.NewGithub())

	ticker := time.NewTicker(5 * time.Hour)
	defer ticker.Stop()

	commitTicker := time.NewTicker(1 * time.Hour)
	defer commitTicker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Println("fetching repository data")
				err = gitService.UpdateRepo(context.Background())
				if err != nil {
					log.Printf("Error updating repository repository data: %v", err)
				}
			case <-commitTicker.C:
				log.Println("search repositories of interest")
				err = gitService.SearchRepos(context.Background(), "cryptocurrency")
				if err != nil {
					log.Printf("Error fetching repository: %v", err)
				}
			}
		}
	}()

	port := 8181

	if os.Getenv("PORT") != "" {
		p, err := strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			log.Fatalf("error parsing port, must be numeric: %v", err)
		}

		port = p
	}

	handler := handlers.NewHandler(gitService)

	router := gin.Default()
	router.GET("/repos/language/:language", handler.FetchByLanguage)
	router.GET("/repos/top/:n", handler.GetTopNRepoByStarCount)
	router.GET("/commit/:owner/:repo", handler.FetchCommit)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		fmt.Printf("Listening on port %d\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting...")

	return
}
