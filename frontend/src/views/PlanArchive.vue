<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title">
        方案存档
        <template v-if="selectedPlans.length === 1">
          <el-button
            type="success"
            style="margin-left: 24px;"
            @click="goTrend"
          >
            查看趋势
          </el-button>
        </template>
        <template v-else>
          <el-button
            type="primary"
            :disabled="selectedPlans.length < 2 || selectedPlans.length > 4"
            style="margin-left: 24px;"
            @click="goCompare"
          >
            开始对比
          </el-button>
        </template>
        <span v-if="selectedPlans.length > 0" style="margin-left: 12px; color: #909399; font-size: 13px;">
          已选 {{ selectedPlans.length }}/4 个方案
        </span>
      </div>

      <div class="filter-bar">
        <el-select
          v-model="filterLine"
          placeholder="按线路筛选"
          clearable
          filterable
          style="width: 200px;"
          @change="loadPlans"
        >
          <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
        </el-select>
        <el-button :icon="sortOrder === 'desc' ? SortDown : SortUp" @click="toggleSort" style="margin-left: 12px;">
          {{ sortOrder === 'desc' ? '最新优先' : '最早优先' }}
        </el-button>
      </div>

      <el-table
        :data="plans"
        border
        stripe
        @selection-change="onSelectionChange"
        style="margin-top: 16px;"
      >
        <el-table-column type="selection" width="50" align="center" />
        <el-table-column label="方案名称" min-width="180">
          <template #default="{ row }">
            <span class="plan-name" @click="goPlanDetail(row.id)">{{ row.name }}</span>
          </template>
        </el-table-column>
        <el-table-column label="模拟类型" width="120" align="center">
          <template #default="{ row }">
            <el-badge :type="row.sim_type === 'single' ? '' : 'warning'" is-dot>
              <el-tag :type="row.sim_type === 'single' ? '' : 'warning'" size="small">
                {{ row.sim_type === 'single' ? '单线路' : '联合' }}
              </el-tag>
            </el-badge>
          </template>
        </el-table-column>
        <el-table-column label="涉及线路" min-width="160">
          <template #default="{ row }">
            <el-tag
              v-for="ln in row.lines.split(',')"
              :key="ln"
              size="small"
              style="margin: 2px;"
            >
              {{ ln }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="daily_trips" label="日总班次" width="100" align="center" />
        <el-table-column label="高峰满载率" width="120" align="center">
          <template #default="{ row }">
            {{ row.peak_load_factor != null ? (row.peak_load_factor * 100).toFixed(1) + '%' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="营运速度" width="120" align="center">
          <template #default="{ row }">
            {{ row.operating_speed != null ? row.operating_speed.toFixed(2) + ' km/h' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="客运强度" width="120" align="center">
          <template #default="{ row }">
            {{ row.passenger_intensity != null ? row.passenger_intensity.toFixed(2) + ' 人/km' : '-' }}
          </template>
        </el-table-column>
        <el-table-column label="创建时间" width="180" align="center">
          <template #default="{ row }">{{ formatDateTime(row.created_at) }}</template>
        </el-table-column>
        <el-table-column label="操作" width="160" align="center" fixed="right">
          <template #default="{ row }">
            <el-popover
              placement="left"
              :width="260"
              trigger="click"
              :visible="renamePopoverVisible[row.id]"
              @update:visible="(v) => renamePopoverVisible[row.id] = v"
            >
              <template #reference>
                <el-button type="primary" link size="small">重命名</el-button>
              </template>
              <div style="display: flex; align-items: center; gap: 8px;">
                <el-input
                  v-model="renameValue"
                  size="default"
                  placeholder="输入新名称"
                  style="flex: 1;"
                  @keyup.enter="confirmRename(row.id)"
                />
                <el-button type="primary" size="default" @click="confirmRename(row.id)">确定</el-button>
              </div>
            </el-popover>
            <el-button type="danger" link size="small" @click="handleDelete(row.id)" style="margin-left: 8px;">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { SortDown, SortUp } from '@element-plus/icons-vue'
import api from '../api'

const router = useRouter()
const plans = ref([])
const lines = ref([])
const sortOrder = ref('desc')
const filterLine = ref('')
const selectedPlans = ref([])
const renamePopoverVisible = reactive({})
const renameValue = ref('')

const formatDateTime = (isoStr) => {
  if (!isoStr) return '-'
  const d = new Date(isoStr)
  if (isNaN(d.getTime())) return isoStr
  const pad = (n) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`
}

const loadPlans = async () => {
  const params = { sort: sortOrder.value }
  if (filterLine.value) params.line = filterLine.value
  plans.value = await api.listPlans(params)
}

const loadLines = async () => {
  lines.value = await api.getLines()
}

const toggleSort = () => {
  sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
  loadPlans()
}

const onSelectionChange = (selection) => {
  selectedPlans.value = selection
}

const goPlanDetail = (id) => {
  router.push({ path: '/plan-compare', query: { ids: String(id) } })
}

const goCompare = () => {
  const ids = selectedPlans.value.map(p => p.id).join(',')
  router.push({ path: '/plan-compare', query: { ids } })
}

const goTrend = () => {
  if (selectedPlans.value.length !== 1) return
  const id = selectedPlans.value[0].id
  router.push({ path: '/plan-compare', query: { ids: String(id), mode: 'trend' } })
}

const handleDelete = async (id) => {
  try {
    await ElMessageBox.confirm('确定删除该方案？删除后不可恢复。', '删除确认', {
      type: 'warning',
      confirmButtonText: '确定',
      cancelButtonText: '取消'
    })
    await api.deletePlan(id)
    ElMessage.success('删除成功')
    loadPlans()
  } catch {}
}

const confirmRename = async (id) => {
  const name = renameValue.value.trim()
  if (!name) {
    ElMessage.warning('名称不能为空')
    return
  }
  await api.renamePlan(id, name)
  ElMessage.success('重命名成功')
  renamePopoverVisible[id] = false
  renameValue.value = ''
  loadPlans()
}

onMounted(() => {
  loadLines()
  loadPlans()
})
</script>

<style scoped>
.filter-bar {
  display: flex;
  align-items: center;
  margin-top: 4px;
}

.plan-name {
  color: #409eff;
  cursor: pointer;
  font-weight: 500;
}

.plan-name:hover {
  text-decoration: underline;
}
</style>
