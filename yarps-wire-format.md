# Yarp's Wire Format

This document describes how YARP encodes data, allowing it to be transferred
between clients and servers.

## Standardising Values
Every value in YARP streams contain enough leading metadata to allow data to be
interpreted or skipped. That said, it is important to notice that the first
three first bits of the first received byte of each value indicates which type
follows:

| Bits  | Type      |
|-------|-----------|
| `000` | Void      |
| `001` | Scalar    |
| `010` | Float     |
| `011` | Array<T>  |
| `100` | Struct    |
| `101` | String    |
| `110` | Map<T, U> |
| `111` | OneOf     |

Remaining bits depend on the type being received.

## Void

The Void type is used to represent no value. It contains no further information
after the first three bits. Remaining five bits should be unset.

## Scalar
YARP considers as Scalar the following types:

- `uint8`
- `uint16`
- `uint32`
- `uint64`
- `int8`
- `int16`
- `int32`
- `int64`
- `boolean`

This means that all those types are encoded like they were the same, what
changes is how those values are interpreted when decoded into a specialised
type. However, signed values will have their fourth bit set, while unsigned ones
won't. The next bits contains bits from the values themselves, except for the
last one, which indicates whether more bits follow.

Given those details, one can consider an arbitrary scalar value to be encoded
as:
<table>
<tbody>
    <tr>
        <th colspan="8">First octet</th>
        <th colspan="8">Other octets</th>
    </tr>
    <tr>
        <th>7</th>
        <th>6</th>
        <th>5</th>
        <th>4</th>
        <th>3</th>
        <th>2</th>
        <th>1</th>
        <th>0</th>
        <th>7</th>
        <th>6</th>
        <th>5</th>
        <th>4</th>
        <th>3</th>
        <th>2</th>
        <th>1</th>
        <th>0</th>
    </tr>
    <tr>
        <th>2<sup>7</sup></th>
        <th>2<sup>6</sup></th>
        <th>2<sup>5</sup></th>
        <th>2<sup>4</sup></th>
        <th>2<sup>3</sup></th>
        <th>2<sup>2</sup></th>
        <th>2<sup>1</sup></th>
        <th>2<sup>0</sup></th>
        <th>2<sup>7</sup></th>
        <th>2<sup>6</sup></th>
        <th>2<sup>5</sup></th>
        <th>2<sup>4</sup></th>
        <th>2<sup>3</sup></th>
        <th>2<sup>2</sup></th>
        <th>2<sup>1</sup></th>
        <th>2<sup>0</sup></th>
    </tr>
    <tr>
        <td colspan="1" style="text-align:center;"><i>A</i></td>
        <td colspan="3" style="text-align:center;"><i>B</i></td>
        <td colspan="1" style="text-align:center;"><i>C</i></td>
        <td colspan="3" style="text-align:center;"><i>D</i></td>
        <td style="text-align:center;"><i>B</i></td>
        <td colspan="7" style="text-align:center;"><i>C<sub>n</sub></i> (<i>n</i> &gt; 0)
    </td>
</tr>
</tbody>
</table>

