// decode-hostname is a command line tool for decoding stateless dns server names
package main

import (
	"fmt"
	"log"
	"os"

	docopt "github.com/docopt/docopt-go"
	"github.com/taskcluster/stateless-dns-go/hostname"
)

var (
	version = "decode-hostname 1.0.6"
	usage   = `
Usage:
  decode-hostname --fqdn FQDN --subdomain SUBDOMAIN --secret SECRET
  decode-hostname --help|-h
  decode-hostname --version

Exit Codes:
   0: Success
   1: Unrecognised command line options

Example:
  $ decode-hostname --fqdn aebagbaaaaadqfbf6nanb2v3zyzdeq27biltfievlqaktog2.foo.com --subdomain foo.com --secret 'Happy Birthday Pete!'
  2017/01/10 18:50:08 IP: 1.2.3.4
  2017/01/10 18:50:08 Expires: 1977-08-19 16:30:00 +0000 UTC
  2017/01/10 18:50:08 Salt: [2]uint8{0xd0, 0xea}
`
)

func main() {

	arguments, err := docopt.Parse(usage, nil, true, version, false, true)
	if err != nil {
		fmt.Println("Error parsing command line arguments!")
		os.Exit(1)
	}

	fqdn := arguments["FQDN"].(string)
	subdomain := arguments["SUBDOMAIN"].(string)
	secret := arguments["SECRET"].(string)

	ip, expires, salt, err := hostname.Decode(fqdn, secret, subdomain)

	if err != nil {
		log.Fatalf("Error occured: %v", err)
	}

	log.Printf("IP: %v", ip)
	log.Printf("Expires: %v", expires)
	log.Printf("Salt: %#v", salt)
}
