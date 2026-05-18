package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"password_hash"`
	Role         string `json:"role"`
	Token        string `json:"token"`
	CreatedAt    string `json:"created_at"`
}

type UserPublic struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Role      string `json:"role"`
	Token     string `json:"token,omitempty"`
	CreatedAt string `json:"created_at"`
}

func (u *User) Public(includeToken bool) UserPublic {
	p := UserPublic{
		ID:        u.ID,
		Username:  u.Username,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
	}
	if includeToken {
		p.Token = u.Token
	}
	return p
}

type usersData struct {
	Users []User `json:"users"`
}

type UserStore struct {
	mu       sync.RWMutex
	dataDir  string
	users    []User
}

func NewUserStore(dataDir string) (*UserStore, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	s := &UserStore{dataDir: dataDir}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *UserStore) filePath() string {
	return filepath.Join(s.dataDir, "users.json")
}

func (s *UserStore) load() error {
	s.users = []User{}
	data, err := os.ReadFile(s.filePath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var d usersData
	if err := json.Unmarshal(data, &d); err != nil {
		return err
	}
	s.users = d.Users
	return nil
}

func (s *UserStore) save() error {
	d := usersData{Users: s.users}
	data, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath(), data, 0644)
}

func (s *UserStore) HasUsers() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.users) > 0
}

func (s *UserStore) CreateUser(username, password, role string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Username == username {
			return nil, fmt.Errorf("username already exists")
		}
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := User{
		ID:           "u_" + generateRandom(8),
		Username:     username,
		PasswordHash: string(hash),
		Role:         role,
		Token:        "tok_" + generateRandom(32),
		CreatedAt:    time.Now().UTC().Format(time.RFC3339),
	}

	s.users = append(s.users, user)
	if err := s.save(); err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *UserStore) Authenticate(username, password string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Username == username {
			if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) == nil {
				return &u
			}
			return nil
		}
	}
	return nil
}

func (s *UserStore) FindByToken(token string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.Token == token {
			return &u
		}
	}
	return nil
}

func (s *UserStore) FindByID(id string) *User {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, u := range s.users {
		if u.ID == id {
			return &u
		}
	}
	return nil
}

func (s *UserStore) ListUsers() []UserPublic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]UserPublic, 0, len(s.users))
	for _, u := range s.users {
		result = append(result, u.Public(false))
	}
	return result
}

func (s *UserStore) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, u := range s.users {
		if u.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return fmt.Errorf("user not found")
	}

	s.users = append(s.users[:idx], s.users[idx+1:]...)
	return s.save()
}

func (s *UserStore) ResetToken(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, u := range s.users {
		if u.ID == id {
			s.users[i].Token = "tok_" + generateRandom(32)
			if err := s.save(); err != nil {
				return "", err
			}
			return s.users[i].Token, nil
		}
	}
	return "", fmt.Errorf("user not found")
}

func generateRandom(length int) string {
	b := make([]byte, length/2+1)
	rand.Read(b)
	return hex.EncodeToString(b)[:length]
}
