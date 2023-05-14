package repository

import "database/sql"

type User struct {
	ID    int
	Name  string
	Email string
}

type UserRepository interface {
	GetAll() ([]User, error)
	GetByID(int) (*User, error)
	Create(*User) error
	Update(*User) error
	Delete(int) error
}

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{db}
}

func (r *UserRepositoryImpl) GetAll() ([]User, error) {
	rows, err := r.db.Query("SELECT * FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		err := rows.Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepositoryImpl) GetByID(id int) (*User, error) {
	var user User
	err := r.db.QueryRow("SELECT * FROM users WHERE id=?", id).Scan(&user.ID, &user.Name, &user.Email)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepositoryImpl) Create(user *User) error {
	result, err := r.db.Exec("INSERT INTO users(name, email) VALUES(?,?)", user.Name, user.Email)
	if err != nil {
		return err
	}

	id, _ := result.LastInsertId()
	user.ID = int(id)

	return nil
}

func (r *UserRepositoryImpl) Update(user *User) error {
	_, err := r.db.Exec("UPDATE users SET name=?, email=? WHERE id=?", user.Name, user.Email, user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRepositoryImpl) Delete(id int) error {
	_, err := r.db.Exec("DELETE FROM users WHERE id=?", id)
	if err != nil {
		return err
	}

	return nil
}
