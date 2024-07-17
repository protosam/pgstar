package modecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type privatekey struct {
	key *ecdsa.PrivateKey
}

func (privkey *privatekey) publicKey(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 0); err != nil {
		return starlark.None, err
	}

	x509publickeybytes, err := x509.MarshalPKIXPublicKey(&privkey.key.PublicKey)
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf(">>>> %s", err))}, nil
	}

	var curve string
	switch privkey.key.PublicKey.Curve {
	case elliptic.P224():
		curve = "P224"
	case elliptic.P256():
		curve = "P256"
	case elliptic.P384():
		curve = "P384"
	case elliptic.P521():
		curve = "P521"
	default:
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("unsupported curve from loaded private key: %s", curve))}, nil
	}

	return starlark.Tuple{
		starlarkPublicKeyStruct(privkey.key.PublicKey, curve, string(x509publickeybytes)),
		starlark.None,
	}, nil
}

func (privkey *privatekey) sign(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var messageHashSumBytes string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "messageSumBytes", &messageHashSumBytes); err != nil {
		return starlark.None, err
	}

	signature, err := ecdsa.SignASN1(rand.Reader, privkey.key, []byte(messageHashSumBytes))
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed signing message hashsum: %s", err))}, nil
	}

	return starlark.Tuple{starlark.String(string(signature)), starlark.None}, nil
}

func (privkey *privatekey) sharedSecret(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x509publickeybytes string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "x509publickeybytes", &x509publickeybytes); err != nil {
		return starlark.None, err
	}

	publickeyAny, err := x509.ParsePKIXPublicKey([]byte(x509publickeybytes))
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed reading private key bytes: %s", err))}, nil
	}

	var publickey ecdsa.PublicKey
	switch pub := publickeyAny.(type) {
	case *ecdsa.PublicKey:
		publickey = *pub
	default:
		return starlark.Tuple{starlark.None, starlark.String("unsupported public key type")}, nil
	}

	if publickey.Curve != privkey.key.Curve {
		return &starlark.Tuple{starlark.None, starlark.String("mismatching curves")}, nil
	}

	privkeyECDH, err := privkey.key.ECDH()
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed to derive ecdh of private key: %s", err))}, nil
	}

	publickeyECDH, err := publickey.ECDH()
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed to derive ecdh of public key: %s", err))}, nil
	}

	sharedSecret, err := privkeyECDH.ECDH(publickeyECDH)
	if err != nil {
		return starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed to derive shared secret: %s", err))}, nil
	}

	return starlark.Tuple{starlark.String(sharedSecret), starlark.None}, nil
}

func starlarkPrivateKeyStruct(privateKey *ecdsa.PrivateKey, curve, x509privatekeybytes string) starlark.Value {
	privkey := &privatekey{key: privateKey}
	return starlarkstruct.FromStringDict(
		starlark.String("ecdsa.privateKey"),
		starlark.StringDict{
			"curve":        starlark.String(curve),
			"x509bytes":    starlark.String(x509privatekeybytes),
			"publicKey":    starlark.NewBuiltin("publicKey", privkey.publicKey),
			"sign":         starlark.NewBuiltin("sign", privkey.sign),
			"sharedSecret": starlark.NewBuiltin("sharedSecret", privkey.sharedSecret),
		})
}
