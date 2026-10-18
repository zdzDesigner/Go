// [输入]: 配置的音频根目录及其文件系统内容。
// [输出]: 音频资源列表、资源 ID、文件元数据和安全的文件路径。
// [定位]: 音频库领域模块，隔离文件扫描和路径安全规则。
// [同步]: CONTEXT.md、README.md。
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

var supportedAudioTypes = map[string]string{
	".aac":  "audio/aac",
	".flac": "audio/flac",
	".m4a":  "audio/mp4",
	".mp3":  "audio/mpeg",
	".oga":  "audio/ogg",
	".ogg":  "audio/ogg",
	".opus": "audio/ogg",
	".wav":  "audio/wav",
	".webm": "audio/webm",
}

// Audio 描述一个可播放的音频资源。
type Audio struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	RelativePath string    `json:"path"`
	Size         int64     `json:"size"`
	ModifiedAt   time.Time `json:"modified_at"`
	ContentType  string    `json:"content_type"`
}

// Library 管理音频资源索引。
type Library struct {
	root  string
	mu    sync.RWMutex
	items map[string]Audio
}

// NewLibrary 创建音频库并校验根目录。
func NewLibrary(root string) (*Library, error) {
	absoluteRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("解析音频根目录 %q 失败: %w", root, err)
	}
	resolvedRoot, err := filepath.EvalSymlinks(absoluteRoot)
	if err != nil {
		return nil, fmt.Errorf("解析音频根目录 %q 的符号链接失败: %w", root, err)
	}
	info, err := os.Stat(resolvedRoot)
	if err != nil {
		return nil, fmt.Errorf("读取音频根目录 %q 失败: %w", resolvedRoot, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("音频根目录 %q 不是目录", resolvedRoot)
	}

	return &Library{
		root:  resolvedRoot,
		items: make(map[string]Audio),
	}, nil
}

// Refresh 递归扫描音频根目录并替换索引。
func (l *Library) Refresh() error {
	items := make(map[string]Audio)
	err := filepath.WalkDir(l.root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("扫描路径 %q 失败: %w", path, walkErr)
		}
		if entry.IsDir() {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("读取音频文件 %q 信息失败: %w", path, err)
		}
		if !info.Mode().IsRegular() {
			return nil
		}

		extension := strings.ToLower(filepath.Ext(entry.Name()))
		contentType, supported := supportedAudioTypes[extension]
		if !supported {
			return nil
		}

		relativePath, err := filepath.Rel(l.root, path)
		if err != nil {
			return fmt.Errorf("计算音频文件 %q 的相对路径失败: %w", path, err)
		}
		relativePath = filepath.ToSlash(relativePath)
		id := resourceID(relativePath)
		items[id] = Audio{
			ID:           id,
			Name:         entry.Name(),
			RelativePath: relativePath,
			Size:         info.Size(),
			ModifiedAt:   info.ModTime().UTC(),
			ContentType:  contentType,
		}
		return nil
	})
	if err != nil {
		return err
	}

	l.mu.Lock()
	l.items = items
	l.mu.Unlock()
	return nil
}

// List 返回按相对路径排序的音频资源列表。
func (l *Library) List() []Audio {
	l.mu.RLock()
	items := make([]Audio, 0, len(l.items))
	for _, item := range l.items {
		items = append(items, item)
	}
	l.mu.RUnlock()

	sort.Slice(items, func(i, j int) bool {
		return items[i].RelativePath < items[j].RelativePath
	})
	return items
}

// Get 返回指定 ID 的音频资源。
func (l *Library) Get(id string) (Audio, bool) {
	l.mu.RLock()
	item, ok := l.items[id]
	l.mu.RUnlock()
	return item, ok
}

// PathFor 返回资源对应的安全文件路径，并拒绝目录穿越或逃逸的符号链接。
func (l *Library) PathFor(item Audio) (string, error) {
	path := filepath.Join(l.root, filepath.FromSlash(item.RelativePath))
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", fmt.Errorf("解析音频资源 id=%s path=%q 失败: %w", item.ID, item.RelativePath, err)
	}
	relativePath, err := filepath.Rel(l.root, resolvedPath)
	if err != nil {
		return "", fmt.Errorf("校验音频资源 id=%s path=%q 失败: %w", item.ID, item.RelativePath, err)
	}
	if relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("音频资源 id=%s path=%q 超出音频根目录", item.ID, item.RelativePath)
	}
	return resolvedPath, nil
}

func resourceID(relativePath string) string {
	digest := sha256.Sum256([]byte(relativePath))
	return hex.EncodeToString(digest[:])
}
