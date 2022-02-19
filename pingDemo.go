/*
 * P I N G D E M O . G O
 *
 * Ping one or more networked devices
 *
 * Send an ICMP Echo request and look for an ICMP echo reply
 *
 * Last Modified on Sat Feb 19 16:54:32 2022
 *
 */

/*
 * 0v5 Specify duraton between multiple pings with pause option -p X.X
 * 0v4 Output ping response delay & Specify multiple pings of the same host with count option -c X 
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
	"encoding/binary" // BigEndian.PutUint16()
	"flag"            // command line options package: Parse(), Args()
	"fmt"             // Printf(), Println()
	"net"             // ResolveIPAddr(), DialTimeout()
	"os"              // Exit()
	"time"            // Now(), ParseDuration()
)

const icmpSize = 28 // Must not be smaller than 8, which is the ICMP Header size

func main() {
	mainStartTime := time.Now()
	countInt := flag.Int("c", 1, "Count number of Echo Requests for each host")
	debugFlag := flag.Bool("D", false, "Turn on Debug output")
	pauseStr := flag.String("p", "0.1", "Pause between Echo requests in seconds")  
	verboseFlag := flag.Bool("v", false, "Turn on verbose output")
	waitStr := flag.String("w", "2", "Individual Echo Reply Timeout wait time in seconds")
	flag.Parse()
	ipAddresses := flag.Args()
	if *verboseFlag || *debugFlag {
		fmt.Println("Welcome to ping demo 0v5, compiled go code")
		fmt.Println("The (Start) time is", mainStartTime)
	}
	if *debugFlag {
		fmt.Println("Debug: count value is ", *countInt)
		fmt.Println("Debug: Debug flag is true")
		fmt.Println("Debug: verbose flag is", *verboseFlag)
		fmt.Println("Debug: pause value is", *pauseStr, "seconds")
		fmt.Println("Debug: wait timeout value is", *waitStr, "seconds")
		fmt.Println("Debug:", len(ipAddresses), "positional args", ipAddresses)
	}
	if len(ipAddresses) < 1 {
		fmt.Println("?? Please specify the name or IP4 address of the device to ping?")
		fmt.Println(" E.g.: ", os.Args[0], "[-c Int] [-D] [-p Float] [-v] [-w Float] name_Or_IP4NumbersOfDeviceToPing")
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
	pauseDuration, err := time.ParseDuration(*pauseStr + "s") //convert pause beween multiple ping sec value
	if err != nil {
		fmt.Println("?? -p <duration> Pause duration error", err.Error())
		os.Exit(4)
	}
	if *debugFlag {
		fmt.Println("Debug: Timeout for ICMP socket creation is", sktTimeOut)
		fmt.Println("Debug: Timeout for ICMP echo reply is", replyTimeOut)
		fmt.Println("Debug: Period between ICMP echo reply is", pauseDuration)
	}
	/*
	 * Loop through internet device names or IP addresses and ping each one count times
	 */
	icmpSeq := int(12345)                         // create ICMP Sequence number
	var pid uint16 = uint16(os.Getpid() & 0xffff) // Use process id as the ICMP identifier
	var icmpPkt [icmpSize]byte
	var icmpReply [512]byte
	for _, ipAddr := range ipAddresses {
		addr, err := net.ResolveIPAddr("ip4", ipAddr)
		if err != nil {
			fmt.Println("??", ipAddr, "IPv4 Address Resolution error:", err.Error())
		} else {
			for indx := 0; indx < *countInt; indx++ {
				// ?? Seems that a new timed socket connection is required for each loop
				// as the time-outs seem ?? to get in the way of reusing the old socket
				conn, err := net.DialTimeout("ip4:icmp", addr.String(), sktTimeOut)
				if err != nil {
					fmt.Println("??", ipAddr, "DialTimeout() function failed:", err.Error()) // handle error
				} else {
					// Successfully created a Network socket, so now setup ICMP Header + Data
					// Fill in ICMP Header except for the checksum
					icmpPkt[0] = 8                                // ICMP Echo Request
					icmpPkt[1] = 0                                // Code 0
					binary.BigEndian.PutUint16(icmpPkt[2:4], 0)   // ICMP Checksum, need 0 for initial calc of checksum
					binary.BigEndian.PutUint16(icmpPkt[4:6], pid) // ICMP Identifier
					// Initialize ICMP data
					for index := 8; index < icmpSize; index++ {
						icmpPkt[index] = byte((index - 8) & 0xff)
					}
					binary.BigEndian.PutUint16(icmpPkt[6:8], uint16(icmpSeq&0xFFFF)) // ICMP Sequence increases with each attempt
					// Calculate ICMP checksum and place it in the ICMP header
					icmpPktChkSum := checkSum(icmpPkt[0:icmpSize])
					binary.BigEndian.PutUint16(icmpPkt[2:4], icmpPktChkSum) //Set ICMP Checksum for Echo Request
					startTime := time.Now()
					// Send ICMP echo to remote device
					wrtLen, err := conn.Write(icmpPkt[0:icmpSize])
					if err != nil {
						fmt.Println("??", wrtLen, " bytes sent and send ICMP Echo Request failed:", err.Error()) // handle send error
					} else {
						// Send ICMP was successful so now look for a corresponding reply
						if *debugFlag {
							fmt.Println("Debug:", wrtLen, "bytes sent")
							fmt.Printf("Debug: % X\n", icmpPkt[0:icmpSize])
						}
						deadLineTime := time.Now().Add(replyTimeOut) //set -w sec timeout (defaults to 2 sec)
						conn.SetReadDeadline(deadLineTime)
						rdLen, rdErr := conn.Read(icmpReply[0:])
						elapsedTime := time.Since(startTime)
						if rdErr != nil {
							if nerr, ok := rdErr.(net.Error); ok && nerr.Timeout() {
								fmt.Println("??", ipAddr, "(", addr.String(), ") did not respond in", elapsedTime)
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
							okFlag = okFlag && checkSeq(replySeq, uint16(icmpSeq&0xFFFF))
							if okFlag {
								fmt.Println(ipAddr, "(", addr.String(), ") responded in", elapsedTime)
							} else {
								fmt.Printf("?? Unanticipated reponse: IP packet was:\n % X\n", icmpReply[:rdLen])
							}
						} // end of succesful read from socket
						icmpSeq += 1	// bump up the ICMP sequence number each time after each ping
						// Only delay for 1 second if the same host will be pinged again
						if (indx + 1) < *countInt {
							time.Sleep(pauseDuration)
						}
					} // end of successful write to socket
					conn.Close() // close the time-out connection after one use
				} // end of successfull time-out socket creation
			} // end of for-loop through count number of pings
		} // end of successful IPv4 DNS lookup
	} // end of for-loop to do each host
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
