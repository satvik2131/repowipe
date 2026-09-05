package controllers

import (
	"net/http"
	"repowipe/services"
	"repowipe/types"

	"github.com/gin-gonic/gin"
)

// StartTransfer enqueues an any-to-any transfer job.
func StartTransfer(c *gin.Context) {
	sessionID, _, ok := requireSession(c)
	if !ok {
		return
	}

	var req types.TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	job, err := services.EnqueueTransfer(sessionID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusAccepted, job)
}

// GetTransfer returns transfer job progress.
func GetTransfer(c *gin.Context) {
	sessionID, _, ok := requireSession(c)
	if !ok {
		return
	}
	id := c.Param("id")
	job, err := services.GetTransferJob(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transfer not found"})
		return
	}
	if job.SessionID != sessionID {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, job)
}
