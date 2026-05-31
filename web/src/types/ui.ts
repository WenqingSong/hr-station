// ─── 前端展示相关类型 ───────────────────────────────────────────────────────────

/** Badge 色调 */
export type BadgeTone = 'default' | 'blue' | 'green' | 'orange' | 'red' | 'gray';

/** 状态标签 */
export interface StatusBadge {
  label: string;
  tone: BadgeTone;
}

/** 页面操作按钮 */
export interface PageAction {
  label: string;
  variant: 'primary' | 'secondary' | 'ghost';
  disabled?: boolean;
}

/** AI 投递决策面板状态 */
export interface DecisionPanelState {
  current_resume_id?: string;
  current_job_id?: string;
  primary_action_label: string;
  helper_text: string;
}
