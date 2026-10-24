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
  if ((v || '') !== cur && !focused.value) {
    el.innerHTML = v || ''
  }
})

function onInput() {
  const html = body.value?.innerHTML || ''
  emit('update:modelValue', isEmptyHTML(html) ? '' : html)
}

function onBlur() {
  focused.value = false
  const el = body.value
  if (!el) return
  const cleaned = pruneBlankTail(el)
  const html = el.innerHTML || ''
  if (cleaned) emit('update:modelValue', isEmptyHTML(html) ? '' : html)
  else if (isEmptyHTML(html)) {
    if (el.innerHTML !== '') el.innerHTML = ''
    emit('update:modelValue', '')
  }
}

// 空节点判断：无可见文本（含全角空格/nbsp）且没有图片或子列表
function isBlankNode(n) {
  if (!n) return false
  if (n.querySelector && n.querySelector('img,ul,ol')) return false
  return (n.textContent || '').replace(/[\s\u00a0\u3000]/g, '') === ''
}

// 退出编辑时移除尾部空列表项 / 空段落，避免保存出空圆点、空行
function pruneBlankTail(el) {
  let cleaned = false
  for (;;) {
    const last = el.lastElementChild
    if (!last) break
    if (last.tagName === 'UL' || last.tagName === 'OL') {
      const li = last.lastElementChild
      if (li && li.tagName === 'LI' && isBlankNode(li)) {
        li.remove()
        if (!last.hasChildNodes()) last.remove()
        cleaned = true
        continue
      }
      break
    }
    if ((last.tagName === 'DIV' || last.tagName === 'P') && isBlankNode(last)) {
      last.remove()
      cleaned = true
      continue
    }
    break
  }
  return cleaned
}

function cmd(name) {
  body.value?.focus()
  document.execCommand(name, false, null)
  syncState()
  onInput()
}

function onKeydown(e) {
  if (e.isComposing || e.keyCode === 229) return // 中文输入法组词过程不处理
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
  } else if (e.key === ' ') {
    // 输入 "- " / "1. " 后按空格 → 自动转为列表；行内 "**重点**" 后按空格 → 加粗
    if (!mdListOnSpace(e)) mdBoldOnSpace()
  } else if (e.key === 'Enter') {
    // 整行输入了 "- 描述内容" 后回车 → 整行转列表
    mdListOnEnter(e)
  }
}

// ===== Markdown 快捷输入 =====
// 取光标所在行的信息：{ block: 当前行所在块, text: 行首到光标的文本 }；结构转换场景需列表外，allowList 用于列表内行内加粗
function lineInfo({ allowList = false } = {}) {
  const el = body.value
  if (!el) return null
  const sel = document.getSelection()
  if (!sel || !sel.rangeCount || !sel.getRangeAt(0).collapsed) return null
  const r = sel.getRangeAt(0)
  if (!el.contains(r.startContainer)) return null
  // 结构转换（- 变列表）只在列表外进行，避免重复转换；行内加粗允许发生在列表项内
  if (!allowList) {
    let p = r.startContainer.nodeType === 1 ? r.startContainer : r.startContainer.parentNode
    while (p && p !== el) {
      if (p.nodeType === 1 && (p.tagName === 'UL' || p.tagName === 'OL' || p.tagName === 'LI')) return null
      p = p.parentNode
    }
  }
  // 最近块级容器（首行文本可能直接挂在 contenteditable 上，无 div 包裹；列表项内则定位到 li）
  let block = r.startContainer
  while (block && block !== el) {
    if (block.nodeType === 1 && /^(DIV|P|H\d|LI)$/.test(block.tagName)) break
    block = block.parentNode
  }
  if (!block) block = el
  const line = document.createRange()
  line.selectNodeContents(block)
  line.setEnd(r.startContainer, r.startOffset)
  return { block, endC: r.startContainer, endO: r.startOffset, text: line.toString() }
}

// 删除 [block 行首, 光标) 的内容，然后执行列表命令
function convertLineTo(cmdName) {
  const info = lineInfo()
  if (!info) return
  const sel = document.getSelection()
  const del = document.createRange()
  del.selectNodeContents(info.block)
  del.setEnd(info.endC, info.endO)
  sel.removeAllRanges()
  sel.addRange(del)
  document.execCommand('delete')
  document.execCommand(cmdName, false, null)
  syncState()
  onInput()
}

// 空格触发：行首恰为 "-" / "*" / "+" → 无序列表；"1." / "1、" / "1)" → 有序列表
function mdListOnSpace(e) {
  const info = lineInfo()
  if (!info) return false
  let cmdName = ''
  if (/^[-*+]+$/.test(info.text)) cmdName = 'insertUnorderedList'
  else if (/^\d{1,3}[.、)]$/.test(info.text)) cmdName = 'insertOrderedList'
  if (!cmdName) return false
  e.preventDefault()
  convertLineTo(cmdName)
  return true
}

// 回车触发：整行 "- 描述" / "1. 描述" → 转列表，光标进入新列表项继续输入
function mdListOnEnter(e) {
  const info = lineInfo()
  if (!info) return false
  let m = info.text.match(/^[-*+][ \t]+([\s\S]*)$/)
  if (m) {
    e.preventDefault()
    convertLineTo('insertUnorderedList')
    fillAndContinue(m[1])
    return true
  }
  m = info.text.match(/^\d{1,3}[.、)][ \t]*([\s\S]*)$/)
  if (m) {
    e.preventDefault()
    convertLineTo('insertOrderedList')
    fillAndContinue(m[1])
    return true
  }
  return false
}

