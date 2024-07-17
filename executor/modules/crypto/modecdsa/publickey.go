package modecdsa

import (
	"crypto/ecdsa"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

type publickey struct {
	key ecdsa.PublicKey
}

func (pubkey *publickey) verify(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var messageHashSumBytes string
	var signature string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "messageSumBytes", &messageHashSumBytes, "signature", &signature); err != nil {
		return starlark.None, err
	}
	return starlark.Bool(ecdsa.VerifyASN1(&pubkey.key, []byte(messageHashSumBytes), []byte(signature))), nil
}

func starlarkPublicKeyStruct(pubKey ecdsa.PublicKey, curve, x509publickeybytes string) starlark.Value {
	pubkey := &publickey{key: pubKey}
	return starlarkstruct.FromStringDict(
		starlark.String("ecdsa.publicKey"),
		starlark.StringDict{
			"curve":     starlark.String(curve),
			"x509bytes": starlark.String(x509publickeybytes),
			"verify":    starlark.NewBuiltin("verify", pubkey.verify),
		})
}
