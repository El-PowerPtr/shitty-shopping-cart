package main

import (
    "context"
    "github.com/valkey-io/valkey-go"
)

type CartManager interface {
    GetCart(id uint64, ctx context.Context) (*Cart, error)
    AddCart(cart Cart, ctx context.Context) error
    RemoveCart(uint64, context.Context) (*Cart, error)
    AddItemInCart(uint64, uint64, int, context.Context) error
    RemoveItemInCart(uint64, uint64, int, context.Context) error
    FulkyRemoveItemFromCart(uint64, uint64, context.Context) error
    ResetCart(uint64, context.Context) (*Cart, error)
}

type ValkeyCartManager struct {
    client valkey.Client
}

func NewValkeyCartManager(address string, user string, password string) (*ValkeyCartManager, error) {
    client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{address}, Username: user, Password: password})
    if err != nil {
        return nil, err
    }
    return &ValkeyCartManager{client}, nil
}