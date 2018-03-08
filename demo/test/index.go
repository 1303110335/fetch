package test

import (
	"time"
	"fmt"
)

var chan1 chan int
var chanLength int = 18
var interval time.Duration = 1500 * time.Millisecond

func main() {
	chan1 = make(chan int, chanLength)
	//fmt.Printf("init chan1:%v\n", chan1)
	go func() {
		for i := 0; i < chanLength ; i++ {
			if i > 0 && i % 3 == 0 {
				close(chan1)
				chan1 = make(chan int, chanLength)
				//fmt.Printf("reset chan1:%v\n", chan1)
			}
			fmt.Printf("Send element %d...\n", i)
			chan1 <- i
			time.Sleep(interval)
		}
		fmt.Println("close chan1...")
		close(chan1)
	}()

	receive()
}

func getChan() chan int {
	//fmt.Printf("get chan1:%v\n", chan1)
	return chan1
}

func receive() {
	fmt.Println("Receive element from chan1...")
	timer := time.After(30 * time.Second)
	Loop:
		for {
			select {
			case e, ok := <-getChan():
				if !ok {
					fmt.Println("--Closed chan1.")
					break
				}
				fmt.Printf("Receive a element: %d\n", e)
				time.Sleep(interval)
			case <-timer:
				fmt.Println("Timeout!")
				break Loop
			}
		}
		fmt.Println("--End.")
}