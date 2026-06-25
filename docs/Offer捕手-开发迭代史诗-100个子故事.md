# Offer捕手 —— 开发迭代史诗（100个子故事）

> 时间跨度：2026年5月28日 → 2026年6月24日  
> 基于 `docs/变更记录.md`（14,460行）全链路分析  
> 这是一个全栈工程师在不到一个月内，将一份空白项目构想落地为生产就绪产品的完整叙事。

---

## 序幕：从零到一的决心

### 故事1：第一行代码之前
5月28日深夜，项目还是一片空白。没有数据库、没有后端、没有前端、没有LLM接入——只有一份写在脑海里的构想：做一个让求职者可以上传简历、获得AI岗位推荐和匹配分析的平台。这将是接下来近一个月疯狂开发的起点。

### 故事2：六张表的诞生
5月29日，第一件要做的事是定义数据世界。你在 `init.sql` 中手写了6张核心表——users、resume_versions、jobs、recommendation_results、match_analyses、version_compare_results。每张表的外键、索引、JSON字段都经过深思熟虑。这不是简单的建表，而是对整个系统数据关系的第一次完整描摹。

### 故事3：写死密码的种子用户
在写种子数据时，你创建了两个Demo用户，密码直接用明文占位。你知道这不安全——但你也知道，3-3阶段才会接入bcrypt。这是一个务实的决策：先让系统跑起来，安全留给下一个迭代。

### 故事4：后端六层架构的蓝图
你在 `阶段3核心业务后端Mock方案.md` 中画出了完整的后端架构蓝图：Model → Repository → Service → Handler → Router → Middleware，六层分明。这不是教科书式的八股文，而是你基于"先Mock后LLM"的渐进式策略，为整个后端系统划定的清晰边界。

### 故事5：用户注册——密码终于安全了
5月29日，你写完了认证系统。JWT签发、bcrypt密码哈希、AuthMiddleware——三件套齐备。种子用户的密码从明文变成了真正的bcrypt hash。你导入 `golang-jwt/jwt/v5` 和 `x/crypto/bcrypt` 两个依赖，用 curl 验证了6种场景：注册、登录、me、无token、错误密码、healthz。全部通过。

### 故事6：简历版本管理——每次保存都是一次历史的定格
你设计了 `resume_versions` 表的版本管理策略：每次POST都是新建版本，历史版本只读，不做更新和删除。用户只有一份主简历，通过版本号追踪演进。这是一个极简但足够优雅的设计——没有resumes主表，没有多简历概念，一切都围绕"版本迭代"展开。

### 故事7：Mock Agent——不用LLM也能打
在接入真实LLM之前，你手写了4个Mock Agent：岗位推荐、岗位匹配、简历优化、版本对比。每个Mock Agent内部都有规则打分引擎——推荐靠城市匹配+方向匹配+技能交集，匹配靠简单规则评分+证据链生成。你甚至为每个Agent写好了逐条证据链和P0/P1优化建议的模板。这些Mock不仅让系统可以独立运行，更成为后续LLM fallback的安全线。

### 故事8：8个种子岗位——手工打磨的数据
你在 `init.sql` 中插入了8个手工编写的种子岗位：后台开发、前端开发、算法工程师、产品经理……每个岗位都有完整的description、requirements、skill_keywords。这些数据不是随便编的——你仔细研究了腾讯招聘的信息结构，确保每个字段都有真实感。

### 故事9：13个接口一夜之间出现
到阶段3收官时，你的后端已经有了13个接口：3个认证接口、5个简历接口、2个岗位接口、1个推荐接口、1个匹配分析接口、1个版本对比接口。每个接口都有完整的Handler→Service→Repository→Model链路。用curl逐个验证，全部返回200。

---

## 第一幕：LLM接入——让AI真正进入系统

### 故事10：Hunyuan API的第一次握手
5月29日，你在 `.env.example` 中新增了14项LLM配置。然后开始写LLM Client抽象层——`llm.Client` 接口、`ChatRequest`/`ChatResponse` 类型、`NoopClient` 测试实现。你甚至为LLM错误定义了9种错误码：从 `LLM_DISABLED` 到 `LLM_API_KEY_MISSING` 到 `LLM_TIMEOUT`，每种都有精准的语义。

### 故事11：JSON Schema——给AI画格子
你为3个Agent各写了一份JSON Schema：match_analysis_schema.json（17个required字段）、recommendation_schema.json（7个required字段）、resume_optimization_schema.json（15个required字段）。这些Schema不是简单的字段定义——每个字段都有type、description、enum约束。你用 `santhosh-tekuri/jsonschema/v5` 做校验，支持draft 2020-12标准。

### 故事12：Prompt Engineering——四份Agent指令书的诞生
你手写了4个Prompt模板：match_analysis_v1.md（24个章节）、recommendation_v1.md（24个章节）、resume_optimization_v1.md（18个章节）、json_repair_v1.md（13个章节）。每个模板都包含角色定义、输入格式、输出结构、禁止行为、边界条件。这不是简单的"请帮我分析"，而是工程化的Agent指令系统。

