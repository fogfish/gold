# Type-safe Linked Data identities for Golang

Linked Data applications require careful handling of identifiers (IRIs, URNs) and relationships between entities. Manual string manipulation is error-prone and lacks compile-time guarantees. The `gold` library solves this by providing type-safe Linked Data identities that leverage Go's type system to prevent schema mismatches and ensure correct identifier usage at compile time.

```bash
go get github.com/fogfish/gold
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/fogfish/gold"
)

// Define domain types
type User struct {
    ID gold.IRI[User] `json:"id"`
}

type Blog struct {
    ID      gold.IRI[Blog] `json:"id"`
    Creator gold.IRI[User] `json:"creator"`
}

func main() {
    // Create type-safe IRIs - schema is derived from type
    userID := gold.ToIRI[User]("alice")     // "user:alice"
    blogID := gold.ToIRI[Blog]("my-blog")   // "blog:my-blog"
    
    user := User{ID: userID}
    blog := Blog{ID: blogID, Creator: userID}
    
    // Compile-time safety: this would fail at compile time
    // blog.Creator = blogID  // ❌ Cannot assign blog ID to user field
}
```

## Core Concepts

### Type-Safe IRIs

IRIs are structured as `{schema}:{id}` where the schema is automatically derived from the Go type name:

```go
type Person struct{}

personID := gold.ToIRI[Person]("john")             // "person:john"
validID, err := gold.AsIRI[Person]("person:jane")  // ✅ Valid
invalidID, err := gold.AsIRI[Person]("user:bob")   // ❌ Error: wrong schema
```

### Hierarchical URNs

For hierarchical identifiers with namespace support:

```go
type MyApp struct{}
type Document struct{}

// Creates "urn:myapp:document:folder:subfolder:file"
urn := gold.ToURN[MyApp, Document]("folder", "subfolder", "file")
iri := urn.ToIRI()  // "document:folder/subfolder/file"
```

### Relationship Modeling

Model relationships between entities using compound types:

```go
// Relationship between User and Blog for compound database keys
type UserBlog = gold.Imply[User, Blog]

relationship := gold.ImplyFrom(
    gold.ToIRI[User]("alice"),
    gold.ToIRI[Blog]("tech-blog")
)
```

`Imply` is a **string**, not a struct. Its text form is `{Domain}|{Range}`:

| Domain | Range | Text                |
| ------ | ----- | ------------------- |
| empty  | empty | `""`                |
| set    | empty | `"user:alice\|"`    |
| set    | set   | `"user:alice\|blog:b"` |
| empty  | set   | error               |

`Imply` implements `encoding.TextMarshaler` / `encoding.TextUnmarshaler`, validating
schemas on both directions. JSON therefore encodes it as a string `"user:a|blog:b"`.

> **Breaking change in v0.1.0:** `Imply` used to be encoded in JSON as `{"domain":…,"range":…}`.

### Text marshaling

`IRI[T]` and `URN[N, T]` implement `encoding.TextMarshaler` and `encoding.TextUnmarshaler`
(the library has no storage dependencies, any codec honouring these interfaces works, e.g. AWS `attributevalue`
with `UseEncodingMarshalers`).

* Decoding validates the schema: reading `blog:x` into `IRI[User]` fails.
* Encoding is strict for `IRI`: `IRI[User]("alice")` (forgotten `.Norm()`) fails instead of persisting a wrong identity.
* An empty value is encoded and decoded as `""`, including JSON (**changed in v0.1.0**: an empty JSON IRI is accepted).

### Composite keys and DynamoDB

```go
// ✅ alias keeps gold's methods, stored as a string
type ThreadID = gold.Imply[Account, Story]

// ⚠️ a defined type drops methods. If a defined type is needed (e.g. to add domain methods),
// forward the text methods explicitly:
type ThreadKey gold.Imply[Account, Story]

func (k ThreadKey) MarshalText() ([]byte, error)  { return gold.Imply[Account, Story](k).MarshalText() }
func (k *ThreadKey) UnmarshalText(b []byte) error { return (*gold.Imply[Account, Story])(k).UnmarshalText(b) }
```

#### Prefix queries

`gold.Any[T]()` is the schema-only IRI `"{schema}:"`. Use it to select every identity of a class
with `begins_with`:

```go
gold.Any[User]()                            // "user:"
gold.ImplyFrom(job, gold.Any[UID]())        // "job:1|uid:"
```

A partial `SortKey` (domain only) is encoded with a trailing separator `"job:1|"`, so a prefix
query for job `1` does not match `job:10|…` (**changed in v0.1.0**). Decoding still accepts the legacy `"job:1"`.
`SortKey` output always equals `Imply.MarshalText`.

### Schema Registry

Schemas are derived automatically (lower-cased type name), but can be customized:

```go
// For complex types, register custom schemas
gold.Register[[]MyType]("my-type-collection")
```

The registry is keyed by the full type identity (import path + name), so same-named types in
different packages do not collide.

#### Schema stability

Renaming a Go type silently changes a derived schema, and therefore stored keys.
Pin persisted schemas in a golden test using `gold.Schemas()`, which returns a copy of every
schema derived or registered so far:

```go
func TestSchemas(t *testing.T) {
    _ = gold.Schema[User]() // make sure the schema is derived
    got := gold.Schemas()
    if got["example.com/app/domain.User"] != "user" {
        t.Fatal("persisted schema of User has changed")
    }
}
```

### Type-Safe URL Templates

URL templates provide compile-time safety for building URLs with typed parameters:

```go
type AccountID = gold.IRI[Account]
type StoryID = gold.IRI[Story]

// Define type-safe URL template
type StoryUrl = gold.Url2[string, AccountID, StoryID]

// Create URL template constants
const CatalogStoryUrl = StoryUrl("/catalog/authors/%s/stories/%s")

// Build URLs with type safety
accountID := gold.ToIRI[Account]("alice")
storyID := gold.ToIRI[Story]("my-story")
url := CatalogStoryUrl.Build(accountID, storyID)
// Result: "/catalog/authors/account:alice/stories/story:my-story"
```

The library provides `Url1` through `Url4` for templates with 1-4 parameters respectively.

## License

See [LICENSE](LICENSE) file for details. 
