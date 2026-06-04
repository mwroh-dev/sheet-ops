package render

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

const maxRendererOutputBytes = 4096

var commandTimeout = 30 * time.Second

type Artifact struct {
	InputWorkbook string        `json:"input_workbook"`
	Status        string        `json:"status"`
	Renderer      RendererTools `json:"renderer"`
	PDFPath       string        `json:"pdf_path,omitempty"`
	PreviewPath   string        `json:"preview_path,omitempty"`
	PageCount     int           `json:"page_count,omitempty"`
	Limitation    string        `json:"limitation"`
	Reasons       []string      `json:"reasons,omitempty"`
	Checks        Checks        `json:"checks"`
}

type RendererTools struct {
	Spreadsheet string `json:"spreadsheet,omitempty"`
	Rasterizer  string `json:"rasterizer,omitempty"`
}

type Checks struct {
	FileOpens           bool `json:"file_opens"`
	PreviewCreated      bool `json:"preview_created"`
	PreviewExists       bool `json:"preview_exists"`
	PreviewNonZeroBytes bool `json:"preview_nonzero_bytes"`
}

func RenderWorkbookPreview(inputWorkbook string, outputDir string) (Artifact, error) {
	artifact := Artifact{
		InputWorkbook: inputWorkbook,
		Status:        "unavailable",
		Renderer:      RendererTools{},
		Limitation:    "render artifact emission is not visual quality verification",
		Checks:        Checks{},
	}
	if err := workbookOpens(inputWorkbook); err != nil {
		artifact.Status = "failed"
		artifact.Reasons = []string{err.Error()}
		return artifact, nil
	}
	artifact.Checks.FileOpens = true

	tools := detectTools()
	artifact.Renderer = tools
	if tools.Spreadsheet == "" {
		artifact.Reasons = append(artifact.Reasons, "spreadsheet renderer not found")
	}
	if tools.Rasterizer == "" {
		artifact.Reasons = append(artifact.Reasons, "PDF rasterizer not found")
	}
	if len(artifact.Reasons) > 0 {
		return artifact, nil
	}

	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return Artifact{}, err
	}
	pdfPath, err := exportPDF(tools.Spreadsheet, inputWorkbook, outputDir)
	if err != nil {
		artifact.Status = "failed"
		artifact.Reasons = []string{err.Error()}
		return artifact, nil
	}
	previewPath, err := renderPNG(tools.Rasterizer, pdfPath, outputDir)
	if err != nil {
		artifact.Status = "failed"
		artifact.PDFPath = pdfPath
		artifact.Reasons = []string{err.Error()}
		return artifact, nil
	}

	artifact.PDFPath = pdfPath
	artifact.PreviewPath = previewPath
	artifact.PageCount = 1
	artifact.Checks.PreviewCreated = true
	if info, err := os.Stat(previewPath); err == nil {
		artifact.Checks.PreviewExists = true
		artifact.Checks.PreviewNonZeroBytes = info.Size() > 0
	}
	if !artifact.Checks.PreviewNonZeroBytes {
		artifact.Status = "failed"
		artifact.Reasons = []string{"PNG render preview is empty"}
		return artifact, nil
	}
	artifact.Status = "rendered"
	return artifact, nil
}

func detectTools() RendererTools {
	return RendererTools{
		Spreadsheet: firstTool("soffice", "libreoffice"),
		Rasterizer:  firstTool("pdftoppm"),
	}
}

func firstTool(names ...string) string {
	for _, name := range names {
		path, err := exec.LookPath(name)
		if err == nil {
			return path
		}
	}
	return ""
}

func workbookOpens(inputWorkbook string) error {
	file, err := excelize.OpenFile(inputWorkbook)
	if err != nil {
		return fmt.Errorf("workbook does not open: %w", err)
	}
	return file.Close()
}

func exportPDF(spreadsheetTool string, inputWorkbook string, outputDir string) (string, error) {
	output, err := runRendererCommand("PDF export", spreadsheetTool, "--headless", "--convert-to", "pdf", "--outdir", outputDir, inputWorkbook)
	if err != nil {
		if strings.Contains(err.Error(), "timed out after") {
			return "", err
		}
		return "", fmt.Errorf("PDF export failed: %v: %s", err, output)
	}
	base := strings.TrimSuffix(filepath.Base(inputWorkbook), filepath.Ext(inputWorkbook))
	pdfPath := filepath.Join(outputDir, base+".pdf")
	if _, err := os.Stat(pdfPath); err != nil {
		return "", fmt.Errorf("PDF export missing %s: %w", pdfPath, err)
	}
	return pdfPath, nil
}

func renderPNG(rasterizer string, pdfPath string, outputDir string) (string, error) {
	prefix := filepath.Join(outputDir, "preview")
	output, err := runRendererCommand("PNG render", rasterizer, "-png", "-f", "1", "-singlefile", pdfPath, prefix)
	if err != nil {
		if strings.Contains(err.Error(), "timed out after") {
			return "", err
		}
		return "", fmt.Errorf("PNG render failed: %v: %s", err, output)
	}
	previewPath := prefix + ".png"
	if _, err := os.Stat(previewPath); err != nil {
		return "", fmt.Errorf("PNG render missing %s: %w", previewPath, err)
	}
	return previewPath, nil
}

func runRendererCommand(label string, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, name, args...)
	output, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("%s timed out after %s", label, commandTimeout)
	}
	return boundedOutput(output), err
}

func boundedOutput(output []byte) string {
	text := strings.TrimSpace(string(output))
	if len(text) <= maxRendererOutputBytes {
		return text
	}
	return text[:maxRendererOutputBytes] + "... [truncated]"
}
