// create-hostname is a command line tool for generating stateless dns server names
package main

import (
	"fmt"
	"net"
	"os"
	"time"

	docopt "github.com/docopt/docopt-go"
	"github.com/taskcluster/stateless-dns-go/hostname"
)

var (
	version = "create-hostname 1.0.6"
	usage   = `
Usage:
  create-hostname --ip IP --subdomain SUBDOMAIN --expires EXPIRES --secret SECRET

Exit Codes:
   0: Success
   1: Unrecognised command line options
  64: Invalid IP given
  65: IP given was an IPv6 IP (IP should be an IPv4 IP)
  66: Invalid SUBDOMAIN given
  67: Invalid EXPIRES given
  68: Invalid SECRET given
  69: Some other problem

Example:
  $ create-hostname --ip 203.115.35.2 --subdomain foo.com --expires 2016-06-04T16:04:03.739Z --secret 'cheese monkey'
  znzsgaqaau2hl7h35f4owqn25s76j4h7apm3fe4qpy6pfxjk.foo.com
`
)

func main() {

	arguments, err := docopt.Parse(usage, nil, true, version, false, true)
	if err != nil {
		fmt.Println("Error parsing command line arguments!")
		os.Exit(1)
	}

	// Validate IP
	ipString := arguments["IP"].(string)
	ip := net.ParseIP(ipString)
	if ip == nil {
		fmt.Fprintf(os.Stderr, "create-hostname: ERR 64: Invalid IP '%s'\n", ipString)
		os.Exit(64)
	}
	ip = ip.To4()
	if ip == nil {
		fmt.Fprintf(os.Stderr, "create-hostname: ERR 65: IPv6 given for IP (should be IPv4) '%s'\n", ipString)
		os.Exit(65)
	}

	// TODO: Validate SUBDOMAIN
	subdomain := arguments["SUBDOMAIN"].(string)

	// Validate EXPIRES
	expiresString := arguments["EXPIRES"].(string)
	expires, err := time.Parse(time.RFC3339Nano, expiresString)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create-hostname: ERR 67: Invalid EXPIRES '%s'\n", expiresString)
		os.Exit(67)
	}

	// TODO: Validate SECRET
	secret := arguments["SECRET"].(string)

	name := hostname.New(ip, subdomain, expires, secret)
	fmt.Println(name)
}
