package main

import (
    "context"
    "strconv"
    "maps"
    "github.com/valkey-io/valkey-go"
)

type CartManager interface {
    GetCart(id uint64, owner uint64, ctx context.Context) (*Cart, error)
    AddCart(cart *Cart, ctx context.Context) error
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

func (this *ValkeyCartManager) getOwner (id *string, ctx context.Context) (uint64, error) {
    getOwner := this.client.B().Hget().Key(*id).Field("owner").Build()
    r := this.client.Do(ctx, getOwner)
    if realOwner, err := r.AsUint64(); err != nil {
        return 0, err
    } else {
        return realOwner, nil
    }
}

func (this *ValkeyCartManager) GetCart(id uint64, owner uint64,ctx context.Context) (*Cart, error) {
    strId := strconv.FormatUint(id, 10)
    
    if ro, err := this.getOwner(&strId, ctx); err != nil {
        return nil, err
    }  else if ro != owner {
        return nil, &DiffOwnerError{id, owner, ro}
    }
    
    cmd := this.client.B().Get().Key(strId).Build()
    result := this.client.Do(ctx, cmd)
    
    if err := result.Error(); err != nil {
        return nil, err
    }
    
    dict, err := result.AsIntMap()
    
    if err != nil {
        return nil, err
    }
    
    items := make(map[uint64]int, len(dict) - 1)
    
    for k, v := range dict {
        if k == "owner" {
            continue
        } else if i, err := strconv.ParseUint(k, 10, 64); err != nil {
            return nil, err
        } else {
            items[i] = int(v)
        }
    }
    
    return &Cart{owner, id, items}, nil
}

func (this *ValkeyCartManager) AddCart(cart *Cart, ctx context.Context) error {
    id := strconv.FormatUint(cart.Id, 10)
    check := this.client.B().Exists().Key(id).Build()
    result := this.client.Do(ctx, check)
    if exists, err := result.AsBool() ; err != nil {
       return err 
    } else if exists {
        return &AlreadyExistingCartError{cart.Id}
    }
    owner := strconv.FormatUint(cart.OwnerId, 10)
    fields := make(map[string]string, len(cart.Items) + 1)
    fields["owner"] = owner
    for k, v := range cart.Items {
        item := strconv.FormatUint(k, 10)
        ammount := strconv.FormatInt(int64(v), 10)
        fields[item] = ammount
    }
    create := this.client.B().Hset().Key(id).FieldValue().FieldValueIter(maps.All(fields)).Build()
    
    r := this.client.Do(ctx, create)
    
    return r.Error()
}