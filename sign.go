// Apla Software includes an integrated development
// environment with a multi-level system for the management
// of access rights to data, interfaces, and Smart contracts. The
// technical characteristics of the Apla Software are indicated in
// Apla Technical Paper.

// Apla Users are granted a permission to deal in the Apla
// Software without restrictions, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of Apla Software, and to permit persons
// to whom Apla Software is furnished to do so, subject to the
// following conditions:
// * the copyright notice of GenesisKernel and EGAAS S.A.
// and this permission notice shall be included in all copies or
// substantial portions of the software;
// * a result of the dealing in Apla Software cannot be
// implemented outside of the Apla Platform environment.

// THE APLA SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY
// OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED
// TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A
// PARTICULAR PURPOSE, ERROR FREE AND NONINFRINGEMENT. IN
// NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR
// THE USE OR OTHER DEALINGS IN THE APLA SOFTWARE.

package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/ioutil"
	"math/big"

	"github.com/google/logger"
)

var (
	nodePrivateKey []byte
	nodePublicKey  []byte
)

// FillLeft is filling slice
func FillLeft(slice []byte) []byte {
	if len(slice) >= 32 {
		return slice
	}
	return append(make([]byte, 32-len(slice)), slice...)
}

func PrivateToPublic(key []byte) []byte {
	var pubkeyCurve elliptic.Curve
	pubkeyCurve = elliptic.P256()

	bi := new(big.Int).SetBytes(key)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi
	priv.PublicKey.X, priv.PublicKey.Y = pubkeyCurve.ScalarBaseMult(key)
	return append(FillLeft(priv.PublicKey.X.Bytes()), FillLeft(priv.PublicKey.Y.Bytes())...)
}

func InitNodeKey() {
	nodeKey, err := ioutil.ReadFile(cfg.Settings.NodePrivateKey)
	if err != nil {
		logger.Fatal(err)
	}
	if nodePrivateKey, err = hex.DecodeString(string(nodeKey)); err != nil {
		logger.Fatal(err)
	}
	nodePublicKey = PrivateToPublic(nodePrivateKey)
}

func SignECDSA(data []byte) (ret []byte, err error) {
	var pubkeyCurve elliptic.Curve

	pubkeyCurve = elliptic.P256()

	bi := new(big.Int).SetBytes(nodePrivateKey)
	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = pubkeyCurve
	priv.D = bi

	signhash := sha256.Sum256(data)
	r, s, err := ecdsa.Sign(crand.Reader, priv, signhash[:])
	if err != nil {
		return
	}
	ret = append(FillLeft(r.Bytes()), FillLeft(s.Bytes())...)
	return
}

func CheckECDSA(data, signature []byte) (bool, error) {
	if len(signature) == 0 {
		return false, nil
	}

	var pubkeyCurve elliptic.Curve
	pubkeyCurve = elliptic.P256()

	hash := sha256.Sum256(data)

	pubkey := new(ecdsa.PublicKey)
	pubkey.Curve = pubkeyCurve
	pubkey.X = new(big.Int).SetBytes(nodePublicKey[:32])
	pubkey.Y = new(big.Int).SetBytes(nodePublicKey[32:])
	r, s, err := parseSign(hex.EncodeToString(signature))
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(pubkey, hash[:], r, s), nil
}

// parseSign converts the hex signature to r and s big number
func parseSign(sign string) (*big.Int, *big.Int, error) {
	var (
		binSign []byte
		err     error
	)
	//	var off int
	parse := func(bsign []byte) []byte {
		blen := int(bsign[1])
		if blen > len(bsign)-2 {
			return nil
		}
		ret := bsign[2 : 2+blen]
		if len(ret) > 32 {
			ret = ret[len(ret)-32:]
		} else if len(ret) < 32 {
			ret = append(bytes.Repeat([]byte{0}, 32-len(ret)), ret...)
		}
		return ret
	}
	if len(sign) > 128 {
		binSign, err = hex.DecodeString(sign)
		if err != nil {
			return nil, nil, err
		}
		left := parse(binSign[2:])
		if left == nil || int(binSign[3])+6 > len(binSign) {
			return nil, nil, errors.New(`wrong left parsing`)
		}
		right := parse(binSign[4+binSign[3]:])
		if right == nil {
			return nil, nil, errors.New(`wrong right parsing`)
		}
		sign = hex.EncodeToString(append(left, right...))
	} else if len(sign) < 128 {
		return nil, nil, fmt.Errorf(`wrong len of signature %d`, len(sign))
	}
	all, err := hex.DecodeString(sign[:])
	if err != nil {
		return nil, nil, err
	}
	return new(big.Int).SetBytes(all[:32]), new(big.Int).SetBytes(all[len(all)-32:]), nil
}
