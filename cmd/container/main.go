package main ;


import (
	"fmt"
	"os"
)

func main(){

	switch os.Args[1] {

	case "run": 
		fmt.Println("Running container")
	default : 
		fmt.Println("Command not found")
	}



}

