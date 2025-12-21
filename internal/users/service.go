package users

import "github.com/sirupsen/logrus"

type Repo interface {
	Get(id int64) (*User, error)
}

type Service struct {
	repo Repo
	log  *logrus.Logger
}
