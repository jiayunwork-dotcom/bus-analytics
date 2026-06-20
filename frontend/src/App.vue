<template>
  <el-container class="layout-container">
    <el-aside width="220px" class="sidebar">
      <div class="logo">
        <el-icon :size="28" color="#409eff"><Van /></el-icon>
        <span>公交运营分析平台</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        router
        background-color="#001529"
        text-color="#b8c4cc"
        active-text-color="#409eff"
      >
        <el-menu-item index="/dashboard">
          <el-icon><DataAnalysis /></el-icon>
          <span>概览仪表盘</span>
        </el-menu-item>
        <el-menu-item index="/upload">
          <el-icon><Upload /></el-icon>
          <span>数据导入</span>
        </el-menu-item>
        <el-menu-item index="/metrics">
          <el-icon><TrendCharts /></el-icon>
          <span>线路效率指标</span>
        </el-menu-item>
        <el-menu-item index="/flow-analysis">
          <el-icon><Connection /></el-icon>
          <span>客流分析</span>
        </el-menu-item>
        <el-menu-item index="/schedule">
          <el-icon><Timer /></el-icon>
          <span>发车间隔优化</span>
        </el-menu-item>
        <el-menu-item index="/vehicles">
          <el-icon><Van /></el-icon>
          <span>车辆周转分析</span>
        </el-menu-item>
        <el-menu-item index="/network">
          <el-icon><Share /></el-icon>
          <span>线网评价</span>
        </el-menu-item>
        <el-menu-item index="/compare">
          <el-icon><Rank /></el-icon>
          <span>对标分析</span>
        </el-menu-item>
        <el-menu-item index="/compare/health">
          <el-icon><Medal /></el-icon>
          <span>线路健康度评分</span>
        </el-menu-item>
        <el-menu-item index="/simulation">
          <el-icon><MagicStick /></el-icon>
          <span>线路调整模拟</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="header">
        <span class="title">公交运营管理部门 · 线路效率评估与调度优化分析平台</span>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const activeMenu = computed(() => route.path)
</script>

<style scoped>
.layout-container {
  height: 100vh;
}

.sidebar {
  background-color: #001529;
  overflow-y: auto;
}

.logo {
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: white;
  font-size: 16px;
  font-weight: 600;
  background: rgba(255, 255, 255, 0.05);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.el-menu {
  border-right: none;
}

.header {
  background: white;
  display: flex;
  align-items: center;
  box-shadow: 0 1px 4px rgba(0, 21, 41, 0.08);
  padding: 0 24px;
}

.title {
  font-size: 16px;
  font-weight: 600;
  color: #1f2937;
}

.el-main {
  background-color: #f0f2f5;
  padding: 0;
  overflow-y: auto;
}
</style>
