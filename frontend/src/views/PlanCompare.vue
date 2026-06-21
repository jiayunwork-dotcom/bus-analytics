<template>
  <div class="page-container">
    <template v-if="isTrendMode">
      <div v-if="!trendData || !trendData.length" class="section-card" style="text-align: center; padding: 60px 20px;">
        <el-empty description="暂无历史方案数据" />
      </div>

      <template v-else>
        <div class="section-card">
          <div class="section-title">
            历史趋势回溯
            <span style="margin-left: 12px; font-size: 13px; color: #909399; font-weight: normal;">
              当前方案: {{ currentPlan?.name }}
            </span>
          </div>
          <div ref="trendChart" class="trend-chart-container"></div>
        </div>

        <div class="section-card">
          <div class="section-title">方案信息</div>
          <el-row :gutter="16">
            <el-col :span="6">
              <div class="info-item">
                <div class="info-label">方案名称</div>
                <div class="info-value">{{ currentPlan?.name }}</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="info-item">
                <div class="info-label">创建时间</div>
                <div class="info-value">{{ formatDateTime(currentPlan?.created_at) }}</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="info-item">
                <div class="info-label">日总班次</div>
                <div class="info-value">{{ currentPlan?.daily_trips }} 班</div>
              </div>
            </el-col>
            <el-col :span="6">
              <div class="info-item">
                <div class="info-label">高峰满载率</div>
                <div class="info-value">{{ currentPlan ? (currentPlan.peak_load_factor * 100).toFixed(1) + '%' : '-' }}</div>
              </div>
            </el-col>
            <el-col :span="6" style="margin-top: 16px;">
              <div class="info-item">
                <div class="info-label">营运速度</div>
                <div class="info-value">{{ currentPlan ? currentPlan.operating_speed.toFixed(2) + ' km/h' : '-' }}</div>
              </div>
            </el-col>
            <el-col :span="6" style="margin-top: 16px;">
              <div class="info-item">
                <div class="info-label">客运强度</div>
                <div class="info-value">{{ currentPlan ? currentPlan.passenger_intensity.toFixed(2) + ' 人/km' : '-' }}</div>
              </div>
            </el-col>
          </el-row>
        </div>

        <div class="section-card kpi-compliance-card" :class="isAllCompliant ? 'compliant' : 'non-compliant'">
          <div class="compliance-title">
            <el-icon v-if="isAllCompliant" class="compliance-icon"><Check /></el-icon>
            <el-icon v-else class="compliance-icon"><Close /></el-icon>
            <span>{{ isAllCompliant ? '方案达标' : '方案未达标' }}</span>
          </div>
          <div class="compliance-items">
            <div class="compliance-item" :class="{ pass: kpiCompliance.dailyTrips.pass, fail: !kpiCompliance.dailyTrips.pass }">
              <el-icon class="item-icon"><Check v-if="kpiCompliance.dailyTrips.pass" /><Close v-else /></el-icon>
              <span class="item-label">日总班次</span>
              <span class="item-value">{{ currentPlan?.daily_trips }} 班</span>
              <span class="item-threshold">(≥ 50班)</span>
            </div>
            <div class="compliance-item" :class="{ pass: kpiCompliance.peakLoadFactor.pass, fail: !kpiCompliance.peakLoadFactor.pass }">
              <el-icon class="item-icon"><Check v-if="kpiCompliance.peakLoadFactor.pass" /><Close v-else /></el-icon>
              <span class="item-label">高峰满载率</span>
              <span class="item-value">{{ currentPlan ? (currentPlan.peak_load_factor * 100).toFixed(1) + '%' : '-' }}</span>
              <span class="item-threshold">(≤ 85%)</span>
            </div>
            <div class="compliance-item" :class="{ pass: kpiCompliance.operatingSpeed.pass, fail: !kpiCompliance.operatingSpeed.pass }">
              <el-icon class="item-icon"><Check v-if="kpiCompliance.operatingSpeed.pass" /><Close v-else /></el-icon>
              <span class="item-label">营运速度</span>
              <span class="item-value">{{ currentPlan ? currentPlan.operating_speed.toFixed(2) + ' km/h' : '-' }}</span>
              <span class="item-threshold">(≥ 12 km/h)</span>
            </div>
            <div class="compliance-item" :class="{ pass: kpiCompliance.passengerIntensity.pass, fail: !kpiCompliance.passengerIntensity.pass }">
              <el-icon class="item-icon"><Check v-if="kpiCompliance.passengerIntensity.pass" /><Close v-else /></el-icon>
              <span class="item-label">客运强度</span>
              <span class="item-value">{{ currentPlan ? currentPlan.passenger_intensity.toFixed(2) + ' 人/km' : '-' }}</span>
              <span class="item-threshold">(≥ 20 人/km)</span>
            </div>
          </div>
          <div v-if="!isAllCompliant && failedItems.length" class="failed-list">
            未达标项: {{ failedItems.join('、') }}
          </div>
        </div>
      </template>
    </template>

    <template v-else>
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
    </template>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import { Check, Close } from '@element-plus/icons-vue'
