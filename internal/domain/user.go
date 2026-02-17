package domain

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
)

type User struct {
	ID       uint   `db:"id"`
	Username string `db:"username"`
	Password string `db:"password"`
	Role     Role   `db:"role"`
	Address  string `db:"address"`
	Phone    string `db:"phone"`
}

// 2. Repository Interface
type UserRepository interface {
	GetByUsername(username string) (*User, error)
	Create(user *User) error

	GetByID(id uint) (*User, error) // 👈 เพิ่ม
	Update(id uint, user *User) error
}

// 3. Service Interface
type UserService interface {
	Register(user *User) error
	Login(username, password string) (string, string, error)

	GetUser(requesterID uint, requesterRole Role, targetID uint) (*User, error)
	UpdateUser(requesterID uint, requesterRole Role, targetID uint, input *User) error
}
