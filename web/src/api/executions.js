import api from './index'

// 获取执行记录列表
export const getExecutions = (params) => {
    return api.get('/executions', { params })
}

// 获取执行记录详情
export const getExecution = (id) => {
    return api.get(`/executions/${id}`)
}
