package posts

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"time"
)

// BlogPost holds all data about a post (duh..)
type BlogPost struct {
	Path          string
	Title         string
	Deleted       bool
	Approved      bool
	Published     bool
	PublishedDate time.Time
	CreatedDate   time.Time
	Body          template.HTML
}

// New creates a new BlogPost
func New(title string, path string, body string) *BlogPost {
	return &BlogPost{Title: title,
		Path:        path,
		Body:        template.HTML(body),
		Approved:    false,
		Deleted:     false,
		Published:   false,
		CreatedDate: time.Now(),
	}
}

// Save BlogPost to it's Path as json
func (p BlogPost) Save() error {
	temp, err := json.Marshal(p)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile("www/posts/"+p.Path+".json", temp, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Load BlogPost from json file at path
func Load(path string) (*BlogPost, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	post := &BlogPost{}
	err = json.Unmarshal(bytes, post)
	return post, err
}

// LoadAll loads all BlogPost structs
func LoadAll() ([]*BlogPost, error) {
	files, err := ioutil.ReadDir("www/posts/")
	if err != nil {
		return nil, err
	}
	posts := make([]*BlogPost, 0, len(files))
	for _, file := range files {
		if !file.IsDir() {
			post, err := Load("www/posts/" + file.Name())
			if err != nil {
				return nil, err
			} else {
				posts = append(posts, post)
			}
		}
	}
	return posts, nil
}