### 故事13：Pipeline——统一的LLM调度引擎
你设计了一个统一的Agent Pipeline：Prompt Render → LLM Call → JSON Parse → Schema Validate → Repair Once → Fallback Decision。这个Pipeline通过Runner接口注入，让所有Agent共享同一套LLM调用链路。Reapir最多执行一次，不递归，不重试——失败后直接进入Mock fallback。

### 故事14：JSON Repair Agent——给AI改错的机会
你单独实现了一个JSONRepairAgent，专门处理LLM返回的畸形JSON。它的任务不是重新做业务分析，而是修JSON结构——补齐缺失的括号、修正字段名拼写、删除多余的逗号。修一次，修不好就直接fallback。这是一个务实的中间层。

### 故事15：安全Key管理——不打印、不泄漏、不追踪
在整个LLM接入过程中，你始终遵循一条铁律：`LLM_API_KEY` 不得写入docs、不得出现在日志中、不得打印真实值。你实现了 `safeErr` 和 `redactSensitive` 两个函数，确保即使发生panic，Key也不会被泄露。`.env` 被 `.gitignore` 排除，`.env.example` 中只有占位符。

### 故事16：真实LLM的第一次调用失败
5月30日，你执行了第一次真实LLM Smoke Test。结果：`LLM_ENABLE=false`，`.env` 不存在，LLM配置不完整。三次核验全部未通过。你没有慌乱——这是预期内的情况。真正的LLM Key需要用户自己配置。

### 故事17：Prompt首轮验收——AI输出了什么？
5月29日，你用真实LLM手动验证了第一份Prompt输出。匹配分析Agent的返回值：JSON合法、字段完整、rewrite_example做得不错——但证据链缺少"missing"判断，部分改写示例超出了简历事实。结论：partial_pass，进入下一轮迭代。

### 故事18：HunyuanClient——Go语言实现的OpenAI兼容客户端
你手写了 `hunyuan_client.go`，实现了OpenAI-compatible的chat completions请求封装。支持BaseURL规范化（兼容 `/v1`、`/v1/`、`/v1/chat/completions`），错误码映射（429→rate limit、5xx→HTTP error、401/403→API key missing），用httptest写了22项单元测试，全部通过。

### 故事19：战略转向——从"推荐工具"到"投递决策系统"
5月30日，你做了一个关键的战略调整。项目的定位从"岗位推荐工具"升级为"投递前决策与简历初筛优化系统"。三个核心杀手锏被确认：投递边界与优化停止条件、JD证据链匹配、Agent输出Quality Gate。这意味着系统将不仅告诉用户"投哪里"，更要告诉用户"还差多远，怎么改"。

### 故事20：Quality Gate——AI说错了谁来兜底
你设计了Agent输出的Quality Gate机制：10项检查维度——jd_grounding_check（JD有依据吗？）、resume_evidence_check（简历有证据吗？）、fabrication_check（编造了吗？）、safety_tone_check（语气安全吗？）……每项都有pass/revise_required/fallback_required/manual_review_suggested四级判定。revise只给一次机会，revise后仍不通过直接fallback。

### 故事21：极端情况处理——当学生简历一片空白
你定义了10种极端学生求职场景：信息不足型、弱简历型、方向不匹配型、海归地点风险型、高潜科研型、方向未定型、真实性存疑型、过度优化风险型……每种场景都有触发条件、推荐输出、禁止表达、与apply_readiness的联动规则。关键底线是：判断"简历×JD"的匹配度，不判断"这个学生行不行"。

---

## 第二幕：前后端联调——数据流动起来

### 故事22：Makefile的一键启动革命
5月29日，你重构了整个开发工具链。从19个零散的Makefile命令升级到40+个标准化命令，所有复杂shell逻辑下沉到 `scripts/dev/` 下的11个脚本中。`make up` 一键启动完整环境：MySQL → Redis → 后端 → 前端。危险操作（rebuild/db-reset/clean-volume）必须 `CONFIRM=yes`。你甚至写了一个 `doctor.sh` 环境检查脚本。

### 故事23：腾讯招聘数据的秘密接口
5月29日，你发现并确认了腾讯招聘的公开接口。通过分析前端JS bundle，找到了 `getJobDetailsByPostId` 的路径。然后用curl验证了列表接口、详情接口、字典接口——三个必要接口全部可用。这为后续46个腾讯岗位的导入铺平了道路。

### 故事24：46个真岗位入库
你写了一个Go程序 `generate_jobs_seed/main.go`，从 `tencent-campus-jobs.json` 生成524KB的SQL文件。46个真实腾讯校招岗位、180个城市、194个事业群、1402个技能关键词、6个筛选分组、1131个选项——全部入库。你用 `seed_versions` 表实现了幂等导入，不用担心重复执行。

### 故事25：前后端第一次握手
6月1日，你执行了前后端API联调。`GET /api/jobs` 返回200，46个岗位完整；`GET /api/jobs/1` 返回完整详情。npm build 1.63s通过。这是前后端通路第一次被验证——从浏览器发出的请求，经过Nginx反向代理，到达Go后端，查询MySQL，再原路返回。

### 故事26：从Mock数据到真实API的切换
6月1日，你开始大规模清理前端Mock数据。6个核心组件一一从硬编码数据切换到API ViewModel：JobList、JobDetailBanner、JobDetailJD、AIDecisionWorkspace、JobDetailAIDecision、SelectResumeModal。grep验证：普通业务页面Mock引入从8处降为0。

