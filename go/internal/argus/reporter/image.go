package reporter

import (
	"fmt"
	"gitlab.bellsoft.net/devops/sre-toolkit/go/internal/argus/domain"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type ImageReporter struct {
	outputDir string
}

func NewImageReporter(outputDir string) *ImageReporter {
	return &ImageReporter{
		outputDir: outputDir,
	}
}

func (r *ImageReporter) Generate(results map[string]domain.AnalysisResult, context, cluster string, startTime time.Time) error {
	hasActionRequired := false
	for _, result := range results {
		if result.ManualResources > 0 {
			hasActionRequired = true
			break
		}
	}

	if !hasActionRequired {
		fmt.Printf("✅ 조치가 필요한 리소스가 없어 이미지를 생성하지 않습니다.\n")
		return nil
	}

	htmlFile := filepath.Join(r.outputDir, fmt.Sprintf("%s.html", startTime.Format("20060102_150405")))

	time.Sleep(100 * time.Millisecond)

	imageFile := filepath.Join(r.outputDir, fmt.Sprintf("%s.png", startTime.Format("20060102_150405")))

	if err := r.convertWithWkhtmltoimage(htmlFile, imageFile); err == nil {
		fmt.Printf("🖼️  이미지 생성 완료: %s\n", imageFile)
		return nil
	}

	fmt.Printf("⚠️  wkhtmltoimage를 찾을 수 없습니다. 대체 방법을 시도합니다.\n")

	if err := r.generateAlternativeInstructions(htmlFile); err != nil {
		return fmt.Errorf("failed to generate alternative instructions: %w", err)
	}

	return nil
}

func (r *ImageReporter) convertWithWkhtmltoimage(htmlFile, imageFile string) error {
	cmd := exec.Command("wkhtmltoimage", "--version")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wkhtmltoimage not found")
	}

	args := []string{
		"--width", "1200",
		"--quality", "100",
		"--format", "png",
		"--enable-local-file-access",
		"--encoding", "utf-8",
		"--no-stop-slow-scripts",
		"--javascript-delay", "1000",
		htmlFile,
		imageFile,
	}

	cmd = exec.Command("wkhtmltoimage", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert HTML to image: %w\nOutput: %s", err, string(output))
	}

	return nil
}

func (r *ImageReporter) generateAlternativeInstructions(htmlFile string) error {
	instructions := `
📸 이미지 변환 도구 설치 안내
================================

HTML 파일이 생성되었습니다: %s

이미지로 변환하려면 다음 도구 중 하나를 설치하세요:

### 1. wkhtmltoimage 설치 (권장)
`

	switch runtime.GOOS {
	case "darwin":
		instructions += `
macOS:
  brew install wkhtmltopdf
  # Korean fonts are supported by default

After installation, run again to automatically generate the image.
`
	case "linux":
		instructions += `
Ubuntu/Debian:
  sudo apt-get install wkhtmltopdf
  # Install Korean fonts (required)
  sudo apt-get install fonts-noto-cjk fonts-nanum

CentOS/RHEL:
  sudo yum install wkhtmltopdf
  # Install Korean fonts (required)
  sudo yum install google-noto-cjk-fonts

After installation, run again to automatically generate the image.
`
	case "windows":
		instructions += `
Windows:
  1. https://wkhtmltopdf.org/downloads.html 에서 다운로드
  2. 설치 후 PATH에 추가
  3. 다시 실행하면 자동으로 이미지가 생성됩니다.
`
	}

	instructions += `
### 2. 수동 변환
브라우저에서 HTML 파일을 열고:
1. 전체 화면 캡처 (Chrome: Ctrl+Shift+P → "Capture full size screenshot")
2. 인쇄 → PDF로 저장 → PDF를 이미지로 변환

### 3. 온라인 변환
HTML 파일을 온라인 HTML-to-Image 변환 서비스에 업로드
`

	fmt.Printf(instructions, htmlFile)

	instructionFile := filepath.Join(r.outputDir, "IMAGE_CONVERSION_GUIDE.txt")
	if err := os.WriteFile(instructionFile, []byte(fmt.Sprintf(instructions, htmlFile)), 0644); err != nil {
		return fmt.Errorf("failed to write instructions: %w", err)
	}

	return nil
}
