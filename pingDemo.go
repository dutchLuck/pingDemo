/*
 * P I N G D E M O . G O
 *
 * Ping a networked device
 *
 * Use the IP4 icmp socket to send an ICMP Echo request
 *
 * Last Modified on Sun Mar  7 20:15:13 2021
 *
 */

/*
 * 0v1 Better handling of timeout error messages.
 * 0v0 Code now times-out when device doesn't respond.
 *     Flattened error checks.
 *
 * This demo code originated from the ideas in section 3.11
 * of the "Socket-level Programming" chapter of the
 * Jan Newmarch ebook; -
 * https://jan.newmarch.name/go/socket/chapter-socket.html
 * I couldn't get the ping code published in the ebook to
 * work for me, but after some editing it began working.
 * 
 */

/*
 * At least "go version go1.15.6 windows/amd64" can compile the code; -
 *  go build pingDemo.go
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
    startTime := time.Now()
	if len(os.Args) != 2 {
		fmt.Println("?? Please specify the name or IP4 address of the device to ping?")
		fmt.Println(" Usage: ", os.Args[0], "name_Or_IP4NumbersOfDeviceToPing")
		os.Exit(1)
	}
	fmt.Println("Welcome to ping demo 0v1, compiled go code")
	fmt.Println("The (Start) time is", startTime )
	addr, err := net.ResolveIPAddr("ip4", os.Args[1])
	if err != nil {
		fmt.Println("Address Resolution error", err.Error())
		os.Exit(1)
	}
	timeOut, err := time.ParseDuration("1s")
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
			fmt.Println(wrtLen, " bytes sent and send ICMP Echo Request failed", err.Error()) // handle send error
		} else {
			deadLineTime := time.Now().Add(5e9) //set 5 sec timeout?
			var icmpReply [512]byte
			conn.SetReadDeadline(deadLineTime)
			rdLen, err := conn.Read(icmpReply[0:])
			if err != nil {
                if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
                    fmt.Println(os.Args[1], "(", addr.String(), ") did not respond in 5 sec:")
                } else {
				    fmt.Println("Ping failed:", rdLen, " bytes read:")
                    fmt.Println(err.Error()) // display error string
                }
			} else {
				okFlag := checkReplyLengthIsEqualOrLongerThanICMP_HdrLength(rdLen)
				okFlag = okFlag && checkReplyChecksum(checkSum(icmpReply[rdLen-8:rdLen]))
				okFlag = okFlag && checkICMP_ReplyIsTypeEchoReply(icmpReply[rdLen-8])
				okFlag = okFlag && checkICMP_Code(icmpReply[rdLen-7])
				replyPid := binary.BigEndian.Uint16(icmpReply[rdLen-4 : rdLen-2])
				okFlag = okFlag && checkPid(replyPid, pid)
				replySeq := binary.BigEndian.Uint16(icmpReply[rdLen-2 : rdLen])
				okFlag = okFlag && checkSeq(replySeq, seq)
				if okFlag {
					fmt.Println(os.Args[1], "(", addr.String(), ") is Alive")
				} else {
					fmt.Printf("% X\n", icmpReply[:rdLen])
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
	return uint16(^sum)
}

func checkSeq(replySeq uint16, seq uint16) bool {
	if replySeq != seq {
		fmt.Printf("ICMP Sequence sent (0x%x) doesn't match received (0x%x)\n", seq, replySeq)
		return false
	}
	return true
}

func checkPid(replyPid uint16, pid uint16) bool {
	if replyPid != pid {
		fmt.Printf("ICMP Identifier sent (0x%x) doesn't match received (0x%x)\n", pid, replyPid)
		return false
	}
	return true
}

func checkICMP_ReplyIsTypeEchoReply(replyType byte) bool {
	if replyType != 0 {
		fmt.Println("Wrong ICMP type, (type: ", replyType, ") returned")
		return false
	}
	return true
}

func checkICMP_Code(replyCode byte) bool {
	if replyCode != 0 {
		fmt.Println("ICMP code: ", replyCode, "), rather than expected 0, returned")
		return true //May still be ok, but is non-standard
	}
	return true
}

func checkReplyChecksum(replyChecksum uint16) bool {
	if replyChecksum != 0 {
		fmt.Println("ICMP Reply checksum check failed (value: ", replyChecksum, ")")
		return false
	}
	return true
}

func checkReplyLengthIsEqualOrLongerThanICMP_HdrLength(replyLength int) bool {
	if replyLength < 8 {
		fmt.Println("ICMP Echo Reply is too short. It should be at least 8 bytes long. It was ", replyLength, " bytes.")
		return false
	}
	return true
}
