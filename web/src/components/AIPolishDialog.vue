<template>
  <el-dialog
    :model-value="modelValue"
    :title="`AI 润色 · ${target?.sectionTitle || ''}`"
    width="720px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
  >
    <!-- 第一步：目标岗位（选填） -->
    <template v-if="state === 'idle'">
      <div class="field">
        <label>目标岗位（选填，让润色更贴合岗位要求）</label>
        <el-input
          v-model="targetRole" :placeholder="`如：${target?.sectionType === 'project' ? '后端开发工程师' : '前端开发实习生'}`"
          @keyup.enter="run"
        />
      </div>
      <div class="tip">AI 会保持原有事实不编造，仅优化表达；点击「开始润色」前可随时取消。</div>
    </template>

    <!-- 第二步：结果对照 -->
    <template v-else>
      <div class="compare">
        <div class="col">
          <div class="col-title">原文</div>
          <div class="content-box" v-html="target?.content || '<i>（空）</i>'"></div>
        </div>
        <div class="col">
          <div class="col-title ai">AI 优化后</div>
          <div class="content-box result" v-html="result"></div>
        </div>
      </div>
    </template>

    <template #footer>
      <template v-if="state === 'idle'">
        <el-button @click="$emit('update:modelValue', false)">取消</el-button>
        <el-button type="primary" :loading="loading" @click="run">开始润色</el-button>
      </template>
      <template v-else>
        <el-button :disabled="loading" @click="run">重新润色</el-button>
        <el-button type="primary" :disabled="loading" @click="apply">应用到简历</el-button>
      </template>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api/resume'

const props = defineProps({
  modelValue: { type: Boolean, default: false },
  // { kind, sectionType, sectionTitle, title, subtitle, content }，content 为原文 HTML
  target: { type: Object, default: null }
})
const emit = defineEmits(['update:modelValue', 'apply'])

const state = ref('idle') // idle | done
const targetRole = ref('')
const result = ref('')
const loading = ref(false)

function run() {
  loading.value = true
  result.value = ''
  localStorage.setItem('aiTargetRole', targetRole.value.trim())
  api
    .aiPolish({
      kind: props.target.kind,
      sectionType: props.target.sectionType,
      sectionTitle: props.target.sectionTitle,
      title: props.target.title,
      subtitle: props.target.subtitle,
      targetRole: targetRole.value.trim(),
      content: props.target.content
    })
    .then((r) => {
      result.value = r.html
      state.value = 'done'
    })
    .catch(() => {})
    .finally(() => (loading.value = false))
}

function apply() {
  emit('apply', result.value)
  emit('update:modelValue', false)
  ElMessage.success('已应用，内容将自动保存')
}

// 每次打开重置到第一步
function open() {
  targetRole.value = localStorage.getItem('aiTargetRole') || ''
  state.value = 'idle'
  result.value = ''
}
defineExpose({ open })
</script>

<style scoped>
.field { margin-bottom: 12px; }
.field label { display: block; font-size: 12px; color: #64748b; margin-bottom: 4px; }
.tip { font-size: 12px; color: #94a3b8; }
.compare { display: flex; gap: 14px; min-height: 220px; }
.col { flex: 1; min-width: 0; }
.col-title { font-size: 12.5px; color: #64748b; margin-bottom: 6px; font-weight: 600; }
.col-title.ai { color: #2563eb; }
.content-box {
  border: 1px solid #e2e8f0; border-radius: 10px; padding: 12px 14px;
  font-size: 13px; line-height: 1.7; color: #1f2937;
  max-height: 320px; overflow: auto; background: #fafbfc;
}
.content-box.result { background: #eff6ff; border-color: #bfdbfe; }
.content-box :deep(ul), .content-box :deep(ol) { margin: 0; padding-left: 20px; }
.content-box :deep(li) { margin: 2px 0; }
</style>
