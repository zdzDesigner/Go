# Spec-Kit 使用复盘

## 1. Spec-Kit 概述

### 什么是 Spec-Kit
Spec-Kit 是 GitHub 开源的 **Spec-Driven Development（规格驱动开发）** 工具包，核心理念是 **"Specifications become executable"** —— 让规格文档不再是静态描述，而是驱动 AI 生成代码的"可执行程序"。

### 核心特性
- **`specify` CLI**：项目初始化和 AI 集成
- **Slash 命令**：`/speckit.specify`（写规格）、`/speckit.plan`（生成计划）、`/speckit.implement`（执行实现）
- **AI Agent 支持**：Claude Code、GitHub Copilot 等
- **分阶段工作流**：定义原则 → 写功能规格 → 规划实现 → 任务分解 → 系统化执行

### 中文版本
- 项目地址：https://github.com/Linfee/spec-kit-cn
- 安装命令：`uv tool install specify-cn-cli --from git+https://github.com/linfee/spec-kit-cn.git`
- 依赖：Python 3.11+、Git、uv

## 2. Spec-Kit 工作原理

### 架构组成

#### 目录结构
```
.project-root/
├── .specify/
│   ├── memory/
│   │   └── constitution.md          # 项目原则/价值观
│   ├── scripts/                     # 工作流脚本
│   └── templates/                   # 规格模板
│       ├── spec-template.md         # 功能规格模板
│       ├── plan-template.md         # 技术方案模板
│       └── tasks-template.md        # 任务分解模板
├── specs/<feature-id>/
│   ├── spec.md                      # 功能规范（what & why）
│   ├── plan.md                      # 技术实施计划（how）
│   ├── tasks.md                     # 任务分解
│   ├── data-model.md                # 数据模型定义（可选）
│   ├── research.md                  # 技术调研（可选）
│   ├── quickstart.md                # 快速启动指南（可选）
│   └── contracts/                   # API 合同（可选）
└── CLAUDE.md                        # AI 交互日志（若使用 Claude Code）
```

#### 核心命令
```md
| 命令                 | 作用                              |
|----------------------+-----------------------------------|
| `specify init`       | 初始化项目，创建 `.specify/` 结构 |
| `/speckit.specify`   | AI 起草功能规格（spec.md）        |
| `/speckit.plan`      | AI 生成技术方案（plan.md）        |
| `/speckit.tasks`     | AI 分解任务列表（tasks.md）       |
| `/speckit.implement` | AI 执行实现                       |
```



### 工作流程

```
阶段 1：定义原则
   ↓
   constitution.md ← 项目价值观、技术约束

阶段 2：编写规格
   ↓
   /speckit.specify → spec.md（功能描述、用户场景）

阶段 3：澄清需求
   ↓
   AI 与用户对话确认细节

阶段 4：生成方案
   ↓
   /speckit.plan → plan.md（架构设计、技术选型）

阶段 5：任务分解
   ↓
   /speckit.tasks → tasks.md（可执行的待办清单）

阶段 6：系统实现
   ↓
   /speckit.implement → 根据 tasks.md 逐步编码
```

### AI 交互机制

- **输入**：AI 读取 `constitution.md`、`spec.md` 等文件作为上下文
- **处理**：通过预定义脚本模板生成结构化文档
- **输出**：写入对应的 `.md` 文件，供下一阶段使用
- **状态管理**：通过 `CLAUDE.md` 等记忆文件跟踪进度

### 关键特性

- **分阶段约束**：每个阶段只关注特定目标（规格 vs 实现分离）
- **模板驱动**：标准化输出格式，便于 AI 解析
- **渐进式细化**：从高层规格逐步细化到可执行代码

## 3. Qoder 如何使用 Spec-Kit

### Qoder 的能力限制

#### ❌ 不支持的功能
- **无法自动读取文档**：不会主动感知 Spec-Kit 生成的文档，需用户明确指示（如"读取 spec.md"）
- **没有内置 slash 命令**：不支持 `/speckit.specify`、`/speckit.plan` 等命令
- **无法自定义命令**：无法添加新的 slash 命令或修改命令系统
- **不会主动感知文件系统**：文件变化后需用户告知才能获取最新内容

#### ✅ 可以做的事情
- **手动模拟工作流**：通过现有工具（Read, Write, Edit, Bash）模拟 Spec-Kit 的完整流程
- **根据模板生成文档**：基于 `.specify/templates/` 中的模板生成规范文档
- **读取并执行文档**：解析 spec/plan/tasks 文档并据此编写代码
- **灵活调整流程**：可跳过、合并或自定义工作流阶段

### Qoder 使用 Spec-Kit 的实际方式

#### 核心原理
Qoder **不能自动使用** Spec-Kit，而是**手动模拟**其工作流程：

```
用户明确指示 → Qoder 执行对应操作 → 生成/读取文档 → 继续下一步
```

#### 具体执行步骤