### 故事27：Agent页面上下文真实化
你发现了一个严重问题：3个Agent页面（推荐、匹配、优化）虽然Agent API走的是真实后端，但页面上下文使用的全是Mock简历和Mock岗位。你逐个重写了3个页面——推荐页从Mock切真实resumeId传递、匹配页接收jobId+versionNo双参数、优化页从Mock resume文本切真实API。grep验证：0 mock import、0 hardcoded IDs。

### 故事28：简历的真实保存
MyResumesPage的保存按钮原来只是一个 `setSavedHint(true)`——不调用任何API，提示文案是"已保存本次演示修改"。你重写了整个保存流程：add模式调 `createResume()`、edit模式调 `updateResume()`、保存后刷新current和versions。"演示"这个词从代码中彻底消失了。

### 故事29：401的破局——Token过期不是世界末日
前端之前对401错误的处理方式是：`withFallback` 静默吞掉，用户永远不知道自己未登录。你建立了完整的Token管理链路：`authStorage.ts` 统一管理localStorage → `client.ts` 自动注入Authorization header → 401时清token + dispatchEvent全局通知 → `viewModels.ts` 10处catch统一跳转登录页。

### 故事30：RequireAuth——路由级别的登录保护
你新增了 `RequireAuth` 组件，保护了7个私有页面（简历4页+AI 3页）。公开页面（岗位列表、详情、登录、注册）不受影响。redirect参数有安全校验：只允许 `/` 开头的站内路径，拒绝 `//`、`http://`、`javascript:` 等外部跳转。

### 故事31：登录弹窗化——告别独立登录页
你把登录/注册从独立页面改为Navbar弹窗。基于Radix Dialog + Portal + 玻璃模糊遮罩。即使弹窗接入后，`/login` 和 `/register` 旧路由仍作为fallback保留。AuthDialog支持login/register双模式切换，成功后在Navbar立即刷新登录态。

### 故事32：评委一键体验——13900139000的秘密
你在AuthDialog中新增了一个"评委一键体验"按钮，点击后直接用Demo账号（13900139000/123456）登录。不需要真实注册，不需要手机号验证。这个按钮让评委可以在3秒内进入系统，看到所有预置的演示数据。

---

## 第三幕：前端大重构——从Demo到产品的蜕变

### 故事33：岗位卡片的三次重生
最初的JobCard像后台管理系统——技能tag堆叠、英文"open"状态、底部"查看详情"按钮。你做了三次重构：第一次去掉了技能和事业群tag、中文化了状态；第二次删除了描述文字和"查看详情"按钮、卡紧padding；第三次建立三层Surface体系——浅灰大容器+白色卡片+卡片内三行信息。最终的效果对标的不是通用设计，而是腾讯招聘官网的列表卡片。

### 故事34：岗位详情页——对齐腾讯招聘的结构
6月3日，用户反馈"当前详情页和腾讯招聘详情页差距很大"。你彻底重写了JobDetailBanner和JobDetailJD。Hero区域精简为标题+标签+slogan，完整信息移至正文白卡。岗位描述、岗位要求、加分项、面试城市——每个模块都有独立的卡片和标题层级。"招聘部门和工作地"成为独立模块，不再被简单tag堆叠。

### 故事35：Hero Banner的CSS绘制
图片Banner被替换为纯CSS绘制的蓝色渐变Hero。多层光斑、圆环、网格纹理——全部用Tailwind CSS实现，零图片依赖。左侧是主标题+副标题+CTA，右侧是CSS绘制的抽象简历卡片+浮动光斑。响应式三档：lg双栏、sm-md单栏、全宽无圆角。

### 故事36：Navbar的透明毛玻璃之旅
Navbar经历了最曲折的视觉进化：V0半透染蓝 → V1纯白割裂 → V2厚重深蓝条 → V3轻薄透明。你最终锁定的参数是 `bg-slate-950/28`（28%暗化）+ `backdrop-blur-[2px]`（轻量柔化）+ `border-white/5`（几乎无边框）。Hero页透明白字，普通内容页白底深字。滚动时上下滑显隐，delta阈值从30px精确到8px。

### 故事37：字体之战——品牌字的诞生
全站只有浏览器默认字体，没有任何字体栈声明。你冻结了一套三栈方案：正文栈（HarmonyOS→Source Han→Noto→PingFang→YaHei→Inter）、品牌栈（Alimama→MiSans→HarmonyOS→Source Han→Inter）、等宽栈（JetBrains→SFMono→Consolas）。所有品牌文字"Offer捕手"统一使用 `font-brand` utility class。

### 故事38：品牌更名——从"offerr鹅"到"Offer捕手"
6月9日，你执行了全链路品牌更名。15个前端文件、Go后端源码、SQL注释、README、Makefile——所有用户可见的"offerr鹅"和"Offer鹅"全部替换为"Offer捕手"。最终grep验证：前品牌残留为0。

