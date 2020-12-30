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
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	fmt.Println("Welcome to the ping demo")
	fmt.Println("The (Start) time is", time.Now())
	if len(os.Args) != 2 {
		fmt.Println("Usage: ", os.Args[0], "nameOrIP4NumbersOfDeviceToPing")
		os.Exit(1)
	}
	addr, err := net.ResolveIPAddr("ip4", os.Args[1])
	if err != nil {
		fmt.Println("Address Resolution error", err.Error())
		os.Exit(1)
	}
	timeOut, err := time.ParseDuration("100ms")
	if err != nil {
		fmt.Println("Duration error", err.Error())
		os.Exit(1)
	}
	conn, err := net.DialTimeout("ip4:icmp", addr.String(), timeOut)
	if err != nil {
		fmt.Println("DialTimeout failed", err.Error()) // handle error
	} else {
		var pid uint16 = uint16(os.Getpid() & 0xffff) //Use process id as the ICMP identifier
		var seq uint16 = uint16(1234)                 //create ICMP Sequence number
		var icmpHdr [16]byte
		icmpHdr[0] = 8                                // ICMP Echo Request
		icmpHdr[1] = 0                                // Code 0
		binary.BigEndian.PutUint16(icmpHdr[2:4], 0)   //ICMP Checksum, need 0 for initial calc of checksum
		binary.BigEndian.PutUint16(icmpHdr[4:6], pid) //ICMP Identifier
		binary.BigEndian.PutUint16(icmpHdr[6:8], seq) //ICMP Sequence
		icmpHdrChkSum := checkSum(icmpHdr[0:8])
		binary.BigEndian.PutUint16(icmpHdr[2:4], icmpHdrChkSum) //Set ICMP Checksum for Echo Request
		wrtLen, err := conn.Write(icmpHdr[0:8])
		if err != nil {
			fmt.Println("Send ICMP failed", err.Error()) // handle error
		} else {
			fmt.Println("Write Length is ", wrtLen)
			var icmpReply [512]byte
			conn.SetReadDeadline(time.Now().Add(2e9))		//2 sec timeout?
			rdLen, err := conn.Read(icmpReply[0:])
			if err != nil {
				fmt.Println("Read ICMP failed", err.Error()) // handle error
			} else {
				fmt.Println("Read Length is ", rdLen)
				replyChkSum := checkSum(icmpReply[rdLen-8 : rdLen])
				if replyChkSum != 0 {
					fmt.Println("ICMP Reply checksum failed (value: ", replyChkSum, ")")
					fmt.Printf("% X\n", icmpReply[:rdLen])
				} else {
					if icmpReply[rdLen-8] != 0 {
						fmt.Println("Wrong ICMP type, (type: ", icmpReply[rdLen-8], ") returned")
						fmt.Printf("% X\n", icmpReply[:rdLen])
					} else {
						replyPid := binary.BigEndian.Uint16(icmpReply[rdLen-4 : rdLen-2])
						if replyPid != pid {
							fmt.Printf("ICMP Identifier sent (0x%x) doesn't match received (0x%x)\n", pid, replyPid)
							fmt.Printf("% X\n", icmpReply[:rdLen])
						} else {
							replySeq := binary.BigEndian.Uint16(icmpReply[rdLen-2 : rdLen])
							if replySeq != seq {
								fmt.Printf("ICMP Sequence sent (0x%x) doesn't match received (0x%x)\n", seq, replySeq)
								fmt.Printf("% X\n", icmpReply[:rdLen])
							} else {
								fmt.Println(os.Args[1], "(", addr.String(), ") is Alive")
							}
						}
					}
				}
			}
		}
		conn.Close()
	}
	fmt.Println("The (Finish) time is", time.Now())
	os.Exit(0)
}

func checkSum(msg []byte) uint16 {
	sum := 0

	// An even number of bytes is assumed
	for n := 0; n < len(msg)-1; n += 2 {
		sum += int(msg[n])*256 + int(msg[n+1])
	}
	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}
