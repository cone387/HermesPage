package storage

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

type Report struct {
	ID         string   `json:"id"`
	Filename   string   `json:"filename"`
	Title      string   `json:"title"`
	Category   string   `json:"category"`
	Tags       []string `json:"tags"`
	Size       int64    `json:"size"`
	CreatedAt  string   `json:"created_at"`
	URL        string   `json:"url"`
	Owner      string   `json:"owner"`
	Visibility string   `json:"visibility"`
}

type Metadata struct {
	Reports []Report `json:"reports"`
}

type Storage struct {
	mu      sync.RWMutex
	dataDir string
	meta    *Metadata
}

func New(dataDir string) (*Storage, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	s := &Storage{dataDir: dataDir}
	if err := s.loadMetadata(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Storage) metadataPath() string {
	return filepath.Join(s.dataDir, "metadata.json")
}

func (s *Storage) loadMetadata() error {
	s.meta = &Metadata{Reports: []Report{}}
	data, err := os.ReadFile(s.metadataPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, s.meta)
}

func (s *Storage) saveMetadata() error {
	data, err := json.MarshalIndent(s.meta, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.metadataPath(), data, 0644)
}

func (s *Storage) generateID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func sanitizeFilename(name string) string {
	name = strings.ReplaceAll(name, "..", "")
	name = strings.ReplaceAll(name, "/", "")
	name = strings.ReplaceAll(name, "\\", "")
	return name
}

func sanitizeCategory(cat string) string {
	cat = strings.ToLower(strings.TrimSpace(cat))
	reg := regexp.MustCompile(`[^a-z0-9\-]`)
	cat = reg.ReplaceAllString(cat, "-")
	if cat == "" {
		cat = "uncategorized"
	}
	return cat
}

var (
	reTitleMeta = regexp.MustCompile(`(?i)<meta\s+name=["']hermes-title["']\s+content=["']([^"']+)["']`)
	reTagsMeta  = regexp.MustCompile(`(?i)<meta\s+name=["']hermes-tags["']\s+content=["']([^"']+)["']`)
	reTitle     = regexp.MustCompile(`(?i)<title>([^<]+)</title>`)
)

func ExtractTitle(html []byte, filename string) string {
	if m := reTitleMeta.FindSubmatch(html); len(m) > 1 {
		return string(m[1])
	}
	if m := reTitle.FindSubmatch(html); len(m) > 1 {
		return string(m[1])
	}
	name := strings.TrimSuffix(filename, filepath.Ext(filename))
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return name
}

func ExtractTags(html []byte) []string {
	if m := reTagsMeta.FindSubmatch(html); len(m) > 1 {
		raw := strings.Split(string(m[1]), ",")
		tags := make([]string, 0, len(raw))
		for _, t := range raw {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
		return tags
	}
	return []string{}
}

func (s *Storage) Save(content []byte, filename, title, category, tagsStr, owner, visibility string) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := s.generateID()
	category = sanitizeCategory(category)
	filename = sanitizeFilename(filename)
	if filename == "" {
		filename = fmt.Sprintf("report-%s.html", id)
	}

	// ensure unique filename
	for _, r := range s.meta.Reports {
		if r.Category == category && r.Filename == filename {
			ext := filepath.Ext(filename)
			base := strings.TrimSuffix(filename, ext)
			filename = fmt.Sprintf("%s-%s%s", base, id, ext)
			break
		}
	}

	// title
	if title == "" {
		title = ExtractTitle(content, filename)
	}

	// tags
	var tags []string
	if tagsStr != "" {
		for _, t := range strings.Split(tagsStr, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	} else {
		tags = ExtractTags(content)
	}

	// write file
	catDir := filepath.Join(s.dataDir, category)
	if err := os.MkdirAll(catDir, 0755); err != nil {
		return nil, err
	}
	filePath := filepath.Join(catDir, filename)
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		return nil, err
	}

	report := Report{
		ID:         id,
		Filename:   filename,
		Title:      title,
		Category:   category,
		Tags:       tags,
		Size:       int64(len(content)),
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		URL:        fmt.Sprintf("/reports/%s/%s", category, filename),
		Owner:      owner,
		Visibility: visibility,
	}

	s.meta.Reports = append([]Report{report}, s.meta.Reports...)
	if err := s.saveMetadata(); err != nil {
		return nil, err
	}

	return &report, nil
}

func (s *Storage) List() []Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.meta.Reports
}

func (s *Storage) Get(id string) *Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.meta.Reports {
		if r.ID == id {
			return &r
		}
	}
	return nil
}

func (s *Storage) Delete(id string) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx := -1
	for i, r := range s.meta.Reports {
		if r.ID == id {
			idx = i
			break
		}
	}
	if idx == -1 {
		return nil, fmt.Errorf("not found")
	}

	report := s.meta.Reports[idx]

	// delete file
	filePath := filepath.Join(s.dataDir, report.Category, report.Filename)
	os.Remove(filePath)

	// remove from metadata
	s.meta.Reports = append(s.meta.Reports[:idx], s.meta.Reports[idx+1:]...)
	if err := s.saveMetadata(); err != nil {
		return nil, err
	}

	return &report, nil
}

func (s *Storage) Categories() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	seen := map[string]bool{}
	for _, r := range s.meta.Reports {
		seen[r.Category] = true
	}
	cats := make([]string, 0, len(seen))
	for c := range seen {
		cats = append(cats, c)
	}
	return cats
}

func (s *Storage) FindByPath(category, filename string) *Report {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, r := range s.meta.Reports {
		if r.Category == category && r.Filename == filename {
			return &r
		}
	}
	return nil
}
