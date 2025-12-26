package main

type EvenManager interface {
    Buy(*Cart) error
}