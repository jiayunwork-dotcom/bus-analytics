<template>
  <div class="page-container">
    <div v-if="!data" class="section-card" style="text-align: center; padding: 60px 20px;">
      <el-empty :description="planIds.length < 2 ? '请选择2~4份方案进行对比' : '请从方案存档页选择方案进行对比'" />
    </div>

    <template v-else>
      <div class="section-card">
        <div class="section-title">方案概览</div>
        <el-row :gutter="16">
          <el-col v-for="(plan, idx) in data.plans" :key="plan.id" :span="Math.max(6, 24 / data.plans.length)">
            <div class="plan-summary-card" :style="{ borderLeftColor: planColors[idx % planColors.length] }">
              <div class="plan-summary-name">
                <span :style="{ color: planColors[idx % planColors.length] }">●</span>
                {{ plan.name }}
              </div>
              <div class="plan-summary-meta">{{ formatDateTime(plan.created_at) }}</div>
              <div class="plan-summary-lines">
                <el-tag
                  v-for="ln in plan.lines.split(',')"
                  :key="ln"
                  size="small"
                  style="margin: 2px;"
                >
                  {{ ln }}
                </el-tag>
              </div>
              <div style="margin-top: 8px;">
                <el-tag :type="plan.sim_type === 'single' ? 'info' : 'warning'" size="small">
                  {{ plan.sim_type === 'single' ? '单线路' : '联合' }}
                </el-tag>
              </div>
            </div>
          </el-col>
        </el-row>
      </div>

      <div class="section-card">
        <div class="section-title">核心指标对比</div>
        <div ref="kpiChart" class="chart-container"></div>
      </div>

      <div v-if="hasJoint" class="section-card">
        <div class="section-title">联合总览对比</div>
        <el-table :data="data.joint_overviews" border stripe>
          <el-table-column label="方案名称" min-width="140">
            <template #default="{ row }">{{ row.plan_name }}</template>
          </el-table-column>
          <el-table-column label="总运力变化(班次)" width="160" align="center">
            <template #default="{ row }">
              <span :class="deltaClass(row.total_trips_delta)">
                {{ row.total_trips_delta > 0 ? '+' : '' }}{{ row.total_trips_delta }}
              </span>
            </template>
          </el-table-column>
          <el-table-column label="总运力变化(%)" width="150" align="center">
            <template #default="{ row }">
              <span :class="deltaClass(row.total_trips_change_pct)">
                {{ row.total_trips_change_pct > 0 ? '+' : '' }}{{ row.total_trips_change_pct.toFixed(1) }}%
              </span>
            </template>
          </el-table-column>
          <el-table-column label="平均满载率(调整前)" width="180" align="center">
            <template #default="{ row }">{{ (row.avg_orig_load_factor * 100).toFixed(1) }}%</template>
          </el-table-column>
          <el-table-column label="平均满载率(调整后)" width="180" align="center">
            <template #default="{ row }">{{ (row.avg_new_load_factor * 100).toFixed(1) }}%</template>
          </el-table-column>
          <el-table-column label="满载率增量" width="140" align="center">
            <template #default="{ row }">
              <span :class="deltaClass(-row.avg_load_factor_delta, true)">
                {{ row.avg_load_factor_delta > 0 ? '+' : '' }}{{ (row.avg_load_factor_delta * 100).toFixed(2) }}%
              </span>
            </template>
          </el-table-column>
        </el-table>
      </div>

      <div v-if="data.param_diffs.length" class="section-card">
        <div class="section-title">参数差异对比</div>
        <el-table :data="data.param_diffs" border stripe class="param-diff-table" :cell-class-name="getParamCellClassName">
          <el-table-column prop="param_name" label="参数名称" width="180" class-name="param-name-col" />
          <el-table-column
            v-for="(plan, idx) in data.plans"
            :key="plan.id"
            :label="plan.name"
            align="center"
            min-width="140"
            :class-name="'plan-col-' + idx"
          >
            <template #default="{ row }">
              <div class="param-cell-text" :style="getParamCellStyle(row, plan.id, idx)">
                {{ getParamValue(row, plan.id) }}
              </div>
            </template>
          </el-table-column>
          <el-table-column label="最优方案推荐" min-width="220" align="center">
            <template #default="{ row, $index }">
              <template v-if="$index === data.param_diffs.length - 1 && data.recommendations.length">
                <template v-for="rec in data.recommendations" :key="rec.plan_id">
                  <div class="recommend-cell">
                    <el-icon color="#67c23a" style="vertical-align: middle; margin-right: 4px;"><Check /></el-icon>
                    <span>{{ rec.plan_name }}: {{ rec.reason }}</span>
                  </div>
                </template>
              </template>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import { Check } from '@element-plus/icons-vue'
