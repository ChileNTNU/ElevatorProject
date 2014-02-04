package main

import(
	"fmt"
	"time"
//	"os"
	"./network"
)

func main(){
	fmt.Println("Hello!")
	go network.Listener()
	for{
		fmt.Println("Doing something else")
		time.Sleep(3000 * time.Millisecond)
	}
}

