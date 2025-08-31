//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package main

import (
	"encoding/json"
	"fmt"

	"github.com/fogfish/gold"
)

//
// Example shows semantic modeling of blogging application
//
// Consider a simple linked data structure where users can create blogs and comment on them.
// Each comment can be anchored to another comment, allowing for threaded discussions.
//

// User Identity is IRI (e.g. user:alice)
type UserID = gold.IRI[User]

// Core domain data for User
type User struct {
	ID UserID `json:"id"`
}

// Projecting user's core domain into a storage is trivial.
// UserID is a "hash key" to be used as-is by the database.
type UserDB struct {
	HKey UserID `json:"hashkey"`
}

// Blog post Identity is IRI (e.g. blog:1234)
type BlogID = gold.IRI[Blog]

// Core domain for User carries blog ID and contains statement about its creator
type Blog struct {
	ID      BlogID `json:"id"`
	Creator UserID `json:"creator,omitempty"`
}

// Projecting blog's core domain into a storage requires two dimensions:
// 1. Hash key - a unique identifier for the user who created the blog
// 2. Sort key - a unique identifier for the blog post itself
//
// Given structure is equally applicable for both SQL and NoSQL databases.
type BlogDB struct {
	HKey UserID `json:"hashkey"`
	SKey BlogID `json:"sortkey"`
}

// Comment Identity is IRI (e.g. comment:1)
type CommentID = gold.IRI[Comment]

// Core domain for comment contains of compound identity. Firstly it has a unique key,
// then it is about particular blog post and can be anchored to another comment.
type Comment struct {
	ID      CommentID `json:"id"`
	Creator UserID    `json:"creator,omitempty"`
	About   BlogID    `json:"about,omitempty"`
	Anchor  CommentID `json:"anchor,omitempty"`
}

// Projecting comment's core domain into a NoSQL storage non-trivial
// in comparision with SQL databases. Comment consists of compound key
//
//	⟨user, blog, comment⟩
//
// Imply type allows us to define a hash key as a combination of user and blog.
type CommentDB struct {
	HKey HashKey   `json:"hashkey"`
	SKey CommentID `json:"sortkey"`
}

type HashKey gold.Imply[User, Blog]

var hashkey = gold.HashKey[HashKey, User, Blog]()

func (k HashKey) MarshalJSON() ([]byte, error) {
	return gold.EncodeJSON(hashkey.Encode(&k))
}

func main() {
	alice := User{ID: UserID("alice").Norm()}
	print("==> user signed up", alice)

	udb := UserDB{HKey: alice.ID}
	print("==> user is written to db", udb)

	blog := Blog{
		ID:      BlogID("1234").Norm(),
		Creator: alice.ID,
	}
	print("==> user created a blog", blog)

	bdb := BlogDB{
		HKey: alice.ID,
		SKey: blog.ID,
	}
	print("==> blog is written to db", bdb)

	bob := User{ID: UserID("bob").Norm()}
	comment := Comment{
		ID:      CommentID("1").Norm(),
		Creator: bob.ID,
		About:   blog.ID,
	}
	print("==> user created a comment", comment)

	cdb := CommentDB{
		HKey: HashKey(gold.ImplyFrom(alice.ID, blog.ID)),
		SKey: comment.ID,
	}
	print("==> comment is written to db", cdb)
}

func print(msg string, obj any) {
	b, _ := json.MarshalIndent(obj, "| ", "  ")
	fmt.Println(msg)
	fmt.Println("| " + string(b))
	fmt.Println()
}
