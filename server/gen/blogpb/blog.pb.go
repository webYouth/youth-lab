// Code generated placeholder for blog.proto.
// Replace this file by running protoc as documented in proto/blog.proto.
package blogpb

// GetPostsRequest represents an empty request.
type GetPostsRequest struct{}

// Post is a single blog entry payload.
type Post struct {
	Id        int32
	Title     string
	Content   string
	CreatedAt string
}

// GetPostsResponse wraps a list of blog posts.
type GetPostsResponse struct {
	Posts []*Post
}
