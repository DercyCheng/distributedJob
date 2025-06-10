import api from './index'

// 获取工作节点列表
export const getWorkers = (params) => {
    return api.get('/workers', { params })
}

// 获取工作节点详情
export const getWorker = (id) => {
    return api.get(`/workers/${id}`)
}
