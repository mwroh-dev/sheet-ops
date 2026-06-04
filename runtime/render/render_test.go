package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/xuri/excelize/v2"
)

func TestRenderWorkbookPreviewReportsUnavailableWhenToolsAreMissing(t *testing.T) {
	t.Setenv("PATH", "")
	inputFile := writeRenderWorkbook(t)

	artifact, err := RenderWorkbookPreview(inputFile, filepath.Join(t.TempDir(), "render"))
	if err != nil {
		t.Fatalf("RenderWorkbookPreview: %v", err)
	}
	if artifact.Status != "unavailable" {
		t.Fatalf("status=%q want unavailable", artifact.Status)
	}
	if artifact.Checks.FileOpens != true {
		t.Fatalf("file_opens=%v want true", artifact.Checks.FileOpens)
	}
	if artifact.Checks.PreviewCreated {
		t.Fatalf("preview_created=true want false")
	}
	if artifact.Checks.PreviewExists {
		t.Fatalf("preview_exists=true want false")
	}
	if artifact.Checks.PreviewNonZeroBytes {
		t.Fatalf("preview_nonzero_bytes=true want false")
	}
	if !strings.Contains(artifact.Limitation, "not visual quality verification") {
		t.Fatalf("limitation=%q must scope render artifact semantics", artifact.Limitation)
	}
	if len(artifact.Reasons) == 0 {
		t.Fatalf("reasons missing for unavailable renderer: %+v", artifact)
	}
}

func TestRenderWorkbookPreviewCapturesPDFAndPNGMetadata(t *testing.T) {
	tempDir := t.TempDir()
	toolDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(toolDir): %v", err)
	}
	writeExecutable(t, filepath.Join(toolDir, "soffice"), `#!/bin/sh
out=""
input=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--outdir" ]; then
    shift
    out="$1"
  else
    input="$1"
  fi
  shift
done
base=$(basename "$input" .xlsx)
mkdir -p "$out"
printf "pdf" > "$out/$base.pdf"
`)
	writeExecutable(t, filepath.Join(toolDir, "pdftoppm"), `#!/bin/sh
for arg in "$@"; do
  prefix="$arg"
done
printf "png" > "$prefix.png"
`)
	t.Setenv("PATH", toolDir+string(os.PathListSeparator)+"/bin"+string(os.PathListSeparator)+"/usr/bin")

	inputFile := writeRenderWorkbook(t)
	renderDir := filepath.Join(tempDir, "render")
	artifact, err := RenderWorkbookPreview(inputFile, renderDir)
	if err != nil {
		t.Fatalf("RenderWorkbookPreview: %v", err)
	}
	if artifact.Status != "rendered" {
		t.Fatalf("status=%q want rendered: %+v", artifact.Status, artifact)
	}
	if artifact.PDFPath == "" || artifact.PreviewPath == "" {
		t.Fatalf("render paths missing: %+v", artifact)
	}
	if _, err := os.Stat(artifact.PDFPath); err != nil {
		t.Fatalf("pdf missing: %v", err)
	}
	if _, err := os.Stat(artifact.PreviewPath); err != nil {
		t.Fatalf("preview missing: %v", err)
	}
	if !artifact.Checks.FileOpens || !artifact.Checks.PreviewCreated || !artifact.Checks.PreviewExists || !artifact.Checks.PreviewNonZeroBytes {
		t.Fatalf("checks=%+v want file_opens, preview_created, preview_exists, and preview_nonzero_bytes", artifact.Checks)
	}
	if !strings.Contains(artifact.Limitation, "not visual quality verification") {
		t.Fatalf("limitation=%q must scope render artifact semantics", artifact.Limitation)
	}
}

func TestRenderWorkbookPreviewFailsWhenPreviewIsEmpty(t *testing.T) {
	tempDir := t.TempDir()
	toolDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(toolDir): %v", err)
	}
	writeExecutable(t, filepath.Join(toolDir, "soffice"), `#!/bin/sh
out=""
input=""
while [ "$#" -gt 0 ]; do
  if [ "$1" = "--outdir" ]; then
    shift
    out="$1"
  else
    input="$1"
  fi
  shift
done
base=$(basename "$input" .xlsx)
mkdir -p "$out"
printf "pdf" > "$out/$base.pdf"
`)
	writeExecutable(t, filepath.Join(toolDir, "pdftoppm"), `#!/bin/sh
for arg in "$@"; do
  prefix="$arg"
done
: > "$prefix.png"
`)
	t.Setenv("PATH", toolDir+string(os.PathListSeparator)+"/bin"+string(os.PathListSeparator)+"/usr/bin")

	artifact, err := RenderWorkbookPreview(writeRenderWorkbook(t), filepath.Join(tempDir, "render"))
	if err != nil {
		t.Fatalf("RenderWorkbookPreview: %v", err)
	}
	if artifact.Status != "failed" {
		t.Fatalf("status=%q want failed for empty preview: %+v", artifact.Status, artifact)
	}
	if !artifact.Checks.PreviewExists {
		t.Fatalf("preview_exists=false want true for empty preview file")
	}
	if artifact.Checks.PreviewNonZeroBytes {
		t.Fatalf("preview_nonzero_bytes=true want false")
	}
	if len(artifact.Reasons) == 0 || !strings.Contains(artifact.Reasons[0], "empty") {
		t.Fatalf("reasons=%v want empty preview reason", artifact.Reasons)
	}
}

