package main

import (
    "github.com/valkey-io/valkey-go"
    "github.com/golang-jwt/jwt/v5"
    "context"
)

type Auth interface {
    ParseToken(ctx context.Context, strToken string) (*User, error)
}

type AuthPort interface {
    RetrievePublicKey() string
}

type JwtAuth struct {
    client valkey.Client
    publicKey []byte
    authPort AuthPort
}

type JwtClaims struct {
    User
    jwt.RegisteredClaims
}

func (this JwtAuth) ParseToken (ctx context.Context, strToken string) (*User, error) {
    cmd := this.client.B().Get().Key(strToken).Build()
    
    if result := this.client.Do(ctx, cmd); result.Error != nil {
        var usr User
        result.DecodeJSON(&usr)
        return &usr, nil
    }
    
    claims := new(JwtClaims)
    parser := jwt.NewParser(jwt.WithExpirationRequired())
    token, err := parser.ParseWithClaims(
        strToken,
        *claims,
        func (t *jwt.Token)(any, error){
            return this.publicKey, nil
        },
    )
    
    if err != nil {
        return nil, err
    } else if token.Valid {
        et, err := claims.GetExpirationTime()
        
        if err != nil {
            return nil, err
        }
        
        so := this.client.B().JsonSet().Key(strToken).Path(".").Value(valkey.JSON(claims.User)).Build()
        ea := this.client.B().Expireat().Key(strToken).Timestamp(et.Unix()).Build()
        
        this.client.Do(ctx, so)
        this.client.Do(ctx, ea)
        
        return &claims.User, err
    } else {
        return nil, &InvalidTokenError{ strToken }
    }
}