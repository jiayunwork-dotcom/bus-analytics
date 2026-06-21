import { createRouter, createWebHistory } from 'vue-router'

const routes = [
  { path: '/', redirect: '/dashboard' },
  { path: '/dashboard', name: 'Dashboard', component: () => import('../views/Dashboard.vue') },
  { path: '/upload', name: 'Upload', component: () => import('../views/Upload.vue') },
  { path: '/metrics', name: 'Metrics', component: () => import('../views/Metrics.vue') },
  { path: '/metrics/:lineNo', name: 'LineDetail', component: () => import('../views/LineDetail.vue') },
  { path: '/flow-analysis', name: 'FlowAnalysis', component: () => import('../views/FlowAnalysis.vue') },
  { path: '/schedule', name: 'Schedule', component: () => import('../views/Schedule.vue') },
  { path: '/vehicles', name: 'Vehicles', component: () => import('../views/Vehicles.vue') },
  { path: '/network', name: 'Network', component: () => import('../views/Network.vue') },
  { path: '/compare', name: 'Compare', component: () => import('../views/Compare.vue') },
  { path: '/compare/health', name: 'LineHealth', component: () => import('../views/LineHealth.vue') },
  { path: '/simulation', name: 'Simulation', component: () => import('../views/Simulation.vue') },
  { path: '/plan-archive', name: 'PlanArchive', component: () => import('../views/PlanArchive.vue') },
  { path: '/plan-compare', name: 'PlanCompare', component: () => import('../views/PlanCompare.vue') },
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
