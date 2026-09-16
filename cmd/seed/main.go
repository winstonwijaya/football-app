// cmd/seed creates the initial user directly in the database, since the
// API deliberately has no registration endpoint — users are provisioned
// out-of-band. Safe to re-run: skips creation if the username already
// exists.
package main

import (
	"context"
	"errors"
	"flag"
	"log"

	"gorm.io/gorm"

	"football-app/internal/config"
	"football-app/internal/model"
	"football-app/internal/repository"
	"football-app/pkg/hash"
)

func main() {
	username := flag.String("username", "admin", "login username")
	password := flag.String("password", "", "login password (required)")
	firstName := flag.String("first-name", "Admin", "user's first name")
	lastName := flag.String("last-name", "User", "user's last name")
	flag.Parse()

	if *password == "" {
		log.Fatal("-password is required")
	}

	cfg := config.Load()
	db, err := config.Connect(cfg.DB)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get underlying sql.DB: %v", err)
	}
	defer sqlDB.Close()

	ctx := context.Background()
	userRepo := repository.NewUserRepository(db)

	if _, err := userRepo.FindByUsername(ctx, *username); err == nil {
		log.Printf("user %q already exists, skipping", *username)
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		log.Fatalf("failed to check for existing user: %v", err)
	}

	passwordHash, err := hash.Hash(*password)
	if err != nil {
		log.Fatalf("failed to hash password: %v", err)
	}

	user := &model.User{
		FirstName:    *firstName,
		LastName:     *lastName,
		Username:     *username,
		PasswordHash: passwordHash,
	}
	if err := userRepo.Create(ctx, user); err != nil {
		log.Fatalf("failed to create user: %v", err)
	}

	log.Printf("created user %q (id=%d)", *username, user.ID)
}
