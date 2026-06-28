package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:unused
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

func validUUID() string { return "12345678-1234-1234-1234-123456789012" }

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: User{
				ID:     validUUID(),
				Name:   "Alice",
				Age:    25,
				Email:  "alice@example.com",
				Role:   "admin",
				Phones: []string{"12345678901"},
			},
			expectedErr: nil,
		},
		{
			in:          App{Version: "1.0.0"},
			expectedErr: nil,
		},
		{
			in:          Token{Header: []byte("h"), Payload: []byte("p"), Signature: []byte("s")},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 200, Body: "OK"},
			expectedErr: nil,
		},
		{
			in:          Response{Code: 404},
			expectedErr: nil,
		},
		{
			in: User{
				ID:    "too-short",
				Age:   25,
				Email: "a@b.c",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		{
			in: User{
				ID:    validUUID(),
				Age:   10,
				Email: "a@b.c",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		{
			in: User{
				ID:    validUUID(),
				Age:   99,
				Email: "a@b.c",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		{
			in: User{
				ID:    validUUID(),
				Age:   25,
				Email: "not-an-email",
				Role:  "admin",
			},
			expectedErr: ValidationErrors{},
		},
		{
			in: User{
				ID:    validUUID(),
				Age:   25,
				Email: "a@b.c",
				Role:  "superuser",
			},
			expectedErr: ValidationErrors{},
		},
		{
			in: User{
				ID:     validUUID(),
				Age:    25,
				Email:  "a@b.c",
				Role:   "admin",
				Phones: []string{"123"},
			},
			expectedErr: ValidationErrors{},
		},
		{
			in:          Response{Code: 301},
			expectedErr: ValidationErrors{},
		},
		{
			in:          App{Version: "v2"},
			expectedErr: ValidationErrors{},
		},
		{
			in:          42,
			expectedErr: ErrNotStruct,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			err := Validate(tt.in)

			switch {
			case tt.expectedErr == nil:
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}

			case errors.Is(tt.expectedErr, ErrNotStruct):
				if !errors.Is(err, ErrNotStruct) {
					t.Errorf("expected ErrNotStruct, got: %v", err)
				}

			default:
				var ve ValidationErrors
				if !errors.As(err, &ve) {
					t.Errorf("expected ValidationErrors, got %T: %v", err, err)
				}
			}
		})
	}
}

func TestValidateSummaryAllErrors(t *testing.T) {
	type S struct {
		A string `validate:"len:5"`
		B int    `validate:"min:10"`
		C string `validate:"in:x,y"`
	}
	var ve ValidationErrors
	err := Validate(S{A: "x", B: 1, C: "z"})
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationErrors, got %T: %v", err, err)
	}
	if len(ve) != 3 {
		t.Errorf("expected 3 validation errors, got %d: %v", len(ve), ve)
	}
}
