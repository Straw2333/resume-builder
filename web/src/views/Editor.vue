<template>
  <div class="editor-page" v-if="loaded">
    <!-- 顶栏 -->
    <header class="topbar">
      <div class="brand" @click="$router.push('/')">
        <span class="brand-mark">简</span>
        <span class="brand-name">简历制作平台</span>
      </div>
      <div class="topbar-actions">
        <span class="save-status" :class="saveState">
          <el-icon v-if="saveState === 'saving'" class="is-loading"><Loading /></el-icon>
          {{ statusText }}
        </span>
        <el-tooltip content="快捷键：Ctrl+S / ⌘S" placement="bottom">
          <el-button class="pill" :loading="saveState === 'saving'" @click="saveNow">
            <el-icon v-if="saveState !== 'saving'"><Check /></el-icon>&nbsp;保存
          </el-button>
        </el-tooltip>
        <el-button class="pill pill-ghost" @click="$router.push(`/preview/${resume.id}`)">
          <el-icon><View /></el-icon>&nbsp;整页预览
        </el-button>
        <el-dropdown trigger="click" @command="download">
          <el-button type="primary" class="pill">
            <el-icon><Download /></el-icon>&nbsp;下载简历
            <el-icon class="caret"><ArrowDown /></el-icon>
          </el-button>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item command="pdf"><el-icon><Document /></el-icon>导出 PDF</el-dropdown-item>
              <el-dropdown-item command="docx"><el-icon><Notebook /></el-icon>导出 Word</el-dropdown-item>
              <el-dropdown-item command="md"><el-icon><Memo /></el-icon>导出 Markdown</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </div>
    </header>

    <div class="editor-body">
      <!-- 图标侧栏 -->
      <aside class="rail">
        <button
          class="rail-btn" :class="{ active: panelOpen && panelTab === 'modules' }"
          title="模块与内容" @click="switchTab('modules')"
        >
          <el-icon :size="18"><Grid /></el-icon><span>模块</span>
        </button>
        <button
          class="rail-btn" :class="{ active: panelOpen && panelTab === 'text' }"
          title="文字设置" @click="switchTab('text')"
        >
          <el-icon :size="18"><EditPen /></el-icon><span>文字</span>
        </button>
        <button
          class="rail-btn" :class="{ active: panelOpen && panelTab === 'style' }"
          title="样式设置" @click="switchTab('style')"
        >
          <el-icon :size="18"><Brush /></el-icon><span>样式</span>
        </button>
      </aside>

      <!-- 内容面板 -->
      <aside class="panel" v-show="panelOpen">
        <!-- 文字设置 -->
        <template v-if="panelTab === 'text'">
          <div class="panel-block">
            <div class="panel-title">文字设置<span class="panel-tip">整页实时生效</span></div>
            <div class="field"><label>中文字体</label>
              <el-select v-model="resume.theme.fontCN" class="grow">
                <el-option v-for="o in fontCNOptions" :key="o.v" :value="o.v" :label="o.l" />
              </el-select>
            </div>
            <div class="field"><label>英文字体（仅作用于英文和数字）</label>
              <el-select v-model="resume.theme.fontEN" class="grow">
                <el-option v-for="o in fontENOptions" :key="o.v" :value="o.v" :label="o.l" />
              </el-select>
            </div>
            <div class="field">
              <label>正文字号 <b class="range-val">{{ resume.theme.fontSize }}px</b></label>
              <el-slider v-model="resume.theme.fontSize" :min="12" :max="18" :step="1" />
            </div>
            <div class="field">
              <label>行间距 <b class="range-val">{{ resume.theme.lineHeight }}</b></label>
              <el-slider v-model="resume.theme.lineHeight" :min="1.2" :max="2.4" :step="0.1" />
            </div>
            <div class="theme-tip">修改后自动保存并刷新预览；中文字体只影响中文，英文字体只影响西文与数字</div>
          </div>
        </template>

        <!-- 样式设置 -->
        <template v-else-if="panelTab === 'style'">
          <div class="panel-block">
            <div class="panel-title">样式设置</div>
            <div class="field"><label>头部布局</label>
              <el-radio-group v-model="resume.theme.headerLayout" class="grow">
                <el-radio-button value="center">居中</el-radio-button>
                <el-radio-button value="left">居左</el-radio-button>
                <el-radio-button value="flat">平铺</el-radio-button>
              </el-radio-group>
            </div>
            <div class="field"><label>信息展示</label>
              <el-radio-group v-model="resume.theme.fieldStyle" class="grow">
                <el-radio-button value="icon">图标</el-radio-button>
                <el-radio-button value="text">文字</el-radio-button>
                <el-radio-button value="plain">纯内容</el-radio-button>
              </el-radio-group>
            </div>
            <div class="field"><label>信息展示说明</label>
              <div class="style-tip-list">
                <span>图标：图标 + 内容（可在基本信息里逐字段改图标）</span>
                <span>文字：以「字段名：内容」展示</span>
                <span>纯内容：只显示内容本身</span>
              </div>
            </div>
          </div>
          <div class="panel-block">
            <div class="panel-title">简历配色</div>
            <div class="theme-row">
              <span>标题底色</span>
              <el-color-picker v-model="resume.theme.dark" :predefine="predefineColors" />
            </div>
            <div class="theme-row">
              <span>强调色</span>
              <el-color-picker v-model="resume.theme.accent" :predefine="predefineColors" />
            </div>
            <div class="theme-presets">
              <div
                v-for="p in themePresets" :key="p.name" class="swatch" :title="p.name"
                @click="applyTheme(p)"
              >
                <span class="swatch-dark" :style="{ background: p.dark }"></span>
                <span class="swatch-accent" :style="{ background: p.accent }"></span>
                <span class="swatch-name">{{ p.name }}</span>
              </div>
            </div>
            <div class="theme-tip">标题底色用于章节丝带，强调色用于求职意向等；修改后自动保存</div>
          </div>
        </template>

        <!-- 模块与内容 -->
        <template v-else>
        <!-- 模块管理 -->
        <div class="panel-block">
          <div class="panel-title">简历模块<span class="panel-tip">选填，可开关/排序</span></div>
          <div
            v-for="(s, i) in resume.sections"
            :key="i"
            class="module-item"
            :class="{ active: activeSection === i }"
            @click="gotoSection(i)"
          >
            <el-switch v-model="s.visible" size="small" @click.stop />
            <span class="module-name">{{ s.title }}</span>
            <span class="module-type">{{ typeLabel(s.type) }}</span>
            <span class="module-ops" @click.stop>
              <el-icon class="op" :class="{ disabled: i === 0 }" @click="move(i, -1)"><ArrowUp /></el-icon>
              <el-icon class="op" :class="{ disabled: i === resume.sections.length - 1 }" @click="move(i, 1)"><ArrowDown /></el-icon>
              <el-icon v-if="s.type === 'custom'" class="op" title="修改模块名称" @click="renameSection(i)"><Pencil /></el-icon>
              <el-icon v-if="s.type === 'custom'" class="op danger" title="删除模块" @click="removeSection(i)"><Delete /></el-icon>
            </span>
          </div>
          <el-button class="add-custom" plain size="small" @click="addCustomSection">
            <el-icon><Plus /></el-icon>&nbsp;添加自定义模块
          </el-button>
        </div>

        <!-- 简历名称 + 基本信息 -->
        <div class="panel-block" id="card-basic">
          <div class="panel-title">简历名称<span class="panel-tip">仅用于列表管理</span></div>
          <div class="field"><el-input v-model="resume.title" placeholder="如：Go后端-张三" /></div>

          <div class="panel-title" style="margin-top: 14px">基本信息<span class="panel-tip">字段可增删/开关/选图标</span></div>
          <div class="field"><label>姓名</label><el-input v-model="basic.name" placeholder="张三" /></div>
          <div class="field"><label>求职意向</label><el-input v-model="basic.intent" placeholder="如：Go 后端开发" /></div>

          <div v-for="(f, i) in basic.fields" :key="i" class="bfield">
            <div class="bfield-row">
              <el-switch v-model="f.visible" size="small" :title="f.visible ? '点击隐藏' : '点击显示'" />
              <el-input v-model="f.label" class="bfield-label" placeholder="字段名" />
              <el-select v-model="f.icon" class="bfield-icon" placeholder="图标">
                <el-option v-for="opt in iconOptions" :key="opt.v" :value="opt.v" :label="opt.l" />
              </el-select>
              <el-icon class="op danger" title="删除字段" @click="basic.fields.splice(i, 1)"><Delete /></el-icon>
            </div>
            <el-input v-model="f.value" :placeholder="'填写' + (f.label || '内容')" class="bfield-value" />
          </div>

          <el-dropdown class="add-field-wrap" trigger="click" @command="addField">
            <el-button plain size="small" class="add-custom">
              <el-icon><Plus /></el-icon>&nbsp;添加字段
            </el-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="p in fieldPresets" :key="p.key" :command="p"
                  :disabled="hasField(p.key)"
                >
                  {{ p.label }}
                  <span v-if="hasField(p.key)" class="preset-added">已添加</span>
                </el-dropdown-item>
                <el-dropdown-item command="custom" divided><el-icon><Pencil /></el-icon>&nbsp;自定义字段…</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>

          <div class="field" style="margin-top: 12px"><label>头像 URL（选填，显示在右上角）</label><el-input v-model="basic.avatar" placeholder="https://…" /></div>
        </div>

        <!-- 各模块表单 -->
        <div
          v-for="(s, i) in resume.sections"
          :key="'form-' + i"
          class="panel-block"
          :class="{ hidden: !s.visible }"
          :id="'card-' + i"
        >
          <div class="panel-title">
            <el-icon v-if="!s.visible" color="#94a3b8"><Hide /></el-icon>
            {{ s.title }}
            <span class="panel-tip" v-if="!s.visible">已关闭，不显示在简历中</span>
            <el-button v-if="s.type === 'custom'" text size="small" @click="renameSection(i)">
              <el-icon><Pencil /></el-icon>
            </el-button>
          </div>

          <!-- 自我评价 -->
          <template v-if="s.type === 'evaluation'">
            <el-input
              v-model="s.content" type="textarea" :rows="5" class="grow"
              placeholder="每行一条，例如：&#10;热爱技术，持续学习&#10;具备良好的沟通与团队协作能力"
            />
          </template>

          <!-- 列表型模块 -->
          <template v-else>
            <div v-for="(item, j) in s.items" :key="j" class="item-box">
              <div class="item-toolbar">
                <span class="item-index">#{{ j + 1 }}</span>
                <el-button-group size="small">
                  <el-button size="small" :disabled="j === 0" @click="moveItem(s, j, -1)"><el-icon><ArrowUp /></el-icon></el-button>
                  <el-button size="small" :disabled="j === s.items.length - 1" @click="moveItem(s, j, 1)"><el-icon><ArrowDown /></el-icon></el-button>
                  <el-button size="small" type="danger" plain @click="s.items.splice(j, 1)"><el-icon><Delete /></el-icon></el-button>
                </el-button-group>
              </div>
              <div class="field"><label>{{ hintOf(s.type).title }}</label><el-input v-model="item.title" :placeholder="hintOf(s.type).titlePh" /></div>
              <div class="field-row">
                <div class="field"><label>{{ hintOf(s.type).subtitle }}</label><el-input v-model="item.subtitle" :placeholder="hintOf(s.type).subtitlePh" /></div>
                <div class="field"><label>时间</label><el-input v-model="item.time" :placeholder="hintOf(s.type).timePh" /></div>
              </div>
              <div class="field">
                <label>描述（支持加粗 / 列表 / 缩进，Ctrl+B 加粗，Tab 与 Shift+Tab 缩进）</label>
                <RichTextEditor v-model="item.desc" :placeholder="hintOf(s.type).descPh" />
              </div>
              <div class="field">
                <label>标签（选填，顿号/逗号分隔）</label>
                <el-input v-model="item.tagsStr" placeholder="Go、Redis、Kafka" />
              </div>
            </div>
            <el-button plain type="primary" size="small" class="add-item" @click="addItem(s)">
              <el-icon><Plus /></el-icon>&nbsp;添加一条
            </el-button>
          </template>
        </div>
        <div class="panel-bottom-space"></div>
        </template>
      </aside>

      <!-- 深色画布：A4 实时预览 -->
      <main class="canvas">
        <div class="paper">
          <iframe v-if="previewSrc" :src="previewSrc" title="简历实时预览"></iframe>
        </div>
      </main>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute, onBeforeRouteLeave } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import RichTextEditor from '../components/RichTextEditor.vue'
