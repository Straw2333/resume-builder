<template>
  <el-dialog
    :model-value="modelValue"
    :title="`AI 润色 · ${target?.sectionTitle || ''}${target?.title ? ' - ' + target.title : ''}`"
    width="820px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <!-- 第一步：目标岗位 + 优化方向指令 -->
    <template v-if="state === 'idle'">
      <div class="field">
        <label>目标岗位（选填，让润色更贴合岗位要求）</label>
        <el-input
          v-model="targetRole"
          :placeholder="`如：${target?.sectionType === 'project' ? '后端开发工程师' : '前端开发实习生'}`"
        />
      </div>

      <div class="field">
        <label>优化方向 / 具体修改要求（选填，告诉 AI 你的侧重点）</label>
        <el-input
          v-model="instruction"
          type="textarea"
          :rows="3"
          placeholder="如：突出高并发架构设计，强化个人主导贡献，把第二条写得更精炼…"
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

      <div class="orig-preview">
        <div class="orig-preview-title">待优化原文预览</div>
        <div class="content-box" v-html="target?.content || '<i>（空）</i>'"></div>
      </div>
      <div class="tip">AI 会保持原有事实不编造，生成后可逐条比对并勾选要采纳的要点，还可继续对话微调。</div>
    </template>

    <!-- 第二步：逐要点比对报告 + 对话追问微调 -->
    <template v-else>
      <div class="report-toolbar">
        <div class="report-summary">
          共比对 <b>{{ pairs.length }}</b> 条要点，已勾选采纳 <b class="accent">{{ selectedCount }}</b> 条 AI 优化
          <span class="sub-hint">（未勾选的要点将保留原文）</span>
        </div>
        <div class="report-actions">
          <el-button size="small" text type="primary" @click="selectAll(true)">全选 AI</el-button>
          <el-button size="small" text @click="selectAll(false)">全部保留原文</el-button>
        </div>
      </div>

      <div class="diff-header">
        <div class="col-check">采纳</div>
        <div class="col-pane">优化前（原文）</div>
        <div class="col-pane ai">AI 优化后（勾选即采纳此条）</div>
      </div>

      <div class="diff-list" v-loading="loading">
        <div
          v-for="p in pairs"
          :key="p.index"
          class="diff-row"
          :class="{ 'is-selected': p.selected }"
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

      <!-- 对话微调窗口 -->
      <div class="chat-box">
        <div class="chat-header">
          <el-icon><ChatDotRound /></el-icon>
          <span>对话式微调（对当前结果不满意？直接告诉 AI 怎么改）</span>
        </div>
        <div v-if="chatHistory.length" class="chat-history">
          <div v-for="(msg, idx) in chatHistory" :key="idx" class="chat-item">
            <span class="chat-tag">{{ msg.type === 'init' ? '初始方向' : `追问 #${idx}` }}</span>
            <span class="chat-text">{{ msg.text }}</span>
          </div>
        </div>
        <div class="chat-input-row">
          <el-input
            v-model="followUp"
            placeholder="输入修改要求，如：把第 2 条改得更简练、补充 Redis 性能指标…（回车发送）"
            :disabled="loading"
            @keyup.enter="sendFollowUp"
          />
          <el-button type="primary" plain :loading="loading" :disabled="!followUp.trim()" @click="sendFollowUp">
            发送微调
          </el-button>
        </div>
        <div class="tag-bar compact">
          <span class="tag-label">快捷追问：</span>
          <button
            v-for="tag in DIRECTION_TAGS"
            :key="tag.label"
            type="button"
            class="dir-chip"
            :disabled="loading"
            @click="quickFollowUp(tag.text)"
          >
            {{ tag.label }}
          </button>
        </div>
      </div>
    </template>

    <template #footer>
      <template v-if="state === 'idle'">
        <el-button @click="$emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="loading" @click="runInitial">开始润色</el-button>
      </template>
      <template v-else>
        <el-button :disabled="loading" @click="state = 'idle'">返回重设方向</el-button>
        <el-button :disabled="loading" @click="runInitial">重新生成</el-button>
        <el-button type="primary" :disabled="loading" @click="apply">
          应用到简历（采纳 {{ selectedCount }} / {{ pairs.length }} 条）
        </el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api/resume'
