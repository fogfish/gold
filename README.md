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

### Schema Registry

Schemas are derived automatically, but can be customized:

```go
// For complex types, register custom schemas
gold.Register[[]MyType]("my-type-collection")
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