import api from '../api'
import * as echarts from 'echarts'

const route = useRoute()
const data = ref(null)
const kpiChart = ref(null)
const planIds = ref([])

const trendData = ref([])
const trendChart = ref(null)
const currentPlanId = ref(null)

const planColors = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c']
const trendColors = {
  dailyTrips: '#409eff',
  peakLoadFactor: '#f56c6c',
  operatingSpeed: '#67c23a',
  passengerIntensity: '#e6a23c'
}

const isTrendMode = computed(() => route.query.mode === 'trend')

const currentPlan = computed(() => {
  if (!trendData.value || !trendData.value.length) return null
  return trendData.value.find(p => p.id === currentPlanId.value) || trendData.value[trendData.value.length - 1]
})

const formatDateTime = (isoStr) => {
  if (!isoStr) return '-'
  const d = new Date(isoStr)
  if (isNaN(d.getTime())) return isoStr
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const formatDateMD = (isoStr) => {
  if (!isoStr) return ''
  const d = new Date(isoStr)
  if (isNaN(d.getTime())) return ''
  const pad = (n) => String(n).padStart(2, '0')
  return `${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

const kpiCompliance = computed(() => {
  if (!currentPlan.value) {
    return {
      dailyTrips: { pass: false, value: 0 },
      peakLoadFactor: { pass: false, value: 0 },
      operatingSpeed: { pass: false, value: 0 },
      passengerIntensity: { pass: false, value: 0 }
    }
  }
  const p = currentPlan.value
  return {
    dailyTrips: { pass: p.daily_trips >= 50, value: p.daily_trips },
    peakLoadFactor: { pass: p.peak_load_factor <= 0.85, value: p.peak_load_factor },
    operatingSpeed: { pass: p.operating_speed >= 12, value: p.operating_speed },
    passengerIntensity: { pass: p.passenger_intensity >= 20, value: p.passenger_intensity }
  }
})

const isAllCompliant = computed(() => {
  const c = kpiCompliance.value
  return c.dailyTrips.pass && c.peakLoadFactor.pass && c.operatingSpeed.pass && c.passengerIntensity.pass
})

const failedItems = computed(() => {
  const items = []
  const c = kpiCompliance.value
  if (!c.dailyTrips.pass) items.push('日总班次')
  if (!c.peakLoadFactor.pass) items.push('高峰满载率')
  if (!c.operatingSpeed.pass) items.push('营运速度')
  if (!c.passengerIntensity.pass) items.push('客运强度')
  return items
})

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

const renderTrendChart = () => {
  if (!trendChart.value || !trendData.value || !trendData.value.length) return

  const chart = echarts.init(trendChart.value)
  const plans = trendData.value
  const dates = plans.map(p => formatDateMD(p.created_at))
  const currentIdx = plans.findIndex(p => p.id === currentPlanId.value)

  const buildSeries = (name, key, color, suffix, isPercent = false) => {
    const data = plans.map((p, idx) => {
      let val = p[key]
      if (isPercent) val = +(val * 100).toFixed(1)
      else val = +val.toFixed(2)
      return {
        value: val,
        itemStyle: {
          color: color,
          borderWidth: idx === currentIdx ? 3 : 0,
          borderColor: '#fff'
        },
        symbolSize: idx === currentIdx ? 16 : 8,
        plan: p
      }
    })

    return {
      name: name,
      type: 'line',
      smooth: true,
      data: data,
      lineStyle: { color: color, width: 2 },
      itemStyle: { color: color },
      symbol: 'circle',
      emphasis: {
        itemStyle: { borderWidth: 2, borderColor: '#fff' }
      }
    }
  }

  const series = [
    buildSeries('日总班次数', 'daily_trips', trendColors.dailyTrips, ' 班'),
    buildSeries('高峰满载率', 'peak_load_factor', trendColors.peakLoadFactor, '%', true),
    buildSeries('营运速度', 'operating_speed', trendColors.operatingSpeed, ' km/h'),
    buildSeries('客运强度', 'passenger_intensity', trendColors.passengerIntensity, ' 人/km')
  ]

  chart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: (params) => {
        if (!params || !params.length) return ''
        const p = params[0].data.plan
        let html = `<div style="font-weight: 600; margin-bottom: 6px;">${p.name}</div>`
        html += `<div style="font-size: 12px; color: #909399; margin-bottom: 8px;">${formatDateTime(p.created_at)}</div>`
        params.forEach(param => {
          let suffix = ''
          if (param.seriesName === '高峰满载率') suffix = '%'
          else if (param.seriesName === '营运速度') suffix = ' km/h'
          else if (param.seriesName === '客运强度') suffix = ' 人/km'
          else if (param.seriesName === '日总班次数') suffix = ' 班'
          html += `${param.marker} ${param.seriesName}: <b>${param.value}</b>${suffix}<br/>`
        })
        return html
      }
    },
    legend: {
      data: ['日总班次数', '高峰满载率', '营运速度', '客运强度'],
      top: 10
    },
    grid: {
      left: 60,
      right: 60,
      top: 60,
      bottom: 40
    },
    xAxis: {
      type: 'category',
      data: dates,
      boundaryGap: false,
      axisLabel: {
        fontSize: 12
      }
    },
    yAxis: {
      type: 'value',
      axisLabel: {
        fontSize: 12
      }
    },
    dataZoom: [
      {
        type: 'inside',
        start: 0,
        end: 100
      }
    ],
    series: series
  })
}

const handleResize = () => {
  setTimeout(() => {
    try {
      if (isTrendMode.value) {
        trendChart.value && echarts.getInstanceByDom(trendChart.value)?.resize()
      } else {
        kpiChart.value && echarts.getInstanceByDom(kpiChart.value)?.resize()
      }
    } catch (e) {}
  }, 100)
}

const loadTrendData = async (planId) => {
  try {
    trendData.value = await api.getPlanHistory(planId)
    await nextTick()
    renderTrendChart()
  } catch (e) {
    console.error('加载趋势数据失败', e)
  }
}

const loadCompareData = async (ids) => {
  try {
    data.value = await api.comparePlans(ids)
    await nextTick()
    renderKpiChart()
  } catch (e) {}
}

onMounted(async () => {
  const idsParam = route.query.ids
  if (!idsParam) return
  const ids = String(idsParam).split(',').map(Number).filter(n => !isNaN(n))
  planIds.value = ids

  if (isTrendMode.value && ids.length === 1) {
    currentPlanId.value = ids[0]
    loadTrendData(ids[0])
  } else if (ids.length >= 2) {
    loadCompareData(ids)
  }

  window.addEventListener('resize', handleResize)
})

onBeforeUnmount(() => {
  window.removeEventListener('resize', handleResize)
  if (kpiChart.value) {
    const chart = echarts.getInstanceByDom(kpiChart.value)
    chart?.dispose()
  }
  if (trendChart.value) {
    const chart = echarts.getInstanceByDom(trendChart.value)
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

.trend-chart-container {
  width: 100%;
  height: 420px;
}

.info-item {
  padding: 12px 16px;
  background: #f5f7fa;
  border-radius: 6px;
}

.info-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.info-value {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.kpi-compliance-card {
  border-radius: 8px;
  padding: 20px 24px;
  transition: background 0.3s;
}

.kpi-compliance-card.compliant {
  background: #f0f9eb;
  border: 1px solid #c2e7b0;
}

.kpi-compliance-card.non-compliant {
  background: #fef0f0;
  border: 1px solid #fbc4c4;
}

.compliance-title {
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

.compliant .compliance-title {
  color: #67c23a;
}

.non-compliant .compliance-title {
  color: #f56c6c;
}

.compliance-icon {
  font-size: 22px;
}

.compliance-items {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 32px;
  margin-bottom: 12px;
}

.compliance-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 14px;
}

.compliance-item.pass {
  color: #67c23a;
}

.compliance-item.fail {
  color: #f56c6c;
}

.item-icon {
  font-size: 16px;
}

.item-label {
  font-weight: 500;
}

.item-value {
  font-weight: 600;
}

.item-threshold {
  font-size: 12px;
  opacity: 0.8;
}

.failed-list {
  font-size: 13px;
  color: #f56c6c;
  margin-top: 8px;
  padding-top: 12px;
  border-top: 1px dashed #fbc4c4;
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
