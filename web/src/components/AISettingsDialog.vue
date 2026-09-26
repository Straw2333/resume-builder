<template>
  <el-dialog
    :model-value="modelValue"
    title="AI 优化设置"
    width="560px"
    :close-on-click-modal="false"
    @update:model-value="$emit('update:modelValue', $event)"
    @open="load"
  >
    <div class="ai-form">
      <div class="field">
        <label>服务商预设（自动填好地址和模型，换 Key 即可）</label>
        <el-select v-model="preset" @change="applyPreset" class="grow">
          <el-option v-for="p in presets" :key="p.v" :value="p.v" :label="p.l" />
        </el-select>
      </div>

      <div class="field">
        <label>接口协议</label>
        <el-radio-group v-model="form.provider">
          <el-radio-button value="openai">OpenAI 兼容</el-radio-button>
          <el-radio-button value="anthropic">Claude（Anthropic）</el-radio-button>
        </el-radio-group>
      </div>

      <div class="field">
        <label>API 地址</label>
        <el-input v-model="form.baseURL" placeholder="如：https://api.openai.com/v1" />
        <div class="tip">OpenAI 兼容填到 /v1 或 /v4 为止；Claude 官方填 https://api.anthropic.com，用中转站则填中转地址</div>
      </div>

      <div class="field">
        <label>模型名称</label>
        <el-input v-model="form.model" placeholder="如：gpt-4o-mini / deepseek-chat / glm-4-flash" />
      </div>

      <div class="field">
        <label>API Key</label>
        <el-input
          v-model="form.apiKey" type="password" show-password
          :placeholder="hasKey ? `已保存（尾号 ${keyTail}），留空表示不修改` : '粘贴服务商提供的 API Key'"
        />
        <div class="tip">Key 仅保存在本机数据库，不会上传到任何第三方</div>
      </div>
    </div>

    <template #footer>
      <el-button @click="test" :loading="testing">测试连接</el-button>
      <el-button @click="$emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" :loading="saving" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import api from '../api/resume'

defineProps({ modelValue: { type: Boolean, default: false } })
const emit = defineEmits(['update:modelValue', 'saved'])

// 预设只负责填表单，协议跟随所选服务商
const presets = [
  { v: 'openai', l: 'OpenAI', provider: 'openai', baseURL: 'https://api.openai.com/v1', model: 'gpt-4o-mini' },
  { v: 'deepseek', l: 'DeepSeek', provider: 'openai', baseURL: 'https://api.deepseek.com/v1', model: 'deepseek-chat' },
  { v: 'zhipu', l: '智谱 GLM', provider: 'openai', baseURL: 'https://open.bigmodel.cn/api/paas/v4', model: 'glm-4-flash' },
  { v: 'anthropic', l: 'Claude（Anthropic）', provider: 'anthropic', baseURL: 'https://api.anthropic.com', model: 'claude-sonnet-4-5' },
  { v: 'custom', l: '自定义 / 中转站', provider: '', baseURL: '', model: '' }
]

const preset = ref('openai')
const form = reactive({ provider: 'openai', baseURL: '', model: '', apiKey: '' })
const hasKey = ref(false)
const keyTail = ref('')
const testing = ref(false)
const saving = ref(false)

function applyPreset(v) {
  const p = presets.find((p) => p.v === v)
  if (!p || v === 'custom') return
  form.provider = p.provider
  form.baseURL = p.baseURL
  form.model = p.model
}

async function load() {
  const s = await api.aiGetSettings()
  form.provider = s.provider || 'openai'
  form.baseURL = s.baseURL || ''
  form.model = s.model || ''
  form.apiKey = ''
  hasKey.value = !!s.hasKey
  keyTail.value = s.keyTail || ''
  // 根据已存配置反推预设；对不上则视为自定义
  preset.value = presets.find((p) => p.v !== 'custom' && p.provider === form.provider && p.baseURL === form.baseURL && p.model === form.model)?.v || 'custom'
}

function test() {
  testing.value = true
  api
    .aiTest({ ...form })
    .then((r) => ElMessage.success(r.message))
    .catch(() => {})
    .finally(() => (testing.value = false))
}

function save() {
  if (!form.provider || !form.baseURL || !form.model) {
    ElMessage.warning('请先选择服务商预设并补全地址和模型名')
    return
  }
  if (!form.apiKey && !hasKey.value) {
    ElMessage.warning('请填写 API Key')
    return
  }
  saving.value = true
  api
    .aiSaveSettings({ ...form })
    .then(() => {
      ElMessage.success('AI 设置已保存')
      emit('saved')
      emit('update:modelValue', false)
    })
    .catch(() => {})
    .finally(() => (saving.value = false))
}
</script>

<style scoped>
.ai-form { display: flex; flex-direction: column; }
.field { margin-bottom: 14px; }
.field label { display: block; font-size: 12px; color: #64748b; margin-bottom: 4px; }
.grow { width: 100%; }
.tip { font-size: 12px; color: #94a3b8; margin-top: 4px; line-height: 1.5; }
</style>
