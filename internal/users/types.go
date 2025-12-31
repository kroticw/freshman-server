package users

type User struct {
	ID           uint64 `db:"id" json:"id"`
	Name         string `db:"name" json:"name"`
	Email        string `db:"email" json:"email"`
	PasswordHash string `db:"password_hash"`
	Salt         string `db:"salt"`
	Password     string `json:"password"`
	CreatedAt    string `db:"created_at"`
	UpdatedAt    string `db:"updated_at"`
	DeletedAt    string `db:"deleted_at"`
}