### 故事39：全局Loading体系的建立
之前前端没有任何统一的loading体验——App启动时白屏、路由跳转无反馈、AI生成时内联灰色文字。你一次性新增了3个统一组件：AppBootLoading（全屏品牌loading）、RouteProgressBar（顶部2px蓝色进度条）、AIThinkingLoading（AI生成中卡片式loading）。然后又经过4轮迭代——从全屏遮罩到轻胶囊到最终放弃遮罩只保留进度条——最终锁定为RouteProgressBar作为唯一全局跳转反馈。

### 故事40：简历编辑页的重生
旧编辑页是958行的"后台系统化"巨兽——左侧目录导航、FormSection+FieldRow布局、无骨架屏、无错误处理和重试。你将其彻底重构：1280px grid双栏布局、8个模块白卡独立、右侧状态栏显示完成度/AI可用性、底部保存区合并两个同义按钮、新增ResumeSkeleton骨架屏、错误态增加重试按钮。

### 故事41：简历字段精简——砍掉22个冗余字段
旧简历编辑页有30+个表单字段，大量ATS专用字段（目标薪资、工作类型、入职时间、出勤天数……）对求职学生毫无意义。你执行了V2字段契约冻结：砍掉22个冗余字段，项目/实习/证书/获奖全部精简为"name + description"核心模型。技能从10分类降为3字段。前端3个文件修改，后端JSON透传零改动。

### 故事42：从JSON大表到8张关系型表
6月6日深夜，你做了一个大胆的决定：推翻K8-5-47的JSON大表方案。resumes主表30+列（含10个JSON列）→ 拆分为1主表+7子表（educations/projects/internships/language_exams/certificates/awards/works）。每张子表有独立的CASCADE删除和sort_order排序。这个决定意味着后端Model/Repository/Service/Handler、前端Types/API/Adapter——全链路需要重写。

### 故事43：中文乱码的幽灵
关系型表重构后，你用mysql CLI（默认latin1连接）INSERT了测试中文数据，造成双重编码乱码。MySQL中看到的全是 `???` 和方块字符。你花了3个子阶段来诊断和修复：DELETE乱码行→Go API重新创建→前端check-encoding验证。从此你知道了：永远用charset=utf-8的连接方式操作中文数据。

### 故事44：404页面的诞生
6月6日，你新增了404专页。蓝白灰SaaS风格，清晰的文案和返回首页链接。同时新增了 `isSafeRedirect` 函数，限制redirect参数只能跳转站内白名单路径——拒绝 `//` 开头、`http://`、`javascript:`。版本号非法时自动跳转到正确页面。

---

## 第四幕：四核心功能深度打造

### 故事45：双模型分层策略——Fast与Quality
6月2日，你冻结了双模型策略：Fast Model（deepseek-v4-flash，高频低成本）负责岗位推荐；Quality Model（deepseek-v4-pro，深度分析）负责匹配分析、简历优化、版本对比。每个层级都有独立的fallback到Mock Agent。这不是按功能固定路由，而是同功能内的渐进式生成——Fast秒返给初步判断，Quality深度分析给完整报告。

### 故事46：generation_tier——Fast和Deep的契约
你设计了 `generation_tier` 参数，支持"fast"和"quality"两个值。后端所有4个核心功能（推荐/匹配/优化/对比）都通过tier参数动态选择runner。前端进入时优先GET /latest读取已有结果，无结果再POST fast生成。这个设计让系统在"秒返"和"深度分析"之间有了明确的用户选择。

### 故事47：深度推荐的4层级联
Deep Recommend采用了4层级联架构：L0 Input Gate → L1 Candidate Discovery（候选发现+SEP语义画像匹配）→ L2 Semantic Rerank（语义重排+精排）→ L3 Per-Job Synthesis（逐岗深度合成，每个岗位1次Hy3 high调用）→ L4 Global Calibration（全局校准+Quality Gate）。5个岗位，每岗独立分析，再全局校对一致性。

### 故事48：岗位匹配5模块Deep分析
Deep Match采用5模块+1Judge架构：M1 Decision（方向主导性判断）、M2 Screening Pass Points（初筛通过点）、M3 Screening Elimination Risks（淘汰风险三分类——硬伤/证据不足/方向错误）、M4 Resume Rewrite Commands（后端改写维度）、M5 Job Strategy（投递策略）。Judge模块负责全局一致性校验。

### 故事49：Fast Match从0到完整引擎
Fast Match最初只是简单标签计数——`min(98, 60 + tagScore*0.8)`，base 60毫无意义。你重写了一个830行的 `match_sep_engine.go`，包含8大模块：100+同义词归一化、5级方向适配分析、4级Hard Gate检查、8项结构化覆盖评估、8项5级证据映射、3规则结构化风险检测、5维度加权评分校准、结果组装+投递决策。

### 故事50：5类示范简历的验证矩阵
你维护了5份典型简历：01-算法研究生、02-211后台开发、03-二本产品经理、04-职业探索型、05-优秀但不适合互联网。每轮算法调整后，你都会用这5份简历×46个岗位跑全量验证，确保：高匹配有高分数和具体证据、弱匹配如实标低分不包装、方向不匹配不"建议投递"。这是你的校准基准线。

### 故事51：学历焦虑的消除
6月14日，你做了一个P0修复：普通后台开发岗位的简历优化建议中，不再输出"建议补充研究生学历"这类学历焦虑建议。但同时，高学历研究岗位合理保留学术背景建议。你改了3个Prompt文件的约束规则、后端8个函数的过滤逻辑、覆盖9个字段的生成路径。5个验收测试用例全部通过。

