# PGStar Modules
TODO: This section needs further explanations. For now, it just attempts to illustrate all the available functions. Some useful features like `fail()` need to be included in examples.

These are all the modules and functions available in `pgstar`

## Builtins
This only covers the built-ins available in PGStar. The language specification includes more and specifics for the Go implementation can be found [here](https://github.com/google/starlark-go/blob/master/doc/spec.md).

- `print(message str)` - Logs to standard output.
- `addRoute(method []str, path str, scriptFile str)` - Only available during configuration, used to configure routes.
- `setGlobal(name string, value any)` - Only available during configuration, used to set a global variable for other scripts to consume.
- `getEnv(name string, default any)` - Only available during configuration, used to get environment variables prefixed with `PGSTAR_ENV`.

## pgstar/postgres
```starlark
load("pgstar/postgres", db="exports")

# selecting a specific row
rows, err = db.query("SELECT * FROM testtable WHERE name = $1", ["alice"])
if err != None:
    pass # TODO: Handle error

# iterate all selected rows
for row in rows:
    print(row)

# automatic savepointing can be enabled to automate rollbacks
# errors can not be handled without this enabled
db.savepoints(True)

# simple query example
rows, err = db.query("SELECT * FROM testtable2", [])
if err != None:
    pass # TODO: Handle error

# iterate all selected rows
for row in rows:
    print(row)

# create a savepoint
_, err = db.exec("SAVEPOINT my_savepoint", [])
if err != None:
    pass # TODO: Handle error

# inserting after a savepoint can be handled
rows_affected, err = db.exec("INSERT INTO testtable2 (name, comment) VALUES ($1, $2)", ["alice", "learned to love cryptography"])
if err != None:
    _, err = db.exec("ROLLBACK TO SAVEPOINT my_savepoint", [])
    if err != None:
        pass # TODO: Handle error

# release a savepoint
_, err = db.exec("RELEASE SAVEPOINT my_savepoint", [])
if err != None:
    pass # TODO: Handle error

# how many rows were affected?
print(rows_affected)
```
## pgstar/http
```starlark
load("pgstar/http", http="exports")

# get the hostname used for the request
http.host()

# get the request protocol (HTTP or HTTPS)
http.protocol()

# get the request method
http.method()

# contains the requester's IP address
http.remoteAddr()

# get cookies set previously
http.cookies()

# get the headers sent in the request
http.headers()

# get variables from the URL path
http.vars()

# get post request data
http.post()

# get query string data
http.query()

# set a cookie
http.setCookie(name, value)
http.setCookie(name, value, expiry)
http.setCookie(name, value, expiry, path)
http.setCookie(name, value, expiry, path, domain)
http.setCookie(name, value, expiry, path, domain, secure)
http.setCookie(name, value, expiry, path, domain, secure, httpOnly)

# write a header
http.setHeader(name, value)

# write a response that will be json encoded
http.write(statusCode, data)
http.write(201, "Hello, World!")
```
## pgstar/time
```starlark
load("pgstar/time", time="exports")

# This is the example formatting used with Go time
exampleFormatting="2006-01-02 15:04:05 MST"

# get the time in unix time
unixTimeNow = time.now()

# format a unix timestamp with a specific formatting for specific timezone
time.timezone(unixTimeNow, exampleFormatting, "America/New_York")

# format a unix time into string
exampleTimeString = time.format(unixTimeNow, exampleFormatting)

# convert a string to epoch
unixTimeAgain = time.epoch(exampleTimeString, exampleFormatting)
```
## pgstar/math

```starlark
# everything in go.starlark.net/lib/math is available
load("pgstar/math", math="exports")

# Returns the ceiling of x, the smallest integer greater than or equal to x.
math.ceil(x)

# Returns a value with the magnitude of x and the sign of y.
math.copysign(x, y)

# Returns the absolute value of x as float.
math.fabs(x)

# Returns the floor of x, the largest integer less than or equal to x.
math.floor(x)

# Returns the floating-point remainder of x/y. The magnitude of the result is less than y and its sign agrees with that of x.
math.mod(x, y)

# Returns x**y, the base-x exponential of y.
math.pow(x, y)

# Returns the IEEE 754 floating-point remainder of x/y.
math.remainder(x, y)

# Returns the nearest integer, rounding half away from zero.
math.round(x)

# Returns e raised to the power x, where e = 2.718281… is the base of natural logarithms.
math.exp(x)

# Returns the square root of x.
math.sqrt(x)

# Returns the arc cosine of x, in radians.
math.acos(x)

# Returns the arc sine of x, in radians.
math.asin(x)

# Returns the arc tangent of x, in radians.
math.atan(x)

# Returns atan(y / x), in radians.
# The result is between -pi and pi.
# The vector in the plane from the origin to point (x, y) makes this angle with the positive X axis.
# The point of atan2() is that the signs of both inputs are known to it, so it can compute the correct
# quadrant for the angle.
# For example, atan(1) and atan2(1, 1) are both pi/4, but atan2(-1, -1) is -3*pi/4.
math.atan2(y, x)

# Returns the cosine of x, in radians.
math.cos(x, y)

# Returns the Euclidean norm, sqrt(x*x + y*y). This is the length of the vector from the origin to point (x, y).
math.hypot(x,

# Returns the sine of x, in radians.
math.sin(x)

# Returns the tangent of x, in radians.
math.tan(x)

# Converts angle x from radians to degrees.
math.degrees(x)

# Converts angle x from degrees to radians.
math.radians(x)

# Returns the inverse hyperbolic cosine of x.
math.acosh(x)

# Returns the inverse hyperbolic sine of x.
math.asinh(x)

# Returns the inverse hyperbolic tangent of x.
math.atanh(x)

# Returns the hyperbolic cosine of x.
math.cosh(x)

# Returns the hyperbolic sine of x.
math.sinh(x)

# Returns the hyperbolic tangent of x.
math.tanh(x, base)

# Returns the logarithm of x in the given base, or natural logarithm by default.
math.log(x,

# Returns the Gamma function of x.
math.gamma(x)

# The base of natural logarithms, approximately 2.71828.
math.e

# The ratio of a circle's circumference to its diameter, approximately 3.14159.
math.pi
```
## pgstar/regex
```starlark
load("pgstar/regex", regex="exports")

# returns true if the string matches the pattern
regex.match(pattern, string)
```
## pgstar/encoding/hex
```starlark
load("pgstar/encoding/hex", hex="exports")

encodedData = hex.encode(data)
stringData, err = hex.decode(encodedData)
```
## pgstar/encoding/base64
```starlark
load("pgstar/encoding/base64", base64="exports")

encodedData = base64.encode(data)
stringData, err = base64.decode(encodedData)
```
## pgstar/encoding/json
```starlark
load("pgstar/encoding/json", json="exports")

jsonString, err = json.encode(data)
jsonData, err = json.decode(jsonString)
```
## pgstar/encoding/yaml
```starlark
load("pgstar/encoding/json", json="exports")

yamlString, err = yaml.encode(data)
yamlData, err = yaml.decode(yamlString)
```
## pgstar/crypto/sha2
```starlark
load("pgstar/crypto/sha2", sha2="exports")

sha2.sum256(data)
sha2.sum512(data)
```
## pgstar/crypto/sha3
```starlark
load("pgstar/crypto/sha3", sha3="exports")

sha3.sum256(data)
sha3.sum384(data)
sha3.sum512(data)
```
## pgstar/crypto/random
```starlark
load("pgstar/crypto/random", random="exports")

random.bytes(numberOfBytes)
random.int(min, max)
```
## pgstar/crypto/ecdsa
```starlark
load("pgstar/crypto/ecdsa", ecdsa="exports")

# this is an example message
message = "Hello, World!"
messageSum = sha2.sum256(message) # hashsums are for signing

# available curves are "P224" "P256" "P384" "P521"
curve = "P256"

# generate a random private key for alice
alicePriv, err = ecdsa.generateKey(curve)
if err != None:
    pass # TODO: handle failures

# signing a message
messageSignature, err = alicePriv.sign(messageSum)
if err != None:
    pass # TODO: handle failures

# get a public key from a private key
alicePub, err = alicePriv.publicKey()
if err != None:
    pass # TODO: handle failures

# verifying a message
alicePub.verify(messageSum, messageSignature)

# get private key bytes
privKeyBytes = alicePriv.x509bytes

# get public key bytes
pubKeyBytes = alicePub.x509bytes

# load a private key from bytes
alicePriv, err = ecdsa.privateKey(alicePriv.x509bytes)
if err != None:
    pass # TODO: handle failures

# load a public key from bytes
alicPub, err = ecdsa.publicKey(alicePub.x509bytes)
if err != None:
    pass # TODO: handle failures

# get a shared secret
sharedSecret, err = alicePriv.sharedSecret(someOtherPub.x509bytes)
if err != None:
    pass # TODO: handle failures
```
## pgstar/crypto/aes
```starlark
load("pgstar/crypto/aes", aes="exports")

# this is an example message
message = "Hello, World!"

# a 16, 24 or 32 byte key is needed for AES
# this determine if the encryption is 128, 192, or 256 bits
secret = random.bytes(16)

# encrypt the message
cipherText, err = aes.encrypt(secret, message)
if err != None:
    pass # TODO: handle failures

# decrypt the message
originalMessage, err = aes.decrypt(sharedSecret2, cipherText)
if err != None:
    pass # TODO: handle failures
```
