<template>
  <el-dialog
    :model-value="modelValue"
    :title="state === 'done' ? 'AI 根据 JD 生成简历 · 预览与微调' : 'AI 根据 JD 自动生成简历'"
    :width="state === 'done' ? '920px' : '680px'"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <!-- 第一步：粘贴 JD + 选填个人背景 -->
    <template v-if="state === 'idle'">
      <div class="intro-banner">
        将招聘网站上的<b>岗位职责、任职要求（JD）</b>粘贴到下方，AI 将自动分析岗位核心能力模型与技术栈关键词，为您生成高匹配度的结构化简历草稿。生成后可预览、勾选模块或通过对话继续微调。
      </div>

      <div class="field">
        <label>招聘岗位 JD 内容 <span class="req">*必填</span></label>
        <el-input
          v-model="jd"
          type="textarea"
          :rows="7"
          placeholder="请粘贴招聘信息中的岗位名称、岗位职责、任职要求等内容…&#10;例如：&#10;岗位：Go 后端开发工程师&#10;职责：1. 负责核心交易系统研发与高并发架构演进；2. 参与微服务治理与性能优化…&#10;要求：熟练掌握 Go、MySQL、Redis、Kafka，具备分布式系统开发经验…"
        />
      </div>

      <div class="field">
        <label>个人真实背景 / 定制要求（选填，让生成的简历更贴合你的实际情况）</label>
        <el-input
          v-model="userBg"
          type="textarea"
          :rows="3"
          placeholder="选填：可简述你的学校专业、真实掌握的技术栈、做过的项目或实习名称（如：25届计算机本科，熟悉Go/Gin/Redis，做过秒杀系统项目）；留空则完全根据 JD 生成标准范本。"
        />
        <div class="tag-bar">
          <span class="tag-label">快捷设定：</span>
          <button
            v-for="tag in bgTags"
            :key="tag.label"
            type="button"
            class="dir-chip"
            @click="userBg = appendDirection(userBg, tag.text)"
          >
            + {{ tag.label }}
          </button>
        </div>
      </div>
    </template>

    <!-- 第二步：结构化预览 + 模块/要点勾选 + 对话追问微调 -->
    <template v-else>
      <div class="meta-bar" v-loading="loading">
        <div class="meta-field">
          <label>简历名称</label>
          <el-input v-model="draft.title" size="small" placeholder="简历名称" />
        </div>
        <div class="meta-field">
          <label>求职意向（自动提取自 JD）</label>
          <el-input v-model="draft.intent" size="small" placeholder="目标岗位" />
        </div>
        <div class="meta-ops">
          <el-button size="small" text type="primary" @click="toggleAllSections(true)">全选模块</el-button>
          <el-button size="small" text @click="toggleAllSections(false)">全部取消</el-button>
        </div>
      </div>

      <div class="preview-tip">
        💡 提示：请勾选需要保留的模块与要点（<span v-if="mode === 'fill'">未勾选的模块将保留您当前简历中的原有内容，不会覆盖；</span>点击确认后可在编辑器中根据个人真实情况进一步修改细节）。
      </div>

      <div class="section-list" v-loading="loading">
        <div
          v-for="sec in activeSections"
          :key="sec.type + '-' + sec.title"
          class="sec-card"
          :class="{ 'is-unchecked': !sec.selected }"
        >
          <div class="sec-header" @click="sec.expanded = !sec.expanded">
            <div class="sec-left" @click.stop>
              <el-checkbox
                v-model="sec.selected"
                @change="(val) => onSectionCheckChange(sec, val)"
              />
              <span class="sec-badge">{{ sec.title }}</span>
              <span class="sec-summary">
                <template v-if="sec.type === 'evaluation'">
                  共 {{ sec.evalPoints.length }} 条评价（已选 {{ sec.evalPoints.filter((p) => p.selected).length }} 条）
                </template>
                <template v-else>
                  共 {{ sec.items.length }} 段经历/分类（已选 {{ selectedPointsInSec(sec) }} 条要点）
                </template>
              </span>
            </div>
            <div class="sec-right">
              <span class="adopt-tag" :class="{ on: sec.selected }">
                {{ sec.selected ? '已勾选采纳' : (mode === 'fill' ? '不覆盖此模块' : '不启用此模块') }}
              </span>
              <el-icon class="expand-icon" :class="{ open: sec.expanded }">
                <ArrowDown />
              </el-icon>
            </div>
          </div>

          <div v-show="sec.expanded" class="sec-body">
            <!-- 自我评价模块 -->
            <template v-if="sec.type === 'evaluation'">
              <div
                v-for="(pt, pIdx) in sec.evalPoints"
                :key="pIdx"
                class="point-row"
                :class="{ off: !pt.selected || !sec.selected }"
                @click="pt.selected = !pt.selected; syncSecCheck(sec)"
              >
                <el-checkbox
                  v-model="pt.selected"
                  @click.stop
                  @change="syncSecCheck(sec)"
                />
                <span class="pt-text">{{ pt.text }}</span>
              </div>
            </template>

            <!-- 列表型模块（教育、技能、实习、项目等） -->
            <template v-else>
              <div
                v-for="(it, iIdx) in sec.items"
                :key="iIdx"
                class="item-card"
                :class="{ off: !it.selected || !sec.selected }"
              >
                <div class="item-head">
                  <div class="item-head-left">
                    <el-checkbox
                      v-model="it.selected"
                      @change="(val) => onItemCheckChange(sec, it, val)"
                    />
                    <b class="it-title">{{ it.title }}</b>
                    <span v-if="it.subtitle" class="it-sub">· {{ it.subtitle }}</span>
                  </div>
                  <div class="item-head-right">
                    <span v-if="it.time" class="it-time">{{ it.time }}</span>
                  </div>
                </div>
                <div v-if="it.tags && it.tags.length" class="item-tags">
                  <span v-for="t in it.tags" :key="t" class="mini-tag">{{ t }}</span>
                </div>
                <div class="item-points">
                  <div
                    v-for="(pt, pIdx) in it.points"
                    :key="pIdx"
                    class="point-row"
                    :class="{ off: !pt.selected || !it.selected || !sec.selected }"
                    @click="pt.selected = !pt.selected; syncItemAndSecCheck(sec, it)"
                  >
                    <el-checkbox
                      v-model="pt.selected"
                      @click.stop
                      @change="syncItemAndSecCheck(sec, it)"
                    />
                    <div class="pt-html" v-html="pt.html"></div>
                  </div>
                </div>
              </div>
            </template>
          </div>
        </div>
      </div>

      <!-- 对话微调窗口 -->
      <div class="chat-box">
        <div class="chat-header">
          <el-icon><ChatDotRound /></el-icon>
          <span>对话式微调（想补充某段经历、调整技术栈或修改风格？直接告诉 AI）</span>
        </div>
        <div v-if="chatHistory.length" class="chat-history">
          <div v-for="(h, idx) in chatHistory" :key="idx" class="chat-item">
            <span class="chat-tag">微调 #{{ idx + 1 }}</span>
            <span>{{ h }}</span>
          </div>
        </div>
        <div class="chat-input-row">
          <el-input
            v-model="followUp"
            placeholder="输入修改要求，如：把第1个项目换成微服务网关相关、专业技能加上 Kubernetes…（回车发送）"
            :disabled="loading"
            @keyup.enter="sendFollowUp"
          />
          <el-button
            type="primary"
            plain
            :loading="loading"
            :disabled="!followUp.trim()"
            @click="sendFollowUp"
          >
            发送微调
          </el-button>
        </div>
        <div class="tag-bar compact">
          <span class="tag-label">快捷调整：</span>
          <button
            v-for="tag in refineTags"
            :key="tag.label"
            type="button"
            class="dir-chip"
            :disabled="loading"
            @click="followUp = appendDirection(followUp, tag.text)"
          >
            {{ tag.label }}
          </button>
        </div>
      </div>
    </template>

    <template #footer>
      <template v-if="state === 'idle'">
        <el-button @click="$emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="loading" :disabled="!jd.trim()" @click="runGenerate">
          <el-icon><MagicStick /></el-icon>&nbsp;开始生成简历
        </el-button>
      </template>
      <template v-else>
        <el-button :disabled="loading" @click="state = 'idle'">返回修改 JD</el-button>
        <el-button :disabled="loading" @click="runGenerate">重新生成</el-button>
        <el-button
          type="primary"
          :disabled="loading || selectedSectionCount === 0"
          @click="confirmApply"
        >
          {{ mode === 'create' ? `创建简历并进入编辑（已选 ${selectedSectionCount} 个模块）` : `应用所选模块到简历（已选 ${selectedSectionCount} 个模块）` }}
        </el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api/resume'
