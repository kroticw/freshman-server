package http

import (
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

// isAllowedAudioExtension validates by filename extension (cheap дополнительный фильтр).
// Не является надежной защитой без sniffing.
func isAllowedAudioExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".mp3", ".wav", ".flac", ".aac", ".m4a", ".ogg", ".opus", ".weba":
		return true
	default:
		return false
	}
}

// sniffContentType reads up to 512 bytes and detects a MIME type.
// NOTE: It does not trust client headers.
func sniffContentType(fh *multipart.FileHeader) (string, error) {
	src, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	buf := make([]byte, 512)
	n, err := io.ReadFull(src, buf)
	if err != nil {
		// io.ReadFull returns an error for short reads; for sniffing that's fine.
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			return "", err
		}
	}

	return http.DetectContentType(buf[:n]), nil
}

func isAllowedAudioMIME(mime string) bool {
	mime = strings.ToLower(strings.TrimSpace(mime))
	// Быстрая проверка по префиксу.
	if strings.HasPrefix(mime, "audio/") {
		return true
	}
	// Некоторые аудио-контейнеры/варианты могут детектиться иначе.
	switch mime {
	case "application/ogg":
		return true
	default:
		return false
	}
}
