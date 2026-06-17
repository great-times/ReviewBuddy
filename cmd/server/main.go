package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"reviewbuddy/internal/api"
	"reviewbuddy/internal/db"
	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
	"reviewbuddy/internal/service/agent"
	"reviewbuddy/internal/service/auth"
	"reviewbuddy/internal/service/dashboard"
	"reviewbuddy/internal/service/guide"
	"reviewbuddy/internal/service/knowledge"
	"reviewbuddy/internal/service/reviewconfig"
	"reviewbuddy/internal/service/settings"
	"reviewbuddy/internal/service/template"
	"reviewbuddy/internal/service/user"
	"reviewbuddy/pkg/config"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	database, err := db.Open(cfg.Database.Path)
	if err != nil {
		log.Fatalf("open db: %v", err)
	}
	defer database.Close()

	// repos
	tplRepo := repo.NewTemplateRepo(database)
	guideRepo := repo.NewGuideRepo(database)
	reviewRepo := repo.NewReviewRepo(database)
	knowledgeRepo := repo.NewKnowledgeRepo(database)
	userRepo := repo.NewUserRepo(database)
	settingsRepo := repo.NewSettingsRepo(database)
	reviewConfigRepo := repo.NewReviewConfigRepo(database)

	settingsSvc := settings.NewService(settingsRepo, cfg.Agent)
	reviewConfigSvc := reviewconfig.NewService(reviewConfigRepo, userRepo)
	dashboardSvc := dashboard.NewService(tplRepo, guideRepo, reviewRepo, knowledgeRepo, reviewConfigRepo)

	// agent (Hermes / OpenAI 兼容；未配置时回退 mock)
	ag := agent.NewDynamicAdapter(settingsSvc)
	log.Printf("agent adapter: %s", ag.Name())

	// services
	knowledgeSvc := knowledge.NewService(knowledgeRepo)
	tplSvc := template.NewService(tplRepo)
	guideSvc := guide.NewService(guideRepo, tplRepo, ag, knowledgeSvc)
	reviewSvc := guide.NewReviewService(reviewRepo, guideRepo, tplRepo, userRepo, knowledgeSvc)
	userSvc := user.NewService(userRepo, reviewConfigSvc)
	authSvc := auth.NewService(userRepo)

	seedTemplates(tplSvc)
	userSvc.SeedDefaults()

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:    []string{"Origin", "Content-Type", "Authorization"},
	}))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "agent": ag.Name()})
	})

	apiGroup := r.Group("/api")
	authHandler := api.NewAuthHandler(authSvc)
	authHandler.Register(apiGroup)

	protected := apiGroup.Group("")
	protected.Use(authHandler.AuthRequired())
	api.NewTemplateHandler(tplSvc).Register(protected)
	api.NewGuideHandler(guideSvc).Register(protected)
	api.NewReviewHandler(reviewSvc).Register(protected)
	api.NewKnowledgeHandler(knowledgeSvc).Register(protected)
	api.NewSettingsHandler(settingsSvc).Register(protected)
	api.NewDashboardHandler(dashboardSvc).Register(protected)
	api.NewUserHandler(userSvc, authSvc).RegisterReadOnly(protected)
	api.NewReviewConfigHandler(reviewConfigSvc).RegisterReadOnly(protected)

	adminGroup := protected.Group("")
	adminGroup.Use(authHandler.AdminRequired())
	api.NewUserHandler(userSvc, authSvc).RegisterAdmin(adminGroup)
	api.NewReviewConfigHandler(reviewConfigSvc).RegisterAdmin(adminGroup)

	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	log.Printf("ReviewBuddy server listening on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// seedTemplates 首次启动写入一个示例模板，方便演示
func seedTemplates(s *template.Service) {
	existing, err := s.List("", "")
	if err != nil || len(existing) > 0 {
		return
	}
	_, _ = s.Create(&model.Template{
		LibraryID:   "default",
		Name:        "标准评审材料模板",
		Category:    "标准",
		Description: "适用于常见评审场景的标准材料模板",
		Variables:   []string{"材料名称", "评审范围", "负责人"},
		Content: `# {{材料名称}} 评审材料

## 一、背景与目标
- 评审范围：{{评审范围}}
- 负责人：{{负责人}}

## 二、核心内容
- [ ] 目标清晰
- [ ] 方案完整
- [ ] 依赖明确

## 三、风险与约束
1.

## 四、评审关注点
1.

## 五、结论与后续动作
- [ ] 结论明确
- [ ] 待办责任人清晰`,
	})
}
