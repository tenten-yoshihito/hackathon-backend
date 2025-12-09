package model

import "errors"

// Item update errors
var (
	ErrNotAuthorized        = errors.New("not authorized to update this item")
	ErrCannotUpdateSoldItem = errors.New("cannot update sold item")
	ErrItemNotFound         = errors.New("item not found")
)

// Validation errors
var (
	ErrInvalidRequest       = errors.New("invalid request")
	ErrInvalidItemRequest   = errors.New("invalid item request")
	ErrInvalidUpdateRequest = errors.New("invalid item update request")
)
