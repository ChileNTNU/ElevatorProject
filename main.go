package main

import(
	"fmt"
	"time"
//	"os"
	"./Network"
)

func main(){
	fmt.Println("Hello!")
	go network.Listener()

	for{
		fmt.Println("Doing something else")
		time.Sleep(3000 * time.Millisecond)
	}

}

