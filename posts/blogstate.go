package posts

import (
	"errors"
	"fmt"
	"sort"
)

// BlogState holds global state for the Blog App
type BlogState struct {
	allPosts []*BlogPost
}

// NewBlogState deas what you would expect
func NewBlogState() *BlogState {
	return &BlogState{
		allPosts: []*BlogPost{},
	}
}

// Load will load all posts from disk and populate allPosts
func (bs *BlogState) Load() error {
	loadedPosts, err := LoadAll()
	if err != nil {
		return err
	}
	bs.allPosts = loadedPosts
	for _, p := range bs.allPosts {
		fmt.Println("Loaded post: " + p.Title)
	}
	return nil
}

func choose(xs []*BlogPost, test func(*BlogPost) bool) (ret []*BlogPost) {
	for _, x := range xs {
		if test(x) {
			ret = append(ret, x)
		}
	}
	return
}

// MainPagePosts returns all published posts to be
// displayed on the main page of the blog, newest post first.
func (bs *BlogState) MainPagePosts() []*BlogPost {
	ps := choose(bs.allPosts, func(p *BlogPost) bool {
		return p.Published && !p.Deleted
	})
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].PublishedDate.After(ps[j].PublishedDate)
	})
	return ps
}

// AdminPagePosts returns all non-deleted posts for
// displaying on the admin page of the blog, newest post first.
func (bs *BlogState) AdminPagePosts() []*BlogPost {
	ps := choose(bs.allPosts, func(p *BlogPost) bool {
		return !p.Deleted
	})
	sort.Slice(ps, func(i, j int) bool {
		return ps[i].CreatedDate.After(ps[j].CreatedDate)
	})
	return ps
}

// FindPost based on matching BlogPost Path
func (bs *BlogState) FindPost(path string) (*BlogPost, error) {
	for i := range bs.allPosts {
		if bs.allPosts[i].Path == path {
			return bs.allPosts[i], nil
		}
	}
	return nil, errors.New("Post not found")
}

// AddAndSave persist the BlogPost and adds it to the BlogState.
func (bs *BlogState) AddAndSave(post *BlogPost) error {
	if err := post.Save(); err != nil {
		return err
	}
	bs.allPosts = append(bs.allPosts, post)
	return nil
}
