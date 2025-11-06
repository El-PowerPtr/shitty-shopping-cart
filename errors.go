package main

import "fmt"

type AlreadyExistingCartError struct {
    id uint64
}

func (e *AlreadyExistingCartError) Error() string {
    return fmt.Sprintf("The cart #%d already exists!\n", e.id)
}

type DiffOwnerError struct {
    cart uint64
    reqOwner uint64
    actualOwner uint64
}

func (e *DiffOwnerError) Error() string {
    return fmt.Sprintf("The cart #%d has user #%d as owner, but user #%d is making the request.\n", e.cart, e.actualOwner, e.reqOwner)
}