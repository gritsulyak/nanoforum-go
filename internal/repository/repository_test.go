package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestPostRepoCreate(t *testing.T) {
	tests := []struct {
		name    string
		exec    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			exec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO posts").
					WithArgs("alice", "hello").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "db error",
			exec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO posts").
					WithArgs("alice", "hello").
					WillReturnError(errors.New("boom"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			tt.exec(mock)

			r := NewPostRepo(db)
			if err := r.Create(context.Background(), "alice", "hello"); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestPostRepoList(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		limit    int
		offset   int
		ids      []int64
		query    func(mock sqlmock.Sqlmock)
		wantErr  bool
		wantLen  int
		wantNext bool
	}{
		{
			name:   "no next page",
			limit:  2,
			offset: 0,
			ids:    []int64{1, 2},
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, content, created_at FROM posts").
					WithArgs(3, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content", "created_at"}).
						AddRow(1, "alice", "a", now).
						AddRow(2, "bob", "b", now))
			},
			wantLen:  2,
			wantNext: false,
		},
		{
			name:   "has next page",
			limit:  2,
			offset: 2,
			ids:    []int64{3, 4},
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, content, created_at FROM posts").
					WithArgs(3, 2).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content", "created_at"}).
						AddRow(3, "carol", "c", now).
						AddRow(4, "dave", "d", now).
						AddRow(5, "erin", "e", now))
			},
			wantLen:  2,
			wantNext: true,
		},
		{
			name:   "query error",
			limit:  2,
			offset: 0,
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, content, created_at FROM posts").
					WithArgs(3, 0).
					WillReturnError(errors.New("boom"))
			},
			wantErr: true,
		},
		{
			name:   "scan error",
			limit:  2,
			offset: 0,
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, content, created_at FROM posts").
					WithArgs(3, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content", "created_at"}).
						AddRow("not-an-int", "alice", "a", now))
			},
			wantErr: true,
		},
		{
			name:   "rows close error",
			limit:  2,
			offset: 0,
			query: func(mock sqlmock.Sqlmock) {
				rows := sqlmock.NewRows([]string{"id", "username", "content", "created_at"}).
					AddRow("not-an-int", "alice", "a", now).
					CloseError(errors.New("close boom"))
				mock.ExpectQuery("SELECT id, username, content, created_at FROM posts").
					WithArgs(3, 0).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
		{
			name:   "empty result",
			limit:  5,
			offset: 0,
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT id, username, content, created_at FROM posts").
					WithArgs(6, 0).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "content", "created_at"}))
			},
			wantLen:  0,
			wantNext: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			tt.query(mock)

			r := NewPostRepo(db)
			got, err := r.List(context.Background(), tt.limit, tt.offset)
			if (err != nil) != tt.wantErr {
				t.Fatalf("List() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatal(err)
				}
				return
			}
			if len(got.Posts) != tt.wantLen {
				t.Errorf("List() returned %d posts, want %d", len(got.Posts), tt.wantLen)
			}
			if got.HasNext != tt.wantNext {
				t.Errorf("List() HasNext = %v, want %v", got.HasNext, tt.wantNext)
			}
			for i, p := range got.Posts {
				if p.ID != tt.ids[i] {
					t.Errorf("List() post %d ID = %d, want %d", i, p.ID, tt.ids[i])
				}
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUserRepoCreate(t *testing.T) {
	tests := []struct {
		name    string
		exec    func(mock sqlmock.Sqlmock)
		wantErr bool
	}{
		{
			name: "success",
			exec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("alice", "$2a$10$hash").
					WillReturnResult(sqlmock.NewResult(1, 1))
			},
		},
		{
			name: "duplicate username",
			exec: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec("INSERT INTO users").
					WithArgs("alice", "$2a$10$hash").
					WillReturnError(errors.New("UNIQUE constraint failed"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			tt.exec(mock)

			r := NewUserRepo(db)
			if err := r.Create(context.Background(), "alice", "$2a$10$hash"); (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestUserRepoGetPasswordHash(t *testing.T) {
	tests := []struct {
		name     string
		username string
		query    func(mock sqlmock.Sqlmock)
		want     string
		wantErr  string
	}{
		{
			name:     "success",
			username: "alice",
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT password_hash FROM users").
					WithArgs("alice").
					WillReturnRows(sqlmock.NewRows([]string{"password_hash"}).AddRow("$2a$10$hash"))
			},
			want: "$2a$10$hash",
		},
		{
			name:     "user not found",
			username: "ghost",
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT password_hash FROM users").
					WithArgs("ghost").
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: "user not found",
		},
		{
			name:     "db error",
			username: "alice",
			query: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT password_hash FROM users").
					WithArgs("alice").
					WillReturnError(errors.New("boom"))
			},
			wantErr: "boom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = db.Close() }()
			tt.query(mock)

			r := NewUserRepo(db)
			got, err := r.GetPasswordHash(context.Background(), tt.username)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Errorf("GetPasswordHash() error = %v, want %q", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("GetPasswordHash() error = %v, want nil", err)
			} else if got != tt.want {
				t.Errorf("GetPasswordHash() = %q, want %q", got, tt.want)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
