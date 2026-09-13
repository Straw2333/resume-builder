<template>
  <div class="rte" :class="{ focused }">
    <div class="rte-toolbar">
      <button type="button" class="tb" :class="{ on: st.bold }" title="加粗 (Ctrl+B)" @mousedown.prevent @click="cmd('bold')">
        <span class="glyph-b">B</span>
      </button>
      <button type="button" class="tb" :class="{ on: st.ul }" title="无序列表" @mousedown.prevent @click="cmd('insertUnorderedList')">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="8" y1="6" x2="20" y2="6"></line><line x1="8" y1="12" x2="20" y2="12"></line><line x1="8" y1="18" x2="20" y2="18"></line><line x1="3.5" y1="6" x2="3.6" y2="6"></line><line x1="3.5" y1="12" x2="3.6" y2="12"></line><line x1="3.5" y1="18" x2="3.6" y2="18"></line></svg>
      </button>
      <button type="button" class="tb" :class="{ on: st.ol }" title="有序列表" @mousedown.prevent @click="cmd('insertOrderedList')">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="10" y1="6" x2="20" y2="6"></line><line x1="10" y1="12" x2="20" y2="12"></line><line x1="10" y1="18" x2="20" y2="18"></line><path d="M4 6h1v4"></path><path d="M4 10h2"></path><path d="M6 18H4c0-1 2-2 2-3s-1-1.5-2-1"></path></svg>
      </button>
      <span class="tb-sep"></span>
      <button type="button" class="tb" title="减少缩进 (Shift+Tab)" @mousedown.prevent @click="cmd('outdent')">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="8 8 4 12 8 16"></polyline><line x1="21" y1="12" x2="13" y2="12"></line><line x1="21" y1="6" x2="13" y2="6"></line><line x1="21" y1="18" x2="13" y2="18"></line><line x1="4" y1="6" x2="11" y2="6"></line><line x1="4" y1="18" x2="11" y2="18"></line></svg>
      </button>
      <button type="button" class="tb" title="增加缩进 (Tab)" @mousedown.prevent @click="cmd('indent')">
        <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><polyline points="4 8 8 12 4 16"></polyline><line x1="21" y1="12" x2="13" y2="12"></line><line x1="21" y1="6" x2="13" y2="6"></line><line x1="21" y1="18" x2="13" y2="18"></line><line x1="11" y1="6" x2="4" y2="6"></line><line x1="11" y1="18" x2="4" y2="18"></line></svg>
      </button>
    </div>
    <div
      ref="body" class="rte-body" contenteditable="true" :data-placeholder="placeholder"
      @input="onInput" @keydown="onKeydown" @keyup="syncState" @mouseup="syncState"
      @focus="focused = true" @blur="onBlur" @paste="onPaste"
    ></div>
  </div>
</template>

<script setup>
import { ref, reactive, watch, onMounted } from 'vue'

const props = defineProps({
  modelValue: { type: String, default: '' },
  placeholder: { type: String, default: '' }
})
const emit = defineEmits(['update:modelValue'])
const body = ref(null)
const focused = ref(false)
const st = reactive({ bold: false, ul: false, ol: false })

const escapeHtml = (s) => s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;')
const isEmptyHTML = (h) => !h || h === '<br>' || h === '<div><br></div>' || h === '<ul></ul>' || h === '<ol></ol>'

onMounted(() => {
  // 旧的纯文本描述（\n 分行）转为无序列表，便于直接编辑
  let v = props.modelValue || ''
  if (v && !v.includes('<')) {
    const lines = v.split('\n').map((s) => s.trim()).filter(Boolean)
    if (lines.length > 0) v = '<ul>' + lines.map((l) => '<li>' + escapeHtml(l) + '</li>').join('') + '</ul>'
  }
  if (body.value) body.value.innerHTML = v
})

watch(() => props.modelValue, (v) => {
  const el = body.value
  if (!el) return
  const cur = el.innerHTML
  if ((v || '') !== cur && !el.contains(document.getSelection()?.anchorNode)) {
    el.innerHTML = v || ''
  }
})

function onInput() {
  const html = body.value?.innerHTML || ''
  emit('update:modelValue', isEmptyHTML(html) ? '' : html)
}

function onBlur() {
  focused.value = false
  const html = body.value?.innerHTML || ''
  if (isEmptyHTML(html)) {
    if (body.value) body.value.innerHTML = ''
    emit('update:modelValue', '')
  }
}

function cmd(name) {
  body.value?.focus()
  document.execCommand(name, false, null)
  syncState()
  onInput()
}

function onKeydown(e) {
  const mod = e.ctrlKey || e.metaKey
  if (mod && (e.key === 'b' || e.key === 'B')) {
    e.preventDefault()
    cmd('bold')
  } else if (e.key === 'Tab') {
    e.preventDefault()
    cmd(e.shiftKey ? 'outdent' : 'indent')
  } else if (mod && e.key === ']') {
    e.preventDefault()
    cmd('indent')
  } else if (mod && e.key === '[') {
    e.preventDefault()
    cmd('outdent')
  }
}

// 粘贴仅保留纯文本，避免外部样式污染
function onPaste(e) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text/plain') || ''
  document.execCommand('insertText', false, text)
}

function syncState() {
  try {
    st.bold = document.queryCommandState('bold')
    st.ul = document.queryCommandState('insertUnorderedList')
    st.ol = document.queryCommandState('insertOrderedList')
  } catch { /* 部分浏览器不支持 */ }
}
</script>

<style scoped>
.rte { border: 1px solid #dcdfe6; border-radius: 6px; background: #fff; transition: border-color .2s; }
.rte.focused, .rte:focus-within { border-color: #2563eb; }
.rte-toolbar {
  display: flex; align-items: center; gap: 2px;
  padding: 4px 6px; border-bottom: 1px solid #f0f2f5; background: #fafbfc;
  border-radius: 6px 6px 0 0;
}
.tb {
  width: 26px; height: 24px; display: inline-flex; align-items: center; justify-content: center;
  border: none; background: transparent; border-radius: 5px; cursor: pointer;
  color: #64748b; transition: .15s;
}
.tb:hover { background: #eef2f7; color: #0f172a; }
.tb.on { background: #dbeafe; color: #2563eb; }
.glyph-b { font-weight: 800; font-size: 13px; font-family: Georgia, serif; }
.tb-sep { width: 1px; height: 14px; background: #e2e8f0; margin: 0 3px; }
.rte-body {
  min-height: 76px; max-height: 260px; overflow-y: auto;
  padding: 8px 10px; font-size: 13px; line-height: 1.7; outline: none; color: #0f172a;
}
.rte-body:empty::before { content: attr(data-placeholder); color: #a8abb2; pointer-events: none; }
.rte-body :deep(ul), .rte-body :deep(ol) { margin: 2px 0; padding-left: 20px; }
</style>
