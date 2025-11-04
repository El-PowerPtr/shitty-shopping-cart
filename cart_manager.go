package main

import (
    "context"
    "strconv"
    "github.com/valkey-io/valkey-go"
)

type CartManager interface {
    GetCart(id uint64, ctx context.Context) (*Cart, error)
    AddCart(cart Cart, ctx context.Context) error
    RemoveCart(uint64, context.Context) (*Cart, error)
    AddItemInCart(uint64, uint64, int, context.Context) error
    RemoveItemInCart(uint64, uint64, int, context.Context) error
    FullyRemoveItemFromCart(uint64, uint64, context.Context) error
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

func (this *ValkeyCartManager) GetCart(id uint64, ctx context.Context) (*Cart, error) {
    strId := strconv.FormatUint(id, 10)
    cmd := this.client.B().Get().Key(strId).Build()
    result := this.client.Do(ctx, cmd)
    
    if err := result.Error(); err != nil {
        return nil, err
    }
    
    dict, err := result.AsMap()
    
    if err != nil {
        return nil, err
    }
    
    owner := dict["owner"]
    ownerId, err := owner.AsUint64()
    
    if err != nil {
        return nil, err
    }
    
    items := dict["items"]
    itemsMap, err := items.AsIntMap()
    
    if err != nil {
        return nil, err
    }
    
    cartItems := make(map[uint64]int, len(itemsMap))
    for k, v := range itemsMap {
        if i, err := strconv.ParseUint(k, 10, 64); err != nil {
            return nil, err
        } else {
            cartItems[i] = int(v)
        }
    }
    
    return &Cart{ownerId, id, cartItems}, nil
}