import api from '../api/resume'

const route = useRoute()
const loaded = ref(false)
const hydrating = ref(true)
const panelOpen = ref(true)
const panelTab = ref('modules') // modules | text | style
const activeSection = ref(0)
const saveState = ref('saved') // saved | dirty | saving
const previewSrc = ref('')

const resume = reactive({
  id: 0, title: '',
  theme: {
    dark: '#16213e', accent: '#2563eb',
    fontCN: 'def', fontEN: 'def',
    fontSize: 14, lineHeight: 1.6,
    headerLayout: 'center', fieldStyle: 'icon'
  },
  sections: []
})
const basic = reactive({ name: '', intent: '', avatar: '', fields: [] })

function switchTab(t) {
  if (panelOpen.value && panelTab.value === t) {
    panelOpen.value = false // 再次点击收起面板
  } else {
    panelTab.value = t
    panelOpen.value = true
  }
}

const typeLabels = {
  education: '教育背景', skill: '专业技能', internship: '实习经历',
  project: '项目经历', honor: '荣誉证书', evaluation: '自我评价', custom: '自定义'
}
const typeLabel = (t) => typeLabels[t] || t

const statusText = computed(() => ({
  saved: '已保存', dirty: '有未保存更改', saving: '保存中…'
}[saveState.value]))

