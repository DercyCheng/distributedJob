import api from './index'

// 获取任务列表
export const getJobs = (params) => {
    return api.get('/jobs', { params })
}

// 获取任务详情
export const getJob = (id) => {
    return api.get(`/jobs/${id}`)
}

// 创建任务
export const createJob = (data) => {
    return api.post('/jobs', data)
}

// 更新任务
export const updateJob = (id, data) => {
    return api.put(`/jobs/${id}`, data)
}

// 删除任务
export const deleteJob = (id) => {
    return api.delete(`/jobs/${id}`)
}

// 手动触发任务
export const triggerJob = (id, params = {}) => {
    return api.post(`/jobs/${id}/trigger`, params)
}
