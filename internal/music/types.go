package music

import "fmt"

type ErrorInvalidParam struct {
	Param string `json:"param"`
}

func (err ErrorInvalidParam) Error() string {
	return fmt.Sprintf("invalid parameter: %s", err.Param)
}

type ErrorGetSong struct {
	SongName string `json:"songName"`
}

func (err ErrorGetSong) Error() string {
	return fmt.Sprintf("error getting song: %s", err.SongName)
}

type Song struct {
	Name    string
	Artists []string
	Albums  []string
	Path    string
	Content []byte
}

func (s *Song) Unmarshal(params map[string][]string, content []byte) error {
	if val, ok := params["name"]; ok && len(val) == 1 {
		s.Name = val[0]
	} else {
		return ErrorInvalidParam{"name"}
	}
	if val, ok := params["artists"]; ok && len(val) > 0 {
		s.Artists = val
	} else {
		return ErrorInvalidParam{"artist"}
	}
	if val, ok := params["albums"]; ok && len(val) > 0 {
		s.Albums = val
	} else {
		return ErrorInvalidParam{"album"}
	}
	if len(content) == 0 {
		return ErrorGetSong{s.Name}
	}
	s.Content = content
	return nil
}
