package main

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"

	"github.com/joaoleau/ezreports/internal/config"
	"github.com/joaoleau/ezreports/internal/service"
	"github.com/joaoleau/ezreports/internal/utils"
)

var RENDERER_URL string
var RENDERER_TIMEOUT string
var CONFIG_PATH string

func loadEnv() {
	_ = godotenv.Load()
}

func main() {
	os.Exit(run())
}

func run() int {
	loadEnv()

	configPath := os.Getenv("CONFIG_PATH")
	rendererToken := os.Getenv("RENDERER_TOKEN")
	rendererURL := os.Getenv("RENDERER_URL")
	rendererTimeoutStr := os.Getenv("RENDERER_TIMEOUT")

	rendererTimeout, err := strconv.Atoi(rendererTimeoutStr)
	if err != nil {
		println("invalid RENDERER_TIMEOUT")
		return 1
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		println(err.Error())
		return 1
	}

	for _, r := range cfg.Reports {
		if !utils.IsValidReceiver(r.Receiver, cfg.Receivers) {
			continue
		}
		reportSvc := service.ReportService{
			RendererURL:     rendererURL,
			RendererTimeout: rendererTimeout,
			RendererToken: rendererToken,
			Template: r.Template,
		}
		files, err := reportSvc.Generate(r)
		if err != nil {
			continue
		}
		rc := service.GetReceiver(r.Receiver, cfg.Receivers); if rc == nil {
			continue
		}

		if rc.EmailConfigs.To != nil {
			smtpSvc := service.SMTPService{
				SMTPConfig: cfg.Global.SMTPConfig,
			}

			smtpSvc.SendMail(
				rc.EmailConfigs.To,
				rc.EmailConfigs.Subject,
				rc.EmailConfigs.HTMLBody,
				files,
			)
		}
	}

	return 0
}