import api from '../api'
import * as echarts from 'echarts'

const route = useRoute()
const data = ref(null)
const kpiChart = ref(null)
const planIds = ref([])

const planColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c']

const formatDateTime = (isoStr) => {
  if (!isoStr) return '-'
  const d = new Date(isoStr)
  if (isNaN(d.getTime())) return isoStr
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const hasJoint = computed(() => {
  if (!data.value) return false
  return data.value.plans.some(p => p.sim_type === 'joint')
})

const deltaClass = (delta, inverse = false) => {
  if (delta === 0) return ''
  const good = inverse ? delta < 0 : delta > 0
  return good ? 'metric-green' : 'metric-red'
}

const getParamValue = (row, planId) => {
  const v = row.values.find(x => x.plan_id === planId)
  return v ? v.value : '-'
}

const getParamCellClassName = ({ row, columnIndex }) => {
  if (!data.value) return ''
  const planCount = data.value.plans.length
  if (columnIndex < 1 || columnIndex > planCount) return ''
  const planIdx = columnIndex - 1
  const plan = data.value.plans[planIdx]
  if (!plan) return ''
  const v = row.values.find(x => x.plan_id === plan.id)
  if (!v || v.same) return 'param-cell-same'
  return 'param-cell-diff-' + (planIdx % planColors.length)
}

const getParamCellStyle = (row, planId, idx) => {
  const v = row.values.find(x => x.plan_id === planId)
  if (!v || v.same) return {}
  return { color: planColors[idx % planColors.length], fontWeight: 700 }
}

const renderKpiChart = () => {
  if (!kpiChart.value || !data.value) return
  const chart = echarts.init(kpiChart.value)
  const plans = data.value.plans
  const categories = ['日总班次数', '高峰满载率', '营运速度', '客运强度']
  const series = plans.map((plan, idx) => ({
    name: plan.name,
    type: 'bar',
    data: [
      plan.kpi.daily_trips,
      +(plan.kpi.peak_load_factor * 100).toFixed(1),
      +plan.kpi.operating_speed.toFixed(2),
      +plan.kpi.passenger_intensity.toFixed(2)
    ],
    itemStyle: { color: planColors[idx % planColors.length] }
  }))
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'shadow' },
      formatter: (params) => {
        let html = params[0].name + '<br/>'
        params.forEach(p => {
          let suffix = ''
          if (p.name === '高峰满载率') suffix = '%'
          else if (p.name === '营运速度') suffix = ' km/h'
          else if (p.name === '客运强度') suffix = ' 人/km'
          else if (p.name === '日总班次数') suffix = ' 班'
          html += p.marker + ' ' + p.seriesName + ': ' + p.value + suffix + '<br/>'
        })
        return html
      }
    },
    legend: { data: plans.map(p => p.name) },
    grid: { left: 60, right: 40, top: 50, bottom: 40 },
    xAxis: { type: 'category', data: categories },
    yAxis: { type: 'value' },
    series
  })
}

const handleResize = () => {
  setTimeout(() => {
    try {
      kpiChart.value && echarts.getInstanceByDom(kpiChart.value)?.resize()
    } catch (e) {}
  }, 100)
}

onMounted(async () => {
  const idsParam = route.query.ids
  if (!idsParam) return
  const ids = String(idsParam).split(',').map(Number).filter(n => !isNaN(n))
  planIds.value = ids
  if (ids.length < 2) return
  try {
    data.value = await api.comparePlans(ids)
    await nextTick()
    renderKpiChart()
  } catch (e) {}
  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (kpiChart.value) {
    const chart = echarts.getInstanceByDom(kpiChart.value)
    chart?.dispose()
  }
})
</script>

<style scoped>
.plan-summary-card {
  padding: 16px 20px;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  border-left: 4px solid #409eff;
  background: #fff;
}

.plan-summary-name {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 6px;
}

.plan-summary-meta {
  font-size: 12px;
  color: #909399;
  margin-bottom: 8px;
}

.plan-summary-lines {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
}

.param-diff-table :deep(.param-cell-same) {
  background: #f5f7fa !important;
  color: #909399 !important;
}

.param-diff-table :deep(.param-cell-diff-0) {
  background: rgba(64, 158, 255, 0.1) !important;
}

.param-diff-table :deep(.param-cell-diff-1) {
  background: rgba(103, 194, 58, 0.1) !important;
}

.param-diff-table :deep(.param-cell-diff-2) {
  background: rgba(230, 162, 60, 0.1) !important;
}

.param-diff-table :deep(.param-cell-diff-3) {
  background: rgba(245, 108, 108, 0.1) !important;
}

.param-cell-text {
  padding: 2px 0;
}

.recommend-cell {
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  color: #67c23a;
  font-weight: 500;
}
</style>
