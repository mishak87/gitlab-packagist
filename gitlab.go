package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

var (
	ErrNotFound = errors.New("Not Found")
)

type GitLab struct {
	Url    string
	Token  string
	Client *http.Client
}

type Arguments map[string]interface{}

type Project struct {
	Id                int       `json:"id"`
	Name              string    `json:"name"`
	SSHURLToRepo      string    `json:"ssh_url_to_repo"`
	LastActivity      time.Time `json:"last_activity_at"`
	PathWithNamespace string    `json:"path_with_namespace"`
}

type Reference struct {
	Name   string `json:"name"`
	Commit Commit `json:"commit"`
}

type Commit struct {
	Id        string    `json:"id"`
	Message   string    `json:"message"`
	Authored  time.Time `json:"authored_date"`
	Committed time.Time `json:"committed_date"`
}

type File struct {
	Name     string `json:"file_name"`
	Path     string `json:"file_path"`
	Encoding string `json:"encoding"`
	Content  string `json:"content"`
	BlobId   string `json:"blob_id"`
	CommitId string `json:"commit_id"`
}

func (g *GitLab) Projects(page int, perPage int) (*[]Project, error) {
	p := new([]Project)
	if err := g.call("projects", p, &Arguments{"page": page, "per_page": perPage}); err != nil {
		return nil, err
	}
	return p, nil
}

func (g *GitLab) Branches(id int) (*[]Reference, error) {
	r := new([]Reference)
	if err := g.callProject(id, "repository/branches", r, nil); err != nil {
		return nil, err
	}
	return r, nil
}

func (g *GitLab) Tags(id int) (*[]Reference, error) {
	r := new([]Reference)
	if err := g.callProject(id, "repository/tags", r, nil); err != nil {
		return nil, err
	}
	return r, nil
}

func (g *GitLab) File(id int, reference string, name string) (*File, error) {
	f := &File{}
	if err := g.callProject(id, "repository/files", f, &Arguments{"file_path": name, "ref": reference}); err != nil {
		return nil, err
	}
	return f, nil
}

func (g *GitLab) callProject(id int, resource string, result interface{}, args *Arguments) error {
	return g.call(fmt.Sprintf("projects/%d/%s", id, resource), result, args)
}

func (g *GitLab) call(resource string, result interface{}, args *Arguments) error {
	p := url.Values{}
	p.Set("private_token", g.Token)
	if args != nil {
		for k, v := range *args {
			p.Set(k, fmt.Sprintf("%v", v))
		}
	}

	url := g.Url + resource + "?" + p.Encode()
	r, err := g.Client.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	if r.StatusCode == 404 {
		return ErrNotFound

	} else if r.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", r.StatusCode)
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, result)
}

func (f *File) DecodeContent() ([]byte, error) {
	return base64.StdEncoding.DecodeString(f.Content)
}
