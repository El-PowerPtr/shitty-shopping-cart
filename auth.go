package main

import (
    "context"
)

type Auth interface {
    ParseToken(ctx context.Context, strToken string) (*User, error)
}

type AuthPort interface {
    RetrievePublicKey() string
}