### 故事52：SEP语义证据画像的诞生
你引入了SEP（Semantic Evidence Profile）层：简历保存后同步生成Fast Profile（规则抽取），异步生成Quality Profile（LLM深度分析）。Profile包含 `dominant_profile`（主导画像类型——algorithm_research/backend_engineering/product_operation/business_solution/career_exploration）和结构化标签。推荐、匹配、优化三大功能全部消费Profile作为核心上下文。

### 故事53：Fast/Deep模式隔离的P0修复
6月14日，你发现了一个隐秘的P0 Bug：`FindLatestByResumeAndJob` 不带tier过滤，当Deep结果在时间上比Fast更新时，Fast页面会误展示Deep的完整报告。你修复了后端查询逻辑（新增tier参数+normalize函数），前端拆分 `fastResult`/`deepResult`，mounted时默认展示Fast。这是模式隔离的关键修复。

### 故事54：岗位匹配信息架构的产品化重构
6月14日，你做了一个重大决定：将匹配页从"10模块完整展示"重构为只回答4问——该不该投、为什么、最大风险、怎么改。第一屏收敛为Hero决策卡+3核心判断+依据摘要。面试追问、Agent Trace、quality_checks不再占据主视野。这标志着页面从"AI报告展示"进化为"投递决策工具"。

### 故事55：匹配页三卡Fast View
Fast模式下Deep专属字段（screeningPassPoints/screeningEliminationRisks/resumeRewriteCommands/jobStrategy）全部为空。你改写了Fast前端为数据感知渲染：核心三卡（为什么匹配/最大短板/现在怎么做）、优先级摘要、匹配依据Top-3。空数据时不展示大空卡，改为摘要卡。页面宽度从1200px扩到1120px。

---

## 第五幕：细节打磨——像素级的追求

### 故事56：密码框双眼睛修复
浏览器原生的密码reveal图标和自定义EyeOff图标同时出现——每个密码框都有"双眼睛"。根因是无 `::-ms-reveal` CSS抑制规则。你在 `theme.css` 中新增了3条隐藏规则，同时将3个页面（AuthDialog/LoginPage/RegisterPage）的showPassword state收敛到统一的PasswordInput组件。

### 故事57：表单控件——从33个原生input到统一组件
全站有33个原生 `<input>`、5个 `<select>`、4个 `<textarea>`，各有各的样式。shadcn的Input/Select/Textarea组件已安装但无一被使用。你改造了4个基础组件（Input/PasswordInput/Select/Textarea），统一design token（h-11 rounded-xl border-slate-200 focus:ring-4），然后逐个页面接入——AuthDialog、MyResumesPage（31个控件）、ResumeComparisonPage。

### 故事58：折叠动画的物理正确性
简历编辑页删除子项时有一个严重bug：AnimatedListItem的exit动画使用 `max-h-[2000px]→0`，相当于从4倍实际高度开始过渡，前75%时间无变化，最后40ms文字瞬间被压缩。你重写了动画：exit改用 `getBoundingClientRect()` 测量真实高度→0（160ms），enter用CSS @keyframes（180ms），移除scale防止GPU亚像素模糊。删除时通过 `removingItemKeys`(Set) + 170ms setTimeout 实现延迟删除。

### 故事59：Navbar与折叠模块的联动异常
模块折叠/展开→页面高度变化→浏览器自动调整scrollY→触发scroll事件→Navbar误判为用户滚动→显隐跳动。根因是Navbar无delta阈值防御。修复：Navbar增加30px delta阈值过滤layout shift，编辑页路由禁用auto-hide固定显示。

### 故事60：简历编辑页保存区的位置之战
保存区最初是 `fixed inset-x-0 bottom-0` 固定底部栏，会遮挡长表单的最后一个字段。你尝试过sticky方案，最终改为页面流末尾的普通白卡模块——标题+文案+取消/保存按钮+loading+成功/失败提示全部整合在卡片内。去掉 `pb-32` 大空白，表单最后一个字段不再被遮挡。

### 故事61：通用弹窗的重构
SelectResumeModal有4个P0问题：遮罩透明度仅30%（背景70%可见）、不使用Portal（渲染在组件树内）、无body scroll lock、z-50与Navbar同级。你新建了ResumePickerModal组件：Portal + `backdrop-blur-sm` + body lock + ARIA + single/compare双模式。旧SelectResumeModal改为thin wrapper，保持JobDetailAIDecision零修改。

### 故事62：简历对比选择弹窗的重构
旧弹窗用数字"1/2"标识选中顺序，用户不理解选择顺序影响分析方向。你重构为顶部双槽位（"上一版简历"蓝色标签+"当前版简历"绿色标签）+ 下方2列grid简历池+交换顺序按钮。9份简历不再需要滚轮——maxHeight 820px + grid布局消除了滚动条。

### 故事63：Footer——一个被遗忘的组件
全站11个页面没有任何Footer。你新增了GlobalFooter组件（50行）：品牌/邮箱/版权/问卷入口。挂载在RootLayout的 `</main>` 之后——12条路由100%覆盖，一次挂载终身有效。