当用户说："用 Spec-Kit 为功能 X 创建文档"时，Qoder 会：

**步骤 1：创建目录结构**
```bash
mkdir -p specs/feature-x
```

**步骤 2：生成 spec.md**
- 使用 Read 工具读取 `.specify/memory/constitution.md`（项目原则）
- 使用 Read 工具读取 `.specify/templates/spec-template.md`（模板）
- 根据用户需求填充模板内容
- 使用 Write 工具写入 `specs/feature-x/spec.md`

**步骤 3：生成 plan.md**
- 使用 Read 工具读取 `specs/feature-x/spec.md`
- 使用 Read 工具读取 `plan-template.md`
- 设计技术方案（架构、技术栈、实现细节）
- 使用 Write 工具写入 `specs/feature-x/plan.md`

**步骤 4：生成 tasks.md**
- 使用 Read 工具读取 `plan.md`
- 分解为可执行任务列表
- 使用 Write 工具写入 `specs/feature-x/tasks.md`

**步骤 5：执行实现**
- 使用 Read 工具读取 `tasks.md`
- 按任务顺序编写代码（使用 Edit/Write 工具）
- 使用 Edit 工具更新 `tasks.md` 中的任务状态（标记为 `[x]`）

#### 与原版 Spec-Kit 的对比
```md
| 对比维度       | 原版 Spec-Kit                    | Qoder 的方式                   |
|----------------|----------------------------------|--------------------------------|
| **触发方式**   | `/speckit.specify` 等 slash 命令 | 自然语言指令（"创建 spec.md"） |
| **自动化程度** | CLI 一键生成                     | 手动调用工具逐步生成           |
| **文档读取**   | 自动读取上下文文件               | 需明确指示读取哪些文件         |
| **状态跟踪**   | 自动更新 CLAUDE.md               | 手动标记任务完成状态           |
| **工作流控制** | 严格按阶段推进                   | 灵活，可跳过或合并阶段         |
| **错误处理**   | 内置验证逻辑                     | 依赖用户审阅确认               |
```



#### 为什么 Qoder 无法完全自动化？

**1. 缺少上下文感知能力**
- 原版 Spec-Kit 通过 CLI 脚本自动检测项目状态
- Qoder 需要用户明确告知当前在哪个阶段

**2. 没有命令系统扩展机制**
- `/speckit.*` 命令是其他 AI 工具（如 Claude Code）通过配置文件注入的
- Qoder 的工具集是固定的，无法动态添加新命令

**3. 文件变化不可见**
- 如果用户手动编辑了 spec.md，Qoder 不会自动感知
- 需要用户说"重新读取 spec.md"才能获取最新内容

#### 实际使用对比示例

**原版 Spec-Kit（在 Claude Code 中）**：
```
用户：/speckit.specify 添加多用户认证
→ 自动读取 constitution.md
→ 自动生成 specs/multi-user-auth/spec.md
→ 更新 CLAUDE.md 记录进度

用户：/speckit.plan
→ 自动读取 spec.md
→ 自动生成 plan.md
```

**Qoder 的方式**：
```
用户："为多用户认证创建 spec.md"
Qoder：
  1. 使用 Read 读取 constitution.md
  2. 使用 Read 读取 spec-template.md
  3. 使用 Write 生成 specs/multi-user-auth/spec.md
  4. 回复："已生成 spec.md，需要生成 plan.md 吗？"

用户："生成 plan.md"
Qoder：
  1. 使用 Read 读取 spec.md
  2. 使用 Read 读取 plan-template.md
  3. 使用 Write 生成 plan.md
```

### 实际使用方式

#### 方式一：通过 Qoder 手动执行

**步骤 1：编写功能规格**
```
用户："为多用户认证功能创建 spec.md"
Qoder：基于 spec-template.md 生成 specs/multi-user-auth/spec.md
```

**步骤 2：生成技术方案**
```
用户："根据 spec.md 生成 plan.md"
Qoder：读取 spec.md，生成 plan.md
```

**步骤 3：分解任务**
```
用户："根据 plan.md 生成 tasks.md"
Qoder：生成任务列表
```

**步骤 4：执行实现**
```
用户："按照 tasks.md 第一个任务开始实现"
Qoder：逐步编写代码
```

#### 方式二：手动编辑（推荐学习阶段）

```bash
# 1. 创建功能目录
mkdir -p specs/feature-name

# 2. 复制模板
cp .specify/templates/spec-template.md specs/feature-name/spec.md

# 3. 手动填写 spec.md，然后告诉 Qoder
"读取 specs/feature-name/spec.md，生成对应的 plan.md"
```

#### 方式三：使用 specify-cn CLI（需安装）

```bash
specify-cn init              # 初始化项目
specify-cn create feature-x  # 创建功能分支
# 然后在其他 AI 工具中使用 /speckit.* 命令
```

### 推荐工作模式

**直接用自然语言描述意图**（最高效）：
```
"为 SFTP 服务器的断点续传功能创建 spec/plan/tasks，然后实现"
```

