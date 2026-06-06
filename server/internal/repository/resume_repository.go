package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"offer-catcher/server/internal/model"
)

type ResumeRepository struct{ db *sql.DB }

func NewResumeRepository(db *sql.DB) *ResumeRepository { return &ResumeRepository{db: db} }

func (r *ResumeRepository) ListByUserID(ctx context.Context, userID uint64) ([]*model.Resume, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id FROM resumes WHERE user_id=? ORDER BY updated_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list resumes: %w", err)
	}
	defer rows.Close()
	var ids []uint64
	for rows.Next() {
		var id uint64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	result := make([]*model.Resume, 0, len(ids))
	for _, id := range ids {
		resume, err := r.GetByID(ctx, userID, id)
		if err != nil {
			return nil, err
		}
		if resume != nil {
			result = append(result, resume)
		}
	}
	return result, nil
}

func (r *ResumeRepository) GetCurrentByUserID(ctx context.Context, userID uint64) (*model.Resume, error) {
	var id uint64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM resumes WHERE user_id=? AND is_current=1 ORDER BY updated_at DESC LIMIT 1`, userID).Scan(&id)
	if err == sql.ErrNoRows {
		err = r.db.QueryRowContext(ctx, `SELECT id FROM resumes WHERE user_id=? ORDER BY updated_at DESC LIMIT 1`, userID).Scan(&id)
	}
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get current resume: %w", err)
	}
	return r.GetByID(ctx, userID, id)
}

func (r *ResumeRepository) GetByID(ctx context.Context, userID, id uint64) (*model.Resume, error) {
	resume := &model.Resume{}
	var rawText, profSkills, aiApp, addInfo sql.NullString
	err := r.db.QueryRowContext(ctx,
		`SELECT id,user_id,name,is_current,real_name,phone,email,city,
		        professional_skills,ai_application,homepage,additional_info,raw_text,created_at,updated_at
		 FROM resumes WHERE id=? AND user_id=?`, id, userID,
	).Scan(&resume.ID, &resume.UserID, &resume.Name, &resume.IsCurrent,
		&resume.RealName, &resume.Phone, &resume.Email, &resume.City,
		&profSkills, &aiApp, &resume.Homepage, &addInfo, &rawText,
		&resume.CreatedAt, &resume.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get resume: %w", err)
	}
	resume.ProfessionalSkills = profSkills.String
	resume.AIApplication = aiApp.String
	resume.AdditionalInfo = addInfo.String
	resume.RawText = rawText.String

	resume.Educations, _ = r.getEducations(ctx, id)
	resume.Projects, _ = r.getProjects(ctx, id)
	resume.Internships, _ = r.getInternships(ctx, id)
	resume.LanguageExams, _ = r.getLanguageExams(ctx, id)
	resume.Certificates, _ = r.getCertificates(ctx, id)
	resume.Awards, _ = r.getAwards(ctx, id)
	resume.Works, _ = r.getWorks(ctx, id)
	return resume, nil
}

func (r *ResumeRepository) Create(ctx context.Context, resume *model.Resume) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var count int
	tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM resumes WHERE user_id=?", resume.UserID).Scan(&count)
	if count == 0 {
		resume.IsCurrent = true
	}
	resume.RawText = buildRawText(resume)

	result, err := tx.ExecContext(ctx,
		`INSERT INTO resumes (user_id,name,is_current,real_name,phone,email,city,
		 professional_skills,ai_application,homepage,additional_info,raw_text)
		 VALUES (?,?,?,?,?,?,?,?,?,?,?,?)`,
		resume.UserID, resume.Name, resume.IsCurrent,
		resume.RealName, resume.Phone, resume.Email, resume.City,
		toNull(resume.ProfessionalSkills), toNull(resume.AIApplication),
		resume.Homepage, toNull(resume.AdditionalInfo), resume.RawText)
	if err != nil {
		return fmt.Errorf("insert resume: %w", err)
	}
	id, _ := result.LastInsertId()
	resume.ID = uint64(id)

	if err := r.insertChildren(tx, resume); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ResumeRepository) Update(ctx context.Context, resume *model.Resume) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	tx.QueryRowContext(ctx, "SELECT 1 FROM resumes WHERE id=? AND user_id=?", resume.ID, resume.UserID).Scan(&exists)
	if !exists {
		return fmt.Errorf("resume not found")
	}

	resume.RawText = buildRawText(resume)
	_, err = tx.ExecContext(ctx,
		`UPDATE resumes SET name=?,is_current=?,real_name=?,phone=?,email=?,city=?,
		 professional_skills=?,ai_application=?,homepage=?,additional_info=?,raw_text=?
		 WHERE id=? AND user_id=?`,
		resume.Name, resume.IsCurrent, resume.RealName, resume.Phone, resume.Email, resume.City,
		toNull(resume.ProfessionalSkills), toNull(resume.AIApplication),
		resume.Homepage, toNull(resume.AdditionalInfo), resume.RawText,
		resume.ID, resume.UserID)
	if err != nil {
		return fmt.Errorf("update resume: %w", err)
	}

	// Delete and re-insert children
	tx.ExecContext(ctx, "DELETE FROM resume_educations WHERE resume_id=?", resume.ID)
	tx.ExecContext(ctx, "DELETE FROM resume_projects WHERE resume_id=?", resume.ID)
	tx.ExecContext(ctx, "DELETE FROM resume_internships WHERE resume_id=?", resume.ID)
	tx.ExecContext(ctx, "DELETE FROM resume_language_exams WHERE resume_id=?", resume.ID)
	tx.ExecContext(ctx, "DELETE FROM resume_certificates WHERE resume_id=?", resume.ID)
	tx.ExecContext(ctx, "DELETE FROM resume_awards WHERE resume_id=?", resume.ID)
	tx.ExecContext(ctx, "DELETE FROM resume_works WHERE resume_id=?", resume.ID)

	if err := r.insertChildren(tx, resume); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *ResumeRepository) Delete(ctx context.Context, userID, id uint64) (wasCurrent bool, err error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var isCur bool
	err = tx.QueryRowContext(ctx, "SELECT is_current FROM resumes WHERE id=? AND user_id=?", id, userID).Scan(&isCur)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check resume: %w", err)
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM resumes WHERE id=? AND user_id=?", id, userID)
	if err != nil {
		return false, fmt.Errorf("delete resume: %w", err)
	}

	if isCur {
		var latestID uint64
		err := tx.QueryRowContext(ctx, "SELECT id FROM resumes WHERE user_id=? ORDER BY updated_at DESC LIMIT 1", userID).Scan(&latestID)
		if err == nil {
			tx.ExecContext(ctx, "UPDATE resumes SET is_current=1 WHERE id=?", latestID)
		}
	}
	return isCur, tx.Commit()
}

func (r *ResumeRepository) SetCurrent(ctx context.Context, userID, id uint64) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	tx.QueryRowContext(ctx, "SELECT 1 FROM resumes WHERE id=? AND user_id=?", id, userID).Scan(&exists)
	if !exists {
		return fmt.Errorf("resume not found")
	}
	tx.ExecContext(ctx, "UPDATE resumes SET is_current=0 WHERE user_id=?", userID)
	tx.ExecContext(ctx, "UPDATE resumes SET is_current=1 WHERE user_id=? AND id=?", userID, id)
	return tx.Commit()
}

func (r *ResumeRepository) CountByUserID(ctx context.Context, userID uint64) (int, error) {
	var c int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM resumes WHERE user_id=?", userID).Scan(&c)
	return c, err
}

// ── Children ──

func (r *ResumeRepository) insertChildren(tx *sql.Tx, resume *model.Resume) error {
	for i, e := range resume.Educations {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_educations(resume_id,school,degree,major,department,start_date,end_date,gpa,gpa_scale,rank_text,core_courses,research_direction,lab,paper_summary,sort_order)
			 VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			resume.ID, e.School, e.Degree, e.Major, e.Department, e.StartDate, e.EndDate,
			e.GPA, e.GPAScale, e.RankText, toNull(e.CoreCourses), e.ResearchDirection, e.Lab, toNull(e.PaperSummary), i)
		if err != nil {
			return fmt.Errorf("insert education: %w", err)
		}
	}
	for i, p := range resume.Projects {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_projects(resume_id,name,role,start_date,end_date,description,sort_order) VALUES(?,?,?,?,?,?,?)`,
			resume.ID, p.Name, p.Role, p.StartDate, p.EndDate, toNull(p.Description), i)
		if err != nil {
			return fmt.Errorf("insert project: %w", err)
		}
	}
	for i, v := range resume.Internships {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_internships(resume_id,company,position,start_date,end_date,description,sort_order) VALUES(?,?,?,?,?,?,?)`,
			resume.ID, v.Company, v.Position, v.StartDate, v.EndDate, toNull(v.Description), i)
		if err != nil {
			return fmt.Errorf("insert internship: %w", err)
		}
	}
	for i, v := range resume.LanguageExams {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_language_exams(resume_id,exam_type,score,sort_order) VALUES(?,?,?,?)`,
			resume.ID, v.ExamType, v.Score, i)
		if err != nil {
			return fmt.Errorf("insert language_exam: %w", err)
		}
	}
	for i, v := range resume.Certificates {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_certificates(resume_id,name,description,sort_order) VALUES(?,?,?,?)`,
			resume.ID, v.Name, toNull(v.Description), i)
		if err != nil {
			return fmt.Errorf("insert certificate: %w", err)
		}
	}
	for i, v := range resume.Awards {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_awards(resume_id,name,description,sort_order) VALUES(?,?,?,?)`,
			resume.ID, v.Name, toNull(v.Description), i)
		if err != nil {
			return fmt.Errorf("insert award: %w", err)
		}
	}
	for i, v := range resume.Works {
		_, err := tx.ExecContext(context.TODO(),
			`INSERT INTO resume_works(resume_id,name,url,password,sort_order) VALUES(?,?,?,?,?)`,
			resume.ID, v.Name, v.URL, v.Password, i)
		if err != nil {
			return fmt.Errorf("insert work: %w", err)
		}
	}
	return nil
}

