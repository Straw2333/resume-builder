<template>
  <el-dialog
    :model-value="modelValue"
    :title="state === 'done' ? 'AI 整份简历优化 · 比对报告与采纳' : 'AI 整份简历优化'"
    :width="state === 'done' ? '920px' : '600px'"
    :close-on-click-modal="false"
    :show-close="state !== 'running'"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <!-- 第一步：说明 + 目标岗位 + 全局优化方向 -->
    <template v-if="state === 'idle'">
      <div class="intro">
        <p>将逐条优化简历中所有已填写的内容（共 <b>{{ tasks.length }}</b> 个模块条目）：</p>
        <ul>
          <li>优化完成后生成<b>前后比对报告</b>，先预览对比、确认后再写入简历</li>
          <li>支持按<b>模块条目</b>或条目内的<b>单条要点（行）</b>自由勾选采纳，未勾选部分保留原文</li>
          <li>在比对报告中可针对任意条目继续<b>对话微调</b>，约需 {{ estimate }}</li>
        </ul>
      </div>
      <template v-if="tasks.length">
        <div class="field">
          <label>目标岗位（选填，让优化更贴合岗位要求）</label>
          <el-input v-model="targetRole" placeholder="如：Go 后端开发工程师" />
        </div>
        <div class="field">
          <label>全局优化方向 / 修改要求（选填，指明整体润色侧重点）</label>
          <el-input
            v-model="instruction"
            type="textarea"
            :rows="3"
            placeholder="如：整体采用 STAR 法则表达，突出高并发与性能量化指标，语言精简干练…"
          />
          <div class="tag-bar">
            <span class="tag-label">快捷方向：</span>
            <button
              v-for="tag in DIRECTION_TAGS"
              :key="tag.label"
              type="button"
              class="dir-chip"
              @click="instruction = appendDirection(instruction, tag.text)"
            >
              + {{ tag.label }}
            </button>
          </div>
        </div>
      </template>
      <div v-else class="empty-tip">当前简历还没有已填写的内容（描述 / 自我评价），无需优化。</div>
    </template>

    <!-- 第二步：进度 -->
    <template v-else-if="state === 'running'">
      <div class="progress-text">
        正在优化第 {{ done + 1 }} / {{ tasks.length }} 条：<b>{{ currentLabel }}</b>
      </div>
      <el-progress :percentage="percent" :stroke-width="10" />
      <div class="progress-tip">
        正在生成优化结果与比对报告，可随时点击下方「提前结束并查看报告」检视已完成的部分。
      </div>
    </template>

    <!-- 第三步：比对报告 + 双层勾选 + 单条对话微调 -->
    <template v-else>
      <div class="report-topbar">
        <div class="report-stat">
          <span class="ok-badge">✅ 优化完成</span>
          <span>
            共生成 <b>{{ okItems.length }}</b> 个模块条目（<b>{{ totalPointCount }}</b> 条要点），当前已勾选采纳
            <b class="accent">{{ selectedPointCount }}</b> 条 AI 要点
          </span>
          <span v-if="failCount" class="fail-badge">（{{ failCount }} 条失败，可展开重试）</span>
        </div>
        <div class="report-global-ops">
          <el-button size="small" text type="primary" @click="toggleAllPoints(true)">全部采纳 AI</el-button>
          <el-button size="small" text @click="toggleAllPoints(false)">全部保留原文</el-button>
          <el-button size="small" text @click="toggleExpandAll">{{ allExpanded ? '全部收起' : '全部展开' }}</el-button>
        </div>
      </div>

      <div class="report-list">
        <div
          v-for="item in reportItems"
          :key="item.id"
          class="report-card"
          :class="{ 'is-fail': item.status === 'fail' }"
        >
          <!-- 条目卡片头部 -->
          <div class="card-header" @click="item.expanded = !item.expanded">
            <div class="card-left" @click.stop>
              <el-checkbox
                v-if="item.status === 'ok'"
                :model-value="isItemAllSelected(item)"
                :indeterminate="isItemIndeterminate(item)"
                @change="(val) => setItemAllSelected(item, val)"
              />
              <span class="sec-tag">{{ item.task.sectionTitle }}</span>
              <span class="item-title">{{ item.task.title || item.task.sectionTitle }}</span>
              <span v-if="item.task.subtitle" class="item-sub">· {{ item.task.subtitle }}</span>
            </div>
            <div class="card-right" @click.stop>
              <template v-if="item.status === 'ok'">
                <span class="point-stat" :class="{ active: itemSelectedCount(item) > 0 }">
                  已采纳 {{ itemSelectedCount(item) }} / {{ item.pairs.length }} 条要点
                </span>
                <el-button
                  size="small"
                  :type="item.chatOpen ? 'primary' : 'default'"
                  plain
                  @click="item.chatOpen = !item.chatOpen; if (item.chatOpen) item.expanded = true"
                >
                  <el-icon><ChatDotRound /></el-icon>&nbsp;对话微调
                </el-button>
              </template>
              <template v-else-if="item.status === 'skipped'">
                <span class="skip-text">已跳过（未执行）</span>
                <el-button size="small" type="primary" plain :loading="item.refining" @click="retryItem(item)">
                  优化此条
                </el-button>
              </template>
              <template v-else>
                <span class="fail-text">优化失败</span>
                <el-button size="small" type="primary" plain :loading="item.refining" @click="retryItem(item)">
                  重试此条
                </el-button>
              </template>
              <el-icon class="expand-icon" :class="{ open: item.expanded }" @click="item.expanded = !item.expanded">
                <ArrowDown />
              </el-icon>
            </div>
          </div>

          <!-- 单条对话微调面板 -->
          <div v-if="item.chatOpen && item.status === 'ok'" class="item-chat-panel">
            <div v-if="item.chatHistory.length" class="chat-history">
              <div v-for="(h, hIdx) in item.chatHistory" :key="hIdx" class="chat-item">
                <span class="chat-tag">微调 #{{ hIdx + 1 }}</span>
                <span>{{ h }}</span>
              </div>
            </div>
            <div class="chat-input-row">
              <el-input
                v-model="item.followUp"
                size="small"
                :placeholder="`针对「${item.task.title || item.task.sectionTitle}」输入修改方向（如：突出架构职责、精简第1条）回车发送…`"
                :disabled="item.refining"
                @keyup.enter="refineSingleItem(item)"
              />
              <el-button
                size="small"
                type="primary"
                :loading="item.refining"
                :disabled="!item.followUp.trim()"
                @click="refineSingleItem(item)"
              >
                发送微调
              </el-button>
            </div>
            <div class="tag-bar compact">
              <span class="tag-label">快捷要求：</span>
              <button
                v-for="tag in DIRECTION_TAGS"
                :key="tag.label"
                type="button"
                class="dir-chip"
                :disabled="item.refining"
                @click="item.followUp = appendDirection(item.followUp, tag.text)"
              >
                {{ tag.label }}
              </button>
            </div>
          </div>

          <!-- 条目内要点（<li>）逐行比对 -->
          <div v-show="item.expanded && item.status === 'ok'" class="card-body" v-loading="item.refining">
            <div class="diff-col-header">
              <div class="col-check">采纳</div>
              <div class="col-pane">优化前（原文）</div>
              <div class="col-pane ai">AI 优化后（点击或勾选即采纳该要点）</div>
            </div>
            <div
              v-for="p in item.pairs"
              :key="p.index"
              class="diff-row"
            >
              <div class="col-check">
                <el-checkbox v-model="p.selected" :disabled="!p.ai && !p.orig" />
                <span class="row-num">#{{ p.index + 1 }}</span>
              </div>
              <div
                class="col-pane pane-box orig-pane"
                :class="{ 'keep-orig': !p.selected && p.orig }"
                @click="p.orig && (p.selected = false)"
              >
                <div class="pane-badge" v-if="!p.selected && p.orig">保留原文</div>
                <div v-if="p.orig" v-html="p.orig"></div>
                <span v-else class="empty-line">（原文无对应项 / AI 新增要点）</span>
              </div>
              <div
                class="col-pane pane-box ai-pane"
                :class="{ 'use-ai': p.selected && p.ai, 'dimmed': !p.selected }"
                @click="p.ai && (p.selected = !p.selected)"
              >
                <div class="pane-badge ai" v-if="p.selected && p.ai">已采纳 AI</div>
                <div v-if="p.ai" v-html="p.ai"></div>
                <span v-else class="empty-line">（AI 已合并/省略此条，取消勾选可保留左侧原文）</span>
              </div>
            </div>
          </div>
        </div>
      </div>
    </template>

    <template #footer>
      <template v-if="state === 'idle'">
        <el-button v-if="hasBackup" @click="restore">恢复上次优化前备份</el-button>
        <el-button @click="$emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :disabled="!tasks.length" @click="run">开始优化</el-button>
      </template>
      <template v-else-if="state === 'running'">
        <el-button @click="stopEarly">提前结束并查看报告</el-button>
      </template>
      <template v-else>
        <el-button @click="$emit('update:modelValue', false)">取消（不修改简历）</el-button>
        <el-button @click="state = 'idle'">返回重设方向</el-button>
        <el-button type="primary" :disabled="!okItems.length" @click="applySelected">
          确认应用所选（采纳 {{ selectedPointCount }} / {{ totalPointCount }} 条要点）
        </el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, onBeforeUnmount, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api/resume'
