package main

import "testing"

func TestApplyTextCorrections(t *testing.T) {
	text := "帮我把这个向木的衣赖从新装一下，然后跑一下单元侧试，如果报全线问题，就看一下接口的权现效验。"
	got, err := applyTextCorrections(text, defaultCorrections())
	if err != nil {
		t.Fatalf("applyTextCorrections returned error: %v", err)
	}
	want := "帮我把这个项目的依赖重新装一下，然后跑一下单元测试，如果报权限问题，就看一下接口的权限校验。"
	if got != want {
		t.Fatalf("corrected text mismatch: got %q, want %q", got, want)
	}
}

func TestApplyTextCorrectionsRejectsInvalidItem(t *testing.T) {
	_, err := applyTextCorrections("测试", "错误格式")
	if err == nil {
		t.Fatal("expected invalid correction item error")
	}
}