// ===== 基本信息字段 =====
const iconOptions = [
  { v: 'none', l: '无图标（显示字段名）' }, { v: 'user', l: '用户' },
  { v: 'calendar', l: '日历' }, { v: 'map-pin', l: '位置' },
  { v: 'phone', l: '电话' }, { v: 'mail', l: '邮件' },
  { v: 'message-circle', l: '消息 / 微信' }, { v: 'github', l: 'GitHub' },
  { v: 'globe', l: '网站' }, { v: 'flag', l: '旗帜' },
  { v: 'award', l: '奖章' }, { v: 'cake', l: '生日' },
  { v: 'graduation-cap', l: '学历' }, { v: 'briefcase', l: '公文包' },
  { v: 'link', l: '链接' }
]
const fieldPresets = [
  { key: 'political', label: '政治面貌', icon: 'flag' },
  { key: 'native', label: '籍贯', icon: 'map-pin' },
  { key: 'birthday', label: '生日', icon: 'cake' },
  { key: 'degree', label: '学历', icon: 'graduation-cap' },
  { key: 'wechat', label: '微信', icon: 'message-circle' },
  { key: 'github', label: 'GitHub', icon: 'github' },
  { key: 'homepage', label: '个人主页', icon: 'globe' },
  { key: 'workyears', label: '工作年限', icon: 'briefcase' }
]
const hasField = (key) => basic.fields.some((f) => f.key === key)

