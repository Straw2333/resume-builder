package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// ExportPDF 用本机 Chrome/Edge 无头浏览器将 HTML 打印为 A4 PDF。
// 优先 Chrome，其次 Edge（Windows 自带，成功率高），找不到则返回错误。
func ExportPDF(ctx context.Context, htmlPath, outputPath string) error {
	browsers := browserCandidates()
	if len(browsers) == 0 {
		return fmt.Errorf("未找到 Chrome 或 Edge 浏览器，无法导出 PDF（Word 导出不受影响）")
	}
	var lastErr error
	for _, b := range browsers {
		// --no-pdf-header-footer 去掉浏览器默认页眉页脚
		cmd := exec.CommandContext(ctx, b,
			"--headless",
			"--disable-gpu",
			"--no-sandbox",
			"--no-pdf-header-footer",
			"--print-to-pdf="+outputPath,
			"file:///"+filepath.ToSlash(htmlPath),
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			lastErr = fmt.Errorf("%s 导出失败: %v: %s", filepath.Base(b), err, string(out))
			continue
		}
		if _, err := os.Stat(outputPath); err == nil {
			return nil
		}
		lastErr = fmt.Errorf("%s 未生成 PDF 文件", filepath.Base(b))
	}
	return lastErr
}

func browserCandidates() []string {
	var dirs []string
	if runtime.GOOS == "windows" {
		for _, env := range []string{"PROGRAM_FILES", "PROGRAM_FILES_X86", "LOCALAPPDATA"} {
			v := os.Getenv(map[string]string{
				"PROGRAM_FILES":    "ProgramFiles",
				"PROGRAM_FILES_X86": "ProgramFiles(x86)",
				"LOCALAPPDATA":     "LOCALAPPDATA",
			}[env])
			if v == "" {
				continue
			}
			dirs = append(dirs,
				filepath.Join(v, "Google", "Chrome", "Application", "chrome.exe"),
				filepath.Join(v, "Microsoft", "Edge", "Application", "msedge.exe"),
			)
		}
	} else {
		dirs = []string{
			"/usr/bin/google-chrome", "/usr/bin/chromium", "/usr/bin/chromium-browser",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
		}
	}
	var found []string
	for _, p := range dirs {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			found = append(found, p)
		}
	}
	return found
}
