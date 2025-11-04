package main

type Cart struct {
    OwnerId uint64
    Id uint64
    Items map[uint64]int
}