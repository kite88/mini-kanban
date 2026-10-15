"use client";

import { memo, useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  closestCorners,
  useSensor,
  useSensors,
  type CollisionDetection,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { useDroppable } from "@dnd-kit/core";
import {
  SortableContext,
  sortableKeyboardCoordinates,
  verticalListSortingStrategy,
} from "@dnd-kit/sortable";
import { useI18n } from "@/components/I18nProvider";
import { TaskCard, TaskCardBody } from "@/components/TaskCard";
import { api } from "@/lib/api";
import { statusLabel } from "@/lib/labels";
import { COLUMNS, type Task, type TaskStatus } from "@/lib/types";

const STATUS_IDS = COLUMNS.map((c) => c.id);

// 空列共享同一个数组，避免每次渲染都生成新引用
const EMPTY_IDS: string[] = [];

// 传感器配置放模块级：useSensor 的 memo 依赖 options 引用，
// 写成内联对象会让 sensors 每次渲染都是新数组，加剧 dnd-kit 的测量/渲染循环。
const POINTER_SENSOR_OPTIONS = { activationConstraint: { distance: 5 } };
const KEYBOARD_SENSOR_OPTIONS = { coordinateGetter: sortableKeyboardCoordinates };

function group(tasks: Task[]): Record<TaskStatus, Task[]> {
  const g: Record<TaskStatus, Task[]> = { todo: [], doing: [], done: [], accepted: [] };
  for (const t of tasks) {
    (g[t.status] ?? g.todo).push(t);
  }
  // position 相同时按 id 兜底排序：id 是 UUID 字符串，必须用 localeCompare（减法会得到 NaN）
  for (const k of STATUS_IDS)
    g[k].sort((a, b) => a.position - b.position || a.id.localeCompare(b.id));
  return g;
}

const Column = memo(function Column({
  status,
  title,
  accent,
  tasks,
  onCardClick,
  onAdd,
  canDrag,
  canCreate,
}: {
  status: TaskStatus;
  title: string;
  accent: string;
  tasks: Task[];
  onCardClick: (t: Task) => void;
  onAdd: (status: TaskStatus) => void;
  canDrag: (t: Task) => boolean;
  canCreate: boolean;
}) {
  const { t } = useI18n();
  const { setNodeRef, isOver } = useDroppable({ id: status });
  // items 必须保持引用稳定：useSortable 会把它塞进 droppable 的 data /
  // resizeObserverConfig，每次渲染换新数组会让测量与 setState 互相触发，
  // 最终抛 "Maximum update depth exceeded"。
  // 这里用「id 序列」作为 memo 键：只要该列的任务顺序没变，就复用同一个数组引用。
  const idsKey = tasks.map((t) => t.id).join(",");
  const itemIds = useMemo(() => (idsKey ? idsKey.split(",") : EMPTY_IDS), [idsKey]);

  return (
    <section className="flex w-[320px] shrink-0 flex-col rounded-2xl bg-slate-200/60 p-3 dark:bg-slate-900/60">
      <header className="mb-3 flex items-center justify-between px-1">
        <div className="flex items-center gap-2">
          <span className={`h-2 w-2 rounded-full ${accent}`} />
          <h2 className="text-sm font-semibold text-slate-700 dark:text-slate-200">{title}</h2>
          <span className="rounded bg-white px-1.5 text-xs text-slate-500 dark:bg-slate-800 dark:text-slate-400">
            {tasks.length}
          </span>
        </div>
        {canCreate && (
          <button
            onClick={() => onAdd(status)}
            className="rounded px-2 text-slate-500 transition hover:bg-white hover:text-indigo-600 dark:text-slate-400 dark:hover:bg-slate-800 dark:hover:text-indigo-400"
            title={t("board.columnAdd")}
          >
            ＋
          </button>
        )}
      </header>

      <div
        ref={setNodeRef}
        className={`flex-1 space-y-2 rounded-xl p-1 transition ${
          isOver ? "bg-indigo-100/70 dark:bg-indigo-500/20" : "bg-transparent"
        }`}
      >
        <SortableContext items={itemIds} strategy={verticalListSortingStrategy}>
          {tasks.map((t) => (
            <TaskCard key={t.id} task={t} onOpen={onCardClick} disabled={!canDrag(t)} />
          ))}
        </SortableContext>

        {tasks.length === 0 && (
          <div className="grid h-24 place-items-center rounded-xl border border-dashed border-slate-300 text-xs text-slate-400 dark:border-slate-700 dark:text-slate-500">
            {t("board.columnEmpty")}
          </div>
        )}
      </div>
    </section>
  );
});

export function Board({
  tasks,
  onReload,
  onEdit,
  onCreate,
  onError,
  canDrag,
  canCreate = true,
}: {
  tasks: Task[];
  onReload: () => void;
  onEdit: (t: Task) => void;
  onCreate: (status: TaskStatus) => void;
  onError?: (msg: string) => void;
  canDrag?: (t: Task) => boolean;
  canCreate?: boolean;
}) {
  const { t } = useI18n();
  // 拖拽属于编辑的一部分：无权限的卡片不允许拖动（后端 MoveTask 兜底校验）
  const canDragTask = canDrag ?? (() => true);
  const [columns, setColumns] = useState<Record<TaskStatus, Task[]>>(() => group(tasks));
  const [activeId, setActiveId] = useState<string | null>(null);
  // 拖拽起点（起始列 + 起始下标）：onDragOver 会乐观改写 columns，
  // 结束后靠它判断「位置是否真的变化」，避免漏掉跨列移动。
  const dragOrigin = useRef<{ status: TaskStatus; index: number } | null>(null);

  useEffect(() => {
    setColumns(group(tasks));
  }, [tasks]);

  const sensors = useSensors(
    useSensor(PointerSensor, POINTER_SENSOR_OPTIONS),
    useSensor(KeyboardSensor, KEYBOARD_SENSOR_OPTIONS),
  );

  const findContainer = useCallback(
    (id: string): TaskStatus | null => {
      if ((STATUS_IDS as string[]).includes(id)) return id as TaskStatus;
      return (
        STATUS_IDS.find((s) => columns[s].some((t) => String(t.id) === id)) ?? null
      );
    },
    [columns],
  );

  const activeTask = useMemo(() => {
    if (!activeId) return null;
    for (const s of STATUS_IDS) {
      const t = columns[s].find((x) => String(x.id) === activeId);
      if (t) return t;
    }
    return null;
  }, [activeId, columns]);

  // 关键：把「被拖拽卡片自身」这个 droppable 从碰撞检测里剔除。
  // 使用 DragOverlay 时，它自身的 rect 会跟随指针移动，与拖拽矩形距离恒为 0，
  // 会和真正的目标列打平并被选中，导致 over 恒等于自己、松手后什么都不发生。
  const collisionDetection = useCallback<CollisionDetection>(
    (args) =>
      closestCorners({
        ...args,
        droppableContainers: args.droppableContainers.filter((c) => c.id !== args.active.id),
      }),
    [],
  );

  const handleDragStart = (e: DragStartEvent) => {
    const id = String(e.active.id);
    setActiveId(id);
    dragOrigin.current = null;
    for (const s of STATUS_IDS) {
      const index = columns[s].findIndex((t) => String(t.id) === id);
      if (index !== -1) {
        dragOrigin.current = { status: s, index };
        break;
      }
    }
  };

  const handleDragEnd = async (e: DragEndEvent) => {
    const { active, over } = e;
    const origin = dragOrigin.current;
    dragOrigin.current = null;
    setActiveId(null);

    const activeTaskId = String(active.id);
    // 落在看板之外，或原地松手：不改变位置
    if (!over || String(over.id) === activeTaskId) {
      onReload();
      return;
    }

    const overId = String(over.id);
    const to = findContainer(overId);
    if (!to) {
      onReload();
      return;
    }

    let task: Task | undefined;
    for (const s of STATUS_IDS) {
      task = columns[s].find((t) => String(t.id) === activeTaskId);
      if (task) break;
    }
    if (!task) return;

    // 目标列去掉自身后的插入下标：落在卡片上按上/下半区决定，落在列的空白处则放到末尾
    const items = columns[to].filter((t) => String(t.id) !== activeTaskId);
    const overIdx = items.findIndex((t) => String(t.id) === overId);
    const translated = active.rect.current.translated;
    const isBelow =
      overIdx !== -1 && !!translated && translated.top > over.rect.top + over.rect.height / 2;
    const index = overIdx === -1 ? items.length : overIdx + (isBelow ? 1 : 0);

    const next = { ...columns };
    for (const s of STATUS_IDS) next[s] = columns[s].filter((t) => String(t.id) !== activeTaskId);
    next[to] = [...next[to].slice(0, index), { ...task, status: to }, ...next[to].slice(index)];
    setColumns(next);

    // 同列且位置没变化时不必请求后端
    if (origin && origin.status === to && origin.index === index) return;

    try {
      await api.moveTask(activeTaskId, to, index);
    } catch (err) {
      // 不要静默吞掉：拖拽失败时把原因交给页面提示，否则看起来就是“拖了没反应”
      const msg = err instanceof Error ? err.message : t("board.moveFailed");
      console.error("[board] moveTask failed:", msg);
      onError?.(msg);
    } finally {
      onReload();
    }
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={collisionDetection}
      onDragStart={handleDragStart}
      // 刻意不监听 onDragOver：跨列乐观移动会让被拖拽卡片在拖拽中被挂到另一列的
      // DOM 里，而 dnd-kit 的碰撞检测正是用这张卡自身的 rect —— 卡片一动、碰撞结果
      // 就翻转，于是「移动 → 翻转 → 移回」在高频 dragOver 下震荡，并不断重挂可拖拽
      // 节点，最终触发 useRect 的测量循环（Maximum update depth exceeded）。
      // 拖拽期间只靠列高亮提示目标列，位置在 onDragEnd 一次性提交。
      onDragEnd={handleDragEnd}
      onDragCancel={() => {
        dragOrigin.current = null;
        setActiveId(null);
        onReload();
      }}
    >
      <div className="flex gap-4 overflow-x-auto pb-4">
        {COLUMNS.map((c) => (
          <Column
            key={c.id}
            status={c.id}
            title={statusLabel(t, c.id)}
            accent={c.accent}
            tasks={columns[c.id]}
            onCardClick={onEdit}
            onAdd={onCreate}
            canDrag={canDragTask}
            canCreate={canCreate}
          />
        ))}
      </div>

      <DragOverlay>
        {activeTask ? (
          <div className="w-[296px] rotate-1">
            <TaskCardBody task={activeTask} dragging />
          </div>
        ) : null}
      </DragOverlay>
    </DndContext>
  );
}
