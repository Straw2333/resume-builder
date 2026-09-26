<template>
  <el-dialog
    :model-value="modelValue"
    title="AI 整份简历优化"
    width="560px"
    :close-on-click-modal="false"
    :show-close="state !== 'running'"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <!-- 第一步：说明 + 目标岗位 -->
    <template v-if="state === 'idle'">
      <div class="intro">
        <p>将逐条优化简历中所有已填写的内容（共 <b>{{ tasks.length }}</b> 条）：</p>
        <ul>
          <li>开始前会自动备份当前内容，完成后可一键恢复</li>
          <li>每条优化后立即回填，约需 {{ estimate }}，请耐心等待</li>
          <li>单条失败会跳过并计入统计，不影响其余条目</li>
        </ul>
      </div>
      <div class="field" v-if="tasks.length">
        <label>目标岗位（选填，让优化更贴合岗位要求）</label>
        <el-input v-model="targetRole" placeholder="如：Go 后端开发工程师" />
      </div>
      <div v-else class="empty-tip">当前简历还没有已填写的内容（描述 / 自我评价），无需优化。</div>
    </template>

    <!-- 第二步：进度 -->
    <template v-else-if="state === 'running'">
      <div class="progress-text">
        正在优化第 {{ done + 1 }} / {{ tasks.length }} 条：{{ currentLabel }}
      </div>
      <el-progress :percentage="percent" :stroke-width="10" />
      <div class="progress-tip">优化过程中请勿关闭此窗口；点击右上角关闭可中止，已完成的部分保留。</div>
    </template>

    <!-- 第三步：完成统计 -->
    <template v-else>
      <div class="result-summary">
        <p class="ok">✅ 优化完成：成功 {{ okCount }} 条<template v-if="failCount">，失败 {{ failCount }} 条</template>。</p>
        <p class="tip">结果已回填到编辑面板，将自动保存。若不满意，可恢复到优化前的内容。</p>
      </div>
    </template>

    <template #footer>
      <template v-if="state === 'idle'">
        <el-button @click="$emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :disabled="!tasks.length" @click="run">开始优化</el-button>
      </template>
      <template v-else-if="state === 'running'">
        <el-button disabled>优化中…</el-button>
      </template>
      <template v-else>
        <el-button @click="restore">恢复优化前内容</el-button>
        <el-button type="primary" @click="$emit('update:modelValue', false)">完成</el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup>
import { computed, onBeforeUnmount, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api/resume'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  // [{ key, kind, sectionType, sectionTitle, title, subtitle, content }]
  tasks: { type: Array, default: () => [] }
})
const emit = defineEmits(['update:modelValue', 'item-done', 'restore'])

const state = ref('idle') // idle | running | done
const targetRole = ref('')
const done = ref(0)
const okCount = ref(0)
const failCount = ref(0)
const currentLabel = ref('')
const stopped = ref(false)

const percent = computed(() =>
  props.tasks.length ? Math.round((done.value / props.tasks.length) * 100) : 0
)
const estimate = computed(() => {
  const s = Math.max(props.tasks.length * 8, 15)
  return s > 60 ? `约 ${Math.round(s / 60)} 分钟` : `约 ${s} 秒`
})

async function run() {
  localStorage.setItem('aiTargetRole', targetRole.value.trim())
  state.value = 'running'
  done.value = 0
  okCount.value = 0
  failCount.value = 0
  stopped.value = false
  for (const task of props.tasks) {
    currentLabel.value = task.title || task.sectionTitle
    try {
      const r = await api.aiPolish({
        kind: task.kind,
        sectionType: task.sectionType,
        sectionTitle: task.sectionTitle,
        title: task.title,
        subtitle: task.subtitle,
        targetRole: targetRole.value.trim(),
        content: task.content
      })
      emit('item-done', task, r.html)
      okCount.value++
    } catch {
      failCount.value++
    }
    done.value++
  }
  state.value = 'done'
  if (stopped.value) emit('update:modelValue', false)
  else ElMessage.success(`整份优化完成：成功 ${okCount.value} 条`)
}

// 关闭弹窗即中止：置标记，当前请求完成后停止循环
function onRequestClose() {
  if (state.value === 'running') stopped.value = true
}

function restore() {
  emit('restore')
  emit('update:modelValue', false)
}

onBeforeUnmount(onRequestClose)

defineExpose({ open: () => (state.value = 'idle') })
</script>

<style scoped>
.intro { font-size: 13.5px; color: #334155; line-height: 1.7; }
.intro ul { margin: 6px 0 0; padding-left: 20px; color: #64748b; font-size: 13px; }
.field { margin-bottom: 12px; margin-top: 14px; }
.field label { display: block; font-size: 12px; color: #64748b; margin-bottom: 4px; }
.empty-tip { font-size: 13.5px; color: #94a3b8; }
.progress-text { font-size: 14px; color: #0f172a; margin-bottom: 14px; }
.progress-tip { font-size: 12px; color: #94a3b8; margin-top: 14px; }
.result-summary .ok { font-size: 14.5px; color: #16a34a; font-weight: 600; }
.result-summary .tip { font-size: 13px; color: #64748b; margin-top: 6px; }
</style>