import { DIRECTION_TAGS, appendDirection, buildPointPairs, mergeSelectedPoints } from '../utils/aiDiff'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  // [{ kind, sectionType, sectionTitle, title, subtitle, content, section, item }]
  tasks: { type: Array, default: () => [] },
  hasBackup: { type: Boolean, default: false }
})
const emit = defineEmits(['update:modelValue', 'apply-report', 'restore'])

const state = ref('idle') // idle | running | done
const targetRole = ref('')
const instruction = ref('')
const done = ref(0)
const currentLabel = ref('')
const stopped = ref(false)
const reportItems = ref([])

const percent = computed(() =>
  props.tasks.length ? Math.round((done.value / props.tasks.length) * 100) : 0
)
const estimate = computed(() => {
  const s = Math.max(props.tasks.length * 8, 15)
  return s > 60 ? `约 ${Math.round(s / 60)} 分钟` : `约 ${s} 秒`
})

const okItems = computed(() => reportItems.value.filter((i) => i.status === 'ok'))
const failCount = computed(() => reportItems.value.filter((i) => i.status === 'fail').length)
const totalPointCount = computed(() =>
  okItems.value.reduce((acc, it) => acc + it.pairs.length, 0)
)
const selectedPointCount = computed(() =>
  okItems.value.reduce((acc, it) => acc + itemSelectedCount(it), 0)
)
const allExpanded = computed(() =>
  reportItems.value.length > 0 && reportItems.value.every((it) => it.expanded)
)