In the first octet:
- `D` would be composed of `001`, as seen in [_Standardising Values_](#standardising-values);
- `C` indicates whether the encoded integer is signed;
- `B` contains the first three bits from the encoded integer
- `A` indicates whether more bytes follow.

In case `A` is set, next octets are encoded as:
- <code>C<sub>n</sub></code> containing following bits from the integer
- `B` indicating whether another byte follow.

Given those rules, we can observe how values are encoded:

| Value      | Hex        | Binary                       |
|------------|------------|------------------------------|
| `uint(3)`  | `0x26`     | `00100110`                   |
| `int(27)`  | `0x3136`   | `00110001 00110110`          |
| `int(512)` | `0x310900` | `00110001 00001001 00000000` |

### Encoding Boolean Values

As already noted, one can only observe the real value of a field after
interpreting it. The same applies to boolean values. Booleans are encoded
by leveraging the sign bit of scalar values. Observe the values for both
encoded values:

| Value    | Hex    | Binary     |
|----------|--------|------------|
| `true`   | `0x30` | `00110000` |
| `false`  | `0x20` | `00100000` |

## Floating-Point Values

YARP supports both `float32` and `float64` values. Both values are encoded as
IEEE754 values with an extra initial byte representing the data type.
For `float32` and `float64`, bits 3 and 4 are used to indicate the size and zero
values, respectively.

- When bit 3 is set, the value that follows have 64-bits. Otherwise, 32-bits.
- When bit 4 is set, the float value is zero. In this case, no more bits follow.

Given those directions, let's observe how π and 0 is encoded:

| Value               | Hex                    | Binary
|---------------------|------------------------|--------
| `float32(3.141593)` | `0x40db0f4940`         | `01000000 11011011 00001111 01001001 01000000`
| `float32(0)`        | `0x48`                 | `01001000`
| `float64(3.141593)` | `0x50182d4454fb210940` | `01010000 00011000 00101101 01000100 01010100 11111011 00100001 00001001 01000000`
| `float64(0)`        | `0x58`                 | `01011000`

## Arrays

Arrays are homogeneous collections of objects. As expected, interpreting values
depends on client and server implementation, and the type itself does not
include much metadata besides whether the array is empty or not. So, for an
array of ints, each item will have their type identifier and every other detail
mentioned so far.

Given those details, arrays are encoded as the following:

1. The three-bits indicating the type
2. A sequence of bits encoded as a Scalar value, indicating how many bytes
comprise the array value
3. N bytes composing the array itself.

For instance, let's observe an of uint8 values containing values `[1, 2, 3]`:

```
00000000  66 22 24 26                                       |f"$&|
00000004
```

Observing it with a better microscope, the following bits composes those four
bytes:

```
01100110   00100010   00100100   00100110
```

Types `011` (Array), and `001` (Scalar) can be noticed across the stream.

## Strings

Strings are always encoded as UTF-8, and are encoded with the common three-bit
type indicator, followed by a varint indicating the string length. Following
the length, from the next byte up to the Nth byte (where N = the read length),
the UTF-8 string follows.

For instance, encoding `こんにちは、YARP！` (pronounced _Kon'nichiwa, YARP!_)
yields us the following byte array:

```
00000000  a1 32 e3 81 93 e3 82 93  e3 81 ab e3 81 a1 e3 81  |.2..............|
00000010  af e3 80 81 59 41 52 50  ef bc 81                 |....YARP...|
```

The first bytes reads:
```
Hex Binary
--- ------
A1  1010 0001
32  0011 0010
```

The first three bits indicates that a String follows (`101`), followed by a
varint indicating its length, 25 bytes, followed by 25 bytes representing the
UTF-8 string.

## Maps

Maps, or associative arrays, associates a key with a given value. In YARP, maps
are defined in the format `map<T, U>` where `T` must represent a primitive type
such as scalars (except `boolean`), or strings. Structs, Maps, Arrays, and
Boolean values cannot be used as a type provided to `T`. Type `U` has no limit
when compared to `T`.

When analysing contents from an encoded map, the following can be observed:
- The length encoded in the map's type header covers all bytes depicting it.
- The first element after its header is an uint value indicating the amount of
bytes comprising all its keys values.
- After the first element, N bytes follow, representing the keys themselves.
- Then, another uint value indicates how many bytes represents all values
present in the map.
- Finally, N bytes follow, representing all values.

Given the associative nature of maps, the amount of items present in the keys
blob will match the amount of items in the values blob.

For instance, consider a `map<string, string>` containing the following entries:

| Key  | Value              |
|------|--------------------|
| `en` | `Hello, YARP!`     |
| `ja` | `こんにちは、YARP！` |
| `it` | `Ciao, YARP!`      |

Encoding it yields the following value:

```
00000000  c1 86 21 12 a4 65 6e a4  6a 61 a4 69 74 21 6c a1  |..!..en.ja.it!l.|
00000010  18 48 65 6c 6c 6f 2c 20  59 41 52 50 21 a1 32 e3  |.Hello, YARP!.2.|
00000020  81 93 e3 82 93 e3 81 ab  e3 81 a1 e3 81 af e3 80  |................|
00000030  81 59 41 52 50 ef bc 81  a1 16 43 69 61 6f 2c 20  |.YARP.....Ciao, |
00000040  59 41 52 50 21                                    |YARP!|
```

> :warning: **Warning!**: Maps ordering are not guaranteed, meaning encoding
maps cannot be done in a deterministic way. This behaviour is inherent of the
current language, so your mileage may vary. Do not expect maps to be in any
specific order.

## oneof

`Oneof` is a mutually-exclusive field, in which a set of possible values are
defined, but only one value can be set at a time. Given those specificities,
a `oneof` value have the following characteristics:

- The first byte contains, along with the identifier bits, an varint containing
the size of the index indicator plus the size of the encoded value following
such indicator.
- The second element contains an unsigned integer value indicating the index of
the field being set in the oneof value.
- The last element contains an encoded value (containing identifier bits
followed by length, so on and so forth)

Observed below is a `oneof` field in which index `3` is set with a 5-byte long
string `"Hello"`:

```
00000000  e1 10 26 a1 0a 48 65 6c  6c 6f                    |..&..Hello|
```

## Struct

Structs are user-defined types created by defining `message` structures on IDL
files. When encoded, the three-bit identifiers are found along with a varint
indicating how many bytes follow.
Past the size indicator, eight bytes representing a 64-bit unsigned integer
uniquely identifies the struct that follow.
Then, a blob of encoded values follow, each value in its own index as defined in
the IDL.

For instance, consider the following message:

```
message SampleStruct {
    id    int64  = 0;
    name  string = 1;
    email string = 2;
}
```

Assuming it has `0x0102030405060708` as its unique identifier, encoding the
following data:

- id: 27
- name: Vito
- email: hey@vito.io

Yields the following blob:

```
00000000  81 3a 08 07 06 05 04 03  02 01 31 36 a1 08 56 69  |.:........16..Vi|
00000010  74 6f a1 16 68 65 79 40  76 69 74 6f 2e 69 6f     |to..hey@vito.io|
```

## Requests
After a connection to a YARP server is established, the client must transfer a
Request object, stating which method it wants to invoke, followed by a set of
optional headers that may provide metadata for the server or RPC method.
Each wire structure are prefixed by a three-byte long magic value indicating the
kind of wire value being transferred. After those headers are pushed, a client
may start streaming the method's argument value.

For `Request`, the magic value is `0x79, 0x79, 0x72`, which when represented in
ASCII, equals `yyr`.
Next, a `Request` object with Method ID `0x0102030405060708`, and a single
header `RequestID` whose value is `First` is depicted:

```
00000000  79 79 72 21 42 23 03 01  c1 81 51 31 1d 10 c1 2c  |yyr!B#....Q1...,|
00000010  21 16 a1 12 52 65 71 75  65 73 74 49 44 21 0e a1  |!...RequestID!..|
00000020  0a 46 69 72 73 74                                 |.First|
```

## Responses
Once a server invokes the handler responsible for a given RPC method, a
`Response` object follows. This object is responsible for indicating that the
server is not reporting an error condition, and also optionally transfers a set
of headers from the server back to the client that issued the request.
Additionally, the object may also indicate whether the server intends to stream
a response. Streamed responses allows the server to transfer a variable number
of objects without disclosing how many objects are intended to be transferred,
allowing services to potentially create results during runtime and push them
immediately, without requiring the server to buffer all objects.

Responses are transferred under the magic value `0x79, 0x79, 0x52`, which stands
for ASCII `yyR`.

Next is an example of a `Response` object with a single header `Header` with
value `Value`, and the stream flag set:

```
00000000  79 79 52 c1 26 21 10 a1  0c 48 65 61 64 65 72 21  |yyR.&!...Header!|
00000010  0e a1 0a 56 61 6c 75 65  30                       |...Value0|
```

## Errors

The server may return an error value instead of a Response in case something
goes awry. Errors are special values that reports an error condition back to a
client. YARP errors were designed to be able to transfer enough context to allow
clients to handle them, which includes the following fields:

- Kind: An integer value indicating which kind of error is being reported.
- Headers: A `map<string, string>` containing metadata regarding the response.
- Identifier: A string value representing the error. Implementors are advised to
use short, repeatable, `snake_case` identifiers, but this is not enforced.
- UserData: A `map<string, string>` containing any kind of information the
service included with the error.

Errors are transferred under the magic value `0x79, 0x79, 0x65`, which stands
for ASCII `yye`.

Kind may assume one of the following values:

| Code | Name                 | Description |
|------|----------------------|-------------|
|  0   | Internal Error       | Internal Error indicates that an internal error prevented the server from performing the requested operation. |
|  1   | Managed Error        | Managed Error indicates that the server returned a user-defined error. See Identifier and UserData fields, and consult the service's documentation for further information. |
|  2   | Request Timeout      | Request Timeout indicates that the server reached a timeout while waiting for the client to transmit headers. |
|  3   | Unimplemented Method | Unimplemented Method indicates that the server does not implement the requested method. |
|  4   | Type Mismatch        | Type Mismatch indicates that the contract between the server and client is out-of-sync, since they could not agree on a type for either a request or response. | |
|  5   | Unauthorized         | Unauthorized indicates that the server refused to perform an operation due to lack of authorization. See Identifier and UserData fields, along with the service's documentation for further information. |
|  6   | Bad Request          | Bad Request indicates that the server refused to continue the operation due to a problem with the incoming request. See Identifier and UserData fields, along with the service's documentation for further information. |

An error with Kind `Unauthorized`, identifier `Identifier`, and no user data or
headers is encoded into the following bytes:

```
00000000  79 79 65 21 0a c0 a1 14  49 64 65 6e 74 69 66 69  |yye!....Identifi|
00000010  65 72 c0                                          |er.|
```
