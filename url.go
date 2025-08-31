//
// Copyright (C) 2025 Dmitry Kolesnikov
//
// This file may be modified and distributed under the terms
// of the MIT license.  See the LICENSE file for details.
// https://github.com/fogfish/gold
//

package gold

import (
	"fmt"
)

// Url1 is a type-safe URL template with one parameter.
// T is the target string type, A is the type of the parameter.
//
// Example:
//
//	type UserUrl = gold.Url1[string, UserID]
//	const UserProfile = UserUrl("/users/%s")
//	url := UserProfile.Build(userID)
type Url1[T ~string, A any] string

// Build constructs the URL by formatting the template with the provided parameter.
func (url Url1[T, A]) Build(a A) T {
	return T(fmt.Sprintf(string(url), a))
}

// Url2 is a type-safe URL template with two parameters.
// T is the target string type, A and B are the types of the parameters.
//
// Example:
//
//	type StoryUrl = gold.Url2[string, AccountID, StoryID]
//	const CatalogStoryUrl = StoryUrl("/catalog/authors/%s/stories/%s")
//	url := CatalogStoryUrl.Build(accountID, storyID)
type Url2[T ~string, A, B any] string

// Build constructs the URL by formatting the template with the provided parameters.
func (url Url2[T, A, B]) Build(a A, b B) T {
	return T(fmt.Sprintf(string(url), a, b))
}

// Url3 is a type-safe URL template with three parameters.
// T is the target string type, A, B, and C are the types of the parameters.
//
// Example:
//
//	type CommentUrl = gold.Url3[string, AccountID, StoryID, CommentID]
//	const StoryCommentUrl = CommentUrl("/catalog/authors/%s/stories/%s/comments/%s")
//	url := StoryCommentUrl.Build(accountID, storyID, commentID)
type Url3[T ~string, A, B, C any] string

// Build constructs the URL by formatting the template with the provided parameters.
func (url Url3[T, A, B, C]) Build(a A, b B, c C) T {
	return T(fmt.Sprintf(string(url), a, b, c))
}

// Url4 is a type-safe URL template with four parameters.
// T is the target string type, A, B, C, and D are the types of the parameters.
//
// Example:
//
//	type ThreadUrl = gold.Url4[string, AccountID, StoryID, CommentID, ThreadID]
//	const CommentThreadUrl = ThreadUrl("/catalog/authors/%s/stories/%s/comments/%s/threads/%s")
//	url := CommentThreadUrl.Build(accountID, storyID, commentID, threadID)
type Url4[T ~string, A, B, C, D any] string

// Build constructs the URL by formatting the template with the provided parameters.
func (url Url4[T, A, B, C, D]) Build(a A, b B, c C, d D) T {
	return T(fmt.Sprintf(string(url), a, b, c, d))
}