func TestRenderWorkbookPreviewFailsDeterministicallyWhenPDFExportTimesOut(t *testing.T) {
	originalTimeout := commandTimeout
	commandTimeout = 10 * time.Millisecond
	defer func() { commandTimeout = originalTimeout }()

	tempDir := t.TempDir()
	toolDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(toolDir): %v", err)
	}
	writeExecutable(t, filepath.Join(toolDir, "soffice"), "#!/bin/sh\nsleep 1\n")
	writeExecutable(t, filepath.Join(toolDir, "pdftoppm"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", toolDir)

	artifact, err := RenderWorkbookPreview(writeRenderWorkbook(t), filepath.Join(tempDir, "render"))
	if err != nil {
		t.Fatalf("RenderWorkbookPreview: %v", err)
	}
	if artifact.Status != "failed" {
		t.Fatalf("status=%q want failed: %+v", artifact.Status, artifact)
	}
	if len(artifact.Reasons) != 1 || !strings.Contains(artifact.Reasons[0], "PDF export timed out after") {
		t.Fatalf("reasons=%v want deterministic PDF timeout", artifact.Reasons)
	}
}

func TestRenderWorkbookPreviewTruncatesLongRendererOutput(t *testing.T) {
	tempDir := t.TempDir()
	toolDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(toolDir): %v", err)
	}
	writeExecutable(t, filepath.Join(toolDir, "soffice"), `#!/bin/sh
python3 - <<'PY'
print("x" * 5000)
PY
exit 2
`)
	writeExecutable(t, filepath.Join(toolDir, "pdftoppm"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", toolDir+string(os.PathListSeparator)+"/bin"+string(os.PathListSeparator)+"/usr/bin")

	artifact, err := RenderWorkbookPreview(writeRenderWorkbook(t), filepath.Join(tempDir, "render"))
	if err != nil {
		t.Fatalf("RenderWorkbookPreview: %v", err)
	}
	if artifact.Status != "failed" {
		t.Fatalf("status=%q want failed: %+v", artifact.Status, artifact)
	}
	if len(artifact.Reasons) != 1 || !strings.Contains(artifact.Reasons[0], "truncated") {
		t.Fatalf("reasons=%v want truncated renderer output marker", artifact.Reasons)
	}
	if len(artifact.Reasons[0]) > 4300 {
		t.Fatalf("reason length=%d want bounded output", len(artifact.Reasons[0]))
	}
}

func TestDetectToolsDoesNotAdvertiseUnsupportedImageMagickRasterizer(t *testing.T) {
	tempDir := t.TempDir()
	toolDir := filepath.Join(tempDir, "bin")
	if err := os.MkdirAll(toolDir, 0o755); err != nil {
		t.Fatalf("MkdirAll(toolDir): %v", err)
	}
	writeExecutable(t, filepath.Join(toolDir, "soffice"), "#!/bin/sh\nexit 0\n")
	writeExecutable(t, filepath.Join(toolDir, "magick"), "#!/bin/sh\nexit 0\n")
	t.Setenv("PATH", toolDir)

	tools := detectTools()
	if tools.Spreadsheet == "" {
		t.Fatalf("spreadsheet tool missing")
	}
	if tools.Rasterizer != "" {
		t.Fatalf("rasterizer=%q want empty when only unsupported magick exists", tools.Rasterizer)
	}
}

func writeRenderWorkbook(t *testing.T) string {
	t.Helper()
	inputFile := filepath.Join(t.TempDir(), "input.xlsx")
	file := excelize.NewFile()
	defer func() { _ = file.Close() }()
	if err := file.SetCellValue("Sheet1", "A1", "status"); err != nil {
		t.Fatalf("SetCellValue(A1): %v", err)
	}
	if err := file.SaveAs(inputFile); err != nil {
		t.Fatalf("SaveAs: %v", err)
	}
	return inputFile
}

func writeExecutable(t *testing.T, path string, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o755); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}