async function addField(preset) {
  if (preset === 'custom') {
    const { value } = await ElMessageBox.prompt('请输入字段名称（如：期望薪资）', '添加自定义字段', {
      inputValue: '',
      inputValidator: (v) => (v && v.trim() ? true : '名称不能为空')
    })
    basic.fields.push({
      key: 'c_' + Date.now().toString(36), label: value.trim(),
      value: '', icon: 'none', visible: true
    })
    return
  }
  basic.fields.push({ key: preset.key, label: preset.label, value: '', icon: preset.icon, visible: true })
}

// ===== 文字/样式选项 =====
const fontCNOptions = [
  { v: 'def', l: '系统默认' },
  { v: 'yahei', l: '微软雅黑' },
  { v: 'ping', l: '苹方' },
  { v: 'source', l: '思源黑体' },
  { v: 'song', l: '宋体' },
  { v: 'hei', l: '黑体' },
  { v: 'kai', l: '楷体' }
]
const fontENOptions = [
  { v: 'def', l: '默认 (IBM Plex Sans)' },
  { v: 'system', l: '系统无衬线' },
  { v: 'arial', l: 'Arial' },
  { v: 'times', l: 'Times New Roman' },
  { v: 'georgia', l: 'Georgia' },
  { v: 'consolas', l: 'Consolas 等宽' }
]

