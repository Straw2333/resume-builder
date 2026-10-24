// AI 优化结果比对与要点级（<li>）合并工具

export const DIRECTION_TAGS = [
  { label: '突出 STAR 法则', text: '按背景(S)-任务(T)-行动(A)-结果(R)逻辑重组表达' },
  { label: '强化成果与价值', text: '用强动作词开头，突出可量化的成果与核心产出价值' },
  { label: '语言精简干练', text: '删减空话套话与冗余表述，用词精炼紧凑' },
  { label: '突出技术深度', text: '重点凸显技术栈落地、架构设计与难点攻克细节' },
  { label: '强调个人贡献', text: '突出个人主导职责与解决复杂问题的能力' }
]

const escHtml = (s) =>
  String(s || '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

// 将纯文本行（可能含简单 **加粗** 或行首列表符）转为安全行内 HTML
function plainLineToInlineHtml(line) {
  let t = line.trim().replace(/^([-*+]|\d+[.、)])\s+/, '')
  t = escHtml(t)
  return t.replace(/\*\*(.+?)\*\*/g, '<b>$1</b>')
}

const ALLOWED_TAGS = new Set(['B', 'STRONG', 'I', 'EM', 'U', 'S', 'BR', 'P', 'UL', 'OL', 'LI'])
const DROP_TAGS = new Set(['SCRIPT', 'STYLE', 'META', 'LINK', 'TITLE', 'THINK'])

// 净化节点为安全白名单 HTML（剥离所有属性与非白名单外壳）
function sanitizeNode(node) {
  if (node.nodeType === Node.TEXT_NODE) {
    return escHtml(node.nodeValue)
  }
  if (node.nodeType !== Node.ELEMENT_NODE) return ''
  const tag = node.tagName.toUpperCase()
  if (DROP_TAGS.has(tag)) return ''
  let inner = ''
  for (const child of node.childNodes) {
    inner += sanitizeNode(child)
  }
  if (ALLOWED_TAGS.has(tag)) {
    const lower = tag.toLowerCase()
    return lower === 'br' ? '<br/>' : `<${lower}>${inner}</${lower}>`
  }
  return inner
}

function sanitizeFragment(html) {
  const doc = new DOMParser().parseFromString(String(html || ''), 'text/html')
  let out = ''
  for (const child of doc.body.childNodes) {
    out += sanitizeNode(child)
  }
  return out.trim()
}

// 将富文本 HTML 或纯文本拆解为要点数组（每项为单条 <li> 内部的 HTML）
export function parseHtmlPoints(raw) {
  const s = String(raw || '').trim()
  if (!s) return []

  // 纯文本（如自我评价或未转 HTML 的 Markdown 文本）
  if (!s.includes('<')) {
    return s
      .split(/\r?\n/)
      .map((l) => l.trim())
      .filter(Boolean)
      .map(plainLineToInlineHtml)
  }

  const doc = new DOMParser().parseFromString(s, 'text/html')
  // 仅提取顶层 <li>（避免嵌套列表中的子 <li> 被重复拆成独立行）
  const lis = Array.from(doc.querySelectorAll('li')).filter(
    (li) => !li.parentElement?.closest('li')
  )
  if (lis.length > 0) {
    return lis
      .map((li) => sanitizeFragment(li.innerHTML))
      .filter(Boolean)
  }

  // 无 <li> 时尝试按 <p> / <div> 切分
  const blocks = Array.from(doc.body.querySelectorAll('p, div'))
  if (blocks.length > 0) {
    const out = []
    for (const b of blocks) {
      const parts = b.innerHTML
        .split(/<br\s*\/?>/i)
        .map((p) => sanitizeFragment(p))
        .filter(Boolean)
      out.push(...parts)
    }
    if (out.length > 0) return out
  }

  // 兜底：按 <br> 或换行切分
  return doc.body.innerHTML
    .split(/<br\s*\/?>|\r?\n/i)
    .map((p) => sanitizeFragment(p))
    .filter(Boolean)
}

// 生成原文与 AI 优化后的逐条对齐数组
export function buildPointPairs(origRaw, aiHtml) {
  const origList = parseHtmlPoints(origRaw)
  const aiList = parseHtmlPoints(aiHtml)
  const len = Math.max(origList.length, aiList.length)
  const pairs = []
  for (let i = 0; i < len; i++) {
    const orig = origList[i] || ''
    const ai = aiList[i] || ''
    pairs.push({
      index: i,
      orig,
      ai,
      // 若有 AI 生成内容默认勾选采纳；若 AI 少了该条则默认不勾选（即保留原文）
      selected: Boolean(ai)
    })
  }
  return pairs
}

// 提取 HTML 纯文本
export function htmlToText(html) {
  if (!html) return ''
  const doc = new DOMParser().parseFromString(html, 'text/html')
  return (doc.body.textContent || '').trim()
}

// 根据勾选状态合并要点：勾选则采用 AI，未勾选则保留原文对应行
export function mergeSelectedPoints(pairs, kind = 'item-desc') {
  const chosen = []
  for (const p of pairs || []) {
    if (p.selected) {
      if (p.ai) chosen.push(p.ai)
    } else {
      if (p.orig) chosen.push(p.orig)
    }
  }
  if (kind === 'section-content') {
    return chosen.map(htmlToText).filter(Boolean).join('\n')
  }
  if (chosen.length === 0) return ''
  return '<ul>' + chosen.map((c) => `<li>${c}</li>`).join('') + '</ul>'
}

// 追加快捷方向指令到现有输入框
export function appendDirection(current, tagText) {
  const cur = String(current || '').trim()
  if (!cur) return tagText
  if (cur.includes(tagText)) return cur
  const sep = /[；;。，,]$/.test(cur) ? ' ' : '；'
  return cur + sep + tagText
}
