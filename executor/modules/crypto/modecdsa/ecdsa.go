package modecdsa

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"fmt"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "ecdsa"
)

type Module struct{}

func Constructor(loader modules.ModuleLoader) (modules.LocalizedModule, error) {
	return &Module{}, nil
}

func (module *Module) Exports() starlark.StringDict {
	return starlark.StringDict{
		"exports": starlarkstruct.FromStringDict(
			starlark.String(ModuleName),
			starlark.StringDict{
				"generateKey": starlark.NewBuiltin("ecdsa.generateKey", generateKey),
				"privateKey":  starlark.NewBuiltin("ecdsa.privateKey", privateKey),
				"publicKey":   starlark.NewBuiltin("ecdsa.publicKey", publicKey),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func generateKey(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var curveOpt string
	if err := starlark.UnpackPositionalArgs(fn.Name(), args, kwargs, 1, &curveOpt); err != nil {
		return starlark.None, err
	}

	var curve elliptic.Curve

	switch curveOpt {
	case "P224":
		curve = elliptic.P224()
	case "P256":
		curve = elliptic.P256()
	case "P384":
		curve = elliptic.P384()
	case "P521":
		curve = elliptic.P521()
	default:
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("invalid curve: %s", curveOpt))}, nil
	}

	privateKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("%s(): %s", fn.Name(), err)
	}

	x509privatekeybytes, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("%s(): %s", fn.Name(), err)
	}

	return &starlark.Tuple{
		starlarkPrivateKeyStruct(privateKey, curveOpt, string(x509privatekeybytes)),
		starlark.None,
	}, nil
}

func privateKey(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x509privatekeybytes string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "x509privatekeybytes", &x509privatekeybytes); err != nil {
		return starlark.None, err
	}

	privateKey, err := x509.ParseECPrivateKey([]byte(x509privatekeybytes))
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed reading private key bytes: %s", err))}, nil
	}

	curve, err := curveCheck(privateKey.Curve)
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	return &starlark.Tuple{
		starlarkPrivateKeyStruct(privateKey, curve, string(x509privatekeybytes)),
		starlark.None,
	}, nil
}

func publicKey(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var x509publickeybytes string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "x509publickeybytes", &x509publickeybytes); err != nil {
		return starlark.None, err
	}

	publickeyAny, err := x509.ParsePKIXPublicKey([]byte(x509publickeybytes))
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("failed reading private key bytes: %s", err))}, nil
	}

	var publickey ecdsa.PublicKey
	switch pub := publickeyAny.(type) {
	case *ecdsa.PublicKey:
		publickey = *pub
	default:
		return &starlark.Tuple{starlark.None, starlark.String("unsupported public key type")}, nil
	}

	curve, err := curveCheck(publickey.Curve)
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	return &starlark.Tuple{
		starlarkPublicKeyStruct(publickey, curve, string(x509publickeybytes)),
		starlark.None,
	}, nil
}

func curveCheck(curve elliptic.Curve) (string, error) {
	var curveType string
	switch curve {
	case elliptic.P224():
		curveType = "P224"
	case elliptic.P256():
		curveType = "P256"
	case elliptic.P384():
		curveType = "P384"
	case elliptic.P521():
		curveType = "P521"
	default:
		return "", fmt.Errorf("unsupported curve: %s", curveType)
	}

	return curveType, nil
}
