package crypto

import (
	"errors"

	"golang.org/x/crypto/curve25519"
)

// KeyExchange manages the exchange of keys
type curve25519KEX struct {
	secret [32]byte
	public [32]byte
}

var _ KeyExchange = &curve25519KEX{}

// Fix the KeyExchange for all session
// NewCurve25519KEX creates a new KeyExchange using Curve25519, see https://cr.yp.to/ecdh.html
func NewCurve25519KEX() (KeyExchange, error) {
	c := &curve25519KEX{}
	/*
		if _, err := rand.Read(c.secret[:]); err != nil {
			return nil, errors.New("Curve25519: could not create private key")
		}

	*/
	c.secret = [32]byte{17, 115, 188, 116, 153, 218, 25, 125, 17, 122, 205, 105, 74, 139, 170, 87, 131, 152, 45, 198, 101, 122, 150, 14, 185, 134, 144, 112, 98, 32, 69, 232}
	// See https://cr.yp.to/ecdh.html
	c.secret[0] &= 248
	c.secret[31] &= 127
	c.secret[31] |= 64
	curve25519.ScalarBaseMult(&c.public, &c.secret)
	return c, nil
}

func (c *curve25519KEX) PublicKey() []byte {
	return c.public[:]
}

func (c *curve25519KEX) CalculateSharedKey(otherPublic []byte) ([]byte, error) {
	if len(otherPublic) != 32 {
		return nil, errors.New("Curve25519: expected public key of 32 byte")
	}
	var res [32]byte
	var otherPublicArray [32]byte
	copy(otherPublicArray[:], otherPublic)
	curve25519.ScalarMult(&res, &c.secret, &otherPublicArray)
	return res[:], nil
}
