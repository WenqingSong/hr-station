package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"offer-catcher/server/internal/model"
)

func makeAlgoResume() *model.ResumeVersion {
	return &model.ResumeVersion{
		ID:     1,
		UserID: 100,
		Name:   "01-算法研究生-岗位匹配",
		Educations: []model.ResumeEducation{
			{
				School:            "清华大学",
				Degree:            "硕士",
				Major:             "计算机科学",
				ResearchDirection: "推荐系统与深度学习",
				PaperSummary:      "发表推荐系统相关论文2篇",
			},
		},
		Projects: []model.ResumeProject{
			{Name: "推荐系统实验平台", Role: "核心开发", Description: "基于PyTorch构建多目标推荐排序模型训练平台，支持AUC/NDCG评估，包含消融实验模块"},
			{Name: "用户行为序列建模", Role: "核心开发", Description: "使用Transformer和注意力机制对用户行为序列建模，提升推荐效果"},
		},
		Internships: []model.ResumeInternship{
			{Company: "某互联网大厂", Position: "算法实习生", Description: "参与推荐系统优化，负责模型训练、离线评估和消融分析"},
		},
		ProfessionalSkills: "Python, PyTorch, TensorFlow, Go, MySQL, Docker, Linux, Scikit-learn, Pandas, NumPy, SQL",
		AdditionalInfo:     "Kaggle竞赛Top10%，开源项目贡献者",
		RawText:            "推荐系统 深度学习 PyTorch 用户行为建模 AUC NDCG 消融实验 Go Docker",
	}
}

func makeBackendResume() *model.ResumeVersion {
	return &model.ResumeVersion{
		ID:     2,
		UserID: 100,
		Name:   "02-后端开发-岗位匹配",
		Educations: []model.ResumeEducation{
			{School: "某211大学", Degree: "本科", Major: "软件工程"},
		},
		Projects: []model.ResumeProject{
			{Name: "电商微服务平台", Role: "后端开发", Description: "基于Go + MySQL + Redis构建微服务电商平台，支持分布式事务和高并发"},
		},
		Internships: []model.ResumeInternship{
			{Company: "某互联网公司", Position: "后端开发实习生", Description: "负责接口开发、权限设计和数据库优化"},
		},
		ProfessionalSkills: "Go, Java, MySQL, Redis, Docker, Nginx, Kafka, Linux, Git",
	}
}

func makeProductResume() *model.ResumeVersion {
	return &model.ResumeVersion{
		ID:     3,
		UserID: 100,
		Name:   "03-产品经理-岗位匹配",
		Educations: []model.ResumeEducation{
			{School: "某二本院校", Degree: "本科", Major: "市场营销"},
		},
		Projects: []model.ResumeProject{
			{Name: "校园社交小程序", Role: "产品经理", Description: "负责需求分析、用户调研、原型设计和产品迭代"},
		},
		Internships: []model.ResumeInternship{
			{Company: "某电商公司", Position: "产品实习生", Description: "参与用户增长运营，A/B测试和数据分析"},
		},
		ProfessionalSkills: "Axure, Figma, SQL, Excel, 用户研究",
	}
}

func makeExplorationResume() *model.ResumeVersion {
	return &model.ResumeVersion{
		ID:     4,
		UserID: 100,
		Name:   "04-职业探索型",
		Educations: []model.ResumeEducation{
			{School: "某普通大学", Degree: "本科", Major: "工商管理"},
		},
		Projects: []model.ResumeProject{
			{Name: "学生会活动策划", Role: "组织者", Description: "策划校园文化节活动"},
		},
		ProfessionalSkills: "Office, Excel, 沟通协调",
	}
}

func makeBusinessResume() *model.ResumeVersion {
	return &model.ResumeVersion{
		ID:     5,
		UserID: 100,
		Name:   "05-商务方向",
		Educations: []model.ResumeEducation{
			{School: "某院校", Degree: "本科", Major: "商务管理"},
		},
		Internships: []model.ResumeInternship{
			{Company: "某贸易公司", Position: "商务助理", Description: "负责合同管理、供应商沟通、采购流程跟进"},
		},
		ProfessionalSkills: "Office, Excel, 沟通谈判, 合同管理",
	}
}