import { appendDirection, parseHtmlPoints } from '../utils/aiDiff'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  // create: 首页新建简历并跳转 | fill: 编辑页填充到当前简历
  mode: { type: String, default: 'create' }
})
const emit = defineEmits(['update:modelValue', 'apply'])

const state = ref('idle') // idle | done
const jd = ref('')
const userBg = ref('')
const followUp = ref('')
const loading = ref(false)
const chatHistory = ref([])

const draft = reactive({
  title: '',
  intent: '',
  sections: []
})

const bgTags = [
  { label: '应届校招风格', text: '我是应届毕业生，请侧重扎实的基础课程、实习潜力与项目技术深度' },
  { label: '社招经验风格', text: '我是社招求职者，请侧重核心系统架构设计、高可用保障与业务量化产出' },
  { label: '突出高并发与性能', text: '重点凸显高并发场景、缓存与数据库调优及性能提升指标' },
  { label: '标注待替换占位数据', text: '对我未提供的具体学校、公司名及量化指标，请用【待替换：...】标注方便我微调' }
]

const refineTags = [
  { label: '再加一段项目经历', text: '请再补充一段贴合 JD 核心技术栈的高含金量项目经历' },
  { label: '强化量化成果指标', text: '请为实习和项目经历补充更具体的性能提升与业务成果量化数据' },
  { label: '精简各模块篇幅', text: '请将各模块描述精简干练一些，方便控制在一页 A4 纸内' },
  { label: '扩充专业技能模块', text: '请根据 JD 要求进一步细化和扩充专业技能模块的要点' }
]