Qoder 会自动：
1. 创建目录结构
2. 生成规范文档
3. 编写代码
4. 运行测试

### 约定关键词方案

用户可以用简短指令触发特定操作：

```bash
用户："spec: 添加公钥认证"
Qoder：→ 生成 spec.md

用户："plan: 公钥认证"  
Qoder：→ 生成 plan.md

用户："implement: tasks.md 第一项"
Qoder：→ 开始编码
```

## 4. 本项目的 Spec-Kit 初始化

### 已完成的工作

1. **创建目录结构**
```
.specify/
├── memory/
│   └── constitution.md          # 项目原则
├── scripts/                     # 工作流脚本（空）
└── templates/
    ├── spec-template.md         # 功能规格模板
    ├── plan-template.md         # 技术方案模板
    └── tasks-template.md        # 任务分解模板
```

2. **编写 constitution.md**
定义了项目的核心价值观和技术约束：
- 安全第一（路径隔离、认证方式）
- 代码质量（简洁、错误处理、注释规范）
- 性能与可靠性（并发、资源限制、异常处理）
- 可维护性（配置管理、日志、职责分离）
- 技术约束（Go 1.20+、架构原则、测试要求）
- 开发规范（Git 提交、文档、发布流程）

3. **创建三个核心模板**
- `spec-template.md`：功能规格编写指南
- `plan-template.md`：技术方案设计框架
- `tasks-template.md`：任务分解标准格式

### 与原版 Spec-Kit 的区别
```md
| 对比项     | 原版 Spec-Kit         | 本项目实践                    |
|------------+-----------------------+-------------------------------|
| 触发方式   | `/speckit.*` 命令     | 自然语言指令                  |
| 自动化程度 | CLI 工具一键生成      | Qoder 手动读写文件            |
| AI 集成    | 内置到 Claude Code 等 | 需明确告知 Qoder              |
| 工作流     | 严格按阶段推进        | 灵活，可跳过或合并阶段        |
| 适用场景   | 大型项目规范化开发    | 学习 Spec-Kit 理念 + 快速原型 |
```



## 5. 使用建议

### 适合使用 Spec-Kit 的场景
- 新功能开发前需要充分讨论和设计
- 多人协作项目，需要统一理解需求
- 复杂功能，需要分阶段推进
- 需要文档驱动开发（文档即代码）

### 不适合的场景
- 简单 bug 修复（直接改代码更快）
- 紧急热修复
- 一次性实验性代码

### 最佳实践
1. **先定义 constitution.md**：明确项目原则，避免后期返工
2. **spec.md 聚焦"为什么"**：不要过早陷入技术细节
3. **plan.md 要具体**：明确文件路径、函数签名、数据结构
4. **tasks.md 要可执行**：每个任务有明确的验收标准
5. **迭代更新文档**：实现过程中发现问题，及时更新 spec/plan

### 示例工作流（完整流程）

假设要为 SFTP 服务器添加"断点续传"功能：

```bash
# 1. 用户发起需求
"用 Spec-Kit 为 SFTP 添加断点续传功能"

# 2. Qoder 创建目录
mkdir -p specs/resume-transfer

# 3. Qoder 生成 spec.md
# - 读取 constitution.md
# - 根据 spec-template.md 格式编写
# - 包含用户故事、验收标准、非功能需求

# 4. 用户审阅并确认
"spec.md 看起来没问题，生成 plan.md"

# 5. Qoder 生成 plan.md
# - 技术选型（如使用 HTTP Range 头）
# - 架构设计（如何存储断点信息）
# - 关键实现细节（续传协议）

# 6. 用户审阅并确认
"生成 tasks.md"

# 7. Qoder 生成 tasks.md
# - 任务 1：设计断点存储结构
# - 任务 2：修改上传接口支持 offset
# - 任务 3：实现续传逻辑
# - 任务 4：添加测试用例

# 8. 开始实现
"按照 tasks.md 开始实现"

# 9. Qoder 逐个任务执行并更新状态
# - 完成任务 1 后勾选 [x]
# - 提交代码
# - 继续下一个任务
```

## 6. 总结

### Spec-Kit 的价值
- **提升协作效率**：统一的文档规范减少沟通成本
- **降低开发风险**：提前发现设计问题
- **知识沉淀**：文档即项目知识库
- **AI 友好**：结构化文档让 AI 更准确理解需求

### Qoder 与 Spec-Kit 的结合
虽然 Qoder 不能自动执行 `/speckit.*` 命令，但通过：
1. **手动模拟工作流**
2. **自然语言交互**
3. **工具链组合使用**

可以实现 Spec-Kit 的核心价值，同时保持灵活性。

### 下一步行动
- 尝试为现有 SFTP 项目的某个新功能编写完整 spec/plan/tasks
- 根据实践反馈优化模板格式
- 探索与 Git 工作流的集成（如 feature 分支与 spec 目录对应）
