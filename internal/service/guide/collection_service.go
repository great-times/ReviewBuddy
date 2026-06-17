package guide

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/textproto"
	"strings"
	"time"

	"github.com/google/uuid"

	"reviewbuddy/internal/model"
	"reviewbuddy/internal/repo"
)

type CollectionService struct {
	repo          *repo.ReviewCollectionRepo
	guides        *repo.GuideRepo
	reviews       *repo.ReviewRepo
	reviewConfigs *repo.ReviewConfigRepo
}

func NewCollectionService(r *repo.ReviewCollectionRepo, guides *repo.GuideRepo, reviews *repo.ReviewRepo, reviewConfigs *repo.ReviewConfigRepo) *CollectionService {
	return &CollectionService{repo: r, guides: guides, reviews: reviews, reviewConfigs: reviewConfigs}
}

func (s *CollectionService) List() ([]model.ReviewCollection, error) { return s.repo.List() }

func (s *CollectionService) Create(item *model.ReviewCollection) (*model.ReviewCollection, error) {
	if strings.TrimSpace(item.Title) == "" {
		return nil, errors.New("title is required")
	}
	if len(item.GuideIDs) == 0 {
		return nil, errors.New("guideIds is required")
	}
	now := time.Now().Format(time.RFC3339)
	item.ID = uuid.NewString()
	if item.DomainID == "" {
		item.DomainID = "default"
	}
	if item.Status == "" {
		item.Status = "pending"
	}
	item.CreatedAt = now
	item.UpdatedAt = now
	if err := s.repo.Create(item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *CollectionService) Update(id string, in *model.ReviewCollection) (*model.ReviewCollection, error) {
	cur, err := s.repo.Get(id)
	if err != nil || cur == nil {
		return cur, err
	}
	if strings.TrimSpace(in.Title) != "" {
		cur.Title = in.Title
	}
	if len(in.GuideIDs) > 0 {
		cur.GuideIDs = in.GuideIDs
	}
	if in.DomainID != "" {
		cur.DomainID = in.DomainID
	}
	if in.Status != "" {
		cur.Status = in.Status
	}
	cur.DecisionNote = in.DecisionNote
	cur.UpdatedAt = time.Now().Format(time.RFC3339)
	if err := s.repo.Update(cur); err != nil {
		return nil, err
	}
	return cur, nil
}

func (s *CollectionService) ExportEML(id string) (string, []byte, error) {
	item, err := s.repo.Get(id)
	if err != nil {
		return "", nil, err
	}
	if item == nil {
		return "", nil, errors.New("collection not found")
	}
	guides := []model.Guide{}
	for _, guideID := range item.GuideIDs {
		if g, err := s.guides.Get(guideID); err == nil && g != nil {
			guides = append(guides, *g)
		}
	}
	domain, err := s.reviewConfigs.GetDomain(item.DomainID)
	if err != nil {
		return "", nil, err
	}
	if domain == nil {
		domain = &model.ReviewDomain{ID: item.DomainID, Name: item.DomainID}
	}
	subjectTemplate := strings.TrimSpace(domain.MailSubjectTemplate)
	if subjectTemplate == "" {
		subjectTemplate = "评审纪要 - {{collectionTitle}}"
	}
	bodyTemplate := strings.TrimSpace(domain.MailBodyTemplate)
	if bodyTemplate == "" {
		bodyTemplate = defaultCollectionMailBodyTemplate()
	}
	vars := s.emailTemplateVars(item, domain, guides)
	subject := renderMailTemplate(subjectTemplate, vars, false)
	body := renderMailTemplate(bodyTemplate, vars, true)
	var buf bytes.Buffer
	headers := textproto.MIMEHeader{}
	headers.Set("Subject", mime.QEncoding.Encode("UTF-8", subject))
	headers.Set("MIME-Version", "1.0")
	headers.Set("Content-Type", `text/html; charset="UTF-8"`)
	headers.Set("Content-Transfer-Encoding", "quoted-printable")
	headers.Set("X-Unsent", "1")
	headers.Set("Date", time.Now().Format(time.RFC1123Z))
	for k, values := range headers {
		for _, v := range values {
			fmt.Fprintf(&buf, "%s: %s\r\n", k, v)
		}
	}
	buf.WriteString("\r\n")
	qp := quotedprintable.NewWriter(&buf)
	_, _ = qp.Write([]byte(body))
	_ = qp.Close()
	filename := strings.NewReplacer("/", "-", "\\", "-", ":", "-", " ", "_").Replace(item.Title) + ".eml"
	return filename, buf.Bytes(), nil
}

func (s *CollectionService) emailTemplateVars(item *model.ReviewCollection, domain *model.ReviewDomain, guides []model.Guide) map[string]string {
	decisionNote := item.DecisionNote
	if strings.TrimSpace(decisionNote) == "" {
		decisionNote = "暂无统一评审意见"
	}
	return map[string]string{
		"collectionTitle": item.Title,
		"domainName":      domain.Name,
		"domainId":        item.DomainID,
		"status":          statusLabel(item.Status),
		"statusCode":      item.Status,
		"decisionNote":    decisionNote,
		"materialsTable":  materialsTable(guides),
		"materialsList":   materialsList(guides),
		"createdBy":       item.CreatedBy,
		"createdAt":       item.CreatedAt,
		"updatedAt":       item.UpdatedAt,
	}
}

func defaultCollectionMailBodyTemplate() string {
	return `<html><body>
<h2>评审纪要：{{collectionTitle}}</h2>
<p><b>领域：</b>{{domainName}}</p>
<p><b>评审状态：</b>{{status}}</p>
<h3>统一评审意见</h3>
<p>{{decisionNote}}</p>
<h3>评审材料清单</h3>
{{materialsTable}}
<p>请各责任人根据评审意见完成后续动作。</p>
</body></html>`
}

func materialsTable(guides []model.Guide) string {
	var b strings.Builder
	b.WriteString("<table border=\"1\" cellspacing=\"0\" cellpadding=\"6\"><tr><th>标题</th><th>状态</th><th>风险</th><th>创建人</th></tr>")
	for _, g := range guides {
		b.WriteString("<tr><td>" + htmlEscape(g.Title) + "</td><td>" + htmlEscape(statusLabel(g.Status)) + "</td><td>" + htmlEscape(g.RiskLevel) + "</td><td>" + htmlEscape(g.CreatedBy) + "</td></tr>")
	}
	b.WriteString("</table>")
	return b.String()
}

func materialsList(guides []model.Guide) string {
	var b strings.Builder
	b.WriteString("<ul>")
	for _, g := range guides {
		b.WriteString("<li>" + htmlEscape(g.Title) + "（" + htmlEscape(statusLabel(g.Status)) + " / " + htmlEscape(g.RiskLevel) + "）</li>")
	}
	b.WriteString("</ul>")
	return b.String()
}

func renderMailTemplate(tpl string, vars map[string]string, html bool) string {
	out := tpl
	for key, value := range vars {
		if html && key != "materialsTable" && key != "materialsList" {
			value = htmlEscape(value)
		}
		out = strings.ReplaceAll(out, "{{"+key+"}}", value)
	}
	return out
}

func statusLabel(status string) string {
	switch status {
	case "approved":
		return "通过"
	case "rejected":
		return "驳回"
	case "follow_up":
		return "待跟进"
	case "pending":
		return "待确认"
	default:
		return status
	}
}

func htmlEscape(s string) string {
	replacer := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;", "\n", "<br/>")
	return replacer.Replace(s)
}
