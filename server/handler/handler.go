package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/project/internal/service"
)

type Handler struct {
	service service.IGitInfo
}

func NewHandler(service service.IGitInfo) *Handler {
	return &Handler{service: service}
}

func (h *Handler) FetchRepo(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctx := context.Background()

	repoData, err := h.service.FetchRepo(ctx, owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, repoData)
}

func (h *Handler) GetTopNRepoByStarCount(c *gin.Context) {
	nStr := c.Param("n")
	n, err := strconv.Atoi(nStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid number"})
		return
	}

	ctx := context.Background()
	repos, err := h.service.GetTopNRepoByStarCount(ctx, n)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, repos)
}

func (h Handler) FetchCommit(c *gin.Context) {
	owner := c.Param("owner")
	repo := c.Param("repo")
	ctx := context.Background()

	commitData, err := h.service.GetCommit(ctx, owner, repo)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, commitData)
}

func (h Handler) FetchByLanguage(c *gin.Context) {
	language := c.Param("language")

	repoData, err := h.service.GetRepoByLanguage(c, language)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, repoData)
}