function fillAndContinue(rest) {
  rest = rest.trim()
  if (rest) {
    document.execCommand('insertText', false, rest)
    document.execCommand('insertParagraph')
  }
  syncState()
  onInput()
}

// 空格触发：列表内/外的行内 "**内容**" → <b>内容</b>（不拦截空格，转换后空格正常输入）
function mdBoldOnSpace() {
  const info = lineInfo({ allowList: true })
  if (!info) return false
  const m = info.text.match(/\*\*([^*]+)\*\*$/)
  if (!m) return false
  const from = info.text.length - m[0].length
  const to = info.text.length
  // 行内文本位置 → DOM 文本节点位置
  const walker = document.createTreeWalker(info.block, NodeFilter.SHOW_TEXT)
  const segs = []
  let acc = 0
  let n
  while ((n = walker.nextNode())) {
    segs.push({ node: n, start: acc, end: acc + n.data.length })
    acc += n.data.length
  }
  const pos = (off) => {
    for (const s of segs) {
      if (off < s.end) return { node: s.node, offset: off - s.start }
    }
    const last = segs[segs.length - 1]
    return { node: last.node, offset: last.node.data.length }
  }
  const sel = document.getSelection()
  const del = document.createRange()
  const a = pos(from)
  const b = pos(to)
  del.setStart(a.node, a.offset)
  del.setEnd(b.node, b.offset)
  sel.removeAllRanges()
  sel.addRange(del)
  document.execCommand('delete')
  const bold = document.createElement('b')
  bold.textContent = m[1]
  sel.getRangeAt(0).insertNode(bold)
  const after = document.createRange()
  after.setStartAfter(bold)
  after.collapse(true)
  sel.removeAllRanges()
  sel.addRange(after)
  syncState()
  onInput()
  return true
}

// 粘贴仅保留纯文本；若带 Markdown 标记（- / 1. / **加粗**）则直接结构化成富文本
function onPaste(e) {
  e.preventDefault()
  const text = e.clipboardData?.getData('text/plain') || ''
  if (!text) return
  const html = pasteMarkdownToHtml(text)
  if (html) {
    document.execCommand('insertHTML', false, html)
  } else {
    document.execCommand('insertText', false, text)
  }
  syncState()
  onInput()
}

// 行内 **加粗** / *斜体* 转换（先转义再替换，避免标签注入）
function mdInlineHtml(s) {
  let out = escapeHtml(s)
  out = out.replace(/\*\*([^*]+)\*\*/g, (_, x) => '<b>' + x + '</b>')
  // 单星对：内部不能以空格开头/结尾，尽量避开 "a * b * c" 这类误判
  out = out.replace(/(^|[^*])\*([^*\s](?:[^*\s]|.*?[^*\s])?)\*(?![\s*]*\*)/g, (mm, pre, inner) => pre + '<i>' + inner + '</i>')
  return out
}

// 解析按行的 Markdown：返回 [{ level, ordered, html, structured }]
function parseMarkdownLines(text) {
  const items = []
  let structured = false
  for (const raw of text.split(/\r\n|\r|\n/)) {
    const t = raw.replace(/\t/g, '  ')
    const spaces = t.length - t.replace(/^ +/, '').length
    const level = Math.min(6, Math.floor(spaces / 2))
    let s = t.replace(/^ +/, '')
    let ordered = false
    if (/^[-*+][ \t]/.test(s)) {
      s = s.replace(/^[-*+][ \t]?/, '')
      structured = true
    } else {
      const mo = s.match(/^(\d{1,3})[.、)]/)
      if (mo) {
        ordered = true
        s = s.slice(mo[0].length).replace(/^[ \t]/, '')
        structured = true
      }
    }
    if (/\*\*[^*]+\*\*/.test(s)) structured = true
    const html = mdInlineHtml(s.trim())
    if (!html) continue
    items.push({ level, ordered, html })
  }
  return structured ? items : null
}

// 把扁平行数组组装成嵌套 ul/ol
function buildNestedList(items) {
  let out = ''
  const stack = []
  const tag = (ordered) => (ordered ? 'ol' : 'ul')
  const top = () => stack[stack.length - 1]
  for (const it of items) {
    while (stack.length && (top().level > it.level || (top().level === it.level && top().ordered !== it.ordered))) {
      if (top().open) out += '</li>'
      out += '</' + tag(top().ordered) + '>'
      stack.pop()
    }
    if (!stack.length || top().level < it.level) {
      out += '<' + tag(it.ordered) + '>'
      stack.push({ level: it.level, ordered: it.ordered, open: false })
    }
    if (top().open) out += '</li>'
    out += '<li>' + it.html
    top().open = true
  }
  while (stack.length) {
    if (top().open) out += '</li>'
    out += '</' + tag(top().ordered) + '>'
    stack.pop()
  }
  return out
}

function pasteMarkdownToHtml(text) {
  const items = parseMarkdownLines(text)
  return items ? buildNestedList(items) : null
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
