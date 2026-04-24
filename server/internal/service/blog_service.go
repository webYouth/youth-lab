// Package service contains gRPC service implementations.
package service

import (
	"context"

	"youthlab/server/gen/blogpb"
)

// BlogService implements blogpb.BlogServiceServer.
type BlogService struct {
	blogpb.UnimplementedBlogServiceServer
}

// NewBlogService creates a BlogService instance.
func NewBlogService() *BlogService {
	return &BlogService{}
}

// GetPosts returns mock blog posts for initialization.
func (s *BlogService) GetPosts(_ context.Context, _ *blogpb.GetPostsRequest) (*blogpb.GetPostsResponse, error) {
	return &blogpb.GetPostsResponse{
		Posts: []*blogpb.Post{
			{
				Id:        1,
				Title:     "Welcome to Youth Blog",
				Content:   "This is the first mock post served over gRPC.",
				CreatedAt: "2026-04-24T10:00:00Z",
			},
			{
				Id:        2,
				Title:     "Second Post",
				Content:   "You can replace these mock values with database content later.",
				CreatedAt: "2026-04-24T11:00:00Z",
			},
		},
	}, nil
}
