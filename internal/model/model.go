package model

// 所有 JSON 字段使用 camelCase，与前端一致。

type Template struct {
	ID             string   `json:"id"`
	LibraryID      string   `json:"libraryId"`
	Name           string   `json:"name"`
	Category       string   `json:"category"`
	Description    string   `json:"description"`
	Content        string   `json:"content"`
	Variables      []string `json:"variables"`
	QualityScore   float64  `json:"qualityScore"`
	UsageCount     int      `json:"usageCount"`
	CurrentVersion int      `json:"currentVersion"`
	Status         string   `json:"status"`
	CreatedBy      string   `json:"createdBy"`
	CreatedAt      string   `json:"createdAt"`
	UpdatedAt      string   `json:"updatedAt"`
}

type TemplateLibrary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type TemplateVersion struct {
	ID         string `json:"id"`
	TemplateID string `json:"templateId"`
	Version    int    `json:"version"`
	Content    string `json:"content"`
	ChangeNote string `json:"changeNote"`
	CreatedBy  string `json:"createdBy"`
	CreatedAt  string `json:"createdAt"`
}

type Guide struct {
	ID             string            `json:"id"`
	Title          string            `json:"title"`
	TemplateID     string            `json:"templateId"`
	Content        string            `json:"content"`
	Variables      map[string]string `json:"variables"`
	Status         string            `json:"status"`
	RiskLevel      string            `json:"riskLevel"`
	CurrentVersion int               `json:"currentVersion"`
	CreatedBy      string            `json:"createdBy"`
	CreatedAt      string            `json:"createdAt"`
	UpdatedAt      string            `json:"updatedAt"`
}

type Review struct {
	ID           string `json:"id"`
	GuideID      string `json:"guideId"`
	GuideVersion int    `json:"guideVersion"`
	Reviewer     string `json:"reviewer"`
	Status       string `json:"status"`
	DecisionNote string `json:"decisionNote"`
	CreatedAt    string `json:"createdAt"`
	FinishedAt   string `json:"finishedAt"`
}

type ReviewComment struct {
	ID        string `json:"id"`
	ReviewID  string `json:"reviewId"`
	Anchor    string `json:"anchor"`
	Severity  string `json:"severity"`
	Category  string `json:"category"`
	Content   string `json:"content"`
	Resolved  bool   `json:"resolved"`
	CreatedAt string `json:"createdAt"`
}

type ReviewIssue struct {
	ID               string `json:"id"`
	SourceReviewID   string `json:"sourceReviewId"`
	Category         string `json:"category"`
	TriggerCondition string `json:"triggerCondition"`
	ProblemDesc      string `json:"problemDesc"`
	CorrectPractice  string `json:"correctPractice"`
	ChangeType       string `json:"changeType"`
	Frequency        int    `json:"frequency"`
	CreatedAt        string `json:"createdAt"`
}

type KnowledgeRule struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	RuleType   string `json:"ruleType"`
	Pattern    string `json:"pattern"`
	Suggestion string `json:"suggestion"`
	Enabled    bool   `json:"enabled"`
	HitCount   int    `json:"hitCount"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
}

type User struct {
	ID           string `json:"id"`
	Username     string `json:"username"`
	PasswordHash string `json:"-"`
	Role         string `json:"role"`
	Enabled      bool   `json:"enabled"`
	CreatedAt    string `json:"createdAt"`
	UpdatedAt    string `json:"updatedAt"`
}

type AuthToken struct {
	Token     string `json:"token"`
	UserID    string `json:"userId"`
	ExpiresAt string `json:"expiresAt"`
	CreatedAt string `json:"createdAt"`
}

type AgentSettings struct {
	Provider       string `json:"provider"`
	BaseURL        string `json:"baseUrl"`
	APIKey         string `json:"apiKey"`
	Model          string `json:"model"`
	EmbeddingModel string `json:"embeddingModel"`
	TimeoutSeconds int    `json:"timeoutSeconds"`
	SystemPrompt   string `json:"systemPrompt"`
}

type AgentType struct {
	Type        string `json:"type"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// AIPrecheckResult AI 预审结果
type AIPrecheckResult struct {
	Findings []PrecheckFinding `json:"findings"`
	Summary  string            `json:"summary"`
}

type PrecheckFinding struct {
	Severity   string `json:"severity"`
	Category   string `json:"category"`
	Excerpt    string `json:"excerpt"`
	Problem    string `json:"problem"`
	Suggestion string `json:"suggestion"`
}
