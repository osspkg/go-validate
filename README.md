# go.osspkg.com/validate

[![Go Reference](https://pkg.go.dev/badge/go.osspkg.com/validate.svg)](https://pkg.go.dev/go.osspkg.com/validate)
[![Go Report Card](https://goreportcard.com/badge/go.osspkg.com/validate)](https://goreportcard.com/report/go.osspkg.com/validate)
[![License](https://img.shields.io/badge/license-BSD--3-blue.svg)](LICENSE)

**validate** is a lightweight, extensible validation library for Go with zero reflection overhead for callbacks, struct tagging support, and optional code generation for type-safe adapters.

## Features

- **Rule‑based validation** – register named rules with custom handlers.
- **Struct validation** – use `validate` struct tags with support for `required` and multiple rules.
- **Callback‑based validation** – validate multiple values in a single pass with `Optional`/`Require`.
- **Type‑safe adapters** – generate boilerplate‑free adapters from your own functions using `govld`.
- **String decoding** – automatically convert string inputs to most built‑in types and common interfaces.
- **Zero‑allocation pools** – internal pooling for callback validators to reduce GC pressure.
- **Generics** – used internally for caches and pools (Go 1.18+).

## Installation

```bash
go get go.osspkg.com/validate
```

To use the code generation tool:

```bash
go install go.osspkg.com/validate/cmd/govld@latest
```

## Quick Start

### 1. Register a rule and validate a struct

```go
package main

import (
    "context"
    "fmt"
    "go.osspkg.com/validate"
)

func main() {
    v := validate.New()

    // Register a rule that checks if an int64 is greater than a reference
    _ = v.Register(validate.Rule{
        Name: "gt",
        Handle: validate.HandlerFunc(func(ctx context.Context, value any, opts ...any) error {
            val, ok := value.(int64)
            if !ok {
                return fmt.Errorf("expected int64, got %T", value)
            }
            if len(opts) != 1 {
                return fmt.Errorf("expected 1 option")
            }
            ref, ok := opts[0].(int64)
            if !ok {
                return fmt.Errorf("option must be int64")
            }
            if val <= ref {
                return fmt.Errorf("value %d must be greater than %d", val, ref)
            }
            return nil
        }),
    })

    type User struct {
        Age int64 `validate:"required;gt=18"`
    }

    u := &User{Age: 25}
    if err := v.ValidateStruct(context.Background(), u); err != nil {
        fmt.Println("Validation failed:", err)
    } else {
        fmt.Println("User is valid")
    }
}
```

### 2. Callback‑based validation

```go
func validateUser(ctx context.Context, v *validate.Validator, userID int64, name string) error {
    return v.Validate(ctx, func(c validate.Callback) {
        c.Require("gt", userID, int64(0))         // userID > 0
        c.Optional("nonempty", name)               // only validated if name != ""
    })
}
```

## Core Concepts

### Rule

A rule consists of a **name** (unique identifier) and a **handler** that implements `validate.Handle`:

```go
type Handle interface {
    ValidateHandle(ctx context.Context, value any, opts ...any) error
}
```

The `validate.HandlerFunc` type allows you to turn any function with the matching signature into a handler.

### Struct Tags

Use the tag key `validate`. Multiple rules are separated by `;`. The special `required` tag makes the field mandatory (zero values are not skipped).

Examples:

```go
type Example struct {
    ID     int     `validate:"required;gt=0"`
    Name   string  `validate:"nonempty;max=64"`
    Score  float64 `validate:"min=-10;max=100"`
}
```

Rules can accept comma‑separated options:

```go
`validate:"in=admin,moderator,user"`
```

### Callback API

- `Require(name, value, opts...)` – always runs the validation. Fails if the rule returns an error.
- `Optional(name, value, opts...)` – only runs the validation when the value is **not** its zero value (see `util.IsDefaultValue`). Useful for partial updates.

## Code Generation (`govld`)

Writing handlers manually with `any` type assertions is verbose. The `govld` tool generates type‑safe adapters from your own functions.

### Step 1: Write a validation function

```go
//go:generate govld -pkg .

//govld:gen
func ValidateUID(ctx context.Context, value int64, min int64) error {
    if value < min {
        return fmt.Errorf("uid %d is less than minimum %d", value, min)
    }
    return nil
}
```

The function must:
- Have at least two parameters: `context.Context` and the value to validate.
- Return only an `error`.
- Be marked with the comment `//govld:gen` (exactly, no spaces).

### Step 2: Run the generator

```bash
go generate ./...
```

This creates `adapt_handlers_gen.go` containing `ValidateUIDAdaptHandler` – a function that matches `validate.HandlerFunc`.

### Step 3: Use the generated adapter

```go
v.Register(validate.Rule{
    Name:   "uid",
    Handle: validate.HandlerFunc(ValidateUIDAdaptHandler),
})
```

Now you can call the rule with proper types:

```go
v.Validate(ctx, func(c validate.Callback) {
    c.Require("uid", int64(123), int64(100))
})
```

The generated adapter automatically converts `any` values and string‑encoded options using `validate.StringDecode`.

## String Decoding

The `StringDecode` function (used internally by adapters) converts a string into many Go types:

- Basic types: `string`, `[]byte`, `int`, `uint`, `float`, `complex`, `bool`
- `time.Duration`, `time.Time` (RFC3339)
- Interfaces: `io.Writer`, `encoding.TextUnmarshaler`, `json.Unmarshaler`, `xml.Unmarshaler`
- Structs, maps, slices, arrays – via `json.Unmarshal`

You can use it directly:

```go
var port int
if err := validate.StringDecode(&port, "8080"); err != nil {
    // handle error
}
```

## Benchmarks

Typical performance on a modern machine (Intel i9-12900KF):

| Operation                      | ns/op     | allocs/op | B/op |
|--------------------------------|-----------|-----------|------|
| `Validate` (callback)          | ~166      | 3         | 48   |
| `ValidateStruct` (simple tags) | ~118      | 2         | 16   |
| `ValidateStruct` with adapters | ~188      | 17        | 376  |

## Full Example

```go
package main

import (
    "context"
    "fmt"
    "go.osspkg.com/validate"
)

//go:generate govld -pkg .

//govld:gen
func positiveInt(ctx context.Context, value int, _ any) error {
    if value <= 0 {
        return fmt.Errorf("value must be positive")
    }
    return nil
}

//govld:gen
func rangeCheck(ctx context.Context, value int, min, max int) error {
    if value < min || value > max {
        return fmt.Errorf("value %d out of range [%d,%d]", value, min, max)
    }
    return nil
}

type Config struct {
    Port    int `validate:"required;positiveInt"`
    Timeout int `validate:"rangeCheck=100,5000"`
}

func main() {
    v := validate.New()
    _ = v.Register(
        validate.Rule{Name: "positiveInt", Handle: validate.HandlerFunc(positiveIntAdaptHandler)},
        validate.Rule{Name: "rangeCheck", Handle: validate.HandlerFunc(rangeCheckAdaptHandler)},
    )

    cfg := &Config{Port: 8080, Timeout: 2000}
    if err := v.ValidateStruct(context.Background(), cfg); err != nil {
        fmt.Println("Invalid config:", err)
    } else {
        fmt.Println("Config OK")
    }
}
```

Run with:

```bash
go generate
go run .
```

## License

BSD 3-Clause – see [LICENSE](LICENSE) file.