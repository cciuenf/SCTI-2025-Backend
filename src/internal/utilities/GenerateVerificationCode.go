package utilities

import (
	"crypto/rand"
)

func GenerateVerificationCode() int {
	min := 100000
	max := 999999

	randomNumber, err := cryptoRand(min, max)
	if err != nil {

	}
	return randomNumber
}

func cryptoRand(min, max int) (int, error) {
	delta := max - min + 1

	buf := make([]byte, 4)
	_, err := rand.Read(buf)
	if err != nil {
		return 0, err
	}

	randomUint32 := uint32(buf[0]) | uint32(buf[1])<<8 | uint32(buf[2])<<16 | uint32(buf[3])<<24

	return min + int(randomUint32%uint32(delta)), nil
}