// 仅展示有生成内容的模块供用户预览和勾选
const activeSections = computed(() =>
  draft.sections.filter((s) =>
    s.type === 'evaluation' ? s.evalPoints.length > 0 : s.items.length > 0
  )
)

const selectedSectionCount = computed(() =>
  activeSections.value.filter((s) => s.selected).length
)

function selectedPointsInSec(sec) {
  return (sec.items || []).reduce(
    (acc, it) => acc + (it.selected ? it.points.filter((p) => p.selected).length : 0),
    0
  )
}

function hydrateDraft(res) {
  draft.title = res.title || '目标岗位简历'
  draft.intent = res.intent || ''
  draft.sections = (res.sections || []).map((s) => {
    if (s.type === 'evaluation') {
      const lines = String(s.content || '')
        .split(/\r?\n/)
        .map((l) => l.trim())
        .filter(Boolean)
      return reactive({
        type: s.type,
        title: s.title,
        selected: lines.length > 0 && s.visible !== false,
        expanded: true,
        evalPoints: lines.map((text) => ({ text, selected: true })),
        items: []
      })
    }
    const items = (s.items || []).map((it) => {
      const pts = parseHtmlPoints(it.desc)
      return {
        title: it.title || '',
        subtitle: it.subtitle || '',
        time: it.time || '',
        tags: it.tags || [],
        selected: true,
        points: pts.map((html) => ({ html, selected: true }))
      }
    })
    return reactive({
      type: s.type,
      title: s.title,
      selected: items.length > 0 && s.visible !== false,
      expanded: true,
      evalPoints: [],
      items
    })
  })
}