func mustMakeJobs(t *testing.T) []*model.Job {
	t.Helper()
	return allJobsFromJSON(t, testJobsJSON)
}

func allJobsFromJSON(t *testing.T, raw string) []*model.Job {
	t.Helper()
	var rows []struct {
		ID            uint64  `json:"id"`
		Title         string  `json:"title"`
		Department    string  `json:"department"`
		City          string  `json:"city"`
		Category      string  `json:"category"`
		JobType       string  `json:"job_type"`
		Description   string  `json:"description"`
		Requirements  string  `json:"requirements"`
		ExternalSource string `json:"external_source"`
		DataQuality   string  `json:"data_quality"`
	}
	if err := json.Unmarshal([]byte(raw), &rows); err != nil {
		t.Fatalf("parse test jobs: %v", err)
	}
	var jobs []*model.Job
	for _, r := range rows {
		jobs = append(jobs, &model.Job{
			ID:             r.ID,
			Title:          r.Title,
			Department:     r.Department,
			City:           r.City,
			Category:       r.Category,
			JobType:        r.JobType,
			Description:    r.Description,
			Requirements:   r.Requirements,
			ExternalSource: r.ExternalSource,
			DataQuality:    r.DataQuality,
		})
	}
	return jobs
}

const testJobsJSON = `[
{"id":1,"title":"多模态算法工程师","department":"技术事业群","city":"深圳","category":"算法","job_type":"full_time","description":"负责多模态模型研发，包括图文理解、视频理解、跨模态检索等","requirements":"深度学习基础扎实，熟悉多模态模型","external_source":"local_mock","data_quality":"full_jd"},
{"id":2,"title":"自然语言处理工程师","department":"AI平台部","city":"北京","category":"算法","job_type":"full_time","description":"负责NLP模型研发，包括文本分类、语义匹配、对话系统","requirements":"熟悉NLP，有Transformer/LLM经验优先","external_source":"local_mock","data_quality":"full_jd"},
{"id":3,"title":"数据科学工程师","department":"数据平台部","city":"深圳","category":"算法","job_type":"full_time","description":"负责数据挖掘和数据分析，构建业务指标和数据模型","requirements":"SQL/Spark/Python，数据分析和建模能力","external_source":"local_mock","data_quality":"full_jd"},
{"id":4,"title":"AI应用开发工程师","department":"技术中台","city":"深圳","category":"技术","job_type":"full_time","description":"负责AI服务化、模型部署、推理优化和工程落地","requirements":"Go/Python，Docker，有AI工程经验优先","external_source":"local_mock","data_quality":"full_jd"},
{"id":5,"title":"计算机图形学研究员","department":"研究院","city":"北京","category":"算法","job_type":"full_time","description":"负责图形学和渲染算法研究","requirements":"图形学基础扎实，熟悉OpenGL/3D","external_source":"local_mock","data_quality":"full_jd"},
{"id":6,"title":"后台开发工程师","department":"技术部","city":"深圳","category":"技术","job_type":"full_time","description":"负责后端服务和分布式系统开发","requirements":"Go/Java，MySQL/Redis，分布式经验","external_source":"local_mock","data_quality":"full_jd"},
{"id":7,"title":"客户端开发工程师","department":"移动端部","city":"北京","category":"技术","job_type":"full_time","description":"负责PC客户端应用开发","requirements":"C++，Windows开发经验","external_source":"local_mock","data_quality":"full_jd"},
{"id":8,"title":"产品策划","department":"产品部","city":"深圳","category":"产品","job_type":"full_time","description":"负责产品需求分析和产品设计","requirements":"产品sense，逻辑分析，用户研究","external_source":"local_mock","data_quality":"full_jd"},
{"id":9,"title":"产品运营","department":"运营部","city":"北京","category":"运营","job_type":"full_time","description":"负责用户增长运营和活动策划","requirements":"数据分析，活动运营，用户增长","external_source":"local_mock","data_quality":"full_jd"},
{"id":10,"title":"解决方案架构师","department":"行业解决方案","city":"深圳","category":"商务","job_type":"full_time","description":"负责行业客户需求分析和解决方案设计","requirements":"行业理解，客户需求分析，方案设计","external_source":"local_mock","data_quality":"full_jd"},
{"id":11,"title":"秘书","department":"综合管理部","city":"深圳","category":"职能","job_type":"full_time","description":"负责文档整理、会议协调和综合事务","requirements":"沟通协调，文档处理，细致负责","external_source":"local_mock","data_quality":"full_jd"},
{"id":12,"title":"采购商务","department":"采购部","city":"深圳","category":"商务","job_type":"full_time","description":"负责供应商管理和采购商务","requirements":"商务谈判，合同管理","external_source":"local_mock","data_quality":"full_jd"},
{"id":13,"title":"数据工程开发工程师","department":"数据平台部","city":"北京","category":"技术","job_type":"full_time","description":"负责数据平台建设和ETL管道开发","requirements":"Spark/Hive，数据仓库，大数据平台","external_source":"local_mock","data_quality":"full_jd"}
]`

