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
    RemoveCart(id uint64, owner uint64, ctx context.Context) (*Cart, error)
    AddItemInCart(id uint64, owner uint64, item uint64, ammount int64, ctx context.Context) (uint64, int, error)
    RemoveItemInCart(id uint64,owner uint64, item uint64, ammount int64, ctx context.Context) (uint64, int64, error)
    FullyRemoveItemFromCart(id uint64, owner uint64, item uint64, ctx context.Context) (uint64, error)
    ResetCart(id uint64, owner uint64, ctx context.Context) error
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

func (this *ValkeyCartManager) getErrIfDiffOwner (idStr *string, id uint64, owner uint64, ctx context.Context) error {
    getOwner := this.client.B().Hget().Key(*idStr).Field("owner").Build()
    r := this.client.Do(ctx, getOwner)
    if realOwner, err := r.AsUint64(); err != nil {
        return err
    } else if realOwner != owner {
        return &DiffOwnerError{id, owner, realOwner}
    } else {
        return nil
    }
}

func (this *ValkeyCartManager) GetCart(id uint64, owner uint64,ctx context.Context) (*Cart, error) {
    strId := strconv.FormatUint(id, 10)
    
    if err := this.getErrIfDiffOwner(&strId, id, owner, ctx); err != nil {
        return nil, err
    }
    
    g := this.client.B().Get().Key(strId).Build()
    r := this.client.Do(ctx, g)
    
    if err := r.Error(); err != nil {
        return nil, err
    }
    
    dict, err := r.AsIntMap()
    
    if err != nil {
        return nil, err
    }
    
    items := make(map[uint64]int64, len(dict) - 1)
    
    for k, v := range dict {
        if k == "owner" {
            continue
        } else if i, err := strconv.ParseUint(k, 10, 64); err != nil {
            return nil, err
        } else {
            items[i] = v
        }
    }
    
    return &Cart{owner, id, items}, nil
}

func (this *ValkeyCartManager) AddCart(cart *Cart, ctx context.Context) error {
    id := strconv.FormatUint(cart.Id, 10)
    e := this.client.B().Exists().Key(id).Build()
    r := this.client.Do(ctx, e)
    
    if exists, err := r.AsBool() ; err != nil {
       return err 
    } else if exists {
        return &AlreadyExistingCartError{cart.Id}
    }
    
    owner := strconv.FormatUint(cart.OwnerId, 10)
    fields := make(map[string]string, len(cart.Items) + 1)
    fields["owner"] = owner
    
    for k, v := range cart.Items {
        item := strconv.FormatUint(k, 10)
        ammount := strconv.FormatInt(v, 10)
        fields[item] = ammount
    }
    
    c := this.client.B().Hset().Key(id).FieldValue().FieldValueIter(maps.All(fields)).Build()
    r = this.client.Do(ctx, c)
    
    return r.Error()
}

func (this *ValkeyCartManager) RemoveCart(id uint64, owner uint64, ctx context.Context) error {
    idStr := strconv.FormatUint(id, 10)
    
    if err := this.getErrIfDiffOwner(&idStr, id, owner, ctx); err != nil {
        return err
    }
    
    d := this.client.B().Del().Key(idStr).Build()
    r := this.client.Do(ctx, d)
    
    return r.Error()
}

func (this *ValkeyCartManager) ResetCart(id uint64, owner uint64, ctx context.Context) error {
    idStr := strconv.FormatUint(id, 10)
    
    if err := this.getErrIfDiffOwner(&idStr, id, owner, ctx); err != nil {
        return err
    }
    
    ownerStr := strconv.FormatUint(owner, 10)
    rst := this.client.B().Set().Key(idStr).Value(ownerStr).Build()
    r := this.client.Do(ctx, rst)
    
    return r.Error()
}

func (this *ValkeyCartManager) AddItemInCart(id uint64, owner uint64, item uint64, ammount int64, ctx context.Context) (uint64, int64, error) {
    idStr := strconv.FormatUint(id, 10)
    
    if err := this.getErrIfDiffOwner(&idStr, id, owner, ctx); err != nil {
        return 0, 0 ,err
    }
    itemStr := strconv.FormatUint(item, 10)
    add := this.client.B().Hincrby().Key(idStr).Field(itemStr).Increment(ammount).Build()
    r := this.client.Do(ctx, add)
    
    if val, err := r.AsInt64(); err != nil {
        return 0, 0, err
    } else {
        return item, val, nil
    }
}

func (this *ValkeyCartManager) RemoveItemInCart(id uint64,owner uint64, item uint64, ammount int64, ctx context.Context) (uint64, int64, error) {
    idStr := strconv.FormatUint(id, 10)
    
    if err := this.getErrIfDiffOwner(&idStr, id, owner, ctx); err != nil {
        return 0, 0 ,err
    }
    
    itemStr := strconv.FormatUint(item, 10)
    ga := this.client.B().Hget().Key(idStr).Field(itemStr).Build()
    r := this.client.Do(ctx, ga)
    
    if current, err := r.AsInt64(); err != nil {
        return 0, 0, err
    } else if current < int64(ammount) {
        // set to 0 if we have not enough items
        return item, 0, nil
    } else {
        d := this.client.B().Hincrby().Key(idStr).Field(itemStr).Increment(-ammount).Build()
        r = this.client.Do(ctx, d)
        
        if val, err := r.AsInt64(); err != nil {
            return 0, 0, err
        } else {
            return item, val, nil
        }
    }
}