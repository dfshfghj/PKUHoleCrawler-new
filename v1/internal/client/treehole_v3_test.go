package client

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFlattenTagsSupportsTagNameChildren(t *testing.T) {
	_, filename, _, _ := runtime.Caller(0)
	root := filepath.Join(filepath.Dir(filename), "../..")
	data, err := os.ReadFile(filepath.Join(root, "doc", "tags.json"))
	if err != nil {
		t.Fatalf("read tags.json: %v", err)
	}
	var envelope struct {
		Data struct {
			List interface{} `json:"list"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal tags.json: %v", err)
	}
	tags := flattenTags(envelope.Data.List, 0)
	if len(tags) == 0 {
		t.Fatal("flattenTags returned no tags")
	}
	foundCategory := false
	foundChild := false
	for _, tag := range tags {
		if tag.ID == 22 && tag.Name == "课程学业" {
			foundCategory = true
		}
		if tag.ID == 1 && tag.Label == "课程心得" && tag.ParentID == 22 {
			foundChild = true
		}
	}
	if !foundCategory {
		t.Fatal("did not find top-level category tag")
	}
	if !foundChild {
		t.Fatal("did not find child tag parsed from tag_name/type_id")
	}
}