### 故事64：BulletList——终结Unicode圆点
4个报告页面存在Unicode `●`/`•` 圆点与文字基线不齐的问题。圆点大小、间距、对齐方式各不统一。你封装了BulletList通用组件——支持6种tone（blue/green/amber/red/purple/slate）、3种size（xs/sm/md）、`items-start` 对齐 + `mt-[0.42em]` 微调。4页面15处列表全部替换，全站Unicode圆点清零。

### 故事65：内部字段的用户可见性清洗
用户页面上出现了"画像"、"标签"、"backend_engineering"、"mock_agent"等内部技术词汇。你做了全站文案清洗：新增 `humanizeProfile`（6种映射：backend_engineering→后端开发相关经历）、新增 `cleanUserText`（画像→简历定位、标签→关键词、内部枚举→自然语言）。4个页面共计40+处文案替换，grep验证用户可见技术词为0。

---

## 第六幕：工程化与部署

### 故事66：Hunyuan到DeepSeek的模型切换
6月初，你完成了从Hunyuan到DeepSeek的Provider切换。新增Hy3 Client作为独立Provider，支持 `reasoning_effort` 参数（high/low/no_think三档）。Fast四功能不走LLM，Deep Agent2用high（决策需要深度推理）、Agent3用low（验证只需轻量）、JSON Repair用no_think。Provider切换对上层Agent完全透明——它们只依赖 `llm.Client` 接口。

### 故事67：.env.example的Key泄漏事件
5月30日，你发现 `.env.example` 中遗留了一个疑似真实LLM API Key。紧急修复：替换为占位符 `sk-your-real-api-key`，同时清除LLM_BASE_URL中的真实URL。从此你强化了安全规则：`LLM_API_KEY` 不得写入docs、日志、Git——在 `.gitignore` 中显式添加 `.env.local`。

### 故事68：init.sql的mysqldump污染
`mysqldump` 命令的stderr warning混入了SQL文件首行，导致 `make up` Step 6报 `ERROR 1064`。你备份了4个污染文件到 `backups/` 目录，移除了warning行，新增了 `scripts/dev/db-backup.sh` 使用 `MYSQL_PWD` 避免stderr污染。这是生产环境必须知悉的教训。

### 故事69：CORS——从硬编码到配置化
CORS Origin最初硬编码在 `cors.go` 中，只允许特定的Origin。你将其改为从环境变量读取 `CORS_ALLOWED_ORIGINS`，支持CSV格式多个Origin，默认包含 `http://localhost`（Nginx同源代理）。`CORS_ALLOW_CREDENTIALS` 默认false——因为当前使用JWT Authorization Header而非Cookie。

### 故事70：Docker生产镜像的补齐
6月11日，你一次性补齐了全部生产镜像化文件：`web/Dockerfile`（node:22-alpine → nginx:1.27-alpine多阶段构建）、`server/Dockerfile`（golang:1.24-alpine → alpine:3.21）、`.dockerignore`、`docker-compose.prod.yml`（4 services，隐藏DB/Redis公网端口）、`.env.prod.example`。`docker compose -f docker-compose.prod.yml config` 验证通过。

### 故事71：部署前全链路只读检查
6月14日上交前，你执行了最后一次全面检查：历史增量SQL已部署、init.sql含 `DROP TABLE IF EXISTS`×24行（P0不可执行标记）、DB表结构完整（25表）、go build/vet/test + npm build全部通过。结论：允许部署但绝对禁止 `mysql < init.sql`、`docker compose down -v`、`make reset`。

### 故事72：开发数据库快照的最终导出
6月8日，你执行了最后一次数据库快照：mysqldump导出→临时库验证→替换init.sql。22张表、2用户、8简历、46岗位、完整SEP profiles——全部落入 `server/sql/init.sql`（798行、1.7MB）。临时库行数全量一致，bad_cnt=0。这是开发环境的最终存档。

---

## 第七幕：简历系统全链路重铸

### 故事73：V2简历字段体系的冻结
你冻结了10模块~90字段的新版简历体系：3个全新模块（证书与语言能力、获奖与竞赛、作品链接与补充信息）+ 7个增强模块（教育+GPA/排名/课程、实习+技术栈/成果、项目+类型/成果/链接、技能+AI技能分层、求职意向+入职/出勤）。舍弃11项ATS专用字段。

### 故事74：Schema Version——V1和V2共存
你在 `resume_versions` 表新增了 `schema_version` 字段，默认1。V2请求携带 `schema_version=2` 时走新的10模块JSON字段；V1请求走旧字段。后端通过 `extractStringFromJSON` 等6个辅助函数从V2字段回填V1兼容字段。前后端各自渐进迁移，不互相阻塞。

### 故事75：Agent Snapshot统一化
四核心Agent（推荐/匹配/优化/对比）的简历输入各自手工拼装DB字段，方式不统一。你新增了 `NormalizeResumeToSnapshotV2` 归一化层：V1→V2映射（TargetRole→meta.targetDirection+jobIntention.targetRoles，SkillsJSON→skills.otherKeywords）和V2→V2映射（JSON字段优先，summary列fallback）。phone/email/user_id排除，不进入Agent上下文。