import { DIRECTION_TAGS, appendDirection, buildPointPairs, mergeSelectedPoints } from '../utils/aiDiff'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  // { kind, sectionType, sectionTitle, title, subtitle, content }，content 为原文 HTML
  target: { type: Object, default: null }
})
const emit = defineEmits(['update:modelValue', 'apply'])

const state = ref('idle') // idle | done
const targetRole = ref('')
const instruction = ref('')
const followUp = ref('')
const result = ref('')
const pairs = ref([])
const chatHistory = ref([])
const loading = ref(false)

const selectedCount = computed(() => pairs.value.filter((p) => p.selected && p.ai).length)

function selectAll(val) {
  pairs.value.forEach((p) => {
    p.selected = val ? Boolean(p.ai) : false
  })
}

function requestPolish({ inst, currentHtml, historyItem }) {
  loading.value = true
  localStorage.setItem('aiTargetRole', targetRole.value.trim())
  api
    .aiPolish({
      kind: props.target.kind,
      sectionType: props.target.sectionType,
      sectionTitle: props.target.sectionTitle,
      title: props.target.title,
      subtitle: props.target.subtitle,
      targetRole: targetRole.value.trim(),
      content: props.target.content,
      instruction: inst,
      currentHtml: currentHtml || ''
    })
    .then((r) => {
      result.value = r.html
      pairs.value = buildPointPairs(props.target.content, r.html)
      if (historyItem) {
        chatHistory.value.push(historyItem)
      }
      state.value = 'done'
    })
    .catch(() => {})
    .finally(() => (loading.value = false))
}

function runInitial() {
  chatHistory.value = []
  const inst = instruction.value.trim()
  requestPolish({
    inst,
    currentHtml: '',
    historyItem: inst ? { type: 'init', text: inst } : null
  })
}

function sendFollowUp() {
  const text = followUp.value.trim()
  if (!text || loading.value) return
  // 以当前已勾选合并的 HTML 作为上下文基础，让 AI 按用户追问继续微调
  const baseHtml = mergeSelectedPoints(pairs.value, 'item-desc') || result.value
  followUp.value = ''
  requestPolish({
    inst: text,
    currentHtml: baseHtml,
    historyItem: { type: 'followup', text }
  })
}

function quickFollowUp(text) {
  followUp.value = appendDirection(followUp.value, text)
}

function apply() {
  const mergedHtml = mergeSelectedPoints(pairs.value, 'item-desc')
  if (!mergedHtml) {
    ElMessage.warning('合并后内容为空，请至少保留一条内容')
    return
  }
  emit('apply', mergedHtml)
  emit('update:modelValue', false)
  ElMessage.success(`已应用所选修改（采纳 ${selectedCount.value} 条 AI 优化，其余保留原文）`)
}

// 每次打开重置到第一步
function open() {
  targetRole.value = localStorage.getItem('aiTargetRole') || ''
  instruction.value = ''
  followUp.value = ''
  state.value = 'idle'
  result.value = ''
  pairs.value = []
  chatHistory.value = []
}
defineExpose({ open })
</script>

