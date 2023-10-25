package controllers 

import (
	_ "fmt"
)

type stateMachine struct {
	currentState string 
	value string 
	actions map[string]func()
	transitions map[string]Transition
}

type Transition struct {
	target string 
	action func()
}

