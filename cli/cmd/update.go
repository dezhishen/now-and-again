package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/dezhishen/now-and-again/cli/internal/action"
	"github.com/spf13/cobra"
)

const ghReleaseAPI = "https://api.github.com/repos/dezhishen/now-and-again/releases/latest"

type ghRelease struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "检查更新并自更新",
	Long: `检查 GitHub 是否有新版本 CLI，以及后端版本是否匹配。

自更新流程:
  na update                  # 检查更新
  na update --yes            # 自动更新（不询问确认）
  na update --check          # 仅检查，不更新`,
	RunE: func(cmd *cobra.Command, args []string) error {
		checkOnly, _ := cmd.Flags().GetBool("check")
		yes, _ := cmd.Flags().GetBool("yes")

		// 1. Check GitHub latest release
		latest, err := fetchLatestRelease()
		if err != nil {
			return fmt.Errorf("获取 GitHub Release 失败: %w", err)
		}

		action.Printf("当前版本: %s", Version)
		action.Printf("最新版本: %s", latest.TagName)

		if Version == latest.TagName && Version != "dev" {
			action.Println("✅ 已是最新版本")
			if err := checkBackendVersion(); err != nil {
				action.Printf("⚠️  %v", err)
			}
			return nil
		}

		if checkOnly {
			action.Printf("📦 新版本可用: %s → %s", Version, latest.TagName)
			return nil
		}

		// 2. Confirm update
		if !yes {
			fmt.Printf("\n更新到 %s ? (y/N): ", latest.TagName)
			var response string
			fmt.Scanln(&response)
			if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
				action.Println("已取消")
				return nil
			}
		}

		// 3. Download and replace
		assetName := fmt.Sprintf("na_%s_%s.tar.gz", runtime.GOOS, runtime.GOARCH)
		var downloadURL string
		for _, a := range latest.Assets {
			if a.Name == assetName {
				downloadURL = a.BrowserDownloadURL
				break
			}
		}
		if downloadURL == "" {
			return fmt.Errorf("未找到匹配的二进制包: %s", assetName)
		}

		exe, err := os.Executable()
		if err != nil {
			return fmt.Errorf("获取当前可执行文件路径失败: %w", err)
		}

		action.Printf("⬇ 下载 %s ...", assetName)
		if err := downloadAndReplace(downloadURL, exe); err != nil {
			return fmt.Errorf("更新失败: %w", err)
		}

		action.Printf("✅ 已更新到 %s", latest.TagName)
		return nil
	},
}

func fetchLatestRelease() (*ghRelease, error) {
	resp, err := http.Get(ghReleaseAPI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub API 返回 %d", resp.StatusCode)
	}

	var r ghRelease
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

func checkBackendVersion() error {
	if na == nil {
		return fmt.Errorf("未初始化，请先运行 na init")
	}
	resp, err := http.Get(na.Config().ServerURL + "/api/system/status")
	if err != nil {
		return fmt.Errorf("无法连接后端: %w", err)
	}
	defer resp.Body.Close()

	var status struct {
		Status  string `json:"status"`
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("解析后端版本失败: %w", err)
	}

	if status.Version == "" || status.Version == "dev" {
		action.Printf("后端版本: %s (开发版)", status.Version)
		return nil
	}

	action.Printf("后端版本: %s", status.Version)
	if status.Version != Version {
		return fmt.Errorf("版本不匹配: CLI=%s, Backend=%s，建议同步更新", Version, status.Version)
	}
	return nil
}

func downloadAndReplace(url, targetPath string) error {
	// ── Pre-flight checks ──────────────────────────────────────
	// On Windows, replacing a running executable is not possible via rename.
	if runtime.GOOS == "windows" {
		return fmt.Errorf("Windows 不支持自更新（运行中的 exe 被锁定）。\n→ 请手动下载: %s", url)
	}

	// Check the target directory is writable
	testFile := targetPath + ".write-test"
	if err := os.WriteFile(testFile, []byte("x"), 0644); err != nil {
		return fmt.Errorf("目标目录无写入权限: %w\n→ 请使用 sudo 重新运行，或手动下载:\n  curl -L %s | tar xz", err, url)
	}
	os.Remove(testFile)

	// ── Download & extract ─────────────────────────────────────
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("下载返回 %d", resp.StatusCode)
	}

	// Extract binary from tar.gz
	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("解压 gzip 失败: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	var binData []byte
	// Look for the CLI binary in the archive (na-linux-amd64, na-darwin-arm64, na-windows-amd64.exe, etc.)
	binPrefix := "na-"
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取 tar 失败: %w", err)
		}
		if hdr.Typeflag == tar.TypeReg && strings.HasPrefix(hdr.Name, binPrefix) {
			binData, err = io.ReadAll(tr)
			if err != nil {
				return fmt.Errorf("读取二进制失败: %w", err)
			}
			break
		}
	}
	if binData == nil {
		return fmt.Errorf("tar.gz 中未找到 CLI 二进制文件")
	}

	// ── Validate binary ────────────────────────────────────────
	if !isValidBinary(binData) {
		return fmt.Errorf("下载的文件不是有效的可执行文件（校验失败），请重试")
	}

	// ── Atomic replace ─────────────────────────────────────────
	// Strategy (safe on Linux/macOS):
	//   1. Write new binary to .new
	//   2. Rename current → .bak   (kernel keeps old inode alive for running process)
	//   3. Rename .new → target    (new invocations get new binary)
	//   4. Remove .bak
	tmpFile := targetPath + ".new"
	if err := os.WriteFile(tmpFile, binData, 0755); err != nil {
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	bakFile := targetPath + ".bak"
	os.Remove(bakFile)
	if err := os.Rename(targetPath, bakFile); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("备份旧文件失败: %w", err)
	}

	if err := os.Rename(tmpFile, targetPath); err != nil {
		// Restore backup
		os.Rename(bakFile, targetPath)
		os.Remove(tmpFile)
		return fmt.Errorf("替换文件失败: %w\n→ 可能是权限问题，请使用 sudo 重试", err)
	}

	os.Remove(bakFile)
	return nil
}

// isValidBinary checks magic bytes for known executable formats.
func isValidBinary(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	// ELF (Linux): 0x7F 'E' 'L' 'F'
	if data[0] == 0x7F && data[1] == 'E' && data[2] == 'L' && data[3] == 'F' {
		return true
	}
	// Mach-O (macOS) 64-bit: 0xCF 0xFA 0xED 0xFE
	if data[0] == 0xCF && data[1] == 0xFA && data[2] == 0xED && data[3] == 0xFE {
		return true
	}
	// Mach-O (macOS) 32-bit: 0xCE 0xFA 0xED 0xFE
	if data[0] == 0xCE && data[1] == 0xFA && data[2] == 0xED && data[3] == 0xFE {
		return true
	}
	// Mach-O fat binary: 0xCA 0xFE 0xBA 0xBE
	if data[0] == 0xCA && data[1] == 0xFE && data[2] == 0xBA && data[3] == 0xBE {
		return true
	}
	// PE (Windows): 'M' 'Z'
	if data[0] == 'M' && data[1] == 'Z' {
		return true
	}
	return false
}

func init() {
	updateCmd.Flags().Bool("check", false, "仅检查更新，不执行")
	updateCmd.Flags().BoolP("yes", "y", false, "跳过确认，直接更新")
	rootCmd.AddCommand(updateCmd)
}