<style scoped>
.field { margin-bottom: 14px; }
.field label { display: block; font-size: 12.5px; color: #475569; margin-bottom: 5px; font-weight: 500; }
.tag-bar { display: flex; flex-wrap: wrap; align-items: center; gap: 6px; margin-top: 7px; }
.tag-bar.compact { margin-top: 8px; }
.tag-label { font-size: 12px; color: #94a3b8; }
.dir-chip {
  border: 1px solid #cbd5e1; background: #f8fafc; color: #334155;
  border-radius: 999px; padding: 2px 10px; font-size: 12px; cursor: pointer;
  transition: all 0.15s;
}
.dir-chip:hover:not(:disabled) { border-color: #3b82f6; color: #2563eb; background: #eff6ff; }
.dir-chip:disabled { opacity: 0.5; cursor: not-allowed; }

.orig-preview { margin-top: 4px; margin-bottom: 10px; }
.orig-preview-title { font-size: 12px; color: #64748b; margin-bottom: 5px; }
.content-box {
  border: 1px solid #e2e8f0; border-radius: 8px; padding: 10px 14px;
  font-size: 13px; line-height: 1.65; color: #1f2937;
  max-height: 180px; overflow: auto; background: #fafbfc;
}
.content-box :deep(ul), .content-box :deep(ol) { margin: 0; padding-left: 20px; }
.content-box :deep(li) { margin: 2px 0; }
.tip { font-size: 12px; color: #94a3b8; }

/* 比对报告 */
.report-toolbar {
  display: flex; justify-content: space-between; align-items: center;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
  padding: 8px 12px; margin-bottom: 10px;
}
.report-summary { font-size: 13px; color: #334155; }
.report-summary .accent { color: #2563eb; }
.report-summary .sub-hint { font-size: 12px; color: #94a3b8; }

.diff-header {
  display: flex; gap: 10px; padding: 0 6px 6px;
  font-size: 12px; font-weight: 600; color: #64748b;
}
.diff-header .ai { color: #2563eb; }
.col-check { width: 54px; flex-shrink: 0; display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 2px; }
.row-num { font-size: 11px; color: #94a3b8; }
.col-pane { flex: 1; min-width: 0; }

.diff-list {
  display: flex; flex-direction: column; gap: 8px;
  max-height: 310px; overflow-y: auto; padding-right: 4px;
}
.diff-row {
  flex-shrink: 0;
  display: flex; gap: 10px; align-items: stretch;
}
.pane-box {
  position: relative; border: 1px solid #e2e8f0; border-radius: 8px;
  padding: 10px 12px; font-size: 13px; line-height: 1.65; color: #1f2937;
  background: #fafbfc; cursor: pointer; transition: all 0.15s;
}
.pane-box.keep-orig {
  border-color: #64748b; background: #f1f5f9;
}
.pane-box.ai-pane.use-ai {
  background: #eff6ff; border-color: #60a5fa;
}
.pane-box.ai-pane.dimmed {
  opacity: 0.55; background: #f8fafc;
}
.pane-badge {
  display: inline-block; font-size: 11px; padding: 1px 6px; border-radius: 4px;
  background: #e2e8f0; color: #475569; margin-bottom: 4px; font-weight: 600;
}
.pane-badge.ai {
  background: #dbeafe; color: #1d4ed8;
}
.empty-line { color: #94a3b8; font-size: 12px; font-style: italic; }

/* 对话式微调 */
.chat-box {
  margin-top: 14px; border: 1px solid #dbeafe; background: #f8fbff;
  border-radius: 10px; padding: 10px 12px;
}
.chat-header {
  display: flex; align-items: center; gap: 6px;
  font-size: 12.5px; font-weight: 600; color: #1e40af; margin-bottom: 8px;
}
.chat-history {
  display: flex; flex-direction: column; gap: 4px;
  max-height: 80px; overflow-y: auto; margin-bottom: 8px;
}
.chat-item {
  font-size: 12px; color: #334155; display: flex; align-items: baseline; gap: 6px;
}
.chat-tag {
  background: #e0e7ff; color: #3730a3; padding: 1px 6px;
  border-radius: 4px; font-size: 11px; flex-shrink: 0;
}
.chat-text { color: #475569; }
.chat-input-row { display: flex; gap: 8px; }
</style>