### 故事76：9份示范简历的精准设计
你为Demo账号设计了9份简历，每份都有明确的使用场景：01-03岗位匹配（算法研/后端/产品）、04-05岗位推荐（探索型/稳定型）、06-07版本对比（基础版vs提升版）、08-09简历优化（接近投递版vs基础薄弱版）。每份简历的自我评价都是80-140字的拟真文本，不是模板生成。

### 故事77：编辑回填——38个V2字段不丢失
编辑模式V2回填有严重问题：adapter丢弃V2 raw字段 → ViewModel不传V2 → MyResumesPage不回填。38个V2 draft字段在编辑时全部丢失。你修复了3个断点：ViewModel新增 `v2?` 可选字段、adapter保留V2 raw模块、MyResumesPage新增80行V2回填逻辑（7个列表+joinList/joinLines helpers）。

### 故事78：简历详情页的诞生
6月6日，你新增了 `ResumeDetailPage.tsx`（~280行）：纯只读详情页，蓝色渐变Hero卡片+8模块卡片化展示。前端10维度计算完善度（教育/项目/实习/证书/获奖/语言/作品/技能/补充信息/AI技能）。简历管理页每行新增"查看详情"按钮。这是简历系统的最后一块拼图。

### 故事79：保存前必填校验与冲突提示
新增简历名+姓名必填校验，缺填时弹出AppConfirmDialog提示。名称重复检测红色化（duplicate/invalid → `bg-red-50 border-red-200 text-red-600`），不再使用降低透明度的弱提示。POST/PUT返回409时自动标记为duplicate并弹窗。

### 故事80：教育经历导师与实验室字段拆分
"导师 / 实验室"原本是合并标签展示——既不好看也不符合腾讯校招简历样式。你将其拆分为两个独立字段：DB新增 `advisor` 列、后端Model/Repository/Service全链路新增Advisor、前端编辑页拆为两个Input、详情页拆为两个InfoItem。0合并标签残留。

---

## 第八幕：最终冲刺

### 故事81：岗位匹配的8组合回归矩阵
6月14日，Deep Match做了最后一次全局校准。8组简历×岗位组合回归矩阵审计：M1新增方向主导性判断（修复算法→后端误判78→52和算法子领域误判48→78）、M5新增算法子领域相邻规则。零硬编码校准，全通过Prompt约束完成。

### 故事82：3个经典案例的锁定
你锁定了3个岗位匹配经典案例：C1-周明远×后台开发 score 78 revise_before_apply（可投，先改简历）、C2-score 33 change_direction（建议换方向）、C3-score 42 not_recommended（不建议投递）。三案例覆盖了匹配分析的全部输出场景。C1数据支撑度24/25为最高。

### 故事83：Hy3 reasoning_effort三档接入
你为四核心Deep Agent的LLM调用点接入 `reasoning_effort` 分级：Agent2（Decision）用high（需要深度推理做判断）、Agent3（Verifier）用low（只需轻量验证）、JSON Repair用no_think（纯结构修复不需思考）。前端的AgentStepLoadingOverlay文案反映真实Agent阶段。

### 故事84：推荐页状态机的9状态契约
你冻结了推荐页的9状态流转：booting→fast_loading→fast_done→deep_loading→deep_done→deep_error→deep_unavailable。每个状态都定义了主区域内容、右侧按钮、顶部warning、deepResult规则。deepResult合法定义：仅≥5 jobs可set。不完整历史设deepHistoryIncomplete，不污染deepResult。

### 故事85：Deep推荐Loading体验的彻底重写
旧Loading弹窗95%视觉空间是静态灰字，只有40px spinner有动效。`noAutoAdvance=true` 关闭假进度后，替代设计全是pending静态步骤。你重写了DeepAgentLoadingOverlay（233行）：96px AI Orb（4层动画：breathe+orbit×2+pulse-dot）、横向4个Agent协同链路shimmer、Top5岗位矩阵breathing dot。"正在深度推荐智能体分析"替代了旧的"深度推荐分析中"。

### 故事86：Visual Language统一审计
6月14日，你做了全站视觉语言审计。抽取首页和推荐页作为"定版参考"，审计匹配页跳脱点：页面宽度1280→1440、Loading风格不统一、骨架屏尺寸不匹配。建议抽取5个统一组件（UnifiedAgentLoadingOverlay、AppButton、SectionCard、StatusBadge、AiNotice）并应用于匹配页。

### 故事87：文本质量兜底——cleanText的进化
UTF-8截断导致 `�`(U+FFFD) 污染用户可见文本。根因是 `focus[:80]` 按字节截断而非rune。你修复了后端 `truncateStr` 函数改用 `[]rune`，前端新增 `cleanText()` 过滤U+FFFD/控制字符/undefined/NaN/[object Object]。覆盖所有用户可见文本字段。

### 故事88：岗位匹配最终文案微调
6月14日，新增 `normalizeDecisionLabel()` 函数，统一 change_direction label为"建议换方向"。修改 `match_deep_agent.go` 一行+新增40行标准化函数。这是上线前最后的文案校准——确保用户不会看到"not_recommended"、"change_direction"等原始枚举值。