// ===== 主题配色 =====
const themePresets = [
  { name: '蓝黑', dark: '#16213e', accent: '#2563eb' },
  { name: '墨黑', dark: '#111827', accent: '#10b981' },
  { name: '青碧', dark: '#0f3b3a', accent: '#14b8a6' },
  { name: '酒红', dark: '#4c1d24', accent: '#e05d6f' }
]
const predefineColors = ['#16213e', '#1f2328', '#111827', '#0f3b3a', '#4c1d24', '#2563eb', '#10b981', '#14b8a6', '#e05d6f', '#c9a227']
function applyTheme(p) {
  resume.theme.dark = p.dark
  resume.theme.accent = p.accent
}

const hints = {
  education:  { title: '学校名称', titlePh: '某某大学', subtitle: '专业 / 学历', subtitlePh: '计算机科学与技术 / 本科', timePh: '2021.09 - 2025.06', descPh: '主修课程、GPA、排名等（选填）' },
  skill:      { title: '技能分类', titlePh: '如：编程语言', subtitle: '熟练度（选填）', subtitlePh: '熟练 / 掌握 / 了解', timePh: '选填', descPh: '技能说明（选填）' },
  internship: { title: '公司名称', titlePh: '某某科技有限公司', subtitle: '职位', subtitlePh: '后端开发实习生', timePh: '2024.06 - 2024.09', descPh: '工作内容与业绩，每行一条，建议用数据量化' },
  project:    { title: '项目名称', titlePh: '高并发秒杀系统', subtitle: '担任角色', subtitlePh: '核心开发', timePh: '2024.03 - 2024.06', descPh: '项目背景、你的职责、技术栈与成果，每行一条' },
  honor:      { title: '证书 / 奖项名称', titlePh: '国家奖学金', subtitle: '颁发机构（选填）', subtitlePh: '教育部', timePh: '2024.05', descPh: '补充说明（选填）' },
  custom:     { title: '标题', titlePh: '标题', subtitle: '副标题（选填）', subtitlePh: '副标题', timePh: '时间（选填）', descPh: '内容描述，每行一条' }
}
const hintOf = (t) => hints[t] || hints.custom

