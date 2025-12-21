package http

import (
	"context"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kroticw/freshman-server/internal/music"
	"github.com/sirupsen/logrus"
)

func SetupRouter(
	ctx context.Context,
	musSvc *music.Service,
	logger *logrus.Logger,
) *gin.Engine {
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.PUT("/api/add", func(c *gin.Context) {
		params := c.Request.URL.Query()
		fh, err := c.FormFile("song")
		if err != nil {
			logger.WithError(err).Error("failed to get file")
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}

		// Дополнительный (быстрый) фильтр по расширению.
		if !isAllowedAudioExtension(fh.Filename) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid file extension",
			})
			return
		}

		mime, err := sniffContentType(fh)
		if err != nil {
			logger.WithError(err).Error("failed to sniff content type")
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		if !isAllowedAudioMIME(mime) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":       "invalid file content type",
				"contentType": mime,
			})
			return
		}

		src, err := fh.Open()
		if err != nil {
			logger.WithError(err).Error("failed to open file")
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		defer src.Close()
		content, err := io.ReadAll(src)
		if err != nil {
			logger.WithError(err).Error("failed to read file")
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		var song music.Song
		if err = song.Unmarshal(params, content); err != nil {
			logger.WithError(err).Error("failed to unmarshal song")
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		err = musSvc.UploadSong(ctx, &song)
		if err != nil {
			logger.WithError(err).Error("failed to upload song")
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"contentType": mime,
		})
	})

	return r
}
