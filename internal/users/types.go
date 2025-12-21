package users

type User struct {
	ID           int64  `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash"`
	Salt         string `db:"salt"`
	Password     string `json:"password"`
}