onMounted(async () => {
  const r = await api.get(route.params.id)
  resume.id = r.id
  resume.title = r.title
  const t = r.theme || {}
  resume.theme.dark = t.dark || '#16213e'
  resume.theme.accent = t.accent || '#2563eb'
  resume.theme.fontCN = t.fontCN || 'def'
  resume.theme.fontEN = t.fontEN || 'def'
  resume.theme.fontSize = t.fontSize || 14
  resume.theme.lineHeight = t.lineHeight || 1.6
  resume.theme.headerLayout = t.headerLayout || 'center'
  resume.theme.fieldStyle = t.fieldStyle || 'icon'
  // 回填 tags → tagsStr，否则再次保存会清空已有标签
  resume.sections = (r.sections || []).map((s) => ({
    ...s,
    items: (s.items || []).map((it) => ({ ...it, tagsStr: (it.tags || []).join('、') }))
  }))
  const b = r.basicInfo || {}
  basic.name = b.name || ''
  basic.intent = b.intent || ''
  basic.avatar = b.avatar || ''
  basic.fields = b.fields || []
  loaded.value = true
  await nextTick()
  hydrating.value = false
  previewSrc.value = `/api/resumes/${resume.id}/preview?t=${Date.now()}`
  window.addEventListener('beforeunload', onBeforeUnload)
  window.addEventListener('keydown', onKeydown)
})
onBeforeUnmount(() => {
  window.removeEventListener('beforeunload', onBeforeUnload)
  window.removeEventListener('keydown', onKeydown)
})

// ===== 保存：自动保存 + Ctrl+S / ⌘S =====
let timer = null
watch([resume, basic], () => {
  if (hydrating.value) return
  saveState.value = 'dirty'
  clearTimeout(timer)
  timer = setTimeout(saveNow, 1200)
}, { deep: true })

function onKeydown(e) {
  if ((e.ctrlKey || e.metaKey) && (e.key === 's' || e.key === 'S')) {
    e.preventDefault()
    saveNow()
  }
}

async function saveNow() {
  if (saveState.value === 'saving') return
  clearTimeout(timer)
  saveState.value = 'saving'
  try {
    const sections = resume.sections.map((s) => ({
      ...s,
      items: (s.items || []).map((it) => ({
        title: it.title || '', subtitle: it.subtitle || '', time: it.time || '',
        desc: it.desc || '',
        tags: (it.tagsStr || '').split(/[,，、\s]+/).filter(Boolean)
      }))
    }))
    await api.update(resume.id, {
      title: resume.title, template: 'classic',
      basicInfo: {
        name: basic.name, intent: basic.intent, avatar: basic.avatar,
        fields: JSON.parse(JSON.stringify(basic.fields))
      },
      sections,
      theme: { ...resume.theme }
    })
    saveState.value = 'saved'
    previewSrc.value = `/api/resumes/${resume.id}/preview?t=${Date.now()}`
  } catch {
    saveState.value = 'dirty'
  }
}

// ===== 下载导出 =====
const extOf = { pdf: '.pdf', docx: '.doc', md: '.md' }

async function download(format) {
  await saveNow()
  try {
    const resp = await fetch(api.exportUrl(resume.id, format))
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}))
      ElMessage.error(err.error || '导出失败')
      return
    }
    const blob = await resp.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `${resume.title || 'resume'}${extOf[format] || ''}`
    a.click()
    URL.revokeObjectURL(a.href)
    ElMessage.success('导出成功')
  } catch { /* 拦截器已提示 */ }
}

// ===== 模块管理 =====
function gotoSection(i) {
  activeSection.value = i
  document.getElementById(`card-${i}`)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
}

function addItem(s) {
  s.items.push({ title: '', subtitle: '', time: '', desc: '', tags: [], tagsStr: '' })
}

function move(i, dir) {
  const j = i + dir
  if (j < 0 || j >= resume.sections.length) return
  ;[resume.sections[i], resume.sections[j]] = [resume.sections[j], resume.sections[i]]
  activeSection.value = j
}

function moveItem(s, i, dir) {
  const j = i + dir
  if (j < 0 || j >= s.items.length) return
  ;[s.items[i], s.items[j]] = [s.items[j], s.items[i]]
}

