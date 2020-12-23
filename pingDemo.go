/*
 * P I N G D E M O
 *
 */

/*
 * Use the IP4 raw socket to send an ICMP Echo request
 *
 */

package main

import (
	"fmt"
	"time"
)

func main() {
	  fmt.Println("Welcome to the ping demo")
    DialTimeout("ip4:icmp", "127.0.0.1", 5 )
	  fmt.Println("The time is", time.Now())
}
