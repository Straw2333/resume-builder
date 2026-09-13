import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', name: 'list', component: () => import('../views/List.vue') },
  { path: '/edit/:id', name: 'editor', component: () => import('../views/Editor.vue') },
  { path: '/preview/:id', name: 'preview', component: () => import('../views/Preview.vue') }
]

export default createRouter({
  history: createWebHistory(),
  routes
})