### 故事89：全站文本一致性——零工程词汇
你执行了全站用户可见区域的工程词汇扫描：画像→简历定位、标签→关键词、mock_agent/mock_quality→空、"规则兜底"→"基础模型"、"规则分析"→"标准匹配分析"、"AI生成"→"智能匹配分析"。全站grep确认：用户可见技术词汇为0。

### 故事90：简历管理页的产品定位纠偏
6月5日，用户明确反馈："/resumes是简历管理页面，不是功能管理页面。"原来页面按Agent功能（匹配/推荐/对比/优化）分组简历，每行主按钮是"去匹配岗位/推荐岗位"。你彻底重构：普通简历列表不再分组、删除4个功能概览卡片和DEFAULT_JOB_MAP硬编码映射、每行仅"编辑"按钮、右侧新增智能求职助手面板。

### 故事91：匹配页与推荐页视觉对齐
匹配页文字尺寸全面缩小对齐推荐页设计标准：Score 48→36px、标题text-base→text-[14px]、正文text-sm→text-[13px]、辅助text-xs→text-[11px]。SectionCard新增titleClassName/descriptionClassName prop，默认标题text-[14px]。两个页面从"各自为政"变为"同一血液"。

### 故事92：推荐页To匹配页的上下文穿帮修复
推荐页跳转匹配页时只传resumeId不传jobId→匹配页显示的是错误的岗位。修复只需要一行：`RecommendationDetail.tsx` L110 navigate添加jobId参数。这一行修复了核心用户流程中的上下文断裂。

### 故事93：简历优化页入口弹窗的实现
用户从简历管理页点"优化"→直接进入general模式（无目标岗位，优化建议空泛）。你新增了JobPicker岗位搜索组件+OptimizationEntryModal入口弹窗：两步式（先选简历再选岗位），target_job成为主路径。简历管理页和详情页的"优化"入口从Link改为弹窗。

### 故事94：版本对比从"画像对比器"到"迭代诊断器"
旧版本对比页面以profile为核心，是"画像方向对比器"。你将其重构为"简历迭代诊断器"：迭代结论→新版提升→新版问题→旧版保留→合并方案→改写示例→初筛影响。Mock Agent 8个函数重写，字段重复"方向方向"bug修复，cleanText去掉无用替换。

### 故事95：SEP Profile资产补齐与预热
7份简历缺少SEP SemanticProfile。你新增了 `SemanticProfileBackfillService`（scan+backfill+dry-run+并发控制+幂等）+ `sep-backfill` CLI。执行backfill：6/7份简历成功生成Profile（1份Hy3输出验证失败），46个岗位全部跳过（已ready）。resume SEP从2/9→8/9。

### 故事96：岗位推荐Fast推荐的非阻塞修复
P0 Bug：`fastRecommendWithSEP()` 中同步调用 `EnsureResumeProfile()`，对无SEP的简历触发同步Hy3 LLM生成（7/9简历受影响，耗时可能数十秒）。修复：改为 `FindLatestByTarget` 只读+ `ErrFastRecommendSEPMissing` 快速降级到TagEvidenceService。新增异步后台补齐+sync.Mutex in-flight去重。

### 故事97：首页右侧登录按钮的弹窗化
6月3日，你发现首页右侧AI卡片中"登录"和"注册新账号"按钮仍用 `navigate("/login")` 跳旧页面。修复：AIDecisionWorkspace接入AuthDialog弹窗——同Navbar一套交互。"评委一键体验"在两处入口都可用。

### 故事98：最终提交准备——5份文档
6月14日，你生成了5份最终提交文档：最终提交说明、演示路径（10条路径全部顺畅）、视频脚本、方案说明、审计日志。服务/构建/API/Demo账号全部通过验收。nginx 502标记为P1（需部署前修复）。整个K9-Final全系列（Nav/Modal/Page-Review/Acceptance）正式收口。

### 故事99：14,460行变更记录的写作
在整个近一个月的开发过程中，你同时维护了 `docs/变更记录.md`。每完成一个阶段或子任务，你都会记录性质、修改范围、验收结果、构建状态。这不是简单的changelog——每个条目都包含根因分析、决策逻辑、修复结果。14,460行，近500个条目，平均每天25条变更记录。

### 故事100：Offer捕手——不是一个项目，是一段旅程
从5月28日的第一行 `init.sql`，到6月24日的部署前最终检查，这不是一个简单的全栈项目。这是一个完整的SaaS产品的诞生：Go后端202个文件（146个.go）、React前端207个文件（139个.tsx）、46个腾讯真实岗位、8张关系型数据库表、4个LLM Agent、4个核心功能（推荐/匹配/优化/对比）、1000篇设计文档、14,460行变更记录。你一个人完成了从数据库设计到AI Prompt Engineering、从后端六层架构到前端像素级打磨、从Mock Agent到真实LLM Pipeline接入、从开发环境到Docker生产镜像的全链路工作。这不是一段代码的历史——这是一个开发者把"投递前决策系统"从想法变成现实的完整史诗。

---

> **尾声**  
> 这100个故事是对14,460行变更记录的浓缩叙事。每个故事背后，都有curl验证的API、grep清零的搜索、npm build通过的控制台输出、go test全部绿色的终端。它们不是虚构，是一个开发者近一个月每天高强度工作的真实印记。
