package modaes

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"

	"github.com/protosam/pgstar/executor/modules"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	ModuleName = "aes"
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
				"encrypt": starlark.NewBuiltin("aes.encrypt", encrypt),
				"decrypt": starlark.NewBuiltin("aes.decrypt", decrypt),
			},
		),
	}
}

func (module *Module) Destroy(loader modules.ModuleLoader) error { return nil }

func (module *Module) Name() string {
	return ModuleName
}

func encrypt(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var secret, message string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "secret", &secret, "message", &message); err != nil {
		return starlark.None, err
	}

	// Generate a new AES cipher
	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	// Create a new GCM cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	// Create a nonce with a length that the GCM requires (12 bytes)
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	// Encrypt the data
	ciphertext := gcm.Seal(nonce, nonce, []byte(message), nil)

	return &starlark.Tuple{
		starlark.String(ciphertext),
		starlark.None,
	}, nil
}

func decrypt(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var secret, ciphertext string
	if err := starlark.UnpackArgs(fn.Name(), args, kwargs, "secret", &secret, "ciphertext", &ciphertext); err != nil {
		return starlark.None, err
	}

	// Generate a new AES cipher using the 256-bit key
	block, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	// Create a new GCM cipher mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	// Get the nonce size
	nonceSize := gcm.NonceSize()

	// Separate the nonce and the actual ciphertext
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Decrypt the data
	message, err := gcm.Open(nil, []byte(nonce), []byte(ciphertext), nil)
	if err != nil {
		return &starlark.Tuple{starlark.None, starlark.String(fmt.Sprintf("%s", err))}, nil
	}

	return &starlark.Tuple{
		starlark.String(message),
		starlark.None,
	}, nil
}
