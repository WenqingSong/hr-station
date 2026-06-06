package model

import "time"

type Resume struct {
	ID            uint64
	UserID        uint64
	Name          string
	IsCurrent     bool
	RealName      string
	Phone         string
	Email         string
	City          string
	ProfessionalSkills string
	AIApplication      string
	Homepage           string
	AdditionalInfo     string
	RawText       string
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Educations     []ResumeEducation
	Projects       []ResumeProject
	Internships    []ResumeInternship
	LanguageExams  []ResumeLanguageExam
	Certificates   []ResumeCertificate
	Awards         []ResumeAward
	Works          []ResumeWork

	// Shim fields for old agent/service compat (K8-5-48-B)
	VersionNo      int
	Location       string
	SelfEvaluation string
	TargetRole     string
	TargetCity     string
	TargetDirection string
	JobType        string
	ExpectedCities string
	School         string
	Degree         string
	Major          string
	GraduationYear string
	SchemaVersion  int
	EducationJSON  []byte
	ExperienceJSON []byte
	ProjectJSON    []byte
	SkillsJSON     []byte
	ProfileJSON      []byte
	JobIntentionJSON []byte
	InternshipJSON   []byte
	CertificatesJSON []byte
	AwardsJSON       []byte
	PortfolioJSON    []byte
	ExtraJSON        []byte
}

type ResumeVersion = Resume

// K8-5-48-F: JSON tags added for relational deserialization
type ResumeEducation struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	School string `json:"school"`
	Degree string `json:"degree"`
	Major string `json:"major"`
	Department string `json:"department"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	GPA string `json:"gpa"`
	GPAScale string `json:"gpa_scale"`
	RankText string `json:"rank_text"`
	CoreCourses string `json:"core_courses"`
	ResearchDirection string `json:"research_direction"`
	Lab string `json:"lab"`
	PaperSummary string `json:"paper_summary"`
	SortOrder int `json:"sort_order,omitempty"`
}

type ResumeProject struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	Name string `json:"name"`
	Role string `json:"role"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	Description string `json:"description"`
	SortOrder int `json:"sort_order,omitempty"`
}

type ResumeInternship struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	Company string `json:"company"`
	Position string `json:"position"`
	StartDate string `json:"start_date"`
	EndDate string `json:"end_date"`
	Description string `json:"description"`
	SortOrder int `json:"sort_order,omitempty"`
}

type ResumeLanguageExam struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	ExamType string `json:"exam_type"`
	Score string `json:"score"`
	SortOrder int `json:"sort_order,omitempty"`
}

type ResumeCertificate struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	Name string `json:"name"`
	Description string `json:"description"`
	SortOrder int `json:"sort_order,omitempty"`
}

type ResumeAward struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	Name string `json:"name"`
	Description string `json:"description"`
	SortOrder int `json:"sort_order,omitempty"`
}

type ResumeWork struct {
	ID uint64 `json:"id,omitempty"`
	ResumeID uint64 `json:"resume_id,omitempty"`
	Name string `json:"name"`
	URL string `json:"url"`
	Password string `json:"password"`
	SortOrder int `json:"sort_order,omitempty"`
}
