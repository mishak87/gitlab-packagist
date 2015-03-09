package main

import (
	"bytes"
	"encoding/json"
	"log"
	"regexp"
	"sync"
	"time"
)

type ComposerRepository struct {
	gitLab *GitLab

	versionRegexp *regexp.Regexp

	sync.RWMutex

	projectsLastActivity       map[int]time.Time
	projectsReferencesCommitId map[int]map[string]string
	cache                      map[string]map[string]interface{}

	content      []byte
	modifiedTime time.Time
}

func NewComposerRepository(g *GitLab) *ComposerRepository {
	return &ComposerRepository{
		gitLab: g,

		versionRegexp: regexp.MustCompile("^v?\\d+\\.\\d+(\\.\\d+)*(\\-(dev|patch|alpha|beta|RC)\\d*)?$"),

		projectsLastActivity:       make(map[int]time.Time),
		projectsReferencesCommitId: make(map[int]map[string]string),
		cache: make(map[string]map[string]interface{}),
	}
}

func (c *ComposerRepository) ModifiedTime() time.Time {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	return c.modifiedTime
}

func (c *ComposerRepository) Content() (*bytes.Reader, error) {
	if c.content == nil {
		if err := c.generateContent(); err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(c.content), nil
}

func (c *ComposerRepository) Update() error {
	page := 1
	perPage := 100
	for {
		projects, err := c.gitLab.Projects(page, perPage)
		if err != nil {
			return err
		}
		for _, p := range *projects {
			if p.LastActivity.After(c.projectsLastActivity[p.Id]) {
				c.refreshProject(&p)
			}
		}
		if len(*projects) < perPage {
			break
		}
		page++
	}
	return nil
}

func (c *ComposerRepository) generateContent() error {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	b, err := json.Marshal(map[string]interface{}{
		"packages": c.cache,
	})
	if err != nil {
		return err
	}

	c.content = b
	c.modifiedTime = time.Now()
	return nil
}

func (c *ComposerRepository) refreshProject(p *Project) {
	log.Printf("%v (%v)", p.PathWithNamespace, p.Id)

	branches, err := c.gitLab.Branches(p.Id)
	if err != nil {
		log.Println(err)
		return
	}
	c.processProjectReferences(p, branches)

	tags, err := c.gitLab.Tags(p.Id)
	if err != nil {
		log.Println(err)
		return
	}
	c.processProjectReferences(p, tags)

	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.projectsLastActivity[p.Id] = p.LastActivity
	c.content = nil
}

func (c *ComposerRepository) processProjectReferences(p *Project, r *[]Reference) {
	for _, x := range *r {
		if c.projectsReferencesCommitId[p.Id][x.Name] != x.Commit.Id {
			c.refreshProjectReference(p, &x)
		}
	}
}

func (c *ComposerRepository) refreshProjectReference(p *Project, r *Reference) {
	version := c.parseVersion(r.Name)
	if version == "" {
		return
	}
	log.Printf("+ %v (%v)", version, r.Name)
	f, err := c.gitLab.File(p.Id, r.Commit.Id, "composer.json")
	if err == ErrNotFound {
		return

	} else if err != nil {
		log.Println("Download Error:", err)
		return
	}

	b, err := f.DecodeContent()
	if err != nil {
		log.Println("Decoding Error:", err)
		return
	}

	var d interface{}
	if err := json.Unmarshal(b, &d); err != nil {
		log.Println("Parsing Error:", err)
		return
	}

	x := d.(map[string]interface{})
	x["version"] = version
	x["source"] = map[string]interface{}{
		"url":       p.SSHURLToRepo,
		"type":      "git",
		"reference": r.Commit.Id,
	}

	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if _, ok := c.projectsReferencesCommitId[p.Id]; !ok {
		c.projectsReferencesCommitId[p.Id] = make(map[string]string)
	}
	c.projectsReferencesCommitId[p.Id][r.Name] = r.Commit.Id
	if _, ok := c.cache[p.PathWithNamespace]; !ok {
		c.cache[p.PathWithNamespace] = make(map[string]interface{})
	}
	c.cache[p.PathWithNamespace][version] = x
	c.content = nil
}

func (c *ComposerRepository) parseVersion(r string) string {
	if c.versionRegexp.MatchString(r) {
		return r
	}

	return "dev-"+r
}