async function addCustomSection() {
  const { value } = await ElMessageBox.prompt('请输入模块名称', '添加自定义模块', {
    inputValue: '自定义模块',
    inputValidator: (v) => (v && v.trim() ? true : '名称不能为空')
  })
  resume.sections.push({ type: 'custom', title: value.trim(), visible: true, items: [] })
  activeSection.value = resume.sections.length - 1
  await nextTick()
  gotoSection(resume.sections.length - 1)
}

async function renameSection(i) {
  const { value } = await ElMessageBox.prompt('修改模块名称', '重命名', {
    inputValue: resume.sections[i].title,
    inputValidator: (v) => (v && v.trim() ? true : '名称不能为空')
  })
  resume.sections[i].title = value.trim()
}

function removeSection(i) {
  resume.sections.splice(i, 1)
}

// ===== 离开保护 =====
function onBeforeUnload(e) {
  if (saveState.value !== 'saved') {
    e.preventDefault()
    e.returnValue = ''
  }
}
onBeforeRouteLeave(async () => {
  if (saveState.value === 'saved') return true
  try {
    await ElMessageBox.confirm('还有未保存的更改，离开将丢失', '提示', {
      type: 'warning', confirmButtonText: '保存并离开', cancelButtonText: '直接离开',
      distinguishCancelAndClose: true
    })
    await saveNow()
    return true
  } catch (action) {
    if (action === 'cancel') return true // 直接离开
    return false                          // 关闭弹窗 = 留在页面
  }
})
</script>

<style scoped>
.editor-page { height: 100vh; display: flex; flex-direction: column; overflow: hidden; }