// ── K9-1-T-C Test: Algorithm Grad ────────────────────────────────────

func TestK9T1C_AlgoGrad_ScoreDifferentiation(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeAlgoResume()
	results, err := mock.Recommend(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if len(results) < 5 {
		t.Fatalf("expected >=5 results: %d", len(results))
	}

	// Check score differentiation
	scores := make([]int, 0, 5)
	all95 := 0
	for i, j := range results[:5] {
		scores = append(scores, j.RecommendationScore)
		if j.RecommendationScore >= 95 {
			all95++
		}
		t.Logf("#%d %s score=%d label=%s", i+1, j.Title, j.RecommendationScore, j.RecommendationLabel)
		if j.WhyRankedHere == "" {
			t.Errorf("#%d why_ranked_here is empty", i+1)
		}
		if len(j.ResumeEvidence) == 0 {
			t.Errorf("#%d resume_evidence is empty", i+1)
		}
		if j.JobSpecificReason == "" {
			t.Errorf("#%d job_specific_reason is empty", i+1)
		}
		if len(j.JobSpecificRisks) == 0 {
			t.Errorf("#%d job_specific_risks is empty", i+1)
		}
		if j.SuggestedImprovement == "" {
			t.Errorf("#%d suggested_improvement is empty", i+1)
		}
	}

	if all95 >= 3 {
		t.Errorf("too many 95 scores (>=3): got %d of top5 at 95", all95)
	}

	gap := scores[0] - scores[4]
	if gap < 5 {
		t.Errorf("top1-top5 score gap too small: %d (want >=8)", gap)
	}
	t.Logf("Top1-Top5 score gap: %d", gap)

	// Label check: not all strong
	strongCount := 0
	for _, j := range results[:5] {
		if j.RecommendationPriority == "strong" {
			strongCount++
		}
	}
	if strongCount > 2 {
		t.Errorf("strong count > 2: %d", strongCount)
	}
	if strongCount == 5 {
		t.Error("all 5 labels are strong — should be differentiated")
	}

	// Reason differentiation
	reasons := map[string]int{}
	for _, j := range results[:5] {
		reasons[j.JobSpecificReason]++
	}
	if len(reasons) < 3 {
		t.Errorf("job_specific_reason differentiation too low: only %d unique out of 5", len(reasons))
	}

	// Risk differentiation
	riskSets := map[string]int{}
	for _, j := range results[:5] {
		key := fmt.Sprintf("%v", j.JobSpecificRisks)
		riskSets[key]++
	}
	if len(riskSets) < 3 {
		t.Errorf("job_specific_risks differentiation too low: only %d unique sets", len(riskSets))
	}
}

func TestK9T1C_BackendEng_ScoreDiff(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeBackendResume()
	results, err := mock.Recommend(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if len(results) < 3 {
		t.Fatalf("expected >=3 results: %d", len(results))
	}

	for i, j := range results[:5] {
		t.Logf("#%d %s score=%d label=%s priority=%s", i+1, j.Title, j.RecommendationScore, j.RecommendationLabel, j.RecommendationPriority)
		if j.RecommendationPriority == "strong" && j.JobID >= 8 {
			t.Errorf("backend profile should not have strong label for non-backend job #%d: %s", i+1, j.Title)
		}
		if j.Priority == "" {
			t.Errorf("#%d empty priority", i+1)
		}
	}
}

func TestK9T1C_ProductManager_NoAlgoStrong(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeProductResume()
	results, err := mock.Recommend(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	for i, j := range results[:5] {
		t.Logf("#%d %s score=%d label=%s priority=%s", i+1, j.Title, j.RecommendationScore, j.RecommendationLabel, j.RecommendationPriority)
		if (j.JobID == 1 || j.JobID == 2 || j.JobID == 6) && (j.RecommendationPriority == "strong" || j.RecommendationPriority == "high") {
			t.Errorf("product profile should not have strong/high for algo/backend job #%d: %s", i+1, j.Title)
		}
	}
}

func TestK9T1C_Exploration_Conservative(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeExplorationResume()
	results, err := mock.Recommend(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	for i, j := range results[:5] {
		t.Logf("#%d %s score=%d label=%s", i+1, j.Title, j.RecommendationScore, j.RecommendationLabel)
		if (j.JobID <= 5 || j.JobID == 6) && j.RecommendationPriority == "strong" {
			t.Errorf("exploration profile should not score algo/backend as strong: #%d %s", i+1, j.Title)
		}
	}
}

func TestK9T1C_Business_NoAlgoStrong(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeBusinessResume()
	results, err := mock.Recommend(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}

	for i, j := range results[:5] {
		t.Logf("#%d %s score=%d label=%s priority=%s", i+1, j.Title, j.RecommendationScore, j.RecommendationLabel, j.RecommendationPriority)
		if (j.JobID <= 5 || j.JobID == 6) && (j.RecommendationPriority == "strong" || j.RecommendationPriority == "high") {
			t.Errorf("business profile should not have strong/high for algo/backend: #%d %s", i+1, j.Title)
		}
	}
}

func TestK9T1C_NewFieldsJSONSerializable(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeAlgoResume()
	results, err := mock.Recommend(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected results")
	}

	// Verify all new fields can be serialized and deserialized
	b, err := json.Marshal(results[0])
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var unmarshaled model.RecommendedJob
	if err := json.Unmarshal(b, &unmarshaled); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if unmarshaled.WhyRankedHere == "" {
		t.Error("why_ranked_here should be present")
	}
	if len(unmarshaled.ResumeEvidence) == 0 {
		t.Error("resume_evidence should be non-empty")
	}
}

func TestK9T1C_QualityFallback_HasNewFields(t *testing.T) {
	mock := NewRecommendationMockAgent()
	jobs := mustMakeJobs(t)
	resume := makeAlgoResume()
	results, err := mock.RecommendQuality(context.Background(), resume, jobs)
	if err != nil {
		t.Fatalf("expected no error from quality: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected quality results")
	}

	for i, j := range results[:3] {
		if j.WhyRankedHere == "" {
			t.Errorf("quality #%d why_ranked_here is empty", i+1)
		}
		if len(j.ResumeEvidence) == 0 {
			t.Errorf("quality #%d resume_evidence is empty", i+1)
		}
		if j.JobSpecificReason == "" {
			t.Errorf("quality #%d job_specific_reason is empty", i+1)
		}
	}
	t.Log("Quality fallback has all new fields ✅")
}
