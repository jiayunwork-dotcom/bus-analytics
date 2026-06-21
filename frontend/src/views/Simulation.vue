<template>
  <div class="page-container">
    <div class="section-card">
      <div class="section-title sim-section-title">
        <span class="title-text">线路调整模拟</span>
        <el-radio-group v-model="simMode" size="default" style="margin-left: 24px;" @change="onModeChange">
          <el-radio-button value="single">单线路调整</el-radio-button>
          <el-radio-button value="joint">多线路联合调整</el-radio-button>
        </el-radio-group>
        <el-button
          type="success"
          :icon="FolderOpened"
          style="margin-left: auto;"
          @click="openSaveDialog(simMode)"
          :disabled="!hasCurrentResult"
        >
          保存当前方案
        </el-button>
      </div>

      <!-- 单线路模式 -->
      <el-row v-if="simMode === 'single'" :gutter="24">
        <el-col :span="12">
          <div class="panel-title">模拟参数</div>

          <el-form :model="form" label-width="140px" label-position="left">
            <el-form-item label="目标线路">
              <el-select v-model="form.line_no" placeholder="选择线路" filterable style="width: 100%;" @change="onLineChange">
                <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
              </el-select>
            </el-form-item>

            <el-form-item label="参考日期">
              <el-date-picker v-model="form.date" type="date" placeholder="选择参考日期" value-format="YYYY-MM-DD" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="高峰发车间隔(分)">
              <el-input-number v-model="form.peak_interval" :min="3" :max="30" :step="1" :controls-position="'right'" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="平峰发车间隔(分)">
              <el-input-number v-model="form.off_peak_interval" :min="3" :max="30" :step="1" :controls-position="'right'" style="width: 100%;" />
            </el-form-item>

            <el-form-item label="站点增减数">
              <el-input-number v-model="form.station_delta" :min="-5" :max="5" :step="1" :controls-position="'right'" style="width: 100%;" />
              <div class="form-tip">正数=新增站点，负数=裁撤站点（范围 -5 ~ 5）</div>
            </el-form-item>

            <el-form-item>
              <el-button type="primary" :icon="RefreshRight" @click="runSingleSimulation" :loading="loading">运行模拟</el-button>
              <el-button @click="resetSingleForm">重置参数</el-button>
            </el-form-item>
          </el-form>
        </el-col>

        <el-col :span="12">
          <div class="panel-title">调整前后对比预览</div>

          <el-row :gutter="12">
            <el-col :span="12">
              <div class="compare-card current">
                <div class="compare-card-title">当前值</div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">站点数</span>
                  <span class="compare-value">{{ singleResult.orig_station_count }} 站</span>
                </div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">高峰间隔</span>
                  <span class="compare-value">{{ singleResult.orig_peak_interval }} 分</span>
                </div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">平峰间隔</span>
                  <span class="compare-value">{{ singleResult.orig_off_peak_interval }} 分</span>
                </div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">日总班次</span>
                  <span class="compare-value">{{ singleResult.orig_total_trips }} 班</span>
                </div>
                <div class="compare-item" v-if="!singleResult">
                  <span class="placeholder">请先运行模拟</span>
                </div>
              </div>
            </el-col>
            <el-col :span="12">
              <div class="compare-card sim">
                <div class="compare-card-title">模拟值</div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">站点数</span>
                  <span class="compare-value" :class="deltaClass(singleResult.new_station_count - singleResult.orig_station_count)">
                    {{ singleResult.new_station_count }} 站
                    <span v-if="singleResult.new_station_count !== singleResult.orig_station_count" class="delta">
                      ({{ singleResult.new_station_count > singleResult.orig_station_count ? '+' : '' }}{{ singleResult.new_station_count - singleResult.orig_station_count }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">高峰间隔</span>
                  <span class="compare-value" :class="deltaClass(singleResult.new_peak_interval - singleResult.orig_peak_interval, true)">
                    {{ singleResult.new_peak_interval }} 分
                    <span v-if="singleResult.new_peak_interval !== singleResult.orig_peak_interval" class="delta">
                      ({{ singleResult.new_peak_interval > singleResult.orig_peak_interval ? '+' : '' }}{{ singleResult.new_peak_interval - singleResult.orig_peak_interval }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">平峰间隔</span>
                  <span class="compare-value" :class="deltaClass(singleResult.new_off_peak_interval - singleResult.orig_off_peak_interval, true)">
                    {{ singleResult.new_off_peak_interval }} 分
                    <span v-if="singleResult.new_off_peak_interval !== singleResult.orig_off_peak_interval" class="delta">
                      ({{ singleResult.new_off_peak_interval > singleResult.orig_off_peak_interval ? '+' : '' }}{{ singleResult.new_off_peak_interval - singleResult.orig_off_peak_interval }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="singleResult">
                  <span class="compare-label">日总班次</span>
                  <span class="compare-value" :class="deltaClass(singleResult.trips_delta)">
                    {{ singleResult.new_total_trips }} 班
                    <span v-if="singleResult.trips_delta !== 0" class="delta">
                      ({{ singleResult.trips_delta > 0 ? '+' : '' }}{{ singleResult.trips_delta }})
                    </span>
                  </span>
                </div>
                <div class="compare-item" v-if="!singleResult">
                  <span class="placeholder">请先运行模拟</span>
                </div>
              </div>
            </el-col>
          </el-row>

          <div class="capacity-box" v-if="singleResult">
            <el-row :gutter="12">
              <el-col :span="12">
                <div class="capacity-label">高峰运力变化</div>
                <div class="capacity-value" :class="deltaClass(singleResult.peak_capacity_change)">
                  {{ singleResult.peak_capacity_change > 0 ? '+' : '' }}{{ singleResult.peak_capacity_change?.toFixed(1) }}%
                </div>
              </el-col>
              <el-col :span="12">
                <div class="capacity-label">可用车辆数</div>
                <div class="capacity-value">
                  {{ singleResult.available_vehicles }} 辆
                </div>
              </el-col>
            </el-row>
            <el-alert v-if="singleResult.capacity_warning" title="运力不足警告：高峰班次超过可用车辆数！" type="error" :closable="false" style="margin-top: 12px;" show-icon />
          </div>
        </el-col>
      </el-row>

      <!-- 多线路联合模式 -->
      <div v-else>
        <el-form label-width="140px" label-position="left">
          <el-row :gutter="24">
            <el-col :span="8">
              <el-form-item label="参考日期">
                <el-date-picker v-model="jointForm.date" type="date" placeholder="选择参考日期" value-format="YYYY-MM-DD" style="width: 100%;" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="选择线路 (2-3条)">
                <el-select
                  v-model="jointForm.selected_line_nos"
                  multiple
                  filterable
                  collapse-tags
                  collapse-tags-tooltip
                  :multiple-limit="3"
                  placeholder="请选择2~3条线路"
                  style="width: 100%;"
                  @change="onJointLinesChange"
                >
                  <el-option v-for="l in lines" :key="l.line_no" :label="l.line_name" :value="l.line_no" />
                </el-select>
                <div class="form-tip">最多选择3条线路进行联合模拟</div>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="操作">
                <el-button type="primary" :icon="RefreshRight" @click="runJointSimulation" :loading="loading" :disabled="jointForm.selected_line_nos.length < 1">
                  运行联合模拟
                </el-button>
                <el-button @click="resetJointForm">重置参数</el-button>
              </el-form-item>
            </el-col>
          </el-row>
        </el-form>

        <div v-if="jointLineParamList.length > 0" style="margin-top: 8px;">
          <div class="panel-title">各线路独立参数</div>
          <el-collapse v-model="jointActivePanels" accordion style="border: none;">
            <el-collapse-item
              v-for="(lp, idx) in jointLineParamList"
              :key="lp.line_no"
              :name="lp.line_no"
            >
              <template #title>
                <div class="card-title-row">
                  <el-tag
                    :color="getLineColorBg(idx)"
                    effect="light"
                    size="small"
                    style="margin-right: 10px;"
                  >
                    <span :style="{ color: getLineColor(idx) }">●</span>
                    线路 {{ idx + 1 }}
                  </el-tag>
                  <span style="font-weight: 600;">{{ lp.line_no }} - {{ getLineName(lp.line_no) }}</span>
                  <span style="margin-left: auto; color: #909399; font-size: 12px; font-weight: normal;">
                    {{ lp.station_delta > 0 ? `+${lp.station_delta}站` : (lp.station_delta < 0 ? `${lp.station_delta}站` : '站点不变') }}
                    &nbsp;|&nbsp;高峰 {{ lp.peak_interval }}分 / 平峰 {{ lp.off_peak_interval }}分
                  </span>
                </div>
              </template>
              <div class="line-card-body">
                <el-form :model="lp" label-width="160px" label-position="left" inline>
                  <el-form-item label="高峰发车间隔(分)">
                    <el-input-number v-model="lp.peak_interval" :min="3" :max="30" :step="1" :controls-position="'right'" style="width: 200px;" />
                  </el-form-item>
                  <el-form-item label="平峰发车间隔(分)">
                    <el-input-number v-model="lp.off_peak_interval" :min="3" :max="30" :step="1" :controls-position="'right'" style="width: 200px;" />
                  </el-form-item>
                  <el-form-item label="站点增减数">
                    <el-input-number v-model="lp.station_delta" :min="-5" :max="5" :step="1" :controls-position="'right'" style="width: 200px;" />
                    <span class="form-tip" style="margin-left: 8px;">正数=新增，负数=裁撤（-5 ~ 5）</span>
                  </el-form-item>
                </el-form>
              </div>
            </el-collapse-item>
          </el-collapse>
        </div>

        <el-alert v-if="jointResult && jointResult.joint_overview?.has_vehicle_conflict" type="error" :closable="false" style="margin-top: 16px;" show-icon>
          <template #title>
            <span style="font-weight: 600;">共用车辆冲突：共 {{ jointResult.joint_overview.shared_vehicle_conflicts.length }} 辆车存在高峰班次超限</span>
          </template>
          <el-table :data="jointResult.joint_overview.shared_vehicle_conflicts" size="small" border style="margin-top: 8px;">
            <el-table-column prop="vehicle_no" label="车辆编号" width="120" align="center" />
            <el-table-column label="涉及线路" min-width="200">
              <template #default="{ row }">
                <el-tag v-for="ln in row.involved_lines" :key="ln" size="small" type="warning" style="margin: 2px;">{{ ln }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="total_peak_trips" label="调整后高峰班次合计" width="180" align="center" />
            <el-table-column prop="capacity_limit" label="承载能力上限" width="140" align="center" />
            <el-table-column label="状态" width="140" align="center">
              <template #default><el-tag type="danger" effect="dark">冲突</el-tag></template>
            </el-table-column>
          </el-table>
        </el-alert>

        <div v-if="jointResult" style="margin-top: 16px;">
          <el-row :gutter="16">
            <el-col :span="8">
              <div class="joint-overview-card">
                <div class="overview-label">总日班次</div>
                <div class="overview-value-row">
                  <span class="overview-orig">{{ jointResult.joint_overview.total_orig_trips }}</span>
                  <el-icon class="overview-arrow"><ArrowRight /></el-icon>
                  <span class="overview-new" :class="deltaClass(jointResult.joint_overview.total_trips_delta)">
                    {{ jointResult.joint_overview.total_new_trips }}
                    <span class="mini-delta" v-if="jointResult.joint_overview.total_trips_delta !== 0">
                      ({{ jointResult.joint_overview.total_trips_delta > 0 ? '+' : '' }}{{ jointResult.joint_overview.total_trips_delta }})
                    </span>
                  </span>
                </div>
                <div class="overview-pct" :class="deltaClass(jointResult.joint_overview.total_trips_change_pct)">
                  变化 {{ jointResult.joint_overview.total_trips_change_pct > 0 ? '+' : '' }}{{ jointResult.joint_overview.total_trips_change_pct?.toFixed(1) }}%
                </div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="joint-overview-card">
                <div class="overview-label">平均高峰满载率</div>
                <div class="overview-value-row">
                  <span class="overview-orig">{{ (jointResult.joint_overview.avg_orig_load_factor * 100).toFixed(1) }}%</span>
                  <el-icon class="overview-arrow"><ArrowRight /></el-icon>
                  <span class="overview-new" :class="deltaClass(-jointResult.joint_overview.avg_load_factor_delta, true)">
                    {{ (jointResult.joint_overview.avg_new_load_factor * 100).toFixed(1) }}%
                    <span class="mini-delta" v-if="jointResult.joint_overview.avg_load_factor_delta !== 0">
                      ({{ jointResult.joint_overview.avg_load_factor_delta > 0 ? '+' : '' }}{{ (jointResult.joint_overview.avg_load_factor_delta * 100).toFixed(2) }}%)
                    </span>
                  </span>
                </div>
                <div class="overview-pct" :class="deltaClass(-jointResult.joint_overview.avg_load_factor_delta, true)">
                  增量 {{ jointResult.joint_overview.avg_load_factor_delta > 0 ? '+' : '' }}{{ (jointResult.joint_overview.avg_load_factor_delta * 100).toFixed(2) }}%
                </div>
              </div>
            </el-col>
            <el-col :span="8">
              <div class="joint-overview-card">
                <div class="overview-label">受影响相邻线路</div>
                <div class="overview-value-row">
                  <span class="overview-new" style="font-size: 28px;">
                    {{ jointResult.joint_overview.merged_adjacent_impacts?.length || 0 }}
                    <span style="font-size: 14px; color: #909399; font-weight: normal;"> 条</span>
                  </span>
                </div>
                <div class="overview-pct" :class="jointResult.joint_overview.adjacent_overload_count > 0 ? 'metric-red' : 'metric-green'">
                  其中 {{ jointResult.joint_overview.adjacent_overload_count }} 条存在过载风险
                </div>
              </div>
            </el-col>
          </el-row>
        </div>
      </div>
    </div>

    <!-- 单线路模式：KPI对比 -->
    <div v-if="simMode === 'single' && singleResult" class="section-card">
      <div class="section-title">目标线路 KPI 对比</div>
      <el-table :data="singleKpiRows" border stripe>
        <el-table-column prop="name" label="KPI 指标" width="180" align="center" />
        <el-table-column label="调整前" align="center">
          <template #default="{ row }">
            <span :class="row.origClass || ''">{{ row.orig }}</span>
          </template>
        </el-table-column>
        <el-table-column label="调整后" align="center">
          <template #default="{ row }">
            <span :class="row.newClass || ''">{{ row.new }}</span>
          </template>
        </el-table-column>
        <el-table-column label="变化量" width="200" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.delta" :type="row.tagType" size="small">
              {{ row.delta }}
            </el-tag>
            <span v-else class="no-change">无变化</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 单线路模式：相邻线路影响 -->
    <div v-if="simMode === 'single' && singleResult" class="section-card">
      <div class="section-title">
        相邻线路影响
        <el-tag v-if="singleAdjOverloadCount > 0" type="danger" style="margin-left: 12px;">
          {{ singleAdjOverloadCount }} 条线路存在过载风险
        </el-tag>
        <el-tag v-else-if="singleResult.adjacent_impacts?.length" type="success" style="margin-left: 12px;">
          {{ singleResult.adjacent_impacts.length }} 条线路受影响
        </el-tag>
      </div>

      <el-table v-if="singleResult.adjacent_impacts?.length" :data="singleResult.adjacent_impacts" border stripe>
        <el-table-column prop="line_no" label="线路编号" width="100" align="center" />
        <el-table-column prop="line_name" label="线路名称" width="180" />
        <el-table-column label="共享站点" min-width="200">
          <template #default="{ row }">
            <el-tag v-for="st in row.shared_stations" :key="st" size="small" style="margin: 2px;">{{ st }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="原高峰满载率" width="140" align="center">
          <template #default="{ row }">{{ (row.orig_peak_load * 100).toFixed(1) }}%</template>
        </el-table-column>
        <el-table-column label="新高峰满载率" width="140" align="center">
          <template #default="{ row }">
            <span :class="row.overload_risk ? 'metric-red' : ''">
              {{ (row.new_peak_load * 100).toFixed(1) }}%
            </span>
          </template>
        </el-table-column>
        <el-table-column label="满载率增量" width="140" align="center">
          <template #default="{ row }">
            <el-tag :type="row.overload_risk ? 'danger' : (row.load_increment > 0 ? 'warning' : 'info')" size="small">
              +{{ (row.load_increment * 100).toFixed(2) }}%
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="风险状态" width="140" align="center">
          <template #default="{ row }">
            <el-tag v-if="row.overload_risk" type="danger" effect="dark">相邻线路过载风险</el-tag>
            <el-tag v-else type="success" effect="plain">正常</el-tag>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-else description="当前调整不影响相邻线路（仅裁撤站点时才会产生相邻影响）" :image-size="100" />
    </div>

    <!-- 联合模式：标签页 -->
    <div v-if="simMode === 'joint' && jointResult" class="section-card">
      <div class="section-title">模拟结果详情</div>
      <el-tabs v-model="jointResultTab" type="card">
        <el-tab-pane
          v-for="(lr, idx) in jointResult.line_results"
          :key="'line-' + lr.line_no"
          :name="'line-' + lr.line_no"
        >
          <template #label>
            <span style="display: flex; align-items: center; gap: 6px;">
              <span :style="{ color: jointResult.line_colors[idx], fontWeight: 'bold' }">●</span>
              {{ lr.line_no }} {{ lr.line_name }}
            </span>
          </template>
          <div class="line-tab-body">
            <div class="panel-title">调整前后对比</div>
            <el-row :gutter="12">
              <el-col :span="12">
                <div class="compare-card current">
                  <div class="compare-card-title">当前值</div>
                  <div class="compare-item"><span class="compare-label">站点数</span><span class="compare-value">{{ lr.orig_station_count }} 站</span></div>
                  <div class="compare-item"><span class="compare-label">高峰间隔</span><span class="compare-value">{{ lr.orig_peak_interval }} 分</span></div>
                  <div class="compare-item"><span class="compare-label">平峰间隔</span><span class="compare-value">{{ lr.orig_off_peak_interval }} 分</span></div>
                  <div class="compare-item"><span class="compare-label">日总班次</span><span class="compare-value">{{ lr.orig_total_trips }} 班</span></div>
                </div>
              </el-col>
              <el-col :span="12">
                <div class="compare-card sim">
                  <div class="compare-card-title">模拟值</div>
                  <div class="compare-item">
                    <span class="compare-label">站点数</span>
                    <span class="compare-value" :class="deltaClass(lr.new_station_count - lr.orig_station_count)">
                      {{ lr.new_station_count }} 站
                      <span v-if="lr.new_station_count !== lr.orig_station_count" class="delta">
                        ({{ lr.new_station_count > lr.orig_station_count ? '+' : '' }}{{ lr.new_station_count - lr.orig_station_count }})
                      </span>
                    </span>
                  </div>
                  <div class="compare-item">
                    <span class="compare-label">高峰间隔</span>
                    <span class="compare-value" :class="deltaClass(lr.new_peak_interval - lr.orig_peak_interval, true)">
                      {{ lr.new_peak_interval }} 分
                      <span v-if="lr.new_peak_interval !== lr.orig_peak_interval" class="delta">
                        ({{ lr.new_peak_interval > lr.orig_peak_interval ? '+' : '' }}{{ lr.new_peak_interval - lr.orig_peak_interval }})
                      </span>
                    </span>
                  </div>
                  <div class="compare-item">
                    <span class="compare-label">平峰间隔</span>
                    <span class="compare-value" :class="deltaClass(lr.new_off_peak_interval - lr.orig_off_peak_interval, true)">
                      {{ lr.new_off_peak_interval }} 分
                      <span v-if="lr.new_off_peak_interval !== lr.orig_off_peak_interval" class="delta">
                        ({{ lr.new_off_peak_interval > lr.orig_off_peak_interval ? '+' : '' }}{{ lr.new_off_peak_interval - lr.orig_off_peak_interval }})
                      </span>
                    </span>
                  </div>
                  <div class="compare-item">
                    <span class="compare-label">日总班次</span>
                    <span class="compare-value" :class="deltaClass(lr.trips_delta)">
                      {{ lr.new_total_trips }} 班
                      <span v-if="lr.trips_delta !== 0" class="delta">
                        ({{ lr.trips_delta > 0 ? '+' : '' }}{{ lr.trips_delta }})
                      </span>
                    </span>
                  </div>
                </div>
              </el-col>
            </el-row>

            <div class="panel-title" style="margin-top: 20px;">KPI 对比</div>
            <el-table :data="buildKpiRows(lr)" border stripe>
              <el-table-column prop="name" label="KPI 指标" width="180" align="center" />
              <el-table-column label="调整前" align="center">
                <template #default="{ row }"><span :class="row.origClass || ''">{{ row.orig }}</span></template>
              </el-table-column>
              <el-table-column label="调整后" align="center">
                <template #default="{ row }"><span :class="row.newClass || ''">{{ row.new }}</span></template>
              </el-table-column>
              <el-table-column label="变化量" width="200" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.delta" :type="row.tagType" size="small">{{ row.delta }}</el-tag>
                  <span v-else class="no-change">无变化</span>
                </template>
              </el-table-column>
            </el-table>

            <div v-if="lr.adjacent_impacts?.length > 0" class="panel-title" style="margin-top: 20px;">该线路对相邻线路的影响</div>
            <el-table v-if="lr.adjacent_impacts?.length > 0" :data="lr.adjacent_impacts" border stripe size="small">
              <el-table-column prop="line_no" label="线路编号" width="100" align="center" />
              <el-table-column prop="line_name" label="线路名称" width="180" />
              <el-table-column label="共享站点" min-width="180">
                <template #default="{ row }">
                  <el-tag v-for="st in row.shared_stations" :key="st" size="small" style="margin: 2px;">{{ st }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="原满载率" width="120" align="center">
                <template #default="{ row }">{{ (row.orig_peak_load * 100).toFixed(1) }}%</template>
              </el-table-column>
              <el-table-column label="新满载率" width="120" align="center">
                <template #default="{ row }">
                  <span :class="row.overload_risk ? 'metric-red' : ''">{{ (row.new_peak_load * 100).toFixed(1) }}%</span>
                </template>
              </el-table-column>
              <el-table-column label="增量" width="120" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.overload_risk ? 'danger' : (row.load_increment > 0 ? 'warning' : 'info')" size="small">
                    +{{ (row.load_increment * 100).toFixed(2) }}%
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="120" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.overload_risk" type="danger" effect="dark" size="small">过载风险</el-tag>
                  <el-tag v-else type="success" effect="plain" size="small">正常</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else-if="lr.adjacent_impacts?.length === 0" description="该线路调整不影响相邻线路" :image-size="80" />
          </div>
        </el-tab-pane>

        <el-tab-pane name="overview" label="联合影响总览">
          <div class="line-tab-body">
            <div class="panel-title">合并的相邻线路影响（累计增量）</div>
            <el-table v-if="jointResult.joint_overview.merged_adjacent_impacts?.length" :data="jointResult.joint_overview.merged_adjacent_impacts" border stripe>
              <el-table-column prop="line_no" label="线路编号" width="100" align="center" />
              <el-table-column prop="line_name" label="线路名称" width="180" />
              <el-table-column label="影响来源" width="180">
                <template #default="{ row }">
                  <el-tag v-for="ln in row.affected_by" :key="ln" size="small" type="primary" style="margin: 2px;">{{ ln }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="共享站点" min-width="180">
                <template #default="{ row }">
                  <el-tag v-for="st in row.shared_stations" :key="st" size="small" style="margin: 2px;">{{ st }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column label="原高峰满载率" width="140" align="center">
                <template #default="{ row }">{{ (row.orig_peak_load * 100).toFixed(1) }}%</template>
              </el-table-column>
              <el-table-column label="新高峰满载率" width="140" align="center">
                <template #default="{ row }">
                  <span :class="row.overload_risk ? 'metric-red' : ''">{{ (row.new_peak_load * 100).toFixed(1) }}%</span>
                </template>
              </el-table-column>
              <el-table-column label="累计增量" width="140" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.overload_risk ? 'danger' : (row.load_increment > 0 ? 'warning' : 'info')" size="small">
                    +{{ (row.load_increment * 100).toFixed(2) }}%
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="风险状态" width="140" align="center">
                <template #default="{ row }">
                  <el-tag v-if="row.overload_risk" type="danger" effect="dark">过载风险</el-tag>
                  <el-tag v-else type="success" effect="plain">正常</el-tag>
                </template>
              </el-table-column>
            </el-table>
            <el-empty v-else description="当前调整不影响相邻线路（仅裁撤站点时才会产生相邻影响）" :image-size="100" />
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <!-- 单线路趋势图 -->
    <div v-if="simMode === 'single' && singleResult" class="section-card">
      <div class="section-title">满载率变化趋势（逐步裁撤站点模拟）</div>
      <div class="trend-desc">横轴：裁撤站点数（0~5），纵轴：高峰满载率</div>
      <div ref="trendChartSingle" class="chart-container"></div>
    </div>

    <!-- 多线路趋势图（叠加） -->
    <div v-if="simMode === 'joint' && jointResult" class="section-card">
      <div class="section-title">满载率变化趋势（逐步裁撤站点模拟） - 多线路叠加</div>
      <div class="trend-desc">横轴：裁撤站点数（0~5），纵轴：高峰满载率。多条线路叠加对比。</div>
      <div ref="trendChartJoint" class="chart-container"></div>
    </div>

    <el-dialog v-model="saveDialogVisible" title="保存当前方案" width="480px" :close-on-click-modal="false">
      <el-form :model="saveForm" label-width="80px">
        <el-form-item label="方案名称">
          <el-input v-model="saveForm.name" placeholder="请输入方案名称" maxlength="50" show-word-limit />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="saveForm.remark" type="textarea" :rows="3" placeholder="可选，填写方案备注" maxlength="200" show-word-limit />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="saveDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="doSavePlan" :loading="saveLoading">确认保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick, watch } from 'vue'
import api from '../api'
import * as echarts from 'echarts'
import { RefreshRight, ArrowRight, FolderOpened } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const lines = ref([])
const loading = ref(false)
const simMode = ref('single')

const form = ref({
  line_no: '',
  peak_interval: 8,
  off_peak_interval: 15,
  station_delta: 0,
  date: ''
})
const singleResult = ref(null)
const trendChartSingle = ref(null)

const jointForm = ref({
  date: '',
  selected_line_nos: []
})
const jointLineParamList = ref([])
const jointActivePanels = ref('')
const jointResult = ref(null)
const jointResultTab = ref('overview')
const trendChartJoint = ref(null)

const lineColorList = ['#409eff', '#67c23a', '#e6a23c', '#f56c6c', '#909399']

const getLineName = (lineNo) => {
  const l = lines.value.find(x => x.line_no === lineNo)
  return l ? l.line_name : lineNo
}

const getLineColor = (idx) => lineColorList[idx] || lineColorList[lineColorList.length - 1]
const getLineColorBg = (idx) => {
  const c = getLineColor(idx)
  return c + '20'
}

const onModeChange = () => {
  singleResult.value = null
  jointResult.value = null
}

const onLineChange = async () => {
  singleResult.value = null
}

const onJointLinesChange = (newVals) => {
  const oldList = jointLineParamList.value
  const oldMap = new Map()
  oldList.forEach(x => oldMap.set(x.line_no, x))

  const newList = []
  newVals.forEach(ln => {
    if (oldMap.has(ln)) {
      newList.push(oldMap.get(ln))
    } else {
      newList.push({
        line_no: ln,
        peak_interval: 8,
        off_peak_interval: 15,
        station_delta: 0
      })
    }
  })
  jointLineParamList.value = newList
  if (newList.length > 0 && !newVals.includes(jointActivePanels.value)) {
    jointActivePanels.value = newList[0].line_no
  }
  jointResult.value = null
}

const deltaClass = (delta, inverse = false) => {
  if (delta === 0) return ''
  const good = inverse ? delta < 0 : delta > 0
  return good ? 'metric-green' : 'metric-red'
}

const singleKpiRows = computed(() => {
  if (!singleResult.value) return []
  return buildKpiRows(singleResult.value)
})

const buildKpiRows = (r) => {
  const deltaTrips = r.trips_delta
  const deltaLoad = r.new_kpi.peak_load_factor - r.orig_kpi.peak_load_factor
  const deltaSpeed = r.new_kpi.operating_speed - r.orig_kpi.operating_speed
  const deltaIntensity = r.new_kpi.passenger_intensity - r.orig_kpi.passenger_intensity
  return [
    {
      name: '日总班次数',
      orig: r.orig_kpi.daily_trips + ' 班',
      new: r.new_kpi.daily_trips + ' 班',
      delta: deltaTrips !== 0 ? (deltaTrips > 0 ? '+' : '') + deltaTrips + ' 班' : '',
      tagType: deltaTrips > 0 ? 'success' : (deltaTrips < 0 ? 'warning' : 'info')
    },
    {
      name: '高峰满载率',
      orig: (r.orig_kpi.peak_load_factor * 100).toFixed(1) + '%',
      new: (r.new_kpi.peak_load_factor * 100).toFixed(1) + '%',
      newClass: r.new_kpi.peak_load_factor > 0.9 ? 'metric-red' : (r.new_kpi.peak_load_factor > 0.7 ? 'metric-yellow' : ''),
      delta: deltaLoad !== 0 ? (deltaLoad > 0 ? '+' : '') + (deltaLoad * 100).toFixed(2) + '%' : '',
      tagType: deltaLoad > 0.1 ? 'danger' : (deltaLoad > 0 ? 'warning' : (deltaLoad < 0 ? 'success' : 'info'))
    },
    {
      name: '营运速度',
      orig: (r.orig_kpi.operating_speed || 0).toFixed(2) + ' km/h',
      new: (r.new_kpi.operating_speed || 0).toFixed(2) + ' km/h',
      delta: deltaSpeed !== 0 ? (deltaSpeed > 0 ? '+' : '') + deltaSpeed.toFixed(2) + ' km/h' : '',
      tagType: deltaSpeed > 0 ? 'success' : (deltaSpeed < 0 ? 'warning' : 'info')
    },
    {
      name: '客运强度',
      orig: (r.orig_kpi.passenger_intensity || 0).toFixed(2) + ' 人/km',
      new: (r.new_kpi.passenger_intensity || 0).toFixed(2) + ' 人/km',
      delta: deltaIntensity !== 0 ? (deltaIntensity > 0 ? '+' : '') + deltaIntensity.toFixed(2) + ' 人/km' : '',
      tagType: deltaIntensity > 0 ? 'success' : (deltaIntensity < 0 ? 'warning' : 'info')
    }
  ]
}

const singleAdjOverloadCount = computed(() => {
  if (!singleResult.value?.adjacent_impacts) return 0
  return singleResult.value.adjacent_impacts.filter(x => x.overload_risk).length
})

const runSingleSimulation = async () => {
  if (!form.value.line_no) return
  loading.value = true
  try {
    const payload = { ...form.value }
    if (!payload.date) delete payload.date
    singleResult.value = await api.runLineSimulation(payload)
    await nextTick()
    renderSingleTrendChart()
  } finally {
    loading.value = false
  }
}

const runJointSimulation = async () => {
  if (jointLineParamList.value.length < 1) return
  loading.value = true
  try {
    const payload = {
      lines: jointLineParamList.value.map(lp => ({
        line_no: lp.line_no,
        peak_interval: lp.peak_interval,
        off_peak_interval: lp.off_peak_interval,
        station_delta: lp.station_delta
      }))
    }
    if (jointForm.value.date) payload.date = jointForm.value.date

    jointResult.value = await api.runJointSimulation(payload)
    if (jointResult.value?.line_results?.length > 0) {
      jointResultTab.value = 'line-' + jointResult.value.line_results[0].line_no
    }
    await nextTick()
    renderJointTrendChart()
  } finally {
    loading.value = false
  }
}

const hasCurrentResult = computed(() => {
  if (simMode.value === 'single') return !!singleResult.value
  return !!jointResult.value
})

const resetSingleForm = () => {
  form.value.peak_interval = 8
  form.value.off_peak_interval = 15
  form.value.station_delta = 0
  singleResult.value = null
}

const resetJointForm = () => {
  jointForm.value.selected_line_nos = []
  jointLineParamList.value = []
  jointActivePanels.value = ''
  jointResult.value = null
}

const saveDialogVisible = ref(false)
const saveLoading = ref(false)
const saveForm = ref({ name: '', remark: '' })
const saveMode = ref('single')

const openSaveDialog = (mode) => {
  saveMode.value = mode
  saveForm.value = { name: '', remark: '' }
  saveDialogVisible.value = true
}

const doSavePlan = async () => {
  if (!saveForm.value.name.trim()) {
    ElMessage.warning('请输入方案名称')
    return
  }
  saveLoading.value = true
  try {
    if (saveMode.value === 'single') {
      await api.createPlan({
        name: saveForm.value.name,
        remark: saveForm.value.remark,
        sim_type: 'single',
        lines: form.value.line_no,
        params: { ...form.value },
        result: singleResult.value
      })
    } else {
      const payload = {
        lines: jointLineParamList.value.map(lp => ({
          line_no: lp.line_no,
          peak_interval: lp.peak_interval,
          off_peak_interval: lp.off_peak_interval,
          station_delta: lp.station_delta
        }))
      }
      if (jointForm.value.date) payload.date = jointForm.value.date
      await api.createPlan({
        name: saveForm.value.name,
        remark: saveForm.value.remark,
        sim_type: 'joint',
        lines: jointForm.value.selected_line_nos.join(','),
        params: payload,
        result: jointResult.value
      })
    }
    ElMessage.success('方案保存成功')
    saveDialogVisible.value = false
  } catch (e) {
    ElMessage.error('保存失败: ' + (e.message || '未知错误'))
  } finally {
    saveLoading.value = false
  }
}

const renderSingleTrendChart = () => {
  if (!trendChartSingle.value || !singleResult.value?.removal_trend) return
  const chart = echarts.init(trendChartSingle.value)
  const data = singleResult.value.removal_trend
  chart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: params => {
        const p = params[0]
        return `裁撤 ${p.name} 个站点<br/>高峰满载率: ${(p.value * 100).toFixed(1)}%`
      }
    },
    grid: { left: 60, right: 40, top: 40, bottom: 50 },
    xAxis: {
      type: 'category',
      name: '裁撤站点数',
      nameLocation: 'middle',
      nameGap: 30,
      data: data.map(d => d.remove_count),
      axisLabel: { formatter: '{value} 个' }
    },
    yAxis: {
      type: 'value',
      name: '高峰满载率',
      min: 0,
      max: 1,
      axisLabel: { formatter: (val) => (val * 100).toFixed(0) + '%' },
      splitLine: { show: true }
    },
    series: [{
      name: '高峰满载率',
      type: 'line',
      smooth: true,
      data: data.map(d => d.peak_load_factor),
      symbol: 'circle',
      symbolSize: 8,
      itemStyle: { color: '#409eff' },
      lineStyle: { width: 3 },
      areaStyle: { opacity: 0.15, color: '#409eff' },
      label: {
        show: true,
        position: 'top',
        formatter: params => (params.value * 100).toFixed(0) + '%'
      },
      markLine: {
        silent: true,
        symbol: 'none',
        lineStyle: { color: '#f56c6c', type: 'dashed', width: 2 },
        data: [{ yAxis: 0.9, label: { formatter: '过载警戒线 90%', position: 'end' } }]
      }
    }]
  })
}

const renderJointTrendChart = () => {
  if (!trendChartJoint.value || !jointResult.value?.line_results) return
  const chart = echarts.init(trendChartJoint.value)
  const lrs = jointResult.value.line_results
  const colors = jointResult.value.line_colors || lineColorList
  const legendData = lrs.map(lr => lr.line_no + ' ' + lr.line_name)
  const series = lrs.map((lr, idx) => ({
    name: lr.line_no + ' ' + lr.line_name,
    type: 'line',
    smooth: true,
    data: (lr.removal_trend || []).map(d => d.peak_load_factor),
    symbol: 'circle',
    symbolSize: 8,
    itemStyle: { color: colors[idx] },
    lineStyle: { width: 3 },
    label: {
      show: true,
      position: 'top',
      formatter: params => (params.value * 100).toFixed(0) + '%',
      fontSize: 10
    }
  }))

  const xData = lrs[0]?.removal_trend?.map(d => d.remove_count) || [0, 1, 2, 3, 4, 5]

  chart.setOption({
    tooltip: {
      trigger: 'axis',
      formatter: params => {
        let html = `裁撤 ${params[0].name} 个站点<br/>`
        params.forEach(p => {
          html += `${p.marker} ${p.seriesName}: ${(p.value * 100).toFixed(1)}%<br/>`
        })
        return html
      }
    },
    legend: {
      data: legendData,
      top: 0,
      type: 'scroll'
    },
    grid: { left: 60, right: 40, top: 60, bottom: 50 },
    xAxis: {
      type: 'category',
      name: '裁撤站点数',
      nameLocation: 'middle',
      nameGap: 30,
      data: xData,
      axisLabel: { formatter: '{value} 个' }
    },
    yAxis: {
      type: 'value',
      name: '高峰满载率',
      min: 0,
      max: 1,
      axisLabel: { formatter: (val) => (val * 100).toFixed(0) + '%' },
      splitLine: { show: true }
    },
    series: [
      ...series,
      {
        name: '过载警戒线',
        type: 'line',
        markLine: {
          silent: true,
          symbol: 'none',
          lineStyle: { color: '#f56c6c', type: 'dashed', width: 2 },
          data: [{ yAxis: 0.9, label: { formatter: '过载警戒线 90%', position: 'end' } }]
        },
        data: []
      }
    ]
  })
}

onMounted(async () => {
  lines.value = await api.getLines()
  const dr = await api.getDateRange()
  if (lines.value.length) {
    form.value.line_no = lines.value[0].line_no
  }
  if (dr.max_date) {
    form.value.date = dr.max_date
    jointForm.value.date = dr.max_date
  }
})

watch(() => singleResult.value, () => {
  nextTick(() => {
    if (singleResult.value) renderSingleTrendChart()
  })
})

watch(() => jointResult.value, () => {
  nextTick(() => {
    if (jointResult.value) renderJointTrendChart()
  })
})

window.addEventListener('resize', () => {
  setTimeout(() => {
    try {
      trendChartSingle.value && echarts.getInstanceByDom(trendChartSingle.value)?.resize()
      trendChartJoint.value && echarts.getInstanceByDom(trendChartJoint.value)?.resize()
    } catch (e) {}
  }, 100)
})
</script>

<style scoped>
.sim-section-title {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
}

.sim-section-title .title-text {
  font-size: 17px;
  font-weight: 600;
  color: #303133;
}

.panel-title {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 16px;
  padding-bottom: 10px;
  border-bottom: 1px dashed #ebeef5;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
}

.compare-card {
  border-radius: 8px;
  padding: 20px;
  border: 1px solid #ebeef5;
}

.compare-card.current {
  background: linear-gradient(135deg, #f5f7fa 0%, #e8ecf1 100%);
}

.compare-card.sim {
  background: linear-gradient(135deg, #ecf5ff 0%, #d9ecff 100%);
  border-color: #a0cfff;
}

.compare-card-title {
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 14px;
  padding-bottom: 10px;
  border-bottom: 1px dashed rgba(0, 0, 0, 0.08);
}

.compare-card.current .compare-card-title { color: #606266; }
.compare-card.sim .compare-card-title { color: #409eff; }

.compare-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  font-size: 13px;
}

.compare-label { color: #606266; }

.compare-value {
  font-weight: 600;
  color: #303133;
  position: relative;
}

.compare-value .delta {
  font-size: 11px;
  font-weight: 500;
  margin-left: 4px;
}

.placeholder {
  color: #c0c4cc;
  font-size: 12px;
  text-align: center;
  display: block;
  padding: 20px 0;
}

.capacity-box {
  margin-top: 20px;
  padding: 16px;
  background: #fafafa;
  border-radius: 8px;
  border: 1px solid #ebeef5;
}

.capacity-label {
  font-size: 12px;
  color: #909399;
  margin-bottom: 6px;
  text-align: center;
}

.capacity-value {
  font-size: 22px;
  font-weight: 700;
  color: #303133;
  text-align: center;
}

.trend-desc {
  font-size: 13px;
  color: #909399;
  margin-bottom: 12px;
  padding-left: 4px;
}

.no-change {
  color: #909399;
  font-size: 13px;
}

.card-title-row {
  display: flex;
  align-items: center;
  width: 100%;
  padding: 4px 0;
}

.line-card-body {
  padding: 8px 4px 8px 20px;
  background: #fafbfc;
  border-radius: 6px;
}

.joint-overview-card {
  padding: 20px;
  border-radius: 8px;
  border: 1px solid #ebeef5;
  background: linear-gradient(135deg, #ffffff 0%, #f9fafb 100%);
}

.overview-label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 10px;
}

.overview-value-row {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 6px;
}

.overview-orig {
  font-size: 20px;
  font-weight: 600;
  color: #606266;
}

.overview-arrow {
  color: #c0c4cc;
  font-size: 18px;
}

.overview-new {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}

.overview-new .mini-delta {
  font-size: 12px;
  font-weight: 500;
  margin-left: 4px;
}

.overview-pct {
  font-size: 12px;
  margin-top: 2px;
  padding-left: 2px;
}

.line-tab-body {
  padding-top: 12px;
}
</style>