function itemSelectedCount(item) {
  return (item.pairs || []).filter((p) => p.selected && p.ai).length
}

function isItemAllSelected(item) {
  const valid = (item.pairs || []).filter((p) => p.ai)
  return valid.length > 0 && valid.every((p) => p.selected)
}

function isItemIndeterminate(item) {
  const valid = (item.pairs || []).filter((p) => p.ai)
  const sel = valid.filter((p) => p.selected).length
  return sel > 0 && sel < valid.length
}

function setItemAllSelected(item, val) {
  ;(item.pairs || []).forEach((p) => {
    p.selected = val ? Boolean(p.ai) : false
  })
}

function toggleAllPoints(val) {
  okItems.value.forEach((item) => setItemAllSelected(item, val))
}

function toggleExpandAll() {
  const target = !allExpanded.value
  reportItems.value.forEach((it) => {
    it.expanded = target
  })
}

async function run() {
  const role = targetRole.value.trim()
  const inst = instruction.value.trim()
  localStorage.setItem('aiTargetRole', role)
  state.value = 'running'
  done.value = 0
  stopped.value = false
  reportItems.value = []

  for (let i = 0; i < props.tasks.length; i++) {
    const task = props.tasks[i]
    if (stopped.value) {
      reportItems.value.push(
        reactive({
          id: i,
          task,
          status: 'skipped',
          aiHtml: '',
          pairs: [],
          expanded: false,
          chatOpen: false,
          followUp: '',
          chatHistory: inst ? [`全局方向：${inst}`] : [],
          refining: false
        })
      )
      continue
    }
    currentLabel.value = task.title || task.sectionTitle
    const row = reactive({
      id: i,
      task,
      status: 'ok',
      aiHtml: '',
      pairs: [],
      expanded: true,
      chatOpen: false,
      followUp: '',
      chatHistory: inst ? [`全局方向：${inst}`] : [],
      refining: false
    })
    try {
      const r = await api.aiPolish({
        kind: task.kind,
        sectionType: task.sectionType,
        sectionTitle: task.sectionTitle,
        title: task.title,
        subtitle: task.subtitle,
        targetRole: role,
        content: task.kind === 'section-content' ? plainToUl(task.content) : task.content,
        instruction: inst
      })
      row.aiHtml = r.html
      row.pairs = buildPointPairs(task.content, r.html)
      row.status = 'ok'
    } catch {
      row.status = 'fail'
    }
    reportItems.value.push(row)
    done.value++
  }
  state.value = 'done'
}