/* 顶栏 */
.topbar {
  height: 64px; flex-shrink: 0; background: #fff; border-bottom: 1px solid #e2e8f0;
  display: flex; justify-content: space-between; align-items: center; padding: 0 24px; z-index: 20;
}
.brand { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.brand-mark {
  width: 36px; height: 36px; border-radius: 10px; background: #1f2328; color: #fff;
  display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 18px;
}
.brand-name { font-size: 20px; font-weight: 700; color: #0f172a; letter-spacing: -0.02em; }
.topbar-actions { display: flex; align-items: center; gap: 12px; }
.save-status { font-size: 13px; color: #94a3b8; display: inline-flex; align-items: center; gap: 4px; margin-right: 2px; }
.save-status.dirty { color: #f59e0b; }
.save-status.saving { color: #2563eb; }
.save-status.saved { color: #16a34a; }
.pill { border-radius: 999px; }
.icon-btn { width: 36px; padding: 8px 0; }
.caret { margin-left: 2px; }

/* 主题面板 */
.theme-title { font-weight: 600; margin-bottom: 10px; }
.range-val { color: #2563eb; font-weight: 600; font-size: 12.5px; }
.style-tip-list { display: flex; flex-direction: column; gap: 4px; font-size: 12px; color: #94a3b8; }
.theme-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; font-size: 13.5px; color: #475569; }
.theme-presets { display: flex; gap: 10px; margin-top: 4px; }
.swatch { cursor: pointer; text-align: center; width: 52px; }
.swatch-dark { display: block; height: 30px; border-radius: 8px 8px 3px 3px; border: 1px solid #e2e8f0; }
.swatch-accent { display: block; height: 8px; border-radius: 0 0 8px 8px; margin-top: -4px; }
.swatch-name { font-size: 11.5px; color: #64748b; }
.swatch:hover .swatch-dark { outline: 2px solid #2563eb; outline-offset: 1px; }
.theme-tip { margin-top: 10px; font-size: 12px; color: #94a3b8; }

/* 主体 */
.editor-body { flex: 1; display: flex; overflow: hidden; }

/* 图标侧栏 */
.rail {
  width: 56px; flex-shrink: 0; display: flex; flex-direction: column; align-items: center;
  padding-top: 14px; gap: 6px; z-index: 5;
  background: linear-gradient(#fafbfc 0%, #f8f9fa 100%); border-right: 1px solid rgba(0,0,0,.06);
}
.rail-btn {
  width: 44px; height: 52px; display: flex; flex-direction: column; align-items: center;
  justify-content: center; gap: 4px; border: 1px solid transparent; border-radius: 10px;
  cursor: pointer; background: transparent; color: #656d76; transition: .2s; font-size: 10px; font-weight: 500;
}
.rail-btn:hover { background: #fff; color: #2563eb; box-shadow: 0 1px 4px rgba(0,0,0,.08); }
.rail-btn.active { background: #fff; color: #2563eb; border-color: #e2e8f0; box-shadow: 0 1px 4px rgba(0,0,0,.08); }

/* 内容面板 */
.panel {
  width: 360px; flex-shrink: 0; overflow-y: auto; background: #fafbfc;
  border-right: 1px solid rgba(0,0,0,.06); padding: 14px 14px 0;
}
.panel-block {
  background: #fff; border: 1px solid #e2e8f0; border-radius: 12px;
  padding: 14px; margin-bottom: 14px;
}
.panel-block.hidden { opacity: .6; }
.panel-title { font-weight: 600; color: #0f172a; margin-bottom: 10px; display: flex; align-items: center; gap: 6px; font-size: 14.5px; }
.panel-tip { font-size: 12px; color: #94a3b8; font-weight: 400; margin-left: auto; }
.field { margin-bottom: 10px; }
.field label { display: block; font-size: 12px; color: #64748b; margin-bottom: 4px; }
.field-row { display: flex; gap: 10px; }
.field-row .field { flex: 1; min-width: 0; }
.grow { width: 100%; }

/* 基本信息字段行 */
.bfield { border: 1px solid #eef2f6; border-radius: 10px; padding: 10px; margin-bottom: 10px; background: #fcfdfe; }
.bfield-row { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
.bfield-label { width: 96px; flex-shrink: 0; }
.bfield-icon { flex: 1; min-width: 0; }
.bfield-value { width: 100%; }
.add-field-wrap { display: block; }
.preset-added { float: right; font-size: 11px; color: #94a3b8; }

.module-item {
  display: flex; align-items: center; gap: 8px; padding: 9px 8px;
  border-radius: 8px; cursor: pointer; border: 1px solid transparent;
}
.module-item:hover { background: #f1f5f9; }
.module-item.active { background: #eff6ff; border-color: #bfdbfe; }
.module-name { flex: 1; font-size: 13.5px; color: #0f172a; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.module-type { font-size: 11.5px; color: #94a3b8; }
.module-ops { display: flex; gap: 6px; }
.op { cursor: pointer; color: #94a3b8; flex-shrink: 0; }
.op:hover { color: #2563eb; }
.op.disabled { color: #e2e8f0; pointer-events: none; }
.op.danger:hover { color: #ef4444; }
.add-custom { width: 100%; border-style: dashed; }

.item-box { border: 1px solid #e2e8f0; border-radius: 10px; padding: 12px; margin-bottom: 12px; background: #fcfdfe; }
.item-toolbar { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.item-index { color: #2563eb; font-weight: 600; font-size: 12.5px; }
.add-item { width: 100%; border-style: dashed; }
.panel-bottom-space { height: 24px; }

/* 深色画布 + A4 纸张 */
.canvas {
  flex: 1; min-width: 0; overflow: auto;
  background: linear-gradient(rgb(62,76,94) 0%, rgb(74,85,104) 50%, rgb(45,55,72) 100%);
  padding: 40px 20px 60px;
  display: flex; justify-content: center; align-items: flex-start;
}
.paper {
  width: 794px; flex-shrink: 0; background: #fff; border-radius: 4px;
  box-shadow: 0 8px 32px rgba(0,0,0,.4), 0 2px 8px rgba(0,0,0,.25); overflow: hidden;
}
.paper iframe { width: 794px; height: 1123px; border: none; display: block; background: #fff; }
</style>
