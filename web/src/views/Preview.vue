<template>
  <div class="preview-page">
    <header class="topbar">
      <div class="brand" @click="$router.push('/')">
        <span class="brand-mark">简</span>
        <span class="brand-name">简历制作平台</span>
      </div>
      <div class="topbar-actions">
        <el-button class="pill pill-ghost" @click="$router.push(`/edit/${route.params.id}`)">
          <el-icon><Edit /></el-icon>&nbsp;返回编辑
        </el-button>
        <el-dropdown trigger="click" @command="download">
          <el-button type="primary" class="pill" :loading="exporting !== ''">
            <el-icon v-if="exporting === ''"><Download /></el-icon>&nbsp;下载简历
            <el-icon v-if="exporting === ''" class="caret"><ArrowDown /></el-icon>
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

    <main class="canvas">
      <div class="paper">
        <iframe v-if="src" :src="src" title="简历预览"></iframe>
      </div>
    </main>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import api from '../api/resume'

const route = useRoute()
const src = ref('')
const exporting = ref('')

onMounted(() => {
  src.value = `/api/resumes/${route.params.id}/preview?t=${Date.now()}`
})

async function download(format) {
  exporting.value = format
  try {
    const resp = await fetch(api.exportUrl(route.params.id, format))
    if (!resp.ok) {
      const err = await resp.json().catch(() => ({}))
      ElMessage.error(err.error || '导出失败')
      return
    }
    const blob = await resp.blob()
    const a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = `resume.${{ pdf: 'pdf', docx: 'doc', md: 'md' }[format] || 'txt'}`
    a.click()
    URL.revokeObjectURL(a.href)
    ElMessage.success('导出成功')
  } finally {
    exporting.value = ''
  }
}
</script>

<style scoped>
.preview-page { height: 100vh; display: flex; flex-direction: column; }
.topbar {
  height: 64px; flex-shrink: 0; background: #fff; border-bottom: 1px solid #e2e8f0;
  display: flex; justify-content: space-between; align-items: center; padding: 0 24px;
}
.brand { display: flex; align-items: center; gap: 10px; cursor: pointer; }
.brand-mark {
  width: 36px; height: 36px; border-radius: 10px; background: #1f2328; color: #fff;
  display: flex; align-items: center; justify-content: center; font-weight: 700; font-size: 18px;
}
.brand-name { font-size: 20px; font-weight: 700; color: #0f172a; letter-spacing: -0.02em; }
.topbar-actions { display: flex; align-items: center; gap: 14px; }
.pill { border-radius: 999px; }
.canvas {
  flex: 1; overflow: auto; padding: 40px 20px 60px;
  background: linear-gradient(rgb(62,76,94) 0%, rgb(74,85,104) 50%, rgb(45,55,72) 100%);
  display: flex; justify-content: center; align-items: flex-start;
}
.paper {
  width: 794px; flex-shrink: 0; background: #fff; border-radius: 4px;
  box-shadow: 0 8px 32px rgba(0,0,0,.4), 0 2px 8px rgba(0,0,0,.25); overflow: hidden;
}
.paper iframe { width: 794px; height: 1123px; border: none; display: block; background: #fff; }
</style>
