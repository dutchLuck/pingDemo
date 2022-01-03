/*
 * P I N G D E M O . G O
 *
 * Ping one or more networked devices
 *
 * Send an ICMP Echo request and look for an ICMP echo reply
 *
 * Last Modified on Mon Jan  3 22:23:37 2022
 *
 */

/*
 * 0v3 Increase size of sent ICMP echo reqest packet & more Debug output
 * 0v2 Better command line option handling: -D(ebug), -v(erbose) -w(ait) <duration>
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
 * At least "go version" go1.17.5 "windows/amd64" can compile this code; -
 *  go build pingDemo.go
 *
 */

package main

import (
	"encoding/binary"   // BigEndian.PutUint16()
	"flag"  // command line options package: Parse(), Args()
	"fmt"   // Printf(), Println()
	"net"   // ResolveIPAddr(), DialTimeout()
	"os"    // Exit()
	"time"  // Now(), ParseDuration()
)

const icmpSize = 28 // Must not be smaller than 8, which is the ICMP Header size

func main() {
	startTime := time.Now()
	verboseFlag := flag.Bool("v", false, "Turn on verbose output")
	debugFlag := flag.Bool("D", false, "Turn on Debug output")
	waitStr := flag.String("w", "2", "Timeout wait time in seconds")
	flag.Parse()
	ipAddresses := flag.Args()
	if *verboseFlag || *debugFlag {
		fmt.Println("Welcome to ping demo 0v3, compiled go code")
		fmt.Println("The (Start) time is", startTime)
	}
	if *debugFlag {
		fmt.Println("Debug: Debug flag is true")
		fmt.Println("Debug: verbose flag is", *verboseFlag)
		fmt.Println("Debug: wait timeout value is", *waitStr, "seconds")
		fmt.Println("Debug:", len(ipAddresses), "positional args", ipAddresses)
	}
	if len(ipAddresses) < 1 {
		fmt.Println("?? Please specify the name or IP4 address of the device to ping?")
		fmt.Println(" E.g.: ", os.Args[0], "[-D] [-v] [-w Int] name_Or_IP4NumbersOfDeviceToPing")
		flag.Usage()
		os.Exit(1)
	}
	sktTimeOut, err := time.ParseDuration("1s") //Probably over-kill, but create 1 sec time out value
	if err != nil {
		fmt.Println("?? Create Socket Time Out duration error", err.Error())
		os.Exit(2)
	}
	replyTimeOut, err := time.ParseDuration(*waitStr + "s") //convert sec time out value
	if err != nil {
		fmt.Println("?? -w <duration> Reply Time Out duration error", err.Error())
		os.Exit(3)
	}
	if *debugFlag {
		fmt.Println("Debug: Timeout for ICMP socket creation is", sktTimeOut)
		fmt.Println("Debug: Timeout for ICMP echo reply is", replyTimeOut)
	}
	/*
	 * Loop through internet device names or IP addresses and ping each one
	 */
	for _, ipAddr := range ipAddresses {
		addr, err := net.ResolveIPAddr("ip4", ipAddr)
		if err != nil {
			fmt.Println("??", ipAddr, "IPv4 Address Resolution error:", err.Error())
		} else {
			conn, err := net.DialTimeout("ip4:icmp", addr.String(), sktTimeOut)
			if err != nil {
				fmt.Println("??", ipAddr, "DialTimeout() function failed:", err.Error()) // handle error
			} else {
				// Successfully created a Network socket, so now setup ICMP Header + Data
				var pid uint16 = uint16(os.Getpid() & 0xffff) // Use process id as the ICMP identifier
				var seq uint16 = uint16(12345)                // create ICMP Sequence number
				var icmpHdr [icmpSize]byte
				// Fill in ICMP Header except for the checksum
				icmpHdr[0] = 8                                // ICMP Echo Request
				icmpHdr[1] = 0                                // Code 0
				binary.BigEndian.PutUint16(icmpHdr[2:4], 0)   // ICMP Checksum, need 0 for initial calc of checksum
				binary.BigEndian.PutUint16(icmpHdr[4:6], pid) // ICMP Identifier
				binary.BigEndian.PutUint16(icmpHdr[6:8], seq) // ICMP Sequence
				// Initialize ICMP data
				for indx := 8; indx < icmpSize; indx++ {
					icmpHdr[indx] = byte((indx - 8) & 0xff)
				}
				// Calculate ICMP checksum and place it in the ICMP header
				icmpHdrChkSum := checkSum(icmpHdr[0:icmpSize])
				binary.BigEndian.PutUint16(icmpHdr[2:4], icmpHdrChkSum) //Set ICMP Checksum for Echo Request
				// Send ICMP echo to remote device
				wrtLen, err := conn.Write(icmpHdr[0:icmpSize])
				if err != nil {
					fmt.Println("??", wrtLen, " bytes sent and send ICMP Echo Request failed:", err.Error()) // handle send error
				} else {
					// Send ICMP was successful so now look for a corresponding reply
					if *debugFlag {
						fmt.Println("Debug:", wrtLen, "bytes sent")
						fmt.Printf("Debug: % X\n", icmpHdr[0:icmpSize])
					}
					deadLineTime := time.Now().Add(replyTimeOut) //set -w sec timeout (defaults to 2 sec)
					var icmpReply [512]byte
					conn.SetReadDeadline(deadLineTime)
					rdLen, err := conn.Read(icmpReply[0:])
					if err != nil {
						if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
							fmt.Println("??", ipAddr, "(", addr.String(), ") did not respond in", *waitStr, "sec:")
						} else {
							fmt.Println("?? Ping failed:", rdLen, " bytes read:")
							fmt.Println(err.Error()) // display error string
						}
					} else {
						if *debugFlag {
							fmt.Println("Debug:", rdLen, "bytes (Standard IP Header = 20 + ICMP) received in reply")
							fmt.Printf("Debug: % X\n", icmpReply[20:rdLen])
						}
						okFlag := checkReplyLengthIsEqualOrLongerThanICMP_HdrLength(rdLen)
						okFlag = okFlag && checkReplyChecksum(checkSum(icmpReply[rdLen-icmpSize:rdLen]))
						okFlag = okFlag && checkICMP_ReplyIsTypeEchoReply(icmpReply[rdLen-icmpSize])
						okFlag = okFlag && checkICMP_Code(icmpReply[rdLen-icmpSize+1])
						replyPid := binary.BigEndian.Uint16(icmpReply[rdLen-icmpSize+4 : rdLen-icmpSize+6])
						okFlag = okFlag && checkPid(replyPid, pid)
						replySeq := binary.BigEndian.Uint16(icmpReply[rdLen-icmpSize+6 : rdLen-icmpSize+8])
						okFlag = okFlag && checkSeq(replySeq, seq)
						if okFlag {
							fmt.Println(ipAddr, "(", addr.String(), ") is Alive")
						} else {
							fmt.Printf("IP packet was: % X\n", icmpReply[:rdLen])
						}
					}
				}
				conn.Close()
			}
		}
	}
	if *verboseFlag || *debugFlag {
		fmt.Println("The (Finish) time is", time.Now())
	}
	os.Exit(0)
}

func checkSum(msg []byte) uint16 {
	lastIndex := len(msg) - 1
	sum := 0
	// An even number of bytes is assumed
	for n := 0; n < lastIndex; n += 2 {
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
