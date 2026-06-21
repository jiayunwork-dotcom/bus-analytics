import axios from 'axios'
import { ElMessage } from 'element-plus'

const api = axios.create({
  baseURL: '/api',
  timeout: 60000
})

api.interceptors.response.use(
  res => res.data,
  err => {
    ElMessage.error(err.message || '请求失败')
    return Promise.reject(err)
  }
)

export default {
  getSummary: () => api.get('/summary'),
  getLines: () => api.get('/lines'),
  getVehicles: () => api.get('/vehicles'),
  getDateRange: () => api.get('/date-range'),

  uploadData: (formData) => api.post('/upload', formData, {
    headers: { 'Content-Type': 'multipart/form-data' }
  }),

  getLineEfficiencies: () => api.get('/metrics/lines'),
  getLineDailyTrend: (lineNo, params) => api.get(`/metrics/lines/${lineNo}/trend`, { params }),

  getSectionFlow: (lineNo, date) => api.get(`/flow/${lineNo}/section`, { params: { date } }),
  getHourlyDistribution: (lineNo, date) => api.get(`/flow/${lineNo}/hourly`, { params: { date } }),
  checkTidalPattern: (lineNo, date) => api.get(`/flow/${lineNo}/tidal`, { params: { date } }),
  inferOD: (lineNo, date) => api.get(`/flow/${lineNo}/od`, { params: { date } }),

  optimizeSchedule: (lineNo, params) => api.get(`/schedule/${lineNo}/optimize`, { params }),

  getVehicleUtilizations: (date) => api.get('/vehicles/utilization', { params: { date } }),
  getVehicleGantt: (vehicleNo, date) => api.get(`/vehicles/${vehicleNo}/gantt`, { params: { date } }),

  getNetworkMetrics: (params) => api.get('/network/metrics', { params }),

  compareLines: (lineNos) => api.post('/compare/lines', { line_nos: lineNos }),
  getHistoricalTrend: (lineNo, params) => api.get(`/compare/lines/${lineNo}/historical`, { params }),
  getLineHealthScores: () => api.get('/compare/health-scores'),

  exportReport: (lineNos) => api.post('/report/export', { line_nos: lineNos }, {
    responseType: 'blob'
  }),

  runLineSimulation: (params) => api.post('/simulation/line', params),
  runJointSimulation: (params) => api.post('/simulation/joint', params),

  createPlan: (data) => api.post('/plans', data),
  listPlans: (params) => api.get('/plans', { params }),
  getPlan: (id) => api.get(`/plans/${id}`),
  deletePlan: (id) => api.delete(`/plans/${id}`),
  renamePlan: (id, name) => api.put(`/plans/${id}/rename`, { name }),
  comparePlans: (planIds) => api.post('/plans/compare', { plan_ids: planIds }),
  getPlanHistory: (planId) => api.get(`/plans/${planId}/history`)
}
