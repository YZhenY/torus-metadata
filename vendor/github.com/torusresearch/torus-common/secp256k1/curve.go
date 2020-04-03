package secp256k1

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/btcsuite/btcd/btcec"
	"github.com/torusresearch/torus-common/common"
	"golang.org/x/crypto/sha3"
)

type KoblitzCurve struct {
	*btcec.KoblitzCurve
}

var (
	Curve = &KoblitzCurve{btcec.S256()}
	// field order, also known as p, usually used for scalars
	FieldOrder = common.HexToBigInt("fffffffffffffffffffffffffffffffffffffffffffffffffffffffefffffc2f")
	// group order, also known as q, it is the number of points in the curve, and is usually used in exponents
	GeneratorOrder = common.HexToBigInt("fffffffffffffffffffffffffffffffebaaedce6af48a03bbfd25e8cd0364141")
	// scalar to the power of this is like square root, eg. y^sqRoot = y^0.5 (if it exists)
	SqRoot = common.HexToBigInt("3fffffffffffffffffffffffffffffffffffffffffffffffffffffffbfffff0c")
	G      = common.Point{X: *Curve.Gx, Y: *Curve.Gy}
	H      = *HashToPoint(G.X.Bytes())
)

func RandomBigInt() *big.Int {
	randomInt, _ := rand.Int(rand.Reader, GeneratorOrder)
	return randomInt
}

func HashToPoint(data []byte) *common.Point {
	keccakHash := Keccak256(data)
	x := new(big.Int)
	x.SetBytes(keccakHash)
	for {
		beta := new(big.Int)
		beta.Exp(x, big.NewInt(3), FieldOrder)
		beta.Add(beta, big.NewInt(7))
		beta.Mod(beta, FieldOrder)
		y := new(big.Int)
		y.Exp(beta, SqRoot, FieldOrder)
		if new(big.Int).Exp(y, big.NewInt(2), FieldOrder).Cmp(beta) == 0 {
			return &common.Point{X: *x, Y: *y}
		} else {
			x.Add(x, big.NewInt(1))
		}
	}
}

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewLegacyKeccak256()
	for _, b := range data {
		n, err := d.Write(b)
		if err != nil {
			fmt.Printf("Could not write n %v err %v", n, err)
		}
	}
	return d.Sum(nil)
}

func Encrypt(pubKey common.Point, input []byte) (encryptedOutput []byte, err error) {
	return btcec.Encrypt(&btcec.PublicKey{
		Curve: Curve,
		X:     &pubKey.X,
		Y:     &pubKey.Y,
	}, input)
}

func Decrypt(privKey big.Int, input []byte) (decryptedOutput []byte, err error) {
	return btcec.Decrypt(&btcec.PrivateKey{
		D: &privKey,
	}, input)
}