// 将当前勾选状态序列化为结构化对象（供多轮微调上下文或最终保存）
function buildStructuredSections(onlySelected = false) {
  return draft.sections.map((s) => {
    if (s.type === 'evaluation') {
      const chosenLines = (s.evalPoints || [])
        .filter((p) => (!onlySelected || (s.selected && p.selected)) && p.text.trim())
        .map((p) => p.text.trim())
      return {
        type: s.type,
        title: s.title,
        visible: s.selected && chosenLines.length > 0,
        content: chosenLines.join('\n'),
        items: []
      }
    }
    const chosenItems = []
    for (const it of s.items || []) {
      if (onlySelected && (!s.selected || !it.selected)) continue
      const pts = (it.points || [])
        .filter((p) => (!onlySelected || p.selected) && p.html.trim())
        .map((p) => `<li>${p.html.trim()}</li>`)
      if (pts.length === 0 && !it.title) continue
      chosenItems.push({
        title: it.title,
        subtitle: it.subtitle,
        time: it.time,
        desc: pts.length ? `<ul>${pts.join('')}</ul>` : '',
        tags: it.tags || []
      })
    }
    return {
      type: s.type,
      title: s.title,
      visible: s.selected && chosenItems.length > 0,
      content: '',
      items: chosenItems
    }
  })
}

function toggleAllSections(val) {
  activeSections.value.forEach((sec) => {
    sec.selected = val
    onSectionCheckChange(sec, val)
  })
}

function onSectionCheckChange(sec, val) {
  if (sec.type === 'evaluation') {
    sec.evalPoints.forEach((p) => (p.selected = val))
  } else {
    sec.items.forEach((it) => {
      it.selected = val
      it.points.forEach((p) => (p.selected = val))
    })
  }
}

function onItemCheckChange(sec, it, val) {
  it.points.forEach((p) => (p.selected = val))
  syncSecCheck(sec)
}

function syncItemAndSecCheck(sec, it) {
  it.selected = it.points.some((p) => p.selected)
  syncSecCheck(sec)
}

function syncSecCheck(sec) {
  if (sec.type === 'evaluation') {
    sec.selected = sec.evalPoints.some((p) => p.selected)
  } else {
    sec.selected = sec.items.some((it) => it.selected && it.points.some((p) => p.selected))
  }
}

function runGenerate() {
  if (!jd.value.trim()) {
    ElMessage.warning('请先粘贴招聘岗位 JD 内容')
    return
  }
  loading.value = true
  chatHistory.value = []
  api
    .aiGenerateFromJD({
      jd: jd.value.trim(),
      userBg: userBg.value.trim()
    })
    .then((res) => {
      hydrateDraft(res)
      state.value = 'done'
    })
    .catch(() => {})
    .finally(() => (loading.value = false))
}

function sendFollowUp() {
  const text = followUp.value.trim()
  if (!text || loading.value) return
  const currentJson = JSON.stringify({
    title: draft.title,
    intent: draft.intent,
    sections: buildStructuredSections(false)
  })
  loading.value = true
  api
    .aiGenerateFromJD({
      jd: jd.value.trim(),
      userBg: userBg.value.trim(),
      instruction: text,
      currentJson
    })
    .then((res) => {
      hydrateDraft(res)
      chatHistory.value.push(text)
      followUp.value = ''
      ElMessage.success('已按您的要求更新简历草稿')
    })
    .catch(() => {})
    .finally(() => (loading.value = false))
}

function confirmApply() {
  const selectedTypes = new Set(
    activeSections.value.filter((s) => s.selected).map((s) => s.type)
  )
  const finalSections = buildStructuredSections(true)
  if (draft.intent.trim()) {
    localStorage.setItem('aiTargetRole', draft.intent.trim())
  }
  emit('apply', {
    title: draft.title.trim() || '我的简历',
    intent: draft.intent.trim(),
    sections: finalSections,
    selectedTypes
  })
  emit('update:modelValue', false)
}

function open() {
  state.value = 'idle'
  followUp.value = ''
  chatHistory.value = []
}

defineExpose({ open })
</script>

