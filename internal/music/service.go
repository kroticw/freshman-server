package music

import (
	"context"

	"github.com/sirupsen/logrus"
)

type Storage interface {
	Exists(ctx context.Context, filename string) (exists bool, err error)
	Upload(ctx context.Context, filename string, file []byte) error
	UploadLinked(ctx context.Context, filename string, sourceFilename string, file []byte) error
	Get(ctx context.Context, filename string) ([]byte, error)
	Delete(ctx context.Context, filename string) error
	IsLinkedExists(ctx context.Context, filename string, sourceFilename string) (exists bool, err error)
}

type Repo interface {
	GetSongByID(ctx context.Context, id int64) (*Song, error)
	GetSongByName(ctx context.Context, name string) (*Song, error)
	GetSongsByArtist(ctx context.Context, artist string) ([]*Song, error)
	GetSongsByAlbum(ctx context.Context, album string) ([]*Song, error)
}

type Service struct {
	storage Storage
	repo    Repo
	log     *logrus.Logger
}

func NewMusicService(storage Storage, repo Repo, log *logrus.Logger) *Service {
	return &Service{
		storage,
		repo,
		log,
	}
}

func (s *Service) UploadSong(ctx context.Context, song *Song) error {
	s.log.Infof("Uploading song %s", song.Name)
	return s.storage.Upload(ctx, song.Name, song.Content)
}

func (s *Service) GetSong(ctx context.Context, songName string) ([]byte, error) {
	s.log.Infof("Getting song %s", songName)
	return s.storage.Get(ctx, songName)
}
