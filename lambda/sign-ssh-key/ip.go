package main

import (
	"fmt"
	"net"

	"github.com/pkg/errors"
)

func ValidateIPRanges(inner, outer []string) ([]string, error) {
	innerIP, err := ParseCIDRS(inner)
	if err != nil {
		return nil, err
	}

	outerIP, err := ParseCIDRS(outer)
	if err != nil {
		return nil, err
	}

	if len(innerIP) == 0 {
		return outer, nil
	}

	if len(outerIP) == 0 {
		return inner, nil
	}

	for i, in := range innerIP {
		found := false
		for _, out := range outerIP {
			if intersect(out, in) {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New(fmt.Sprintf("%s is not a subset of allowed source addresses.", inner[i]))
		}
	}

	return inner, nil
}

func ParseCIDRS(input []string) ([]*net.IPNet, error) {
	res := make([]*net.IPNet, len(input))
	for i, cidr := range input {
		_, value, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, err
		}
		res[i] = value
	}
	return res, nil
}

// https://stackoverflow.com/a/34729993
func intersect(sup, sub *net.IPNet) bool {
	if len(sup.IP) != len(sub.IP) {
		return false
	}

	for i := range sup.IP {
		if sup.IP[i]&sup.Mask[i] != sub.IP[i]&sub.Mask[i]&sup.Mask[i] {
			return false
		}
	}
	return true
}