<style scoped>
.intro-banner {
  background: #f0f7ff; border: 1px solid #dbeafe; border-radius: 8px;
  padding: 10px 14px; font-size: 13px; color: #1e3a8a; line-height: 1.65;
  margin-bottom: 14px;
}
.field { margin-bottom: 14px; }
.field label { display: block; font-size: 12.5px; color: #475569; margin-bottom: 5px; font-weight: 500; }
.req { color: #dc2626; font-size: 12px; margin-left: 4px; }
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

/* 预览顶栏 */
.meta-bar {
  display: flex; align-items: flex-end; gap: 12px; flex-wrap: wrap;
  background: #f8fafc; border: 1px solid #e2e8f0; border-radius: 8px;
  padding: 10px 14px; margin-bottom: 8px;
}
.meta-field { flex: 1; min-width: 180px; }
.meta-field label { display: block; font-size: 11.5px; color: #64748b; margin-bottom: 4px; }
.meta-ops { display: flex; align-items: center; gap: 4px; }

.preview-tip { font-size: 12px; color: #64748b; margin-bottom: 10px; }

/* 模块卡片列表 */
.section-list {
  display: flex; flex-direction: column; gap: 10px;
  max-height: 52vh; overflow-y: auto; padding-right: 4px;
}
.sec-card {
  flex-shrink: 0;
  border: 1px solid #cbd5e1; border-radius: 10px; background: #fff; overflow: hidden;
  transition: opacity 0.15s;
}
.sec-card.is-unchecked { opacity: 0.65; border-color: #e2e8f0; }

.sec-header {
  display: flex; justify-content: space-between; align-items: center; gap: 12px;
  padding: 10px 14px; min-height: 44px; box-sizing: border-box;
  background: #f8fafc; cursor: pointer; user-select: none;
}
.sec-left { display: flex; align-items: center; gap: 8px; }
.sec-badge {
  font-size: 13px; font-weight: 700; color: #0f172a;
}
.sec-summary { font-size: 12px; color: #64748b; }
.sec-right { display: flex; align-items: center; gap: 10px; }
.adopt-tag {
  font-size: 11.5px; padding: 2px 8px; border-radius: 999px;
  background: #f1f5f9; color: #64748b;
}
.adopt-tag.on { background: #dbeafe; color: #1d4ed8; font-weight: 600; }
.expand-icon { color: #64748b; transition: transform 0.2s; }
.expand-icon.open { transform: rotate(180deg); }

.sec-body {
  padding: 10px 14px; border-top: 1px solid #f1f5f9;
  display: flex; flex-direction: column; gap: 10px;
}
.item-card {
  border: 1px solid #e2e8f0; border-radius: 8px; padding: 8px 12px; background: #fafbfc;
}
.item-card.off { opacity: 0.55; }
.item-head {
  display: flex; justify-content: space-between; align-items: center; gap: 8px;
  margin-bottom: 4px;
}
.item-head-left { display: flex; align-items: center; gap: 6px; }
.it-title { font-size: 13.5px; color: #0f172a; }
.it-sub { font-size: 12.5px; color: #475569; }
.it-time { font-size: 12px; color: #64748b; }
.item-tags { display: flex; flex-wrap: wrap; gap: 5px; margin: 4px 0 6px 22px; }
.mini-tag {
  font-size: 11px; padding: 1px 7px; border-radius: 4px;
  background: #eff6ff; color: #2563eb; border: 1px solid #bfdbfe;
}
.item-points { display: flex; flex-direction: column; gap: 4px; margin-left: 8px; }

.point-row {
  display: flex; align-items: flex-start; gap: 8px;
  padding: 4px 8px; border-radius: 6px; cursor: pointer;
  font-size: 13px; line-height: 1.6; color: #1f2937;
}
.point-row:hover { background: #f1f5f9; }
.point-row.off { opacity: 0.5; text-decoration: line-through; }
.pt-html :deep(b), .pt-html :deep(strong) { color: #1e40af; }

/* 对话式微调 */
.chat-box {
  margin-top: 12px; border: 1px solid #dbeafe; background: #f8fbff;
  border-radius: 10px; padding: 10px 12px;
}
.chat-header {
  display: flex; align-items: center; gap: 6px;
  font-size: 12.5px; font-weight: 600; color: #1e40af; margin-bottom: 8px;
}
.chat-history {
  display: flex; flex-direction: column; gap: 4px;
  max-height: 70px; overflow-y: auto; margin-bottom: 8px;
}
.chat-item { font-size: 12px; color: #334155; display: flex; align-items: baseline; gap: 6px; }
.chat-tag {
  background: #e0e7ff; color: #3730a3; padding: 1px 6px;
  border-radius: 4px; font-size: 11px; flex-shrink: 0;
}
.chat-input-row { display: flex; gap: 8px; }
</style>