function plainToUl(text) {
  const esc = (s) => String(s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
  return (
    '<ul>' +
    String(text || '')
      .split('\n')
      .filter((l) => l.trim())
      .map((l) => `<li>${esc(l.trim())}</li>`)
      .join('') +
    '</ul>'
  )
}

async function retryItem(item) {
  item.refining = true
  try {
    const task = item.task
    const r = await api.aiPolish({
      kind: task.kind,
      sectionType: task.sectionType,
      sectionTitle: task.sectionTitle,
      title: task.title,
      subtitle: task.subtitle,
      targetRole: targetRole.value.trim(),
      content: task.kind === 'section-content' ? plainToUl(task.content) : task.content,
      instruction: instruction.value.trim()
    })
    item.aiHtml = r.html
    item.pairs = buildPointPairs(task.content, r.html)
    item.status = 'ok'
    item.expanded = true
  } catch {
    // 拦截器已提示
  } finally {
    item.refining = false
  }
}

async function refineSingleItem(item) {
  const text = item.followUp.trim()
  if (!text || item.refining) return
  item.refining = true
  const task = item.task
  const baseHtml = mergeSelectedPoints(item.pairs, 'item-desc') || item.aiHtml
  try {
    const r = await api.aiPolish({
      kind: task.kind,
      sectionType: task.sectionType,
      sectionTitle: task.sectionTitle,
      title: task.title,
      subtitle: task.subtitle,
      targetRole: targetRole.value.trim(),
      content: task.kind === 'section-content' ? plainToUl(task.content) : task.content,
      instruction: text,
      currentHtml: baseHtml
    })
    item.aiHtml = r.html
    item.pairs = buildPointPairs(task.content, r.html)
    item.chatHistory.push(text)
    item.followUp = ''
    ElMessage.success(`「${task.title || task.sectionTitle}」已按新要求微调更新`)
  } catch {
    // 拦截器已提示
  } finally {
    item.refining = false
  }
}

function stopEarly() {
  stopped.value = true
}

function applySelected() {
  const payload = okItems.value.map((item) => ({
    task: item.task,
    merged: mergeSelectedPoints(item.pairs, item.task.kind)
  }))
  emit('apply-report', payload)
  emit('update:modelValue', false)
  ElMessage.success(`已应用所选修改（共采纳 ${selectedPointCount.value} 条 AI 优化要点）`)
}

function restore() {
  emit('restore')
  emit('update:modelValue', false)
}

onBeforeUnmount(() => {
  if (state.value === 'running') stopped.value = true
})

function open() {
  targetRole.value = localStorage.getItem('aiTargetRole') || ''
  instruction.value = ''
  state.value = 'idle'
  reportItems.value = []
}

defineExpose({ open })
</script>

<style scoped>
.intro { font-size: 13.5px; color: #334155; line-height: 1.7; }
.intro ul { margin: 6px 0 0; padding-left: 20px; color: #64748b; font-size: 13px; }
.field { margin-bottom: 12px; margin-top: 14px; }
.field label { display: block; font-size: 12.5px; color: #475569; margin-bottom: 5px; font-weight: 500; }
.tag-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin-top: 7px; }
.tag-bar.compact { margin-top: 6px; }
.tag-label { font-size: 12px; color: #94a3b8; }
.dir-chip {
  border: 1px solid #cbd5e1; background: #f8fafc; color: #334155;
  border-radius: 999px; padding: 2px 10px; font-size: 12px; cursor: pointer;
  transition: all 0.15s;
}
.dir-chip:hover:not(:disabled) { border-color: #3b82f6; color: #2563eb; background: #eff6ff; }
.dir-chip:disabled { opacity: 0.5; cursor: not-allowed; }

.empty-tip { font-size: 13.5px; color: #94a3b8; margin-top: 12px; }
.progress-text { font-size: 14px; color: #0f172a; margin-bottom: 14px; }
.progress-tip { font-size: 12px; color: #94a3b8; margin-top: 14px; }

/* 比对报告顶栏 */
.report-topbar {
  display: flex; justify-content: space-between; align-items: center; flex-wrap: wrap; gap: 8px;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
  padding: 8px 12px; margin-bottom: 12px;
}
.report-stat { font-size: 13px; color: #334155; display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.ok-badge { color: #16a34a; font-weight: 600; }
.fail-badge { color: #dc2626; font-size: 12px; }
.accent { color: #2563eb; }

/* 报告列表 */
.report-list {
  display: flex; flex-direction: column; gap: 10px;
  max-height: 62vh; overflow-y: auto; padding-right: 4px;
}
.report-card {
  flex-shrink: 0;
  border: 1px solid #e2e8f0; border-radius: 10px; background: #ffffff;
  overflow: hidden;
}
.report-card.is-fail { border-color: #fecaca; }

.card-header {
  display: flex; justify-content: space-between; align-items: center; gap: 12px;
  padding: 10px 14px; min-height: 44px; box-sizing: border-box;
  background: #f8fafc; cursor: pointer; user-select: none;
}
.card-left { display: flex; align-items: center; gap: 8px; min-width: 0; }
.sec-tag {
  font-size: 11.5px; padding: 2px 7px; border-radius: 5px;
  background: #e2e8f0; color: #334155; font-weight: 600; flex-shrink: 0;
}
.item-title { font-size: 13.5px; font-weight: 600; color: #0f172a; }
.item-sub { font-size: 12.5px; color: #64748b; }

.card-right { display: flex; align-items: center; gap: 10px; flex-shrink: 0; }
.point-stat { font-size: 12px; color: #94a3b8; }
.point-stat.active { color: #2563eb; font-weight: 500; }
.skip-text { font-size: 12px; color: #64748b; }
.fail-text { font-size: 12px; color: #dc2626; }
.expand-icon { color: #64748b; transition: transform 0.2s; cursor: pointer; }
.expand-icon.open { transform: rotate(180deg); }

/* 单条对话微调面板 */
.item-chat-panel {
  padding: 10px 14px; background: #f0f7ff; border-top: 1px solid #dbeafe;
}
.chat-history {
  display: flex; flex-direction: column; gap: 4px; margin-bottom: 6px;
}
.chat-item { font-size: 12px; color: #334155; display: flex; align-items: baseline; gap: 6px; }
.chat-tag {
  background: #dbeafe; color: #1e40af; padding: 1px 6px; border-radius: 4px; font-size: 11px; flex-shrink: 0;
}
.chat-input-row { display: flex; gap: 8px; }

/* 卡片主体：要点对比 */
.card-body {
  padding: 10px 14px 12px; border-top: 1px solid #f1f5f9;
  display: flex; flex-direction: column; gap: 8px;
}
.diff-col-header {
  display: flex; gap: 10px; font-size: 12px; font-weight: 600; color: #64748b;
  padding-bottom: 2px;
}
.diff-col-header .ai { color: #2563eb; }
.col-check { width: 50px; flex-shrink: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2px; }
.row-num { font-size: 11px; color: #94a3b8; }
.col-pane { flex: 1; min-width: 0; }

.diff-row { display: flex; gap: 10px; align-items: stretch; }
.pane-box {
  border: 1px solid #e2e8f0; border-radius: 8px;
  padding: 8px 12px; font-size: 13px; line-height: 1.65; color: #1f2937;
  background: #fafbfc; cursor: pointer; transition: all 0.15s;
}
.pane-box.keep-orig { border-color: #64748b; background: #f1f5f9; }
.pane-box.ai-pane.use-ai { background: #eff6ff; border-color: #60a5fa; }
.pane-box.ai-pane.dimmed { opacity: 0.55; background: #f8fafc; }
.pane-badge {
  display: inline-block; font-size: 11px; padding: 1px 6px; border-radius: 4px;
  background: #e2e8f0; color: #475569; margin-bottom: 4px; font-weight: 600;
}
.pane-badge.ai { background: #dbeafe; color: #1d4ed8; }
.empty-line { color: #94a3b8; font-size: 12px; font-style: italic; }
</style>
