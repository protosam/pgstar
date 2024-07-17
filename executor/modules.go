package executor

import (
	"github.com/protosam/pgstar/executor/modules"
	"github.com/protosam/pgstar/executor/modules/crypto/modaes"
	"github.com/protosam/pgstar/executor/modules/crypto/modecdsa"
	"github.com/protosam/pgstar/executor/modules/crypto/modrandom"
	"github.com/protosam/pgstar/executor/modules/crypto/modsha2"
	"github.com/protosam/pgstar/executor/modules/crypto/modsha3"
	"github.com/protosam/pgstar/executor/modules/encoding/modbase64"
	"github.com/protosam/pgstar/executor/modules/encoding/modhex"
	"github.com/protosam/pgstar/executor/modules/encoding/modjson"
	"github.com/protosam/pgstar/executor/modules/modhttp"
	"github.com/protosam/pgstar/executor/modules/modmath"
	"github.com/protosam/pgstar/executor/modules/modpostgres"
	"github.com/protosam/pgstar/executor/modules/modregex"
	"github.com/protosam/pgstar/executor/modules/modtime"
)

var Modules = map[string]modules.ModuleExporterFn{
	"pgstar/postgres":        modpostgres.Constructor,
	"pgstar/http":            modhttp.Constructor,
	"pgstar/math":            modmath.Constructor,
	"pgstar/time":            modtime.Constructor,
	"pgstar/regex":           modregex.Constructor,
	"pgstar/crypto/sha2":     modsha2.Constructor,
	"pgstar/crypto/sha3":     modsha3.Constructor,
	"pgstar/crypto/random":   modrandom.Constructor,
	"pgstar/crypto/ecdsa":    modecdsa.Constructor,
	"pgstar/crypto/aes":      modaes.Constructor,
	"pgstar/encoding/base64": modbase64.Constructor,
	"pgstar/encoding/hex":    modhex.Constructor,
	"pgstar/encoding/json":   modjson.Constructor,
}
