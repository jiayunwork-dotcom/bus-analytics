<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">发车间隔优化</div>
      <el-row :gutter="16" style="margin-bottom: 20px;">
        <el-col :span="5">
          <el-select v-model="selectedLine" placeholder="选择线路" filterable>
            <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
          </el-select>
        </el-col>
        <el-col :span="5">
          <el-date-picker v-model="selectedDate" type="date" placeholder="选择参考日期" value-format="YYYY-MM-DD" />
        </el-col>
        <el-col :span="4">
          <el-input-number v-model="totalVehicles" :min="1" :max="100" label="可用车辆" />
        </el-col>
        <el-col :span="4">
          <el-button type="primary" @click="optimize">计算优化方案</el-button>
        </el-col>
      </el-row>

      <el-alert
        title="优化说明"
        type="info"
        :closable="false"
        style="margin-bottom: 20px;"
      >
        <p>优化目标：最小化乘客平均等待时间（平均等待时间 ≈ 发车间隔/2）</p>
        <p>约束条件：最小间隔3分钟、最大间隔30分钟、高峰间隔 ≤ 平峰间隔</p>
      </el-alert>
    </div>

    <div v-if="result" class="section-card">
      <div class="section-title">优化结果概览</div>
      <el-row :gutter="20">
        <el-col :span="6">
          <div class="result-card">
            <div class="result-label">推荐总班次数</div>
            <div class="result-value">{{ result.total_trips }} 班</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="result-card">
            <div class="result-label">现行平均等待</div>
            <div class="result-value">{{ result.current_avg_wait_time?.toFixed(1) }} 分钟</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="result-card">
            <div class="result-label">优化后平均等待</div>
            <div class="result-value primary">{{ result.optimized_avg_wait_time?.toFixed(1) }} 分钟</div>
          </div>
        </el-col>
        <el-col :span="6">
          <div class="result-card">
            <div class="result-label">预计改善</div>
            <div class="result-value" :class="result.improvement_percent > 0 ? 'green' : ''">
              {{ result.improvement_percent > 0 ? '+' : '' }}{{ result.improvement_percent?.toFixed(1) }}%
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <div v-if="result" class="section-card">
      <div class="section-title">各时段优化方案详情</div>
      <el-table :data="result.optimizations" border stripe>
        <el-table-column prop="period_name" label="时段" width="100" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.is_peak" type="danger" effect="dark">{{ row.period_name }}</el-tag>
            <el-tag v-else type="success">{{ row.period_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="time_range" label="时间范围" width="140" align="center" />
        <el-table-column label="现行间隔 (分钟)" align="center">
          <template #default="{ row }">{{ row.current_interval || '-' }}</template>
        </el-table-column>
        <el-table-column label="推荐间隔 (分钟)" align="center">
          <template #default="{ row }">
            <strong :class="row.current_interval !== row.recommended_interval ? 'metric-green' : ''">
              {{ row.recommended_interval }}
            </strong>
          </template>
        </el-table-column>
        <el-table-column prop="recommended_trips" label="推荐班次" width="120" align="center" />
        <el-table-column label="间隔变化" width="150" align="center">
          <template #default="{ row }">
            <span v-if="row.current_interval && row.current_interval !== row.recommended_interval">
              <el-tag v-if="row.recommended_interval < row.current_interval" type="success" size="small">
                ↓ {{ row.current_interval - row.recommended_interval }}分钟 (加密)
              </el-tag>
              <el-tag v-else type="warning" size="small">
                ↑ {{ row.recommended_interval - row.current_interval }}分钟 (拉大)
              </el-tag>
            </span>
            <span v-else class="no-change">无需调整</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="result" class="section-card">
      <div class="section-title">间隔对比可视化</div>
      <div ref="compareChart" class="chart-container"></div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, nextTick } from 'vue'
import api from '../api'
import * as echarts from 'echarts'

const lines = ref([])
const selectedLine = ref('')
const selectedDate = ref('')
const totalVehicles = ref(20)
const result = ref(null)
const compareChart = ref(null)

const optimize = async () => {
  if (!selectedLine.value || !selectedDate.value) return
  result.value = await api.optimizeSchedule(selectedLine.value, {
    date: selectedDate.value,
    total_vehicles: totalVehicles.value
  })
  await nextTick()
  renderCompareChart()
}

const renderCompareChart = () => {
  if (!compareChart.value || !result.value) return
  const chart = echarts.init(compareChart.value)
  const names = result.value.optimizations.map(o => o.period_name + '\n' + o.time_range)
  chart.setOption({
    tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
    legend: { data: ['现行间隔', '推荐间隔'] },
    xAxis: { type: 'category', data: names },
    yAxis: { type: 'value', name: '分钟', min: 0 },
    series: [
      {
        name: '现行间隔',
        type: 'bar',
        data: result.value.optimizations.map(o => o.current_interval),
        itemStyle: { color: '#909399' },
        label: { show: true, position: 'top', formatter: '{c}分' }
      },
      {
        name: '推荐间隔',
        type: 'bar',
        data: result.value.optimizations.map(o => o.recommended_interval),
        itemStyle: { color: '#409eff' },
        label: { show: true, position: 'top', formatter: '{c}分' }
      }
    ]
  })
}

onMounted(async () => {
  lines.value = await api.getLines()
  const dr = await api.getDateRange()
  if (lines.value.length) selectedLine.value = lines.value[0].line_no
  if (dr.max_date) selectedDate.value = dr.max_date
})
</script>

<style scoped>
.result-card {
  text-align: center;
  padding: 24px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
  border-radius: 8px;
}
.result-label {
  color: #606266;
  font-size: 14px;
  margin-bottom: 8px;
}
.result-value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
}
.result-value.primary { color: #409eff; }
.result-value.green { color: #67c23a; }
.no-change {
  color: #909399;
  font-size: 13px;
}
</style>
