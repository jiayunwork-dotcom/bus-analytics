<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">
        线路效率指标面板
        <el-tooltip content="绿色=超过全网均值20% | 黄色=一般 | 红色=低于全网均值30%" placement="right">
          <el-icon color="#909399" style="margin-left: 8px; cursor: pointer;"><QuestionFilled /></el-icon>
        </el-tooltip>
      </div>
      <div class="metrics-legend">
        <span class="legend-item"><span class="dot green"></span>优秀 (>均值120%)</span>
        <span class="legend-item"><span class="dot yellow"></span>一般</span>
        <span class="legend-item"><span class="dot red"></span>较差 (<均值70%)</span>
      </div>
      <el-table
        :data="metrics"
        stripe
        :default-sort="{ prop: 'passenger_intensity', order: 'descending' }"
        @row-click="goToDetail"
        style="cursor: pointer;"
      >
        <el-table-column prop="line_no" label="线路编号" width="100" sortable fixed="left" />
        <el-table-column prop="line_name" label="线路名称" min-width="160" />
        <el-table-column label="客运强度 (人次/km)" width="160" sortable prop="passenger_intensity">
          <template #default="{ row }">
            <span :class="getLevel(row, 'passenger_intensity')">{{ row.passenger_intensity?.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="高峰满载率" width="140" sortable prop="peak_load_factor">
          <template #default="{ row }">
            <span :class="getLevel(row, 'peak_load_factor')">{{ (row.peak_load_factor * 100)?.toFixed(1) }}%</span>
          </template>
        </el-table-column>
        <el-table-column label="平峰满载率" width="140" sortable prop="off_peak_load_factor">
          <template #default="{ row }">
            <span :class="getLevel(row, 'off_peak_load_factor')">{{ (row.off_peak_load_factor * 100)?.toFixed(1) }}%</span>
          </template>
        </el-table-column>
        <el-table-column label="营运速度 (km/h)" width="150" sortable prop="operating_speed">
          <template #default="{ row }">
            <span :class="getLevel(row, 'operating_speed')">{{ row.operating_speed?.toFixed(2) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="准点率 (%)" width="130" sortable prop="on_time_rate">
          <template #default="{ row }">
            <span :class="getLevel(row, 'on_time_rate')">{{ row.on_time_rate?.toFixed(2) }}%</span>
          </template>
        </el-table-column>
        <el-table-column label="综合评级" width="120" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.level === 'green'" type="success" effect="dark">优秀</el-tag>
            <el-tag v-else-if="row.level === 'red'" type="danger" effect="dark">较差</el-tag>
            <el-tag v-else type="warning" effect="dark">一般</el-tag>
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 12px; color: #909399; font-size: 13px;">
        💡 点击任意行查看该线路的日维度趋势详情
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import api from '../api'

const router = useRouter()
const metrics = ref([])

const averages = computed(() => {
  if (!metrics.value.length) return {}
  const keys = ['passenger_intensity', 'peak_load_factor', 'off_peak_load_factor', 'operating_speed', 'on_time_rate']
  const avgs = {}
  keys.forEach(k => {
    avgs[k] = metrics.value.reduce((s, x) => s + (x[k] || 0), 0) / metrics.value.length
  })
  return avgs
})

const getLevel = (row, key) => {
  const avg = averages.value[key]
  if (!avg) return 'metric-yellow'
  if (row[key] > avg * 1.2) return 'metric-green'
  if (row[key] < avg * 0.7) return 'metric-red'
  return 'metric-yellow'
}

const goToDetail = (row) => {
  router.push(`/metrics/${row.line_no}`)
}

onMounted(async () => {
  metrics.value = await api.getLineEfficiencies()
})
</script>

<style scoped>
.metrics-legend {
  display: flex;
  gap: 20px;
  margin-bottom: 16px;
  font-size: 13px;
  color: #606266;
}
.legend-item {
  display: flex;
  align-items: center;
  gap: 6px;
}
.dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}
.dot.green { background: #67c23a; }
.dot.yellow { background: #e6a23c; }
.dot.red { background: #f56c6c; }
</style>
