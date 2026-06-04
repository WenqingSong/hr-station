package model

// ResumeSnapshotV2 is the unified snapshot representation of a resume
// consumed by all downstream Agents (Recommend, Match, Optimize, Compare).
// It normalizes V1/V2 ResumeVersion data into a single stable structure.
//
// Design frozen in K8-5-21-D.
type ResumeSnapshotV2 struct {
	Meta         ResumeSnapshotMeta          `json:"meta"`
	Profile      ResumeSnapshotProfile       `json:"profile"`
	JobIntention ResumeSnapshotJobIntention  `json:"job_intention"`
	Educations   []ResumeSnapshotEducation   `json:"educations"`
	Internships  []ResumeSnapshotInternship  `json:"internships"`
	Projects     []ResumeSnapshotProject     `json:"projects"`
	Skills       ResumeSnapshotSkills        `json:"skills"`
	Certificates []ResumeSnapshotCertificate `json:"certificates"`
	Awards       []ResumeSnapshotAward       `json:"awards"`
	Portfolio    ResumeSnapshotPortfolio     `json:"portfolio"`
	ExtraInfo    ResumeSnapshotExtraInfo     `json:"extra_info"`
}

// ResumeSnapshotMeta contains version metadata (not a page module).
type ResumeSnapshotMeta struct {
	SchemaVersion   int    `json:"schema_version"`
	ResumeVersionNo int    `json:"resume_version_no"`
	ResumeName      string `json:"resume_name"`
	TargetDirection string `json:"target_direction"`
}

// ResumeSnapshotProfile — module 1: basic information.
// phone / email are excluded from Snapshot per privacy policy.
type ResumeSnapshotProfile struct {
	Name           string `json:"name,omitempty"`
	School         string `json:"school,omitempty"`
	Degree         string `json:"degree,omitempty"`
	Major          string `json:"major,omitempty"`
	GraduationYear string `json:"graduation_year,omitempty"`
	CurrentCity    string `json:"current_city,omitempty"`
}

// ResumeSnapshotJobIntention — module 2: job intention.
type ResumeSnapshotJobIntention struct {
	TargetRoles        []string `json:"target_roles"`
	JobType            string   `json:"job_type,omitempty"`
	ExpectedCities     []string `json:"expected_cities"`
	AcceptOtherCities  bool     `json:"accept_other_cities,omitempty"`
	EarliestStartDate  string   `json:"earliest_start_date,omitempty"`
	InternshipDuration string   `json:"internship_duration,omitempty"`
	WeeklyAvailableDays int    `json:"weekly_available_days,omitempty"`
	TargetIndustries   []string `json:"target_industries,omitempty"`
	TargetCompanies    []string `json:"target_companies,omitempty"`
}

// ResumeSnapshotEducation — module 3: education entries.
type ResumeSnapshotEducation struct {
	School            string   `json:"school"`
	Degree            string   `json:"degree,omitempty"`
	Major             string   `json:"major,omitempty"`
	Department        string   `json:"department,omitempty"`
	StartDate         string   `json:"start_date,omitempty"`
	EndDate           string   `json:"end_date,omitempty"`
	GPA               string   `json:"gpa,omitempty"`
	GPAScale          string   `json:"gpa_scale,omitempty"`
	Rank              string   `json:"rank,omitempty"`
	CoreCourses       []string `json:"core_courses,omitempty"`
	ResearchDirection string   `json:"research_direction,omitempty"`
	Lab               string   `json:"lab,omitempty"`
	PaperSummary      string   `json:"paper_summary,omitempty"`
}

// ResumeSnapshotInternship — module 4: internship experience.
type ResumeSnapshotInternship struct {
	Company     string   `json:"company"`
	Role        string   `json:"role"`
	StartDate   string   `json:"start_date,omitempty"`
	EndDate     string   `json:"end_date,omitempty"`
	WorkContent string   `json:"work_content,omitempty"`
	TechStack   []string `json:"tech_stack,omitempty"`
	Achievements []string `json:"achievements,omitempty"`
	Metrics     string   `json:"metrics,omitempty"`
}

// ResumeSnapshotProject — module 5: project experience.
type ResumeSnapshotProject struct {
	Name            string   `json:"name"`
	Role            string   `json:"role,omitempty"`
	Type            string   `json:"type,omitempty"`
	StartDate       string   `json:"start_date,omitempty"`
	EndDate         string   `json:"end_date,omitempty"`
	TechStack       []string `json:"tech_stack,omitempty"`
	Description     string   `json:"description,omitempty"`
	Responsibilities string  `json:"responsibilities,omitempty"`
	Achievements    []string `json:"achievements,omitempty"`
	Link            string   `json:"link,omitempty"`
}

// ResumeSnapshotSkills — module 6: skill categories.
type ResumeSnapshotSkills struct {
	ProgrammingLanguages  []string `json:"programming_languages,omitempty"`
	BackendFrameworks     []string `json:"backend_frameworks,omitempty"`
	FrontendTechnologies  []string `json:"frontend_technologies,omitempty"`
	Databases             []string `json:"databases,omitempty"`
	Middleware            []string `json:"middleware,omitempty"`
	CloudPlatforms        []string `json:"cloud_platforms,omitempty"`
	DevTools              []string `json:"dev_tools,omitempty"`
	AITools               []string `json:"ai_tools,omitempty"`
	AIApplicationScenarios []string `json:"ai_application_scenarios,omitempty"`
	OtherKeywords         []string `json:"other_keywords,omitempty"`
}

// ResumeSnapshotCertificate — module 7: certificates & languages.
type ResumeSnapshotCertificate struct {
	Name     string `json:"name"`
	Type     string `json:"type,omitempty"`
	Level    string `json:"level,omitempty"`
	Score    string `json:"score,omitempty"`
	IssuedAt string `json:"issued_at,omitempty"`
}

// ResumeSnapshotAward — module 8: awards & competitions.
type ResumeSnapshotAward struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Level       string `json:"level,omitempty"`
	AwardedAt   string `json:"awarded_at,omitempty"`
	Description string `json:"description,omitempty"`
}

// ResumeSnapshotPortfolio — module 9: portfolio links.
type ResumeSnapshotPortfolio struct {
	Github         string   `json:"github,omitempty"`
	Gitee          string   `json:"gitee,omitempty"`
	CNB            string   `json:"cnb,omitempty"`
	Blog           string   `json:"blog,omitempty"`
	DemoLinks      []string `json:"demo_links,omitempty"`
	PortfolioLinks []string `json:"portfolio_links,omitempty"`
	Homepage       string   `json:"homepage,omitempty"`
}

// ResumeSnapshotExtraInfo — module 10: extra / supplementary info.
type ResumeSnapshotExtraInfo struct {
	SelfEvaluation string `json:"self_evaluation,omitempty"`
	AdditionalInfo string `json:"additional_info,omitempty"`
}