func (r *ResumeRepository) getEducations(ctx context.Context, resumeID uint64) ([]model.ResumeEducation, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id,resume_id,school,degree,major,department,start_date,end_date,gpa,gpa_scale,rank_text,COALESCE(core_courses,''),research_direction,lab,COALESCE(paper_summary,''),sort_order FROM resume_educations WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []model.ResumeEducation
	for rows.Next() {
		var e model.ResumeEducation
		if err := rows.Scan(&e.ID, &e.ResumeID, &e.School, &e.Degree, &e.Major, &e.Department, &e.StartDate, &e.EndDate, &e.GPA, &e.GPAScale, &e.RankText, &e.CoreCourses, &e.ResearchDirection, &e.Lab, &e.PaperSummary, &e.SortOrder); err != nil {
			return nil, err
		}
		result = append(result, e)
	}
	return result, rows.Err()
}

func (r *ResumeRepository) getProjects(ctx context.Context, resumeID uint64) ([]model.ResumeProject, error) {
	rows, _ := r.db.QueryContext(ctx, `SELECT id,resume_id,name,role,start_date,end_date,COALESCE(description,''),sort_order FROM resume_projects WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	var result []model.ResumeProject
	for rows.Next() {
		var p model.ResumeProject
		rows.Scan(&p.ID, &p.ResumeID, &p.Name, &p.Role, &p.StartDate, &p.EndDate, &p.Description, &p.SortOrder)
		result = append(result, p)
	}
	return result, nil
}

func (r *ResumeRepository) getInternships(ctx context.Context, resumeID uint64) ([]model.ResumeInternship, error) {
	rows, _ := r.db.QueryContext(ctx, `SELECT id,resume_id,company,position,start_date,end_date,COALESCE(description,''),sort_order FROM resume_internships WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	var result []model.ResumeInternship
	for rows.Next() {
		var v model.ResumeInternship
		rows.Scan(&v.ID, &v.ResumeID, &v.Company, &v.Position, &v.StartDate, &v.EndDate, &v.Description, &v.SortOrder)
		result = append(result, v)
	}
	return result, nil
}

func (r *ResumeRepository) getLanguageExams(ctx context.Context, resumeID uint64) ([]model.ResumeLanguageExam, error) {
	rows, _ := r.db.QueryContext(ctx, `SELECT id,resume_id,exam_type,score,sort_order FROM resume_language_exams WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	var result []model.ResumeLanguageExam
	for rows.Next() {
		var v model.ResumeLanguageExam
		rows.Scan(&v.ID, &v.ResumeID, &v.ExamType, &v.Score, &v.SortOrder)
		result = append(result, v)
	}
	return result, nil
}

func (r *ResumeRepository) getCertificates(ctx context.Context, resumeID uint64) ([]model.ResumeCertificate, error) {
	rows, _ := r.db.QueryContext(ctx, `SELECT id,resume_id,name,COALESCE(description,''),sort_order FROM resume_certificates WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	var result []model.ResumeCertificate
	for rows.Next() {
		var v model.ResumeCertificate
		rows.Scan(&v.ID, &v.ResumeID, &v.Name, &v.Description, &v.SortOrder)
		result = append(result, v)
	}
	return result, nil
}

func (r *ResumeRepository) getAwards(ctx context.Context, resumeID uint64) ([]model.ResumeAward, error) {
	rows, _ := r.db.QueryContext(ctx, `SELECT id,resume_id,name,COALESCE(description,''),sort_order FROM resume_awards WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	var result []model.ResumeAward
	for rows.Next() {
		var v model.ResumeAward
		rows.Scan(&v.ID, &v.ResumeID, &v.Name, &v.Description, &v.SortOrder)
		result = append(result, v)
	}
	return result, nil
}

func (r *ResumeRepository) getWorks(ctx context.Context, resumeID uint64) ([]model.ResumeWork, error) {
	rows, _ := r.db.QueryContext(ctx, `SELECT id,resume_id,name,url,password,sort_order FROM resume_works WHERE resume_id=? ORDER BY sort_order,id`, resumeID)
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()
	var result []model.ResumeWork
	for rows.Next() {
		var v model.ResumeWork
		rows.Scan(&v.ID, &v.ResumeID, &v.Name, &v.URL, &v.Password, &v.SortOrder)
		result = append(result, v)
	}
	return result, nil
}

func toNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// buildRawText generates AI text snapshot from main table + children.
func buildRawText(r *model.Resume) string {
	var b strings.Builder
	fmt.Fprintf(&b, "【简历名称】%s\n", r.Name)
	fmt.Fprintf(&b, "【基础信息】姓名：%s｜手机：%s｜邮箱：%s｜城市：%s\n", r.RealName, r.Phone, r.Email, r.City)
	if r.ProfessionalSkills != "" {
		fmt.Fprintf(&b, "【专业技能】%s\n", r.ProfessionalSkills)
	}
	if r.AIApplication != "" {
		fmt.Fprintf(&b, "【AI应用技能】%s\n", r.AIApplication)
	}
	for _, e := range r.Educations {
		fmt.Fprintf(&b, "【教育经历】%s｜%s｜%s｜%s｜%s-%s", e.School, e.Degree, e.Major, e.Department, e.StartDate, e.EndDate)
		if e.GPA != "" {
			fmt.Fprintf(&b, "｜GPA %s/%s", e.GPA, e.GPAScale)
		}
		if e.CoreCourses != "" {
			fmt.Fprintf(&b, "｜主修：%s", e.CoreCourses)
		}
		if e.ResearchDirection != "" {
			fmt.Fprintf(&b, "｜研究方向：%s", e.ResearchDirection)
		}
		if e.PaperSummary != "" {
			fmt.Fprintf(&b, "｜论文：%s", e.PaperSummary)
		}
		b.WriteString("\n")
	}
	for _, p := range r.Projects {
		fmt.Fprintf(&b, "【项目经历】%s｜角色：%s｜%s-%s｜%s\n", p.Name, p.Role, p.StartDate, p.EndDate, p.Description)
	}
	for _, v := range r.Internships {
		fmt.Fprintf(&b, "【实习经历】%s｜%s｜%s-%s｜%s\n", v.Company, v.Position, v.StartDate, v.EndDate, v.Description)
	}
	for _, v := range r.LanguageExams {
		fmt.Fprintf(&b, "【外语考试】%s：%s\n", v.ExamType, v.Score)
	}
	for _, v := range r.Certificates {
		fmt.Fprintf(&b, "【证书】%s｜%s\n", v.Name, v.Description)
	}
	for _, v := range r.Awards {
		fmt.Fprintf(&b, "【获奖】%s｜%s\n", v.Name, v.Description)
	}
	for _, v := range r.Works {
		fmt.Fprintf(&b, "【作品】%s｜%s", v.Name, v.URL)
		if v.Password != "" {
			fmt.Fprintf(&b, "｜密码：%s", v.Password)
		}
		b.WriteString("\n")
	}
	if r.Homepage != "" {
		fmt.Fprintf(&b, "【个人主页】%s\n", r.Homepage)
	}
	if r.AdditionalInfo != "" {
		fmt.Fprintf(&b, "【补充信息】%s\n", r.AdditionalInfo)
	}
	return b.String()
}

// ExistsByName checks if another resume with the same name exists for the user.
// excludeID can be 0 to exclude nothing (for creation).
func (r *ResumeRepository) ExistsByName(ctx context.Context, userID uint64, name string, excludeID uint64) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		"SELECT COUNT(1) FROM resumes WHERE user_id=? AND name=? AND (?=0 OR id<>?)",
		userID, name, excludeID, excludeID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check name exists: %w", err)
	}
	return count > 0, nil
}

// ── Compat wrappers for old service code (K8-5-48-B shim) ──

func (r *ResumeRepository) FindByID(ctx context.Context, id uint64) (*model.ResumeVersion, error) {
	res, err := r.GetByID(ctx, id, id)
	if res == nil || err != nil { return nil, err }
	return &model.ResumeVersion{ID: res.ID, UserID: res.UserID, Name: res.Name}, nil
}
func (r *ResumeRepository) FindByUserIDAndID(ctx context.Context, userID, id uint64) (*model.ResumeVersion, error) {
	res, err := r.GetByID(ctx, userID, id)
	if res == nil || err != nil { return nil, err }
	return &model.ResumeVersion{ID: res.ID, UserID: res.UserID, Name: res.Name}, nil
}
func (r *ResumeRepository) CreateVersion(ctx context.Context, rv *model.ResumeVersion) error {
rr := &model.Resume{UserID: rv.UserID, Name: rv.Name, RealName: rv.Name}
	if err := r.Create(ctx, rr); err != nil { return err }
	rv.ID = uint64(uint64(rr.ID))
	return nil
}
