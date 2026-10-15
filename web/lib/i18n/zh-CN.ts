/**
 * 简体中文文案（源语言）。
 *
 * 约定：
 * - 键名用「域.用途」的扁平点号形式，服务端 go 侧的错误目录用同一套语义（见 README）；
 * - `{name}` 是插值占位符，由 t(key, { name }) 替换；
 * - 这里新增 / 改名一个键，en.ts 会因类型不匹配而编译报错，避免漏翻。
 */
export const zhCN = {
  // 站点级
  "app.title": "任务看板",
  "app.description": "Next.js + Go Gin 拖拽式任务看板",

  // 通用
  "common.loading": "加载中...",
  "common.save": "保存",
  "common.saving": "保存中...",
  "common.cancel": "取消",
  "common.close": "关闭",
  "common.confirm": "确认",
  "common.confirmAction": "确定",
  "common.delete": "删除",
  "common.edit": "编辑",
  "common.view": "查看",
  "common.reset": "重置",
  "common.query": "查询",
  "common.clear": "清除",
  "common.empty": "—",
  "common.unknown": "未知",
  "common.notSet": "未设置",
  "common.unassigned": "未指派",
  "common.loadFailed": "加载失败",
  "common.saveFailed": "保存失败",
  "common.deleteFailed": "删除失败",
  "common.operationFailed": "操作失败",
  "common.uploadFailed": "上传失败",
  "common.requestFailed": "请求失败 ({status})",
  "common.uploadRequestFailed": "上传失败 ({status})",
  /** 列表连接符：如「2 个任务、1 条资料」 */
  "common.listSeparator": "、",

  // 导航与身份
  "nav.projects": "项目看板",
  "nav.trash": "回收箱",
  "nav.users": "用户管理",
  "nav.departments": "部门管理",
  "nav.logs": "操作日志",
  "nav.logout": "退出",

  "role.admin": "管理员",
  "role.member": "成员",

  // 看板列（与后端 models.TaskStatusLabels 对应）
  "column.todo": "待办",
  "column.doing": "进行中",
  "column.done": "已完成",
  "column.accepted": "已验收",

  // 主题与语言
  "theme.label": "主题模式",
  "theme.light": "浅色",
  "theme.system": "跟随系统",
  "theme.dark": "深色",
  "language.label": "界面语言",

  // 登录
  "login.title": "登录任务看板",
  "login.subtitle": "请使用管理员分配的账号登录",
  "login.email": "邮箱",
  "login.password": "密码",
  "login.submit": "登录",
  "login.submitting": "登录中...",
  "login.failed": "登录失败",

  // 用户管理
  "users.title": "用户管理",
  "users.subtitle": "仅管理员可新增 / 修改 / 删除用户",
  "users.create": "新建用户",
  "users.edit": "编辑用户",
  "users.col.id": "ID",
  "users.col.username": "用户名",
  "users.col.email": "邮箱（登录账号）",
  "users.col.department": "部门",
  "users.col.role": "角色",
  "users.col.createdAt": "创建时间",
  "users.col.actions": "操作",
  "users.field.username": "用户名",
  "users.field.usernamePlaceholder": "张三",
  "users.field.department": "部门",
  "users.field.email": "邮箱（登录账号）",
  "users.field.password": "密码",
  "users.field.passwordKeep": "密码（留空则不修改）",
  "users.field.role": "角色",
  "users.unassigned": "未分配",
  "users.empty": "暂无用户",
  "users.deleteTitle": "删除用户",
  "users.deleteDesc": "确认删除用户「{name}」？",
  "users.deleteNote":
    "用户会被物理删除且无法恢复：被指派的任务会解除关联，软删除项目的归属会转给操作人。",

  // 部门管理
  "departments.title": "部门管理",
  "departments.subtitle":
    "部门独立成表，用户在「用户管理」里关联；有成员的部门需先调整归属才能删除",
  "departments.create": "新建部门",
  "departments.edit": "编辑部门",
  "departments.col.id": "ID",
  "departments.col.name": "部门名称",
  "departments.col.remark": "备注",
  "departments.col.userCount": "人数",
  "departments.col.createdAt": "创建时间",
  "departments.col.actions": "操作",
  "departments.field.name": "部门名称",
  "departments.field.namePlaceholder": "如 研发部",
  "departments.field.remark": "备注",
  "departments.field.remarkPlaceholder": "负责人、职能说明等（可留空）",
  "departments.nameRequired": "部门名称不能为空",
  "departments.empty": "暂无部门",
  "departments.deleteTitle": "删除部门",
  "departments.deleteDesc": "确认删除部门「{name}」？",
  "departments.deleteNote": "部门下还有用户时无法删除，需先把这些用户的归属调整到别的部门。",

  // 项目列表
  "projects.title": "项目",
  "projects.subtitle": "每个项目拥有一张可拖拽的任务看板",
  "projects.create": "新建项目",
  "projects.empty": "还没有项目，点击右上角「新建项目」开始",
  "projects.noDescription": "暂无描述",
  "projects.summary": "{tasks} 个任务 · {members} 位成员",

  // 项目表单 / 删除
  "project.createTitle": "新建项目",
  "project.editTitle": "编辑项目",
  "project.field.title": "标题",
  "project.field.titlePlaceholder": "例如：官网改版",
  "project.field.description": "描述",
  "project.field.descriptionPlaceholder": "项目背景、目标等",
  "project.field.members": "关联成员（只有被关联的人和管理员能看到该项目）",
  "project.deleteTitle": "删除项目",
  "project.deleteDesc": "确认删除项目「{title}」？该项目下所有任务会被一并删除。",
  "project.deleteNote": "可在回收箱中恢复。",

  // 看板
  "board.backToProjects": "← 返回项目列表",
  "board.docs": "项目资料",
  "board.editProject": "编辑项目",
  "board.deleteProject": "删除项目",
  "board.createTask": "+ 新建任务",
  "board.filter.all": "全部",
  "board.filter.mine": "我创建的",
  "board.filter.assigned": "指派给我的",
  "board.taskSummary": "共 {visible} / {total} 个任务",
  "board.emptyAll": "还没有任务",
  "board.emptyFiltered": "当前筛选下没有任务",
  "board.members": "关联成员 {count} 人：",
  "board.deleteTaskTitle": "删除任务",
  "board.deleteTaskDesc": "确认删除任务「{title}」？",
  "board.columnAdd": "在此列新建任务",
  "board.columnEmpty": "拖拽任务到这里",
  "board.moveFailed": "移动失败",

  // 任务
  "task.createTitle": "新建任务",
  "task.editTitle": "编辑任务",
  "task.detailTitle": "任务详情",
  "task.field.title": "标题",
  "task.field.titlePlaceholder": "任务标题",
  "task.field.description": "描述",
  "task.field.descriptionPlaceholder":
    "补充信息：支持加粗、列表、引用、链接与图片（可直接粘贴截图）",
  "task.field.attachments": "附件",
  "task.field.assignee": "关联人",
  "task.field.status": "所属列",
  "task.field.dueDate": "任务到期时间",
  "task.field.creator": "创建人",
  "task.field.dueDatePlaceholder": "点击选择日期时间",
  "task.titleRequired": "标题不能为空",
  "task.unknownUser": "用户 #{id}",
  "task.me": "（我）",
  "task.readonlyNote":
    "只读模式：仅管理员、任务创建人、被指派人和项目创建人可编辑该任务（拖拽改状态同样受限）",
  "task.deleteNote": "仅管理员、任务创建人、被指派人和项目创建人可删除",
  "task.noDescription": "暂无描述",
  "task.attachmentCount": "{count} 个附件",
  "task.createdByMe": "我创建",
  "task.dueAt": "截止 {time}",
  "task.overdueAt": "已逾期 · {time}",

  // 项目资料
  "docs.title": "项目资料",
  "docs.createTitle": "新建资料",
  "docs.editTitle": "编辑资料",
  "docs.detailTitle": "资料详情",
  "docs.create": "+ 新建资料",
  "docs.searchPlaceholder": "按标题搜索资料",
  "docs.count": "共 {visible} / {total} 条",
  "docs.empty": "还没有项目资料",
  "docs.emptyHint":
    "，点击右上角「新建资料」添加需求文档、会议纪要、设计稿、项目密钥、测试账密等",
  "docs.noMatch": "没有标题包含「{keyword}」的资料",
  "docs.noContent": "暂无内容",
  "docs.updatedAt": "更新于 {time}",
  "docs.field.title": "标题",
  "docs.field.titlePlaceholder": "资料标题，如：需求文档 / 10 月会议纪要",
  "docs.field.content": "内容",
  "docs.field.contentPlaceholder":
    "补充资料内容：支持加粗、列表、引用、链接与图片（可直接粘贴截图）",
  "docs.field.attachments": "附件",
  "docs.field.creator": "创建人",
  "docs.field.updatedAt": "更新时间",
  "docs.titleRequired": "标题不能为空",
  "docs.backToList": "返回列表",
  "docs.deleteTitle": "删除项目资料",
  "docs.deleteDesc": "确认删除资料「{title}」？",
  "docs.deleteNote": "可在回收箱中恢复。",

  // 附件
  "attachment.empty": "暂无附件",
  "attachment.remove": "移除",
  "attachment.removeTitle": "移除附件",
  "attachment.uploading": "上传中…",
  "attachment.add": "＋ 添加附件",
  "attachment.hint": "可多选，支持图片 / 文档 / 压缩包等",

  // 用户下拉
  "userSelect.placeholder": "搜索用户名 / 邮箱 / 部门",
  "userSelect.noMatch": "没有匹配的用户",

  // 成员选择器
  "memberPicker.noDepartment": "未分配部门",
  "memberPicker.departmentFallback": "部门 #{id}",
  "memberPicker.empty": "暂无可选用户",
  "memberPicker.creator": "创建人",
  "memberPicker.selected": "已选 {count} 人；勾选部门名称可整部门选中",

  // 富文本编辑器
  "editor.placeholder": "补充信息",
  "editor.bold": "加粗",
  "editor.italic": "斜体",
  "editor.underline": "下划线",
  "editor.strike": "删除线",
  "editor.heading": "小标题",
  "editor.unorderedList": "无序列表",
  "editor.orderedList": "有序列表",
  "editor.quote": "引用",
  "editor.code": "代码块",
  "editor.link": "插入链接",
  "editor.image": "插入图片（也可直接粘贴截图）",
  "editor.removeFormat": "清除格式",
  "editor.uploading": "图片上传中…",
  "editor.imageUploadFailed": "图片上传失败",

  // 操作日志
  "logs.title": "操作日志",
  "logs.subtitle":
    "记录项目 / 任务 / 项目资料 / 用户 / 部门的操作留痕；「删除用户」「删除部门」「回收箱彻底删除」属于物理删除，日志里会附带被删数据的完整快照。",
  "logs.filter.allTargets": "全部类型",
  "logs.filter.allActions": "全部操作",
  "logs.searchPlaceholder": "搜索对象名称 / 摘要 / 操作人",
  "logs.col.time": "时间",
  "logs.col.operator": "操作人",
  "logs.col.targetType": "类型",
  "logs.col.action": "操作",
  "logs.col.target": "对象",
  "logs.col.summary": "变更摘要",
  "logs.col.detail": "详情",
  "logs.empty": "没有符合条件的操作日志",
  "logs.detailFailed": "加载详情失败",
  "logs.count": "共 {total} 条 · 第 {page} / {pages} 页",
  "logs.prev": "上一页",
  "logs.next": "下一页",
  "logs.detailTitle": "操作日志详情",
  "logs.snapshot": "数据快照",
  "logs.snapshotEmpty":
    "该操作没有数据快照。只有物理删除（删除用户 / 删除部门 / 回收箱彻底删除）会留存快照，其余操作的数据仍在库中，可直接在对应页面查看。",

  // 日志枚举：对象类型 / 动作
  "logTarget.project": "项目",
  "logTarget.task": "任务",
  "logTarget.doc": "项目资料",
  "logTarget.user": "用户",
  "logTarget.department": "部门",
  "logAction.create": "新建",
  "logAction.update": "修改",
  "logAction.move": "移动",
  "logAction.delete": "删除",
  "logAction.restore": "恢复",
  "logAction.purge": "彻底删除",

  // 回收箱
  "trash.title": "回收箱",
  "trash.subtitle":
    "删除的项目、任务与项目资料会先放在这里，可以随时恢复；「彻底删除」之后数据不可找回。恢复项目时，随它一起被删的任务与项目资料也会一并回来。",
  "trash.empty": "回收箱是空的",
  "trash.projects": "已删除的项目",
  "trash.tasks": "已删除的任务",
  "trash.docs": "已删除的项目资料",
  "trash.noProjects": "没有已删除的项目",
  "trash.noTasks": "没有单独删除的任务（随项目一起删除的任务在上面的项目条目里）",
  "trash.noDocs": "没有单独删除的项目资料（随项目一起删除的资料在上面的项目条目里）",
  "trash.unknownProject": "未知项目",
  "trash.projectSummary": "{tasks} 个任务 · {docs} 条资料 · 删除于 {time}",
  "trash.projectSummaryCreator": "创建人 {name} · ",
  "trash.taskSummary": "{project} · 原状态 {status} · 删除于 {time}",
  "trash.docSummary": "{project} · 删除于 {time}",
  "trash.attachmentSuffix": " · {count} 个附件",
  "trash.restore": "恢复",
  "trash.purge": "彻底删除",
  "trash.countTasks": "{count} 个任务",
  "trash.countDocs": "{count} 条资料",
  "trash.purgeProjectTitle": "彻底删除项目",
  "trash.purgeProjectDesc": "确认彻底删除项目「{title}」{inner}？",
  "trash.purgeProjectInner": "及其下 {items}",
  "trash.purgeProjectNote": "物理删除，删除后无法恢复，也不会进入回收箱；附件文件同样会被清掉。",
  "trash.purgeTaskTitle": "彻底删除任务",
  "trash.purgeTaskDesc": "确认彻底删除任务「{title}」？",
  "trash.purgeDocTitle": "彻底删除项目资料",
  "trash.purgeDocDesc": "确认彻底删除资料「{title}」？",
  "trash.purgeNote": "物理删除，删除后无法恢复，也不会进入回收箱。",

  // 表单校验
  "validate.emailInvalid": "邮箱格式不正确，示例：you@example.com",
} as const;

/** 全部文案键：en.ts 必须逐一覆盖，否则编译报错 */
export type MessageKey = keyof typeof zhCN;
