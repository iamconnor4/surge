package web

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iamconnor4/surge/internal/domain"
)

func TestHandleUserByID(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
	id, err := domain.ParseUserID(testUserID)
	if err != nil {
		t.Fatalf("parse test user ID: %v", err)
	}

	tests := []struct {
		name            string
		path            string
		result          domain.User
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
	}{
		{
			name: "success",
			path: "/users/" + testUserID,
			result: domain.User{
				ID: id, Email: "connor@example.com", FirstName: "Connor", LastName: "Smith", CreatedAt: createdAt,
			},
			wantStatus: http.StatusOK, wantBody: `"email":"connor@example.com"`, wantServiceCall: true,
		},
		{
			name: "invalid ID", path: "/users/not-a-uuid",
			wantStatus: http.StatusBadRequest, wantBody: `"error":"Invalid user ID"`,
		},
		{
			name: "not found", path: "/users/" + testUserID, serviceErr: domain.ErrUserNotFound,
			wantStatus: http.StatusNotFound, wantBody: `"error":"User not found"`, wantServiceCall: true,
		},
		{
			name: "service failure", path: "/users/" + testUserID, serviceErr: errors.New("database unavailable"),
			wantStatus: http.StatusInternalServerError, wantBody: `"error":"The server encountered a problem`, wantServiceCall: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Users: fakeUserManager{
					userByID: func(_ context.Context, gotID domain.UserID) (domain.User, error) {
						called = true
						if gotID != id {
							t.Errorf("UserByID() ID = %s, want %s", gotID, id)
						}
						return tt.result, tt.serviceErr
					},
				},
			})

			response := httptest.NewRecorder()
			server.routes().ServeHTTP(response, httptest.NewRequest(http.MethodGet, tt.path, nil))

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", response.Body.String(), tt.wantBody)
			}
			if called != tt.wantServiceCall {
				t.Errorf("service called = %t, want %t", called, tt.wantServiceCall)
			}
		})
	}
}

func TestHandleCreateUser(t *testing.T) {
	t.Parallel()

	createdAt := time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
	id, err := domain.ParseUserID(testUserID)
	if err != nil {
		t.Fatalf("parse test user ID: %v", err)
	}

	createdUser := domain.User{
		ID:        id,
		Email:     "connor@example.com",
		FirstName: "Connor",
		LastName:  "Smith",
		CreatedAt: createdAt,
	}

	tests := []struct {
		name            string
		body            string
		serviceErr      error
		wantStatus      int
		wantBody        string
		wantServiceCall bool
		wantInput       domain.CreateUserInput
	}{
		{
			name:            "success",
			body:            `{"email":"connor@example.com","firstName":"Connor","lastName":"Smith"}`,
			wantStatus:      http.StatusCreated,
			wantBody:        `"id":"` + testUserID + `"`,
			wantServiceCall: true,
			wantInput:       domain.CreateUserInput{Email: "connor@example.com", FirstName: "Connor", LastName: "Smith"},
		},
		{
			name:       "invalid JSON",
			body:       `{"email":`,
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"Invalid request body"`,
		},
		{
			name:            "duplicate email",
			body:            `{"email":"connor@example.com","firstName":"Connor","lastName":"Smith"}`,
			serviceErr:      domain.ErrUserAlreadyExists,
			wantStatus:      http.StatusConflict,
			wantBody:        `"error":"A user with this email already exists"`,
			wantServiceCall: true,
			wantInput:       domain.CreateUserInput{Email: "connor@example.com", FirstName: "Connor", LastName: "Smith"},
		},
		{
			name:            "invalid email",
			body:            `{"email":"","firstName":"Connor","lastName":"Smith"}`,
			serviceErr:      domain.ErrUserInvalidEmail,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"Email is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateUserInput{FirstName: "Connor", LastName: "Smith"},
		},
		{
			name:            "invalid first name",
			body:            `{"email":"connor@example.com","firstName":"","lastName":"Smith"}`,
			serviceErr:      domain.ErrUserInvalidFirstName,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"First name is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateUserInput{Email: "connor@example.com", LastName: "Smith"},
		},
		{
			name:            "invalid last name",
			body:            `{"email":"connor@example.com","firstName":"Connor","lastName":""}`,
			serviceErr:      domain.ErrUserInvalidLastName,
			wantStatus:      http.StatusUnprocessableEntity,
			wantBody:        `"error":"Last name is required"`,
			wantServiceCall: true,
			wantInput:       domain.CreateUserInput{Email: "connor@example.com", FirstName: "Connor"},
		},
		{
			name:            "service failure",
			body:            `{"email":"connor@example.com","firstName":"Connor","lastName":"Smith"}`,
			serviceErr:      errors.New("database unavailable"),
			wantStatus:      http.StatusInternalServerError,
			wantBody:        `"error":"The server encountered a problem`,
			wantServiceCall: true,
			wantInput:       domain.CreateUserInput{Email: "connor@example.com", FirstName: "Connor", LastName: "Smith"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			called := false
			server := testServer(Dependencies{
				Users: fakeUserManager{
					createUser: func(_ context.Context, input domain.CreateUserInput) (domain.User, error) {
						called = true
						if input != tt.wantInput {
							t.Errorf("CreateUser() input = %#v, want %#v", input, tt.wantInput)
						}

						return createdUser, tt.serviceErr
					},
				},
			})

			response := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(tt.body))
			server.routes().ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if !strings.Contains(response.Body.String(), tt.wantBody) {
				t.Errorf("body = %q, want it to contain %q", response.Body.String(), tt.wantBody)
			}
			if called != tt.wantServiceCall {
				t.Errorf("service called = %t, want %t", called, tt.wantServiceCall)
			}
		})
	}
}
