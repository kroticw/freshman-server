package sql

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kroticw/freshman-server/internal/music"
)

type MusicRepo struct {
	pool *pgxpool.Pool
	tx   *pgxpool.Tx
}

func (r *MusicRepo) GetSongByID(ctx context.Context, id int64) (*music.Song, error) {
	//TODO implement me
	panic("implement me")
}

func (r *MusicRepo) GetSongByName(ctx context.Context, name string) (*music.Song, error) {
	//TODO implement me
	panic("implement me")
}

func (r *MusicRepo) GetSongsByArtist(ctx context.Context, artist string) ([]*music.Song, error) {
	//TODO implement me
	panic("implement me")
}

func (r *MusicRepo) GetSongsByAlbum(ctx context.Context, album string) ([]*music.Song, error) {
	//TODO implement me
	panic("implement me")
}

func NewMusicRepo(pool *pgxpool.Pool) *MusicRepo {
	return &MusicRepo{pool: pool}
}
