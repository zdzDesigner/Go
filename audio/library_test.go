package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLibraryRefreshAndPathFor(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "有声书")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("创建测试目录失败: %v", err)
	}
	filePath := filepath.Join(nested, "第一章.mp3")
	if err := os.WriteFile(filePath, []byte("audio"), 0o644); err != nil {
		t.Fatalf("创建测试音频失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "忽略.txt"), []byte("text"), 0o644); err != nil {
		t.Fatalf("创建测试文本失败: %v", err)
	}

	library, err := NewLibrary(root)
	if err != nil {
		t.Fatalf("创建音频库失败: %v", err)
	}
	if err := library.Refresh(); err != nil {
		t.Fatalf("扫描音频库失败: %v", err)
	}
	items := library.List()
	if len(items) != 1 {
		t.Fatalf("音频数量=%d，期望=1", len(items))
	}
	item := items[0]
	if item.RelativePath != "有声书/第一章.mp3" {
		t.Fatalf("相对路径=%q，期望=%q", item.RelativePath, "有声书/第一章.mp3")
	}
	if item.ContentType != "audio/mpeg" {
		t.Fatalf("Content-Type=%q，期望=%q", item.ContentType, "audio/mpeg")
	}
	resolvedPath, err := library.PathFor(item)
	if err != nil {
		t.Fatalf("解析音频路径失败: %v", err)
	}
	if resolvedPath != filePath {
		t.Fatalf("解析路径=%q，期望=%q", resolvedPath, filePath)
	}

	firstID := item.ID
	if err := library.Refresh(); err != nil {
		t.Fatalf("二次扫描音频库失败: %v", err)
	}
	if library.List()[0].ID != firstID {
		t.Fatalf("同一相对路径的资源 ID 不稳定")
	}
}

func TestLibraryRejectsFileRoot(t *testing.T) {
	filePath := filepath.Join(t.TempDir(), "not-a-directory.mp3")
	if err := os.WriteFile(filePath, []byte("audio"), 0o644); err != nil {
		t.Fatalf("创建测试文件失败: %v", err)
	}
	if _, err := NewLibrary(filePath); err == nil {
		t.Fatal("文件路径应被拒绝为音频根目录")
	}
}
