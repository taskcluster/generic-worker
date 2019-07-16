package hostname_test

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"github.com/taskcluster/stateless-dns-go/hostname"
)

func ExampleNew() {
	ip := net.IPv4(byte(203), byte(43), byte(55), byte(2))
	subdomain := "foo.com"
	expires := time.Now().Add(15 * time.Minute)
	secret := "turnip4tea"
	fmt.Println(hostname.New(ip, subdomain, expires, secret))
}

func TestEncodeDecode(t *testing.T) {
	ip := net.IPv4(byte(203), byte(43), byte(55), byte(2))
	subdomain := "foo.com"
	expires := time.Now().Add(15 * time.Minute)
	secret := "turnip4tea"
	// encode
	fqdn := hostname.New(ip, subdomain, expires, secret)
	// decode
	ip2, expires2, _, err := hostname.Decode(fqdn, secret, subdomain)
	if err != nil {
		t.Fatalf("Error when creating hostname: %v", err)
	}
	if ip.String() != ip2.String() {
		t.Fatalf("Incorrect IP - got %v but was expecting %v", ip, ip2)
	}
	if expires.Unix() != expires2.Unix() {
		t.Fatalf("Incorrect Expiry - got %v but was expecting %v", expires, expires2)
	}
}

func ExampleDecode() {
	ip, expires, salt, err := hostname.Decode("zmvtoaqaaaavkjlja2i2n2ligiol2idykqa3t7vk4vfakdv6.foo.com", "turnip4tea", "foo.com")
	if err != nil {
		log.Fatalf("Not able to decode example hostname")
	}
	fmt.Println(ip)
	fmt.Println(expires)
	fmt.Println(salt)
	// Output:
	// 203.43.55.2
	// 2016-06-06 11:11:27.889 +0000 UTC
	// [166 233]
}

func TestBadName(t *testing.T) {
	_, _, _, err := hostname.Decode("zmvtoaqaaaavkjlja2i2n2ligiol2idykqa3t7vk4vfakdw6.foo.com", "turnip4tea", "foo.com")
	if err == nil {
		log.Fatalf("Was expecting an error as hash is invalid")
	}
}
