package service

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/joaoleau/ezreports/internal/config"
	"github.com/joaoleau/ezreports/internal/utils"
)

type ReportData struct {
	Path       string
	Title      string
	Receiver   string
	PanelsData []PanelData
}

type PanelData struct {
	Path        string
	Title       string
	Description string
}

type ReportService struct {
	RendererURL     string
	RendererTimeout int
	RendererToken   string
	Template        string
}

func (r *ReportService) renderer(panel config.Panel) (*http.Response, error) {
	rawURL := r.RendererURL + "/render/d/" + panel.DashboardID
	u, err := url.Parse(rawURL)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	for _, p := range panel.Params {
		parts := strings.SplitN(p, "=", 2)
		if len(parts) != 2 {
			continue
		}
		q.Set(parts[0], parts[1])
	}

	u.RawQuery = q.Encode()

	client := &http.Client{
		Timeout: time.Duration(r.RendererTimeout) * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if r.RendererToken != "" {
		req.Header.Set("Authorization", "Bearer "+r.RendererToken)
	}
	req.Header.Set("Accept", "image/png")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		loc := resp.Header.Get("Location")
		return nil, fmt.Errorf("renderer redirected to %s", loc)
	}

	ct := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(ct, "image/png") {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("renderer did not return PNG, content-type=%s body=%s", ct, body)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("renderer error %d: %s", resp.StatusCode, body)
	}

	return resp, nil
}

func (r *ReportService) generatePanel(resp *http.Response, fileName string) error {
	defer resp.Body.Close()

	out, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func (r *ReportService) generateReportFolder(reportFolderName string) string {
	basePath := "reports"
	reportPath := basePath + "/" + reportFolderName
	_ = os.MkdirAll(basePath, 0755)

	err := os.MkdirAll(reportPath, 0755)
	if err != nil {
		panic(err)
	}

	return reportPath + "/"
}

func (r *ReportService) buildReportData(
	report config.Report,
	reportPath string,
	panels []PanelData,
) ReportData {
	return ReportData{
		Path:       reportPath,
		Title:      report.Title,
		Receiver:   report.Receiver,
		PanelsData: panels,
	}
}

func (r *ReportService) generateHTML(data ReportData) (string, error) {
	funcMap := template.FuncMap{
		"now": func() string {
			return time.Now().Format("02/01/2006 15:04")
		},
	}

	tpl, err := template.New("report").
		Funcs(funcMap).
		Parse(r.Template)
	if err != nil {
		return "", err
	}

	htmlPath := data.Path + "report.html"

	file, err := os.Create(htmlPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	if err := tpl.Execute(file, data); err != nil {
		return "", err
	}

	return htmlPath, nil
}

func (r *ReportService) exportPNG(htmlPath string) (string, error) {
	pngPath := strings.Replace(htmlPath, ".html", ".png", 1)

	cmd := exec.Command(
		"wkhtmltoimage",
		"--enable-local-file-access",
		htmlPath,
		pngPath,
	)

	return pngPath, cmd.Run()
}

func (r *ReportService) exportPDF(htmlPath string) (string, error) {
	pdfPath := strings.Replace(htmlPath, ".html", ".pdf", 1)

	cmd := exec.Command(
		"wkhtmltopdf",
		"--enable-local-file-access",
		htmlPath,
		pdfPath,
	)

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return pdfPath, nil
}

func (r *ReportService) Generate(report config.Report) ([]string, error) {
	var outputs []string
	reportFolder := utils.SlugPanelName(report.Title, false)
	reportPath := r.generateReportFolder(reportFolder)

	var panels []PanelData

	for _, p := range report.Panels {
		resp, err := r.renderer(p)
		if err != nil {
			return nil, err
		}

		panelName := utils.SlugPanelName(p.Title, true) + ".png"
		panelPath := reportPath + panelName

		if err := r.generatePanel(resp, panelPath); err != nil {
			return nil, err
		}

		panels = append(panels, PanelData{
			Path:        panelName,
			Title:       p.Title,
			Description: p.Description,
		})
	}

	reportData := r.buildReportData(report, reportPath, panels)

	htmlPath, err := r.generateHTML(reportData)
	if err != nil {
		return nil, err
	}

	if report.Format == "pdf" {
		pdf, err := r.exportPDF(htmlPath)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, pdf)
	}

	if report.Format == "png" {
		png, err := r.exportPNG(htmlPath)
		if err != nil {
			return nil, err
		}
		outputs = append(outputs, png)
	}

	return outputs, nil
}
