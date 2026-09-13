<template>
  <div class="list-page">
    <header class="topbar">
      <div class="topbar-inner">
        <div class="brand">
          <span class="brand-mark">简</span>
          <span class="brand-name">简历制作平台</span>
        </div>
        <el-button type="primary" class="pill" size="large" @click="createResume">
          <el-icon><Plus /></el-icon>&nbsp;新建简历
        </el-button>
      </div>
    </header>

    <main class="content">
      <el-empty v-if="!loading && resumes.length === 0" description="还没有简历，点击右上角新建一份吧">
        <el-button type="primary" @click="createResume">新建简历</el-button>
      </el-empty>

      <el-row :gutter="20">
        <el-col v-for="r in resumes" :key="r.id" :xs="24" :sm="12" :md="8" :lg="6">
          <el-card class="resume-card" shadow="hover">
            <div class="card-body">
              <div class="card-title" :title="r.title">{{ r.title }}</div>
              <div class="card-meta">
                <span class="card-name">{{ nameOf(r) || '未填写姓名' }}</span>
              </div>
              <div class="card-time">更新于 {{ formatTime(r.updatedAt) }}</div>
            </div>
            <div class="card-actions">
              <el-button text type="primary" @click="$router.push(`/edit/${r.id}`)">
                <el-icon><Edit /></el-icon>&nbsp;编辑
              </el-button>
              <el-button text type="primary" @click="$router.push(`/preview/${r.id}`)">
                <el-icon><View /></el-icon>&nbsp;预览
              </el-button>
              <el-dropdown trigger="click" @command="(cmd) => onCommand(cmd, r)">
                <el-button text type="primary">
                  更多<el-icon><ArrowDown /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="duplicate"><el-icon><CopyDocument /></el-icon>复制</el-dropdown-item>
                    <el-dropdown-item command="pdf" divided><el-icon><Document /></el-icon>导出 PDF</el-dropdown-item>
                    <el-dropdown-item command="docx"><el-icon><Notebook /></el-icon>导出 Word</el-dropdown-item>
                    <el-dropdown-item command="md"><el-icon><Memo /></el-icon>导出 Markdown</el-dropdown-item>
                    <el-dropdown-item command="delete" divided><el-icon><Delete /></el-icon>删除</el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </div>
          </el-card>
        </el-col>
      </el-row>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import api from '../api/resume'

const router = useRouter()
const resumes = ref([])
const loading = ref(false)

onMounted(load)

async function load() {
  loading.value = true
  try {
    resumes.value = await api.list()
  } finally {
    loading.value = false
  }
}

async function createResume() {
  const { value } = await ElMessageBox.prompt('请输入简历名称', '新建简历', {
    inputValue: '我的简历',
    inputValidator: (v) => (v && v.trim() ? true : '名称不能为空')
  })
  const r = await api.create({ title: value.trim() })
  ElMessage.success('创建成功')
  router.push(`/edit/${r.id}`)
}

function nameOf(r) {
  try {
    const b = typeof r.basicInfo === 'string' ? JSON.parse(r.basicInfo) : r.basicInfo
    return b?.name || ''
  } catch {
    return ''
  }
}

function formatTime(t) {
  return new Date(t).toLocaleString('zh-CN', { hour12: false })
}

async function onCommand(cmd, r) {
  if (cmd === 'duplicate') {
    await api.duplicate(r.id)
    ElMessage.success('复制成功')
    load()
  } else if (cmd === 'delete') {
    await ElMessageBox.confirm(`确定删除「${r.title}」吗？删除后不可恢复`, '删除简历', {
      type: 'warning',
      confirmButtonText: '删除',
      cancelButtonText: '取消'
    })
    await api.remove(r.id)
    ElMessage.success('已删除')
    load()
  } else {
    // pdf / docx 下载
    window.open(api.exportUrl(r.id, cmd), '_blank')
  }
}
</script>

<style scoped>
.topbar { background: #fff; border-bottom: 1px solid #e2e8f0; position: sticky; top: 0; z-index: 10; }
.topbar-inner {
  max-width: 1280px; margin: 0 auto; padding: 14px 24px;
  display: flex; justify-content: space-between; align-items: center;
}
.brand { display: flex; align-items: center; gap: 10px; }
.brand-mark {
  width: 40px; height: 40px; border-radius: 10px; background: #1f2328; color: #fff;
  display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 20px;
}
.brand-name { font-size: 22px; font-weight: 700; color: #0f172a; letter-spacing: -0.02em; }
.pill { border-radius: 999px; }
.content { max-width: 1280px; margin: 28px auto; padding: 0 24px; min-height: calc(100vh - 140px); }
.resume-card { margin-bottom: 20px; border-radius: 12px; border: 1px solid #e2e8f0; }
.card-title { font-size: 17px; font-weight: 700; color: #0f172a; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-meta { margin-top: 8px; color: #475569; }
.card-name { font-size: 13.5px; }
.card-time { margin-top: 6px; color: #94a3b8; font-size: 12.5px; }
.card-actions { display: flex; justify-content: space-between; margin-top: 14px; border-top: 1px solid #f1f5f9; padding-top: 10px; }
</style